# NIMB Mobile - Termux Edition

Port of NIMB and NIMB Search Proxy for Android via Termux.

## Requirements

- Android device with Termux installed
- Go cross-compiled binaries for `linux/arm64`
- `cloudflared` (optional, for tunneling)

## Building

On your desktop machine, cross-compile for ARM64:

```bash
# Build NIMB Mobile
cd nimb-mobile/nimb
GOOS=linux GOARCH=arm64 go build -o nimb-mobile .

# Build Search Proxy Mobile
cd ../search-proxy
GOOS=linux GOARCH=arm64 go build -o nimb-search-proxy-mobile .
```

## Installation

1. Transfer the `nimb-mobile` folder to your Android device
2. Open Termux and navigate to the folder
3. Make scripts executable: `chmod +x start.sh nimb/nimb-mobile search-proxy/nimb-search-proxy-mobile`

## Usage

```bash
./start.sh
```

This will:
- Start NIMB on port 3000
- Start Search Proxy on port 4000
- Run both in background tmux sessions
- Apply wake lock to prevent process termination

## Access

- NIMB UI: `http://localhost:3000`
- Search Proxy UI: `http://localhost:4000`
- API endpoints work the same as desktop version

## Managing Sessions

```bash
# View NIMB logs
tmux attach -t nimb

# View Search Proxy logs
tmux attach -t search-proxy

# Stop NIMB
tmux kill-session -t nimb

# Stop Search Proxy
tmux kill-session -t search-proxy
```
