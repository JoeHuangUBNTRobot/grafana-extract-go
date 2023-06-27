# grafana-extract-go
# Run webhook server
go run main.go  
# POST message to webhook server
curl -X POST -H "Content-Type: application/json" -d '{"alertName":"Example Alert","state":"firing"}' http://localhost:6688/webhook
# get the local excel file in /cmd/main, EX: CrashLogs-UNVR-3.1.9-2023-06-15.xlsx
