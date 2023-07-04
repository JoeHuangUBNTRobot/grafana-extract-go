package googleapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"grafana-extract-go/internal/app/crashlog"
	"grafana-extract-go/internal/crashlogutil"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type GoogleAPI struct {
	ctx           context.Context
	sheetsSvc     *sheets.Service
	spreadsheet   *sheets.Spreadsheet
	sheetName     string
	spreadsheetID string
}

func parseToken(token []byte) *oauth2.Token {
	strToken := strings.TrimSpace(string(token))
	return &oauth2.Token{
		AccessToken: strToken,
	}
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config, tokenFile string) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tok, _ := tokenFromFile(tokenFile)
	return config.Client(context.Background(), tok)
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func (g *GoogleAPI) CreateSpreadsheet(spreadsheetName string) error {
	// Create a new spreadsheet
	spreadsheet := &sheets.Spreadsheet{
		Properties: &sheets.SpreadsheetProperties{
			Title: spreadsheetName,
		},
	}

	// Call the Sheets API to create the spreadsheet
	spreadsheet, err := g.sheetsSvc.Spreadsheets.Create(spreadsheet).Context(g.ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to create spreadsheet: %v", err)
	}

	g.spreadsheet = spreadsheet
	g.spreadsheetID = spreadsheet.SpreadsheetId

	return nil
}

func (g *GoogleAPI) WriteData(data [][]interface{}, sheetName string) error {
	writeRange := sheetName + "!A1" // Specify the sheet name and cell range

	var vr sheets.ValueRange

	// Iterate over each column in the data slice
	for _, col := range data {
		// Filter out empty values in the column
		filteredCol := make([]interface{}, 0)
		for _, value := range col {
			if value != "" {
				filteredCol = append(filteredCol, value)
			}
		}
		// Append the filtered column to the ValueRange
		vr.Values = append(vr.Values, filteredCol)
	}

	_, err := g.sheetsSvc.Spreadsheets.Values.Update(g.spreadsheetID, writeRange, &vr).ValueInputOption("RAW").Do()
	if err != nil {
		log.Fatalf("Unable to update data in sheet. %v", err)
	}

	return nil
}

func (g *GoogleAPI) CreateSheet(sheetName string) error {
	// Create the sheet properties
	sheetProperties := &sheets.SheetProperties{
		Title: sheetName,
	}

	// Create the request to add a sheet
	addSheetRequest := &sheets.Request{
		AddSheet: &sheets.AddSheetRequest{
			Properties: sheetProperties,
		},
	}

	// Create the batch update spreadsheet request
	batchUpdateRequest := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{addSheetRequest},
	}

	// Execute the batch update request
	_, err := g.sheetsSvc.Spreadsheets.BatchUpdate(g.spreadsheetID, batchUpdateRequest).Context(g.ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to create sheet: %v", err)
	}

	return nil
}

func NewGoogleAPI(credentialsFile, tokenFile string) (*GoogleAPI, error) {
	ctx := context.Background()

	// Read the credentials file
	credentials, err := ioutil.ReadFile(credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read credentials file: %v", err)
	}

	// Parse the credentials file
	config, err := google.ConfigFromJSON(credentials, sheets.SpreadsheetsScope)
	if err != nil {
		return nil, fmt.Errorf("failed to parse credentials file: %v", err)
	}

	client := getClient(config, tokenFile)

	sheetsSvc, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	return &GoogleAPI{
		ctx:       ctx,
		sheetsSvc: sheetsSvc,
	}, nil

}

func WriteCrashLogs(crashLogs []crashlog.CrashLog) error {
	if len(crashLogs) == 0 {
		return errors.New("crashLogs slice is empty")
	}

	// Extract the year and date from the first crash log entry
	firstLogSystemTime := crashLogs[0].SystemTime
	yearDate := firstLogSystemTime.Format("2006-01-02")

	// Extract the version number from crashLogs[0].Version
	version, err := crashlogutil.ExtractVersion(crashLogs[0].Version)
	if err != nil {
		return fmt.Errorf("failed to extract version: %s", err)
	}

	// Generate the spreadsheet and sheet names
	spreadsheetName := fmt.Sprintf("CrashLogs-%s-%s-%s", crashLogs[0].Model, version, yearDate)
	sheetNamePrefix := "CrashLog"

	// Initialize the google sheet client API
	// Users/joehuang/Desktop/grafana-extract-go/credential
	api, err := NewGoogleAPI("../../credential/credentials.json", "../../credential/token.json")
	if err != nil {
		return err
	}
	log.Println("Creating Spreadsheet")
	// Create the spreadsheet
	err = api.CreateSpreadsheet(spreadsheetName)
	if err != nil {
		log.Println("Failed to creating srpeadsheet")
		return err
	}

	// Extract unique crash logs
	uniqueCrashLogs := crashlogutil.ExtractUniqueCrashLogs(crashLogs)

	// Create sheets for each unique crash log
	for i, crashLog := range uniqueCrashLogs {
		sheetName := fmt.Sprintf("%s%d", sheetNamePrefix, i+1)

		// Create a new sheet within the spreadsheet
		err = api.CreateSheet(sheetName)
		if err != nil {
			return fmt.Errorf("Failed to create sheet: %v", err)
		}

		// Prepare the crash log data
		crashLogData := crashlogutil.FilterCrashLogByValue(crashLogs, crashLog)

		// Populate the crash log data in the sheet
		for _, log := range crashLogData {
			// Apply the regex pattern to the crash log
			cleanLog := crashlogutil.ApplyRegex(log.CrashLog)

			// Split the cleaned crash log into lines
			lines := strings.Split(cleanLog, "\n")

			// Create a slice to store the column-wise data
			columnData := make([][]interface{}, 0)

			// Identify kernel panic
			kpType := crashlog.IdentifyKernelPanic(lines)
			strReason := "Reason: "
			//fmt.Println("kpType:", kpType)
			// Convert kernel panic type into a column
			columnData = append(columnData, []interface{}{strReason + kpType})

			// Convert each non-empty line of the crash log into a column
			for _, line := range lines {
				if line != "" {
					columnData = append(columnData, []interface{}{line})
				}
			}
			// Write the column-wise data to the sheet
			err = api.WriteData(columnData, sheetName)
			if err != nil {
				return fmt.Errorf("failed to write crash log line: %v", err)
			}
		}
	}

	return nil
}
