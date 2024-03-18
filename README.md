# grafana-extract-go

# CLI

  - -d string
    	The date, ex: 2023_06_15
  - -m string
    	The model, ex: UDM,UDMPRO,UDMPROSE,UDR,UDW,UDWPRO,UNASPRO,UCKG2,UCKP,UCKENT,UNVR,UNVRPRO
  - -mode string
    	The mode, ex: google mean googlesheet, excel or webhook mean waiting for notify from Grafana, default is webhook 
  - -p string
    	The product line, ex: product or network
  - -s int
    	The size(the total crash log counts), ex: 10 (default 10)
  - -v string
    	The version, ex: 3.1.9 or v3.1.9
    
# Writing crashlog into googlesheet
go run main.go -mode google -p network -d 2023_07_02 -v v3.0.18 -m UDMPROSE -s 10
# Writing crashlog into local excel
go run main.go -mode excel -p network -d 2023_07_02 -v v3.0.18 -m UDMPROSE -s 10
# Checking local excel file in /cmd/main
EX:  /cmd/main/CrashLogs-UNVR-3.1.9-2023-06-15.xlsx


# TODO(TBD): Run webhook server
go run main.go  
  # POST message to webhook server by curl
    curl -X POST -H "Content-Type: application/json" -d '{"alertName":"Example Alert","state":"firing"}' http://localhost:6688/webhook    
  # Notify from Telegram
    Partial nitification from Telegram
    "
    Console alert
    admin
    **Firing**
    
    Value: B=3.4482758620689653, C=1
    Labels:
     - alertname = Alert - Kernel Crash - Network Platform
     - dev_fw_version = 3.0.17
     - grafana_folder = Storage Team
     - model = UCKP
     - networkSeverity = urgency
