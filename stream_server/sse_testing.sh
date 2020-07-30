# get streaming data
curl -N -H "Content-Type: text/event-stream" -H "Connection: keep-alive" http://localhost:3000/v0/stream/subscribe/customer_count
