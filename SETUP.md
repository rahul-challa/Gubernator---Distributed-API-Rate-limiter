# Setup Guide for Gubernator

## Prerequisites Installation

### Option 1: Using Docker (Recommended - Easiest)

If you have Docker Desktop installed, this is the easiest way:

1. **Install Docker Desktop for Windows**
   - Download from: https://www.docker.com/products/docker-desktop/
   - Install and restart your computer
   - Make sure Docker Desktop is running

2. **Skip to "Running with Docker" section below**

### Option 2: Manual Setup (Go + Redis)

#### Install Go

1. **Download Go for Windows**
   - Visit: https://go.dev/dl/
   - Download the latest Windows installer (e.g., `go1.21.x.windows-amd64.msi`)

2. **Install Go**
   - Run the installer
   - It will install to `C:\Program Files\Go` by default
   - **Important**: Restart your terminal/PowerShell after installation

3. **Verify Installation**
   ```powershell
   go version
   ```

#### Install Redis

**Option A: Using Docker (if you install Docker)**
```powershell
docker run -d -p 6379:6379 --name redis redis:7-alpine
```

**Option B: Using WSL2 (Windows Subsystem for Linux)**
```bash
# In WSL2 terminal
sudo apt update
sudo apt install redis-server
redis-server
```

**Option C: Using Windows Port**
- Download Redis for Windows from: https://github.com/microsoftarchive/redis/releases
- Or use Memurai (Redis-compatible): https://www.memurai.com/

**Option D: Cloud Redis (Free Tier)**
- Use Redis Cloud (free tier): https://redis.com/try-free/
- Or Upstash: https://upstash.com/

## Project Setup

### Step 1: Install Go Dependencies

```powershell
cd /path/to/gubernator
go mod download
```

### Step 2: Start Redis

**If using Docker:**
```powershell
docker run -d -p 6379:6379 --name gubernator-redis redis:7-alpine
```

**If using local Redis:**
- Make sure Redis is running on `localhost:6379`

### Step 3: Run the Server

```powershell
go run cmd/server/main.go
```

You should see:
```
HTTP server starting on port 8080
gRPC server starting on port 9090
```

### Step 4: Test the Server

Open a new terminal and run:

```powershell
# Health check
curl http://localhost:8080/health

# Test endpoint (will be rate limited)
curl http://localhost:8080/api/v1/test
```

## Running with Docker Compose (Easiest)

If you have Docker installed:

```powershell
# Build and start everything
docker-compose up --build

# Or run in background
docker-compose up -d --build

# View logs
docker-compose logs -f gubernator

# Stop everything
docker-compose down
```

## Running the Visualization

1. **Install Python dependencies:**
```powershell
pip install -r scripts/requirements.txt
```

2. **Make sure the server is running**, then:
```powershell
python scripts/flood_test.py --rps 50 --duration 30
```

## Troubleshooting

### "go: command not found"
- Make sure Go is installed and PATH is set
- Restart your terminal after installing Go
- Verify with: `go version`

### "Cannot connect to Redis"
- Make sure Redis is running: `docker ps` (if using Docker)
- Check Redis is on port 6379
- Try: `redis-cli ping` (if redis-cli is available)

### Port already in use
- Change the port: `go run cmd/server/main.go -http-port 8081`
- Or stop the process using port 8080

### Docker not starting
- Make sure Docker Desktop is running
- Check Windows WSL2 is enabled (for Docker Desktop)

