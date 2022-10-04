package input

import (
	"encoding/json"
	"errors"
	// "fmt"
	// "net"
	// "net/http"
	// "net/url"
	// "os"
)

// ----------------------------------------------------------------------------
type Record struct {
	DataSource string `json:"DATA_SOURCE"`
	RecordId string `json:"RECORD_ID"`
}

// ----------------------------------------------------------------------------
func ValidateLine(line string) (bool, error) {
	var record Record
	valid := json.Unmarshal([]byte(line), &record) == nil
	if valid {
		return ValidateRecord(record)
	}
	return valid, errors.New("JSON-line not well formed.")
}

// ----------------------------------------------------------------------------
func ValidateRecord(record Record) (bool, error) {
	// FIXME: errors should be specific to the input method
	//  ala rabbitmq message ID?
	if record.DataSource == "" {
		return false, errors.New("A DATA_SOURCE field is required.")
	}
	if record.RecordId == "" {
		return false, errors.New("A RECORD_ID field is required.")
	}
	return true, nil
}
