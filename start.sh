#!/bin/bash
# NIMB Mobile - Termux Startup Script
# =====================================
# This script starts both NIMB and NIMB Search Proxy
# in separate tmux sessions for Android/Termux usage.

set -e

echo "============================================="
echo "        NIMB Mobile - Termux Edition        "
echo "============================================="
echo ""

# Check for required commands
if ! command -v tmux &> /dev/null; then
    echo "Installing tmux..."
    pkg install -y tmux
fi

if ! command -v cloudflared &> /dev/null; then
    echo "Note: cloudflared not found. Tunnels won't work."
    echo "Install with: pkg install cloudflared"
fi

# Apply wake lock to prevent Termux from being killed
echo "Acquiring wake lock..."
termux-wake-lock 2>/dev/null || true

# Get the script's directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Kill any existing sessions
tmux kill-session -t nimb 2>/dev/null || true
tmux kill-session -t search-proxy 2>/dev/null || true

# Start NIMB (port 3000)
echo "Starting NIMB on port 3000..."
tmux new-session -d -s nimb -c "$SCRIPT_DIR/nimb" "./nimb-mobile"

# Start Search Proxy (port 4000)
echo "Starting Search Proxy on port 4000..."
tmux new-session -d -s search-proxy -c "$SCRIPT_DIR/search-proxy" "./nimb-search-proxy-mobile"

# Wait a moment for servers to start
sleep 2

echo ""
echo "============================================="
echo "                 RUNNING!                    "
echo "============================================="
echo ""
echo "  NIMB:         http://localhost:3000"
echo "  Search Proxy: http://localhost:4000"
echo ""
echo "  Access from phone browser or LAN:"
PHONE_IP=$(ip -4 addr show wlan0 2>/dev/null | grep -oP '(?<=inet\s)\d+(\.\d+){3}' | head -1)
if [ -n "$PHONE_IP" ]; then
    echo "  NIMB:         http://$PHONE_IP:3000"
    echo "  Search Proxy: http://$PHONE_IP:4000"
fi
echo ""
echo "============================================="
echo "To view logs:"
echo "  tmux attach -t nimb"
echo "  tmux attach -t search-proxy"
echo ""
echo "To stop:"
echo "  tmux kill-session -t nimb"
echo "  tmux kill-session -t search-proxy"
echo "============================================="
echo ""
echo "Servers are running in background."
echo "You can close this terminal safely."
