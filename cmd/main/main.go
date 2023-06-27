package main

import (
	"fmt"
	"grafana-extract-go/internal/app/crashlog"
	"grafana-extract-go/internal/localexcel"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

func main() {
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
	//err = http.ListenAndServe(addr, nil)
	err = http.ListenAndServe(addr, r)
	if err != nil {
		log.Fatal("Failed to start webhook server:", err)
	}
}

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

	//crashLogs, err := crashlog.FetchCrashLogs()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = localexcel.CreateExcel(crashLogs)
	if err != nil {
		log.Println("Create excel failed with: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with the generated file
	// http.ServeFile(w, r, filename)

	//err = googleapi.WriteCrashLogsToGoogleSheet(crashLogs)
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//	return
	//}

	// Respond with a success message
	//w.WriteHeader(http.StatusOK)
	//w.Write([]byte("Crash logs written to Google Sheets"))
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	// Handle the webhook request here
	// Retrieve necessary information, perform actions, etc.
	// Example: Parsing the request body and extracting required data
	log.Println("Step into crashlogHandler")
	// Call crashlogHandler to retrieve crash logs and process them
	crashlogHandler(w, r)
}
