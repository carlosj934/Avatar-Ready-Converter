# Avatar Ready Converter

A fast, efficient web service that transforms images into standardized avatar formats. Upload any common image format and get back a perfectly sized, square profile picture in JPG or PNG format.

## 🎯 What It Does

Avatar Ready Converter solves the common problem of preparing profile pictures for various platforms. Instead of manually cropping, resizing, and converting images, simply upload your photo and get back a ready-to-use avatar in seconds.

**Key Features:**
- Accepts JPG, PNG, GIF, and WebP formats
- Automatic square cropping with smart centering
- Configurable output sizes (default: 512x512)
- Format conversion to universal JPG/PNG formats
- Lightning-fast processing (typically under 3 seconds)
- No signup required - completely stateless

## 🏗️ Architecture

This project uses a dual-service architecture optimized for performance and simplicity:

```
┌─────────────┐
│   Browser   │
└──────┬──────┘
       │ HTTP (multipart/form-data)
       ▼
┌─────────────────┐
│  Rails API      │ ← HTTP Gateway (Port 3000)
│  (Gateway)      │   Validation & Routing
└────────┬────────┘
         │ HTTP (localhost)
         ▼
┌─────────────────┐
│  Go Service     │ ← Image Processing Engine (Port 8080)
│  (Processor)    │   Fast Image Transformations
└─────────────────┘
```

### Technology Stack

**Rails API (Gateway)**
- Ruby on Rails 8.1+ (API-only mode)
- HTTP gem for service communication
- Rack CORS for browser support
- No database required

**Go Service (Image Processor)**
- Go 1.21+
- [imaging](https://github.com/disintegration/imaging) library for transformations
- [golang.org/x/image/webp](https://pkg.go.dev/golang.org/x/image/webp) for WebP support
- Efficient in-memory processing

### Why This Architecture?

1. **Separation of Concerns**: Rails handles HTTP/validation logic, Go handles heavy image processing
2. **Performance**: Go's blazing speed for CPU-intensive image operations
3. **Simplicity**: Single container deployment with both services
4. **Cost-Effective**: No external storage needed - direct memory transfer between services
5. **Scalability**: Can easily move to separate containers if needed

## 🚀 Getting Started

### Prerequisites

- Docker (for containerized deployment)
- OR locally:
  - Ruby 3.2+
  - Go 1.21+
  - Bundler

### Quick Start (Docker)

The project includes a Docker setup that runs both services in a single container:

```bash
# Build and run
docker build -t avatar-converter .
docker run -p 3000:3000 avatar-converter
```

Visit `http://localhost:3000` to see the simple upload interface.

### Local Development

**Terminal 1 - Start Go Service:**
```bash
cd go-processor
go run main.go
# Runs on http://localhost:8080
```

**Terminal 2 - Start Rails API:**
```bash
cd rails-api
bundle install
rails server
# Runs on http://localhost:3000
```

### Configuration

Create a `.env` file in the `rails-api` directory:

```env
# Rails Configuration
SECRET_KEY_BASE=your_secret_key_here
PORT=3000

# Go Service URL (for Rails to communicate with Go)
GO_SERVICE_URL=http://localhost:8080

# Processing Limits
MAX_FILE_SIZE=10485760  # 10MB in bytes
REQUEST_TIMEOUT=30      # seconds
```

The Go service is configured through environment variables (see `go-processor/config/config.go`):

```env
# Go Service Configuration
GO_PORT=8080
MAX_FILE_SIZE=10485760
PROCESSING_TIMEOUT=30
MAX_MEMORY_MB=512
```

## 📡 API Usage

### Endpoint

```
POST /api/process
```

### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `file` | File | Yes | - | Image file to process (JPG, PNG, GIF, WebP) |
| `output_format` | String | No | "jpg" | Output format: "jpg" or "png" |
| `size` | Integer | No | 512 | Target dimensions in pixels (creates square image) |

### Example Request

**Using cURL:**
```bash
curl -X POST http://localhost:3000/api/process \
  -F "file=@/path/to/your/image.jpg" \
  -F "output_format=png" \
  -F "size=256" \
  --output avatar.png
```

**Using JavaScript:**
```javascript
const formData = new FormData();
formData.append('file', fileInput.files[0]);
formData.append('output_format', 'jpg');
formData.append('size', 512);

fetch('http://localhost:3000/api/process', {
  method: 'POST',
  body: formData
})
.then(response => response.blob())
.then(blob => {
  // Handle the processed image
  const url = URL.createObjectURL(blob);
  document.getElementById('avatar').src = url;
});
```

### Response

**Success (200 OK):**
- Returns the processed image file directly
- Content-Type: `image/jpeg` or `image/png`
- Content-Disposition: `attachment; filename="avatar.jpg"`

**Error Responses:**
```json
{
  "error": "No file provided"
}
// Status: 400 Bad Request

{
  "error": "File too large. Maximum size is 10MB"
}
// Status: 400 Bad Request

{
  "error": "Invalid format. Must be jpg or png"
}
// Status: 400 Bad Request

{
  "error": "Processing timeout"
}
// Status: 408 Request Timeout
```

## 🔒 Security & Limits

- **File Size Limit**: 10MB maximum
- **Allowed Input Formats**: JPG, PNG, GIF, WebP
- **Allowed Output Formats**: JPG, PNG only
- **Request Timeout**: 30 seconds
- **Rate Limiting**: Configured via rack-attack (100 requests per hour per IP in production)
- **CORS**: Enabled for cross-origin browser requests

## 🏥 Health Checks

Both services expose health check endpoints:

**Rails API:**
```bash
curl http://localhost:3000/health
# {"status":"ok"}
```

**Go Service:**
```bash
curl http://localhost:8080/health
# {"status":"healthy","service":"go-processor"}
```

## 🐳 Deployment

The project is designed for easy deployment to Railway, Fly.io, Render, or any container platform.

### Deployment Configuration

The included `supervisord.conf` manages both services in a single container:

```ini
[program:rails]
command=bundle exec rails server -b 0.0.0.0 -p 3000

[program:go]
command=/app/go-processor/processor
```

### Environment Variables for Production

Set these in your hosting platform:

```env
RAILS_ENV=production
SECRET_KEY_BASE=<generate_with_rails_secret>
GO_SERVICE_URL=http://localhost:8080
PORT=3000
```

### Graceful Shutdown

Both services support graceful shutdown via SIGTERM signals, properly handling in-flight requests during deployment.

## 🛠️ Development

### Project Structure

```
Avatar-Ready-Converter/
├── Dockerfile              # Single container for both services
├── supervisord.conf        # Process manager config
├── railway.json           # Railway deployment config
├── go-processor/          # Go image processing service
│   ├── main.go           # Server setup with health checks
│   ├── config/           # Configuration loading
│   ├── handlers/         # HTTP request handlers
│   ├── services/         # Image processing logic
│   └── models/           # Data structures
└── rails-api/            # Rails HTTP gateway
    ├── app/
    │   └── controllers/
    │       └── api/
    │           └── images_controller.rb
    ├── config/
    │   ├── routes.rb
    │   └── initializers/
    │       ├── cors.rb          # CORS configuration
    │       └── rack_attack.rb   # Rate limiting
    └── public/
        └── index.html           # Simple test UI
```

### Adding New Features

**To add new image formats:**
1. Add decoder in `go-processor/services/image_processor.go`
2. Update `SupportedFormats` map
3. Update documentation

**To add new transformations:**
1. Extend `ProcessImage` function in `services/image_processor.go`
2. Add parameters to `models/options.go`
3. Update Rails controller to accept new parameters

### Testing

**Test the Go service directly:**
```bash
cd go-processor
go test ./...
```

**Test the Rails API:**
```bash
cd rails-api
rails test
```

## 📊 Performance

- **Processing Time**: Typically 1-3 seconds for standard images
- **Memory Usage**: ~100MB per concurrent request
- **Throughput**: Handles 10-50 concurrent users comfortably on basic hosting
- **No Storage Overhead**: All processing done in memory

## 🤝 Contributing

This is a personal learning project demonstrating polyglot architecture, but suggestions are welcome!

## 📝 License

MIT License - feel free to use this code for your own projects.

## 🙏 Acknowledgments

Built with:
- [disintegration/imaging](https://github.com/disintegration/imaging) - Excellent Go image processing library
- [Ruby on Rails](https://rubyonrails.org/) - Powerful web framework
- [http.rb](https://github.com/httprb/http) - Clean HTTP client for Ruby

---

**Note**: This service is stateless and doesn't store any uploaded images. All processing happens in memory and images are immediately returned to the client.