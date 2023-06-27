# grafana-extract-go
# run webhook server
go run main.go  
# POST message to webhook server
curl -X POST -H "Content-Type: application/json" -d '{"alertName":"Example Alert","state":"firing"}' http://localhost:6688/webhook

