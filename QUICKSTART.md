# Quick Start Guide

## Fastest Way to Get Started

### Step 1: Install Go (Required)

1. **Download Go for Windows:**
   - Go to: https://go.dev/dl/
   - Download: `go1.21.x.windows-amd64.msi` (or latest version)
   - Run the installer
   - **Restart your terminal/PowerShell after installation**

2. **Verify:**
   ```powershell
   go version
   ```
   Should show: `go version go1.21.x windows/amd64`

### Step 2: Choose Redis Option

#### Option A: Docker (Recommended if you have Docker)
```powershell
docker run -d -p 6379:6379 --name gubernator-redis redis:7-alpine
```

#### Option B: Cloud Redis (No installation needed!)
Use a free cloud Redis service:
- **Upstash** (Free tier): https://upstash.com/
- **Redis Cloud** (Free tier): https://redis.com/try-free/

After signing up, you'll get a connection string like:
```
redis://default:password@hostname:port
```

Update the connection in `cmd/server/main.go` or use environment variables.

#### Option C: WSL2 (If you have WSL installed)
```bash
# In WSL terminal
sudo apt update && sudo apt install -y redis-server
redis-server
```

### Step 3: Install Dependencies & Run

```powershell
# Navigate to project directory
cd /path/to/gubernator

# Download Go dependencies
go mod download

# Run the server (default: localhost:6379)
go run cmd/server/main.go
```

### Step 4: Test It!

Open a **new terminal** and run:

```powershell
# Test health endpoint
curl http://localhost:8080/health

# Test rate-limited endpoint (try multiple times)
curl http://localhost:8080/api/v1/test
curl http://localhost:8080/api/v1/test
curl http://localhost:8080/api/v1/test
# ... after 10 requests, you'll get 429 errors
```

### Step 5: Run Visualization (Optional)

```powershell
# Install Python dependencies
pip install requests

# Run the flood test (make sure server is running!)
python scripts/flood_test.py
```

## Using Docker Compose (If Docker is installed)

```powershell
# Start everything (Redis + Server)
docker-compose up --build

# In another terminal, test it
curl http://localhost:8080/health
```

## Configuration

You can customize the rate limiter:

```powershell
go run cmd/server/main.go -capacity 20 -refill-rate 2.0 -redis-addr localhost:6379
```

Options:
- `-capacity`: Token bucket size (default: 10)
- `-refill-rate`: Tokens per second (default: 1.0)
- `-redis-addr`: Redis address (default: localhost:6379)
- `-http-port`: HTTP server port (default: 8080)

## Need Help?

1. **Go not found?** → Restart terminal after installing Go
2. **Redis connection error?** → Make sure Redis is running on port 6379
3. **Port in use?** → Change port with `-http-port 8081`

