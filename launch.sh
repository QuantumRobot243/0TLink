#!/bin/bash

SESSION="0TLink-Dev"

tmux kill-session -t $SESSION 2>/dev/null

tmux new-session -d -s $SESSION -n 'Infrastructure'

tmux send-keys -t $SESSION:0.0 "go run cmd/server/main.go" C-m

tmux split-window -h -t $SESSION:0.0
tmux send-keys -t $SESSION:0.1 "python3 -m http.server 3000" C-m

tmux split-window -v -t $SESSION:0.1
tmux send-keys -t $SESSION:0.2 "sleep 2 && go run cmd/client/main.go" C-m

tmux split-window -v -t $SESSION:0.0
tmux send-keys -t $SESSION:0.3 "echo 'Waiting for Agent...'; sleep 4; echo 'Try: curl http://localhost:8081'" C-m

tmux select-pane -t $SESSION:0.0

tmux attach-session -t $SESSION
