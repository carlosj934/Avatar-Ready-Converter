package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"context"
	"strings"
	"time"
	"log"

	"avatar-ready-converter/processor/config"
	"avatar-ready-converter/processor/models"
	"avatar-ready-converter/processor/services"
)

type ProcessHandler struct {
	processor *services.ImageProcessor
	config *config.Config
}

func NewProcessHandler(cfg *config.Config) *ProcessHandler {
	return &ProcessHandler {
		processor: services.NewImageProcessor(),
		config: cfg,
	}
}

func (h *ProcessHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(h.config.ProcessingTimeout)*time.Second)
	defer cancel()

	// process with timeout
	done := make(chan error, 1)
	var processedData []byte

	go func() {
		var err error
		processedData, err = h.processImage(r)
		done <- err
	}()

	select {
	case err := <-done:
		if err != nil {
			log.Printf("Processing error: %v", err)
			h.handleError(w, err)
			return
		}

		// get output format for content type
		opts := models.ProcessOptions{
			Format: r.FormValue("format"),
		}
		opts.SetDefaults()

		// set response headers
		contentType := "image/jpeg"
		if opts.Format == "png" {
			contentType = "image/png"
		}

		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Content-Length", strconv.Itoa(len(processedData)))
		w.WriteHeader(http.StatusOK)
		w.Write(processedData)

	case <-ctx.Done():
		log.Printf("Request timeout exceeded")
		http.Error(w, "Processing timeout exceeded", http.StatusRequestTimeout)
		return
	}
}

func (h *ProcessHandler) processImage(r *http.Request) ([]byte, error) {
	// parse multipart form
	err := r.ParseMultipartForm(h.config.MaxFileSize)
	if err != nil {
		return nil, fmt.Errorf("file too large or invalid: %w", err)
	}

	// Get uploaded file
	file, header, err := r.FormFile("file")
	if err != nil {
		return nil, fmt.Errorf("no file uploaded: %w", err)
	}
	defer file.Close()

	// validate file size
	if header.Size > h.config.MaxFileSize {
		return nil, fmt.Errorf("file size %d exceeds max %d bytes", header.Size, h.config.MaxFileSize)
	}

	// log file info
	log.Printf("Processing file: %s, size: %d bytes", header.Filename, header.Size)

	// read file data with size limit
	imageData, err := io.ReadAll(io.LimitReader(file, h.config.MaxFileSize))
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// get processing options
	opts := models.ProcessOptions{
		Format: r.FormValue("format"),
		Size: h.parseSize(r.FormValue("size")),
	}

	// process the image
	processedData, err := h.processor.Process(imageData, opts)
	if err != nil {
		return nil, fmt.Errorf("processing failed: %w", err)
	}

	log.Printf("Successfully processed image: %s -> %s, output size: %d bytes",
		header.Filename, opts.Format, len(processedData))
	
	return processedData, nil
}

func (h *ProcessHandler) parseSize(sizeStr string) int {
	if sizeStr == "" {
		return 0
	}
	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		log.Printf("Invalid size parameter: %s, using default", sizeStr)
		return 0
	}
	return size
}

// handleError provides appropriate HTTP response based on error types
func (h *ProcessHandler) handleError(w http.ResponseWriter, err error) {
	errMsg := err.Error()

	// determine appropriate status code based on error
	switch {
	case strings.Contains(errMsg, "no file uploaded"):
		http.Error(w, "No file uploaded", http.StatusBadRequest)
	case strings.Contains(errMsg, "file too large"), strings.Contains(errMsg, "exceeds maximum"):
		http.Error(w, "File too large. Max size is 10MB", http.StatusBadRequest)
	case strings.Contains(errMsg, "failed to decode"), strings.Contains(errMsg, "unsupported format"):
		http.Error(w, "Invalid or unsupported image format", http.StatusBadRequest)
	case strings.Contains(errMsg, "failed to read"):
		http.Error(w, "Failed to read uploaded file", http.StatusBadRequest)
	default:
		http.Error(w, "Image processing failed", http.StatusInternalServerError)
	}
}
