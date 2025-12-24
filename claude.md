# Avatar-Ready-Converter - Design Document

## Project Overview
A simple web application that allows users to upload profile pictures and receive back processed, avatar-ready images in JPG or PNG format. The application focuses on simplicity with no user authentication or account management.

## Technology Stack

### Backend
- **Go**: Image processing service (fast, efficient, great for image manipulation)
- **Ruby on Rails**: Web API and file handling
- **Architecture**: Single container deployment with both services (Rails on port 3000, Go on port 8080)
- **Communication**: HTTP via localhost between Rails and Go

### Frontend
- Simple HTML/CSS/JavaScript (can be served from Rails)
- Drag-and-drop file upload interface
- Real-time processing feedback

### Infrastructure
- **Deployment**: Single Docker container hosting both Rails and Go services
- **Storage**: No external storage needed (direct memory transfer)
- **Compute**: Single container on Railway, Fly.io, DigitalOcean, or AWS
- **Process Management**: Supervisord or Foreman to run both services

## Core Features

### Supported Image Formats

**Phase 1 - Current (MVP):**
- ✅ **Input**: JPEG (.jpg, .jpeg), PNG (.png), GIF (.gif), WebP (.webp)
- ✅ **Output**: JPEG (.jpg), PNG (.png)
- ✅ **Processing**: Square crop, resize, format conversion

**Phase 2 - Mobile Support (Coming Soon):**
- 📱 HEIC (.heic) - iPhone photos
- 📱 Better handling of high-resolution mobile images

**Phase 3 - Advanced Formats (Future):**
- 🎨 SVG (.svg) - Vector graphics (requires rendering)
- 📷 TIFF (.tif, .tiff) - Professional photography
- 📄 PDF (.pdf) - First page conversion

**Format Strategy:**
The goal is to accept ANY image format users might have and convert it to universally-supported avatar formats (JPG/PNG). We're building a universal translator for profile pictures, starting with the most common formats (80% of use cases) and expanding based on user demand.

### MVP Features
1. **Image Upload**
   - Accept common image formats (JPG, PNG, GIF, WebP)
   - File size validation (max 10MB)
   - Client-side image preview

2. **Image Processing**
   - Automatic cropping to square aspect ratio
   - Face detection and centering (optional)
   - Resize to standard avatar sizes (e.g., 512x512, 256x256)
   - Background removal (optional)
   - Format conversion (JPG/PNG output)

3. **Image Download**
   - Immediate download of processed image
   - Option to download multiple sizes
   - Temporary storage (auto-delete after 1 hour)

## System Architecture

```
┌─────────────┐
│   Browser   │
└──────┬──────┘
       │ HTTP (multipart/form-data)
       ▼
┌─────────────────┐
│  Rails API      │ ← Handles upload, validation, orchestration
│  (API Gateway)  │
└────────┬────────┘
         │ HTTP (multipart/form-data)
         ▼
┌─────────────────┐
│  Go Service     │ ← Image processing engine
│  (Processor)    │ ← Returns processed image directly
└─────────────────┘
```

### Communication Flow
**Single Container, Dual Process (Synchronous)**
- Browser uploads image to Rails (port 3000, public)
- Rails forwards image to Go service via HTTP POST to localhost:8080 (internal)
- Go processes image in memory
- Go returns processed image in HTTP response
- Rails streams result directly back to browser
- **No storage layer needed**
- **Both services in same container = zero network latency**

## API Design

### Endpoints

### Rails API Endpoints

#### POST /api/process
Upload and process an image in a single synchronous request.

**Request:**
- `Content-Type: multipart/form-data`
- `file`: Image file (required)
- `output_format`: "jpg" or "png" (optional, default: "jpg")
- `size`: Target size in pixels (optional, default: 512)

**Response:**
- Returns the processed image file directly
- `Content-Type: image/jpeg` or `image/png`
- `Content-Disposition: attachment; filename="avatar.jpg"`

**Example Rails Controller:**
```ruby
def process
  uploaded_file = params[:file]
  output_format = params[:output_format] || 'jpg'
  size = params[:size] || 512
  
  # Forward to Go service
  response = HTTP.post("#{ENV['GO_SERVICE_URL']}/process",
    form: {
      file: HTTP::FormData::File.new(uploaded_file.path),
      format: output_format,
      size: size
    }
  )
  
  # Stream result back to browser
  send_data response.body,
    type: "image/#{output_format}",
    filename: "avatar.#{output_format}",
    disposition: 'attachment'
end
```

### Go Service Endpoints

#### POST /process
Process an image and return the result.

**Request:**
- `Content-Type: multipart/form-data`
- `file`: Image file (required)
- `format`: "jpg" or "png" (optional)
- `size`: Target size in pixels (optional)

**Response:**
- Returns the processed image bytes
- `Content-Type: image/jpeg` or `image/png`

## Go Processing Service

### Responsibilities
- Receive image processing requests from Rails via HTTP
- Perform image transformations:
  - Load and decode images
  - Face detection (using OpenCV or similar)
  - Crop and resize
  - Background removal (using ML model or simple algorithms)
  - Format conversion
- Return processed image directly in HTTP response

### Libraries
- `github.com/disintegration/imaging` - Image manipulation (resize, crop, rotate)
- `golang.org/x/image/webp` - WebP format support
- `gocv.io/x/gocv` - Computer vision (face detection, optional for Phase 2)
- Standard library `net/http` and `mime/multipart` for HTTP handling
- Standard library `image/jpeg`, `image/png`, `image/gif` for format encoding/decoding

### Service Structure
```
go-processor/
├── main.go
├── handlers/
│   └── process.go
├── services/
│   ├── image_processor.go
│   └── face_detector.go
├── models/
│   └── options.go
└── config/
    └── config.go
```

## Rails API Service

### Responsibilities
- **HTTP Gateway**: Accept image uploads from browsers via multipart/form-data
- **Input Validation**: Validate file presence, size (10MB max), and output format requests
- **Format Orchestration**: Accept any supported input format (JPG, PNG, GIF, WebP) and coordinate conversion to standard output formats (JPG or PNG)
- **Proxy Layer**: Forward validated uploads to Go service via HTTP localhost connection
- **Response Streaming**: Stream processed images directly back to client without intermediate storage
- **Error Handling**: Handle timeouts, processing failures, and provide clear error messages
- **CORS Management**: Enable cross-origin requests for browser-based frontends

### Key Design Principles
- **No Format Discrimination on Input**: Rails accepts any image format the user uploads - format validation happens in Go during decode
- **Standardized Output**: Rails validates that users can only request JPG or PNG output (universal avatar formats)
- **Stateless Operation**: No database, no file storage, no session management
- **Direct Passthrough**: Acts as a thin validation and routing layer, not processing logic

### Libraries
- `http` gem (~> 5.2) - HTTP client for calling Go service with multipart form support
- `rack-cors` - CORS middleware for cross-origin browser requests
- Rails 8.1+ with API-only mode (no ActiveRecord, ActionCable, or ActiveStorage)

### Service Structure
```
rails-api/
├── app/
│   └── controllers/
│       └── api/
│           └── images_controller.rb  # Main upload endpoint
├── config/
│   ├── routes.rb                     # POST /api/process route
│   ├── initializers/
│   │   └── cors.rb                   # CORS configuration
│   └── environments/
│       └── development.rb
├── public/
│   └── index.html                    # Simple test frontend
└── .env                              # Local environment variables
```

### Input/Output Format Handling

**What Rails Does:**
```ruby
# Rails validates OUTPUT format (what user wants back)
unless ['jpg', 'jpeg', 'png'].include?(output_format.downcase)
  return render json: { error: 'Invalid format. Must be jpg or png' }
end
```

**What Rails Doesn't Do:**
- Does NOT validate input file format (Go handles this during decode)
- Does NOT reject WebP, GIF, or other supported input formats
- Does NOT process images (that's Go's job)

**The Flow:**
```
User uploads WebP file, requests PNG output
    ↓
Rails: ✅ File present? Yes
Rails: ✅ Size under 10MB? Yes  
Rails: ✅ Output format valid (png)? Yes
    ↓
Rails forwards WebP file to Go with format=png
    ↓
Go: Decodes WebP ✅
Go: Processes image ✅
Go: Encodes as PNG ✅
    ↓
Rails streams PNG back to user
```

## Data Flow (Synchronous)

**Single Request/Response Cycle:**

1. User selects image in browser
2. Browser validates file size/type client-side
3. Browser POSTs to Rails `/api/process`
4. Rails validates uploaded file
5. Rails forwards image to Go service via HTTP POST
6. Go service:
   - Receives image in memory
   - Decodes image
   - Performs transformations (resize, crop, format conversion)
   - Encodes result
   - Returns processed image bytes
7. Rails receives processed image
8. Rails streams image directly to browser
9. Browser initiates download

**Total time: 2-7 seconds depending on image size and processing**

**No storage, no job tracking, no cleanup needed**

## Cost Optimization Strategies

### Infrastructure
1. **Compute**
   - Single server/container running both Rails and Go
   - Rails listens on port 3000 (public)
   - Go listens on port 8080 (internal only)
   - Both services on same machine = no network latency

2. **Storage**
   - **None needed!** All processing done in memory
   - Saves $5-20/month on S3 costs

3. **Database**
   - **None needed!** No job tracking or persistence
   - Saves $15-50/month on RDS/database costs

### Development Costs
**Single Container Deployment Options:**
- **Option 1**: Railway ($5/month with $5 free credits monthly) - **RECOMMENDED**
  - Easy deployment: `railway up`
  - Auto-scaling
  - Good free tier
- **Option 2**: Fly.io ($0-10/month)
  - Free tier: 3 shared CPUs, 256MB RAM
  - May need paid tier for both services
- **Option 3**: DigitalOcean App Platform ($5-12/month)
  - Reliable, good performance
  - Simple deployment
- **Option 4**: Render ($7/month)
  - Free tier available but limited
  - Good for starting out
- **Option 5**: AWS ECS Fargate ($15-20/month)
  - More complex setup
  - Better for production scale

**Total monthly cost: $0-12 depending on hosting choice**

**Cost Savings from Single Container:**
- No separate container orchestration needed
- No load balancer between services needed
- Simpler networking = lower resource usage
- Can fit in smaller instance sizes

### Alternative Cheap Hosting Options
1. **Fly.io**: Free tier for small apps
2. **Railway**: $5/month with credits
3. **Render**: Free tier for web services
4. **DigitalOcean App Platform**: $5/month

## Security Considerations

1. **File Upload**
   - Validate file types (magic number checking, not just extension)
   - Limit file size (10MB max)
   - Rate limiting per IP (prevent abuse)
   - CORS configuration for frontend

2. **Processing**
   - Request timeout limits (30 seconds max)
   - Memory limits in Go service
   - Input sanitization (prevent malicious images)
   - Reject files that fail to decode

3. **Network Security**
   - Go service only accessible from Rails (internal network)
   - Rails is the only public-facing service
   - Use environment variables for service URLs

## Performance Targets

- **Total Request Time**: < 7 seconds for 5MB file (upload + process + download)
- **Processing Only**: < 3 seconds for standard operations
- **Concurrent Users**: 10-50 initially
- **Memory Usage**: < 100MB per request in Go service
- **No file retention needed** (everything in memory)

## Deployment Architecture

### Single Container Structure
```
Docker Container
├── Rails App (Port 3000, Public)
│   ├── Handles incoming HTTP requests
│   ├── Serves frontend
│   └── Proxies to Go service
├── Go Service (Port 8080, Internal Only)
│   ├── Listens on localhost:8080
│   └── Processes images
└── Process Manager (Supervisord/Foreman)
    ├── Starts Rails
    ├── Starts Go
    └── Manages both processes
```

### Dockerfile Structure
```dockerfile
FROM ruby:3.2 as base

# Install Go
RUN wget https://go.dev/dl/go1.21.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.21.linux-amd64.tar.gz
ENV PATH="/usr/local/go/bin:${PATH}"

# Install ImageMagick for Go image processing
RUN apt-get update && apt-get install -y imagemagick

# Install supervisord
RUN apt-get install -y supervisor

# Copy and build Go service
WORKDIR /app/go-processor
COPY go-processor/ .
RUN go mod download
RUN go build -o processor main.go

# Copy and setup Rails
WORKDIR /app/rails-api
COPY rails-api/Gemfile rails-api/Gemfile.lock ./
RUN bundle install
COPY rails-api/ .

# Copy supervisord config
COPY supervisord.conf /etc/supervisor/conf.d/supervisord.conf

# Expose only Rails port (Go is internal)
EXPOSE 3000

# Start both services
CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/conf.d/supervisord.conf"]
```

### Supervisord Configuration
```ini
[supervisord]
nodaemon=true

[program:rails]
command=bundle exec rails server -b 0.0.0.0 -p 3000
directory=/app/rails-api
autostart=true
autorestart=true
stdout_logfile=/dev/stdout
stdout_logfile_maxbytes=0
stderr_logfile=/dev/stderr
stderr_logfile_maxbytes=0

[program:go]
command=/app/go-processor/processor
directory=/app/go-processor
autostart=true
autorestart=true
stdout_logfile=/dev/stdout
stdout_logfile_maxbytes=0
stderr_logfile=/dev/stderr
stderr_logfile_maxbytes=0
```

### Local Development
For development, you can run both services separately:

```bash
# Terminal 1: Start Go service
cd go-processor
go run main.go

# Terminal 2: Start Rails
cd rails-api
rails server
```

Or use Foreman with a Procfile:
```
web: cd rails-api && bundle exec rails server -p 3000
processor: cd go-processor && go run main.go
```

## Development Phases

### Phase 1: MVP (Week 1-2)
- [x] Basic Rails API setup (Rails 8.1, API-only mode)
- [x] Go processing service with basic resize and crop
- [x] WebP input format support (golang.org/x/image/webp)
- [x] HTTP communication between Rails and Go (localhost)
- [x] Simple file upload endpoint with validation and passthrough
- [x] Basic frontend (HTML form with drag-and-drop)
- [x] CORS configuration for browser requests
- [ ] Create Dockerfile with both services
- [ ] Setup supervisord configuration
- [ ] End-to-end testing with multiple formats

### Phase 2: Core Features (Week 3)
- [ ] Face detection and centering (optional)
- [ ] Multiple output format support
- [ ] Better UI with drag-and-drop
- [ ] Progress indicator during processing

### Phase 3: Polish (Week 4)
- [ ] Background removal (optional)
- [ ] Comprehensive error handling
- [ ] Rate limiting
- [ ] Health check endpoints for both services
- [ ] Graceful shutdown handling
- [ ] Deploy to production (Railway/Fly.io)
- [ ] Setup monitoring and logging

### Phase 4: Optimization (Future)
- [ ] Performance tuning
- [ ] Cost monitoring
- [ ] Analytics (optional)
- [ ] Additional image formats

## Environment Variables

```bash
# Shared/Container Level
RAILS_ENV=production
PORT=3000  # External port for Rails

# Rails Specific
SECRET_KEY_BASE=xxx
GO_SERVICE_URL=http://localhost:8080  # Internal communication
MAX_FILE_SIZE=10485760
REQUEST_TIMEOUT=30

# Go Specific
GO_PORT=8080  # Internal port for Go
MAX_FILE_SIZE=10485760
PROCESSING_TIMEOUT=30
MAX_MEMORY_MB=512
```

## Monitoring and Logging

### Health Checks
- **Rails**: `GET /health` - Returns 200 if Rails is up
- **Go**: `GET /health` - Returns 200 if Go is up  
- **Container**: Both must be healthy for container to be healthy

### Logging
- Both services log to stdout/stderr
- Supervisord captures and forwards all logs
- Centralized logging through hosting provider
- Error tracking: Rollbar or Sentry (free tier)

### Metrics
- Request count and processing time
- Memory usage per service
- Error rates
- Image processing success/failure rates

### Monitoring Tools
- Railway/Fly.io built-in metrics
- Optional: New Relic free tier
- Optional: DataDog free tier

## Future Enhancements

- Batch processing multiple images
- Custom filters and effects
- Video thumbnail extraction
- Advanced rate limiting per IP
- Social media specific presets (LinkedIn, Twitter, etc.)
- Async processing with job queue (if request times exceed 30s)
- Caching layer (Redis) for frequently processed similar images

## Open Questions

1. What specific avatar requirements are we targeting?
2. Should we support animated GIFs/WebP?
3. Do we need background removal or just cropping?
4. What's the expected traffic volume?
5. Any specific image quality requirements?

---

**Last Updated**: December 2024
**Version**: 1.0
