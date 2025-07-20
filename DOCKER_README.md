# CodeCamp Coolify Deployment

This is a Go web application optimized for Coolify deployment.

## Coolify Configuration

### Application Settings
- **Port**: 8082
- **Build Pack**: Dockerfile
- **Health Check Path**: `/` (optional)

### Environment Variables
Set these in Coolify if needed:
- `ENV=production`
- `TZ=America/Denver`

### Domain Setup
Coolify will handle:
- SSL termination
- Domain routing
- Reverse proxy
- Load balancing (if scaled)

## Local Development

```bash
# Run locally
go run main.go

# Or build and run
go build -o main .
./main
```

Access at http://localhost:8082

## Docker Build Test

```bash
# Test Docker build locally
docker build -t codecamp .
docker run -p 8082:8082 codecamp
```

## File Structure

- `templates/` - HTML templates
- `archive/` - Static assets and images
- `main.go` - Application entry point
- `Dockerfile` - Container build instructions

## Deployment Notes

- Coolify automatically handles SSL certificates
- No need for docker-compose files
- Application runs on port 8082 internally
- Coolify manages scaling and health checks
- Static files are served directly by the Go application
