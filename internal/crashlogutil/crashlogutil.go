package crashlogutil

import (
	"fmt"
	"grafana-extract-go/internal/app/crashlog"
	"regexp"
	"sort"
)

func ExtractVersion(input string) (string, error) {
	// Define the regular expression pattern to match the version
	pattern := `v(\d+\.\d+\.\d+)`

	// Compile the regular expression
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return "", fmt.Errorf("failed to compile regex pattern: %s", err)
	}

	// Find the first match of the pattern in the input string
	match := regex.FindStringSubmatch(input)
	if len(match) < 2 {
		return "", fmt.Errorf("no version found in input")
	}

	// Extract the version from the matched group
	version := match[1]

	return version, nil
}

func ExtractUniqueCrashLogs(data []crashlog.CrashLog) []string {
	uniqueCrashLogs := make(map[string]bool)
	for _, log := range data {
		uniqueCrashLogs[log.CrashLog] = true
	}
	crashLogs := make([]string, 0, len(uniqueCrashLogs))
	for log := range uniqueCrashLogs {
		crashLogs = append(crashLogs, log)
	}
	sort.Strings(crashLogs)
	return crashLogs
}

func FilterCrashLogByValue(data []crashlog.CrashLog, value string) []crashlog.CrashLog {
	var filteredData []crashlog.CrashLog
	for _, log := range data {
		if log.CrashLog == value {
			filteredData = append(filteredData, log)
		}
	}
	return filteredData
}

func ApplyRegex(crashLog string) string {
	// Define the regex pattern
	regexPattern := `<\d{1,3}>`

	// Create a regex object with the pattern
	regex := regexp.MustCompile(regexPattern)

	// Replace the matched patterns with a newline character
	cleanLog := regex.ReplaceAllString(crashLog, "\n")

	return cleanLog
}
