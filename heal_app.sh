#!/bin/bash

LOG_FILE="healer.log"
TIMESTAMP=$(date "+%Y-%m-%d %H:%M:%S")

echo "[$TIMESTAMP] Sidecar detected a failure. Attempting restart..." >> $LOG_FILE

pkill -f "python3 -m http.server 3000"
nohup python3 -m http.server 3000 > /dev/null 2>&1 &

if [ $? -eq 0 ]; then
    echo "[$TIMESTAMP] Healing successful: App restarted." >> $LOG_FILE
else
    echo "[$TIMESTAMP] Healing FAILED." >> $LOG_FILE
fi
