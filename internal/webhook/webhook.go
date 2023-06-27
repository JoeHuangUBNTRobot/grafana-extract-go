package webhook

import (
	"encoding/json"
	"log"
	"net/http"
)

func HandleWebhook(w http.ResponseWriter, r *http.Request) {
	var payload map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		log.Println("Failed to decode JSON payload:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	alertName := payload["alertName"].(string)
	state := payload["state"].(string)

	log.Println("Received alert:", alertName)
	log.Println("Alert state:", state)

	// Call a function from the excel package to export data to Excel
	//excel.ExportDataToExcel()

	//w.WriteHeader(http.StatusOK)
}
