package crawler

import (
	"code/internal/crawler/models"
	"encoding/json"
)

func ReportFormat(report models.Report) ([]byte, error) {
	return json.MarshalIndent(report, "", "    ")
}
