package crashlog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	// Elasticsearch base URL
	ESBaseURL = "https://search-crash-manual-t332rijsqlg3hz7pk5pu7atqla.us-west-2.es.amazonaws.com"
)

type CrashLog struct {
	Type                string    `json:"type"`
	SystemTime          time.Time `json:"system_time"`
	AnonymousDeviceID   string    `json:"anonymous_device_id"`
	Model               string    `json:"model"`
	Version             string    `json:"version"`
	BomRev              string    `json:"bomrev"`
	IsDefault           bool      `json:"is_default"`
	ProductLine         string    `json:"product_line"`
	BootTime            time.Time `json:"boot_time"`
	Uptime              int       `json:"uptime"`
	HumanReadableUptime string    `json:"human_readable_uptime"`
	KernelVersion       string    `json:"kernel_version"`
	Architecture        string    `json:"architecture"`
	LoadAverage         string    `json:"load_average"`
	CrashLog            string    `json:"crash_log"`
	IsInternal          string    `json:"is_internal"`
	Signal              int       `json:"signal"`
	APIVersion          string    `json:"apiVersion"`
	CleanVersion        string    `json:"clean_version"`
	SortableVersion     int       `json:"sortable_version"`
}

// func FetchCrashLogs(productLine, date string) ([]CrashLog, error) {
func FetchCrashLogs(productLine, date, version, model string, size int) ([]CrashLog, error) {
	// Check if productLine or date is empty
	if productLine == "" || date == "" || version == "" {
		//TODO: comment for testing prupose
		//return nil, fmt.Errorf("productLine and date arguments are required")
		productLine = "protect"
		date = "2023_06_15"
		version = "v3.1.9"
		model = "UNVR"

	}
	if size <= 0 {
		size = 10
	}

	// Debug output
	log.Printf("productLine: %s, date: %s, version: %s, model: %s, size: %d\n", productLine, date, version, model, size)

	//https://search-crash-manual-t332rijsqlg3hz7pk5pu7atqla.us-west-2.es.amazonaws.com/network_controller_logs_2023_06_15/_search
	//Network product line: network_controller_logs_year_month_date EX: network_controller_logs_2023_06_15
	//https://search-crash-manual-t332rijsqlg3hz7pk5pu7atqla.us-west-2.es.amazonaws.com/protect_logs_2023_06_15/_search
	//Protect product line: protect_logs_year_month_date EX: protect_logs_2023_06_15
	// Construct the Elasticsearch URL based on the product line and date
	url := fmt.Sprintf("%s/%s_logs_%s/_search", ESBaseURL, productLine, date)
	log.Println("Elasticsearch URL:", url) // Print the Elasticsearch URL for debugging

	// curl --location --request GET 'https://search-crash-manual-t332rijsqlg3hz7pk5pu7atqla.us-west-2.es.amazonaws.com/protect_logs_2023_06_15/_search' --header 'Content-Type: application/json' --data '{
	// 	"query": {
	// 	  "bool": {
	// 		"must": [
	// 		  {
	// 			"term": {
	// 			  "body.type": "kernel_crash"
	// 			}
	// 		  },
	// 		  {
	// 			"term": {
	// 			  "body.version": "v3.1.9"
	// 			}
	// 		  },
	// 		  {
	// 			"terms": {
	// 			  "body.model.keyword": ["UNVRPRO"]
	// 			}
	// 		  }
	// 		]
	// 	  }
	// 	},
	// 	"size": 10,
	// 	"aggs": {
	// 		"distinct_counts": {
	// 		  "cardinality": {
	// 			"field": "body.anonymous_device_id.keyword"
	// 		  }
	// 		}
	// 	  }
	//   }' | jq

	// Construct the request body
	requestBody := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{
						"term": map[string]interface{}{
							"body.type": "kernel_crash",
						},
					},
					{
						"term": map[string]interface{}{
							"body.version": version,
						},
					},
					{
						"terms": map[string]interface{}{
							"body.model.keyword": []string{model},
						},
					},
				},
			},
		},
		"size": size,
		"aggs": map[string]interface{}{
			"distinct_counts": map[string]interface{}{
				"cardinality": map[string]interface{}{
					"field": "body.anonymous_device_id.keyword",
				},
			},
		},
	}

	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %s", err)
	}

	// Send the request to Elasticsearch
	//resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch crash logs: %s", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %s", err)
	}

	var response struct {
		Hits struct {
			Hits []struct {
				Source struct {
					CrashLog `json:"body"`
				} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %s", err)
	}

	if len(response.Hits.Hits) == 0 {
		log.Println("No crash logs found") // Print a debug message when no crash logs are found
	}

	crashLogs := make([]CrashLog, len(response.Hits.Hits))
	for i, hit := range response.Hits.Hits {
		crashLogs[i] = hit.Source.CrashLog
	}

	return crashLogs, nil
}
