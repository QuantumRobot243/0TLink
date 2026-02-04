#!/bin/bash

SESSION="0TLink-Dev"

echo "[0TLink] Initiating shutdown..."

if tmux has-session -t "$SESSION" 2>/dev/null; then
    tmux kill-session -t "$SESSION"
    echo " Tmux session '$SESSION' terminated."
fi

pkill -f "go run cmd/server/main.go"
pkill -f "go run cmd/client/main.go"
pkill -f "python3 -m http.server 3000"

rm -f certs/*.srl
rm -rf ~/.local/share/sidecar-net/*.tmp

echo "All ports (7000, 8081, 3000) have been released."
