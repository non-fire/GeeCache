#!/bin/bash
trap "kill 0" EXIT

go build -o server
./server -port=8001 &
./server -port=8002 &
./server -port=8003 -api=1 &
# api(9999) and server(8003) are managed in a same group and tied together
# only when this server(8003) miss, other nodes wil be visited
# when user visit 1 api => visit the maincache in this group => if not hit, choose peer according to concurrent hash
# a group represents a peer

sleep 2
echo ">>> start test"
curl "http://localhost:9999/api?key=Tom" &
curl "http://localhost:9999/api?key=Tom" &
curl "http://localhost:9999/api?key=Tom" &

wait