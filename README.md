# NIMB Mobile - Termux Edition

Run NIMB on your Android device via Termux.

## Requirements
- Android device with Termux installed
- `cloudflared`

## Installation

### Step 1: Download the Release

Go to the **Releases** page and download `nimb-mobile-termux.zip`

### Step 2: Setup Termux

Open Termux on your Android device and run:

```bash
# Update Termux and install dependencies
pkg update && pkg install -y tmux unzip cloudflared

# Grant storage access (tap "Allow" when prompted)
termux-setup-storage
```

### Step 3: Extract and Run

```bash
# Create directory and navigate to it
mkdir -p ~/nimb-mobile && cd ~/nimb-mobile

# Extract the release from Downloads
unzip ~/storage/downloads/nimb-mobile-termux.zip

# Make everything executable
chmod +x nimb-mobile start.sh

# Start NIMB!
./start.sh
```

That's it! NIMB is now running on your phone.

## Usage

Start NIMB anytime with:

```bash
cd ~/nimb-mobile
./start.sh
```

The script will:
- Start NIMB on port 3000
- Run in background
- Apply wake lock to keep it running

## Accessing NIMB

After starting, you'll see:

```
=============================================
                 RUNNING!                    
=============================================

  NIMB: http://localhost:3000

  LAN Access: http://192.168.x.x:3000

=============================================
```

- **On your phone:** Open browser → `http://localhost:3000`
- **From other devices:** Use the LAN cloudflared tunnel without v1/chat/completions

## Managing NIMB

```bash
# View live logs
tmux attach -t nimb

# Exit logs (press these keys)
Ctrl+B, then D

# Stop NIMB
tmux kill-session -t nimb

# Check if running
tmux list-sessions
```

## Termux Basics

New to Termux? Here are essential commands:

```bash
# Where am I?
pwd

# What's here?
ls

# Go to NIMB folder
cd ~/nimb-mobile

# Go back one folder
cd ..

# Read a file
cat filename
```

## Troubleshooting

**"unzip: command not found"**
```bash
pkg install unzip
```

**"Permission denied"**
```bash
chmod +x nimb-mobile start.sh
# or try:
bash start.sh
```

**Can't find the zip file?**
```bash
# First, make sure you have storage access:
termux-setup-storage

# Then check if file is there:
ls ~/storage/downloads/

# Make sure it's named: nimb-mobile-termux.zip
```

**Port already in use?**
```bash
tmux kill-session -t nimb
# or:
pkill -f nimb-mobile
```

**NIMB stops when I close Termux?**

This shouldn't happen because start.sh runs NIMB in tmux. If it does:
- Don't force-close Termux from Recent Apps
- Check that wake lock is enabled (start.sh does this automatically)


## Building from Source

Want to compile it yourself instead of using the pre-compiled release?

### On Your Desktop Machine

Cross-compile the binary for ARM64:

```bash
# Navigate to your NIMB project
cd nimb-mobile

# Build for Android (ARM64)
GOOS=linux GOARCH=arm64 go build -o nimb-mobile .
```

This creates a `nimb-mobile` binary ready for Android.

### Transfer to Android

1. Transfer the compiled `nimb-mobile` binary to your Android device's Downloads folder (via USB, cloud storage, etc.)
2. Also download or create the `start.sh` script from this repository

### Install on Termux

```bash
# Setup Termux (if you haven't already)
pkg update && pkg install -y tmux cloudflared
termux-setup-storage

# Create directory
mkdir -p ~/nimb-mobile && cd ~/nimb-mobile

# Copy files from Downloads
cp ~/storage/downloads/nimb-mobile ~/nimb-mobile/
cp ~/storage/downloads/start.sh ~/nimb-mobile/

# Make executable
chmod +x nimb-mobile start.sh

# Run!
./start.sh
```

---

**Pro Tip:** NIMB runs in the background via tmux, so you can close Termux and it keeps running. Access it anytime at `http://localhost:3000`
