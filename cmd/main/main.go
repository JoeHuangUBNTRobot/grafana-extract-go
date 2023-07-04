package main

import (
	"flag"
	"fmt"
	"grafana-extract-go/internal/app/crashlog"
	"grafana-extract-go/internal/googleapi"
	"grafana-extract-go/internal/localexcel"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			ip := ipNet.IP.String()

			// Check if IP has the prefix "192."
			if strings.HasPrefix(ip, "192.") {
				return ip, nil
			}
		}
	}

	// If no matching IP is found, set it to "127.0.0.1"
	return "127.0.0.1", nil
}

func crashlogHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: parser response from Grafana webhook
	// Get the product line and date from query parameters
	productLine := r.URL.Query().Get("productLine")
	date := r.URL.Query().Get("date")
	version := r.URL.Query().Get("version")
	model := r.URL.Query().Get("model")
	sizeStr := r.URL.Query().Get("size")
	size, err := strconv.Atoi(sizeStr)
	// if err != nil {
	// 	// Handle the error condition and return an error
	// 	log.Println("Get size error")
	// 	return
	// }
	// Fetch crash logs based on the product line and date
	crashLogs, err := crashlog.FetchCrashLogs(productLine, date, version, model, size)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to fetch crash logs: %s", err), http.StatusInternalServerError)
		return
	}
	// Check if crashLogs slice is empty
	if len(crashLogs) == 0 {
		http.Error(w, "no crash logs found", http.StatusNotFound)
		return
	}

	// Attempt to write crash logs to Google Sheets
	err = googleapi.WriteCrashLogs(crashLogs)
	if err == nil {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Crash logs written to Google Sheets"))
		//return
	} else {
		log.Println("Create Google Sheets failed with: ", err)
	}

	// If writing to Google Sheets failed, create a local Excel file
	err = localexcel.CreateExcel(crashLogs)
	if err != nil {
		log.Println("Create excel failed with: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Crash logs written to local Excel file"))
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	// Handle the webhook request here
	// Retrieve necessary information, perform actions, etc.
	// Example: Parsing the request body and extracting required data
	log.Println("Step into crashlogHandler")
	// Call crashlogHandler to retrieve crash logs and process them
	crashlogHandler(w, r)
}

func writeCrashLogsToExcel(productLine, date, version, model string, size int) error {
	// Fetch crash logs
	crashLogs, err := crashlog.FetchCrashLogs(productLine, date, version, model, size)
	if err != nil {
		return fmt.Errorf("failed to fetch crash logs: %s", err)
	}

	// Write crash logs to Excel
	err = localexcel.CreateExcel(crashLogs)
	if err != nil {
		return fmt.Errorf("failed to create Excel: %s", err)
	}

	fmt.Println("Crash logs written to Excel")
	return nil
}

func writeCrashLogsToGoogleSheets(productLine, date, version, model string, size int) error {
	// Fetch crash logs
	crashLogs, err := crashlog.FetchCrashLogs(productLine, date, version, model, size)
	if err != nil {
		return fmt.Errorf("failed to fetch crash logs: %s", err)
	}

	// Write crash logs to Google Sheets
	err = googleapi.WriteCrashLogs(crashLogs)
	if err != nil {
		return fmt.Errorf("failed to write crash logs to Google Sheets: %s", err)
	}

	fmt.Println("Crash logs written to Google Sheets")
	return nil
}

func main() {
	// Define command-line flags
	mode := flag.String("mode", "", "The mode, ex: google mean googlesheet, excel or webhook mean waiting for notify from Grafana, default is webhook ")
	productLine := flag.String("p", "", "The product line, ex: product or network")
	date := flag.String("d", "", "The date, ex: 2023_06_15")
	version := flag.String("v", "", "The version, ex: 3.1.9 or v3.1.9")
	model := flag.String("m", "", "The model, ex: UDM,UDMPRO,UDMPROSE,UDR,UDW,UDWPRO,UNASPRO,UCKG2,UCKP,UCKENT,UNVR,UNVRPRO")
	size := flag.Int("s", 10, "The size(the total crash log counts), ex: 10")

	// Parse command-line flags
	flag.Parse()

	// Attach prefix 'v' to version if it's not present
	if *version != "" && !strings.HasPrefix(*version, "v") {
		*version = "v" + *version
	}
	*version = *version + "*"

	// Check if a command-line mode flag is provided
	if *mode != "" {
		// Debug output
		log.Printf("Parse CLI: mode: %s, productLine: %s, date: %s, version: %s, model: %s, size: %d\n", *mode, *productLine, *date, *version, *model, *size)

		// Call the CLI function based on the provided command
		switch *mode {
		case "excel":
			err := writeCrashLogsToExcel(*productLine, *date, *version, *model, *size)
			if err != nil {
				fmt.Println("Error writing crash logs to Excel:", err)
			}
		case "google":
			err := writeCrashLogsToGoogleSheets(*productLine, *date, *version, *model, *size)
			if err != nil {
				fmt.Println("Error writing crash logs to Google Sheets:", err)
			}
		default:
			log.Println("Invalid command. Usage: go run main.go [command]")
			log.Println("Available commands:")
			log.Println("  excel - Write crash logs to Excel")
			log.Println("  google - Write crash logs to Google Sheets")
		}
	} else {
		// Webhook mode
		r := mux.NewRouter()
		r.HandleFunc("/webhook", webhookHandler).Methods(http.MethodPost)
		r.HandleFunc("/crashlogs", crashlogHandler).Methods(http.MethodGet)
		//http.HandleFunc("/crashlogs", crashlogHandler)
		//http.HandleFunc("/webhook", webhook.HandleWebhook)

		localIP, err := getLocalIP()
		if err != nil {
			log.Fatal("Failed to get local IP:", err)
		}

		port := "6688"
		addr := localIP + ":" + port
		webhookURL := "http://" + addr + "/webhook"

		log.Println("Webhook server started at:", webhookURL)

		err = http.ListenAndServe(addr, r)
		if err != nil {
			log.Fatal("Failed to start webhook server:", err)
		}
	}

}
