package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

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

	// Parse multipart form
	err := r.ParseMultipartForm(h.config.MaxFileSize)
	if err != nil {
		http.Error(w, "File too large or invalid", http.StatusBadRequest)
		return
	}

	// Get uploaded file
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "No file uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read file data
	imageData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	// Get processing options from form
	opts := models.ProcessOptions {
		Format: r.FormValue("format"),
		Size: h.parseSize(r.FormValue("size")),
	}

	// Process the image
	processedData, err := h.processor.Process(imageData, opts)
	if err != nil {
		http.Error(w, fmt.Sprintf("Processing failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Set response headers
	contentType := "image/jpeg"
	if opts.Format == "png" {
		contentType = "image/png"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.Itoa(len(processedData)))
	w.WriteHeader(http.StatusOK)
	w.Write(processedData)
}

func (h *ProcessHandler) parseSize(sizeStr string) int {
	if sizeStr == "" {
		return 0
	}
	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		return 0
	}
	return size
}
