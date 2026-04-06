package crawler

import (
	"encoding/json"
)

func ReportFormat(report Report) ([]byte, error) {
	return json.Marshal(report)
}
