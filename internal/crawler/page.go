package crawler

type Page struct {
	Url        string `json:"url"`
	Depth      int    `json:"depth"`
	HttpStatus int    `json:"http_status"`
	Status     string `json:"status"`
	Error      string `json:"error"`
}
