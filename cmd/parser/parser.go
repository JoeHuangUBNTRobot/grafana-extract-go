package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/tealeg/xlsx"
)

func main() {
	// Get keywords from command line arguments
	keyword1Ptr := flag.String("keyword1", "", "First keyword for searching in Excel")
	keyword2Ptr := flag.String("keyword2", "", "Second keyword for searching in Excel")
	filePathPtr := flag.String("file", "", "Excel file path")
	flag.Parse()

	// Check if required parameters are provided
	if *keyword1Ptr == "" || *filePathPtr == "" {
		fmt.Println("Please provide the first keyword and Excel file path")
		os.Exit(1)
	}

	// Open the Excel file
	xlFile, err := xlsx.OpenFile(*filePathPtr)
	if err != nil {
		fmt.Println("Unable to open Excel file:", err)
		os.Exit(1)
	}

	// Initialize counts for each keyword
	keyword1Count := 0
	keyword2Count := 0
	bothKeywordsCount := 0

	// #1
	// Iterate through each sheet in the Excel file
	//for _, sheet := range xlFile.Sheets {
	// #2
	// Iterate through each sheet, starting from 1
	for i := 1; i < len(xlFile.Sheets)-1; i++ {
		// Get the current sheet
		sheet := xlFile.Sheets[i]

		// Flags to track the presence of each keyword in the current sheet
		keyword1Found := false
		keyword2Found := false

		// Iterate through each row and cell in the sheet
		for _, row := range sheet.Rows {
			// If the second keyword is provided, check both keywords in each cell
			for _, cell := range row.Cells {
				cellValue := strings.TrimSpace(cell.String())

				if strings.Contains(cellValue, *keyword1Ptr) {
					// First keyword found in the cell
					keyword1Found = true
				}
				if *keyword2Ptr != "" && strings.Contains(cellValue, *keyword2Ptr) {
					// Second keyword found in the cell
					keyword2Found = true
				}
			}
		}

		// Update counts based on keyword presence in the sheet
		if keyword1Found {
			keyword1Count++
		}
		if keyword2Found {
			keyword2Count++
		}
		if keyword1Found && keyword2Found {
			bothKeywordsCount++
		}
	}

	// Output the results
	if *keyword2Ptr == "" {
		fmt.Printf("Number of sheets where the keyword '%s' appears in the Excel file: %d\n", *keyword1Ptr, keyword1Count)
	} else {
		fmt.Printf("Number of sheets where the keyword '%s' appears in the Excel file: %d\n", *keyword1Ptr, keyword1Count)
		fmt.Printf("Number of sheets where the keyword '%s' appears in the Excel file: %d\n", *keyword2Ptr, keyword2Count)
		fmt.Printf("Number of sheets where both keywords appear in the Excel file: %d\n", bothKeywordsCount)
	}
}
