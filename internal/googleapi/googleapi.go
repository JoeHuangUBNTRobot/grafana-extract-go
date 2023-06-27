package googleapi

import (
	"context"
	"fmt"

	"grafana-extract-go/internal/app/crashlog"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

const (
	credentialsFile = "path/to/credentials.json" // Replace with the path to your Google Sheets credentials file
	spreadsheetID   = "your-spreadsheet-id"      // Replace with the ID of your Google Sheets spreadsheet
	sheetName       = "Sheet1"                   // Replace with the name of your Google Sheets sheet
)

func WriteCrashLogsToGoogleSheet(crashLogs []crashlog.CrashLog) error {
	ctx := context.Background()

	// Load Google Sheets credentials
	creds, err := option.CredentialsFile(ctx, credentialsFile)
	if err != nil {
		return fmt.Errorf("failed to load Google Sheets credentials: %s", err)
	}

	// Create a Google Sheets client
	service, err := sheets.NewService(ctx, creds)
	if err != nil {
		return fmt.Errorf("failed to create Google Sheets client: %s", err)
	}

	// Prepare the data for writing to the sheet
	var values [][]interface{}
	for _, log := range crashLogs {
		row := []interface{}{log.Type, log.Version, log.Model}
		values = append(values, row)
	}

	// Define the range where the data will be written
	writeRange := fmt.Sprintf("%s!A2:C%d", sheetName, len(crashLogs)+1)

	// Prepare the update request
	req := &sheets.ValueRange{
		Values: values,
	}

	// Execute the update request
	_, err = service.Spreadsheets.Values.Update(spreadsheetID, writeRange, req).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		return fmt.Errorf("failed to write crash logs to Google Sheets: %s", err)
	}

	return nil
}
