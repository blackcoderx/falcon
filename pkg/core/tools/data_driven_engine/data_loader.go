package data_driven_engine

import (
	"fmt"
)

// DataLoader loads data from files or generates fake data.
type DataLoader struct {
	Source string
}

// Load retrieves rows of data as maps.
func (l *DataLoader) Load(variables []string, maxRows int) ([]map[string]interface{}, error) {
	if l.Source == "fake" || l.Source == "random" {
		return l.generateFakeData(variables, maxRows), nil
	}

	// For simulation, if file doesn't exist, return error
	// In real tool, we would parse CSV or JSON.
	return nil, fmt.Errorf("data source file not found: %s", l.Source)
}

func (l *DataLoader) generateFakeData(variables []string, count int) []map[string]interface{} {
	if count <= 0 {
		count = 10
	}

	var rows []map[string]interface{}
	for i := 0; i < count; i++ {
		row := make(map[string]interface{})
		for _, v := range variables {
			row[v] = fmt.Sprintf("fake_%s_%d", v, i)
		}
		rows = append(rows, row)
	}
	return rows
}
