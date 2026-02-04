#!/bin/bash

echo -e "\n--- 0TLink System Status ---"


for port in 7000 8081 3000; do
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null ; then
        echo -e "Port $port: \e[32mONLINE\e[0m"
    else
        echo -e "Port $port: \e[31mOFFLINE\e[0m"
    fi
done

if [ -f "certs/client.crt" ]; then
    EXPIRY=$(openssl x509 -enddate -noout -in certs/client.crt | cut -d= -f2)
    echo "Identity: Valid until $EXPIRY"
fi
echo "----------------------------"
