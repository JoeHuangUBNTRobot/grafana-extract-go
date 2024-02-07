package localexcel

import (
	"errors"
	"fmt"
	"grafana-extract-go/internal/app/crashlog"
	"grafana-extract-go/internal/crashlogutil"
	"strings"

	"github.com/xuri/excelize/v2"
)

func CreateExcel(data []crashlog.CrashLog, unique bool) error {
	if len(data) == 0 {
		return errors.New("data slice is empty")
	}
	// Create a new Excel file
	file := excelize.NewFile()

	// Extract unique crash logs
	crashLogs := crashlogutil.ExtractUniqueCrashLogs(data)

	// Start from the five row in default Sheet1
	sheet1row := 5

	// Create a map to store AnonymousDeviceID
	processedIDs := make(map[string]bool)

	// Create sheets for each unique crash log
	for i, crashLog := range crashLogs {
		sheetName := fmt.Sprintf("CrashLog%d", i+1)
		index, err := file.NewSheet(sheetName)
		if err != nil {
			return fmt.Errorf("failed to create new sheet: %s", err)
		}

		// Populate the crash log data in the sheet
		crashLogData := crashlogutil.FilterCrashLogByValue(data, crashLog)
		// Start from the third row in indivudual crash sheet
		row := 3 

		for _, log := range crashLogData {
			if processedIDs[log.AnonymousDeviceID] && unique {
				err := file.DeleteSheet(sheetName)
				if err != nil {
					return fmt.Errorf("failed to delete new sheet: %s", err)
				}
				continue
			}

			// Apply the regex pattern to the crash log
			cleanLog := crashlogutil.ApplyRegex(log.CrashLog)

			// Write each line of the cleaned crash log to the same column but different rows
			lines := strings.Split(cleanLog, "\n")
			// Identify kernel panic
			kpType := crashlog.IdentifyKernelPanic(lines)
			strReason := "Reason: "
			strTitle := "AnonymousDeviceID: "
			// Set the header column
			file.SetCellValue(sheetName, "A1", strReason + kpType)
			// Set the AnonymousDevice ID
			file.SetCellValue(sheetName, "A2", strTitle + log.AnonymousDeviceID)
			// calculate total AnonymousDevice ID
			sheet1cell := fmt.Sprintf("A%d", sheet1row)
			file.SetCellValue("Sheet1", sheet1cell, log.AnonymousDeviceID)
			sheet1row++
			// Take a record for handled AnonymousDeviceID
			processedIDs[log.AnonymousDeviceID] = true
			
			// Reverse the order of lines
			for j := len(lines) - 1; j >= 0; j-- {
				line := lines[j]
				if line != "" {
					cell := fmt.Sprintf("A%d", row)
					file.SetCellValue(sheetName, cell, line)
					row++
				}
			}
		}

		// Set the active sheet
		file.SetActiveSheet(index)

	}

	// Extract the year and date from the first crash log entry
	firstLogSystemTime := data[0].SystemTime
	yearDate := firstLogSystemTime.Format("2006-01-02")

	// Extract the version number from crashLog.Version
	version, err := crashlogutil.ExtractVersion(data[0].Version)
	if err != nil {
		return fmt.Errorf("failed to extract version: %s", err)
	}

	fmt.Println("Extracted version:", version)

	// Generate the file name
	fileName := fmt.Sprintf("CrashLogs-%s-%s-%s.xlsx", data[0].Model, version, yearDate)

	// Save the Excel file with the custom name
	err = file.SaveAs(fileName)
	if err != nil {
		return fmt.Errorf("failed to save Excel file: %s", err)
	}

	return nil
}

// func extractVersion(input string) (string, error) {
// 	// Define the regular expression pattern to match the version
// 	pattern := `v(\d+\.\d+\.\d+)`

// 	// Compile the regular expression
// 	regex, err := regexp.Compile(pattern)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to compile regex pattern: %s", err)
// 	}

// 	// Find the first match of the pattern in the input string
// 	match := regex.FindStringSubmatch(input)
// 	if len(match) < 2 {
// 		return "", fmt.Errorf("no version found in input")
// 	}

// 	// Extract the version from the matched group
// 	version := match[1]

// 	return version, nil
// }

// func extractUniqueCrashLogs(data []crashlog.CrashLog) []string {
// 	uniqueCrashLogs := make(map[string]bool)
// 	for _, log := range data {
// 		uniqueCrashLogs[log.CrashLog] = true
// 	}
// 	crashLogs := make([]string, 0, len(uniqueCrashLogs))
// 	for log := range uniqueCrashLogs {
// 		crashLogs = append(crashLogs, log)
// 	}
// 	sort.Strings(crashLogs)
// 	return crashLogs
// }

// func filterCrashLogByValue(data []crashlog.CrashLog, value string) []crashlog.CrashLog {
// 	var filteredData []crashlog.CrashLog
// 	for _, log := range data {
// 		if log.CrashLog == value {
// 			filteredData = append(filteredData, log)
// 		}
// 	}
// 	return filteredData
// }

// func applyRegex(crashLog string) string {
// 	// Define the regex pattern
// 	regexPattern := `<\d{1,3}>`

// 	// Create a regex object with the pattern
// 	regex := regexp.MustCompile(regexPattern)

// 	// Replace the matched patterns with a newline character
// 	cleanLog := regex.ReplaceAllString(crashLog, "\n")

// 	return cleanLog
// }
