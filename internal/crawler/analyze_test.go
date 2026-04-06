package crawler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestAnalyze_Success(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<html><body>Test page</body></html>`))
	}))
	defer server.Close()

	// Create options
	options := Options{
		Client:  &http.Client{Timeout: 10 * time.Second},
		Retries: 3,
		Depths:  2,
		Delay:   1 * time.Second,
		Url:     server.URL,
	}

	// Run Analyze
	reportJSON, err := Analyze(context.Background(), options)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if reportJSON == "" {
		t.Error("Expected non-empty report JSON, got empty string")
	}

	// Try to parse JSON to verify it's valid
	var report Report
	if err := json.Unmarshal([]byte(reportJSON), &report); err != nil {
		t.Errorf("Expected valid JSON, got parsing error: %v", err)
	}

	// Verify report content
	if report.RootUrl != server.URL {
		t.Errorf("Expected RootUrl %s, got %s", server.URL, report.RootUrl)
	}

	if len(report.Pages) == 0 {
		t.Error("Expected at least one page in report")
	}
}

func TestAnalyze_InvalidURL(t *testing.T) {
	options := Options{
		Client:  &http.Client{},
		Retries: 3,
		Depths:  2,
		Delay:   1 * time.Second,
		Url:     "://invalid-url",
	}

	reportJSON, err := Analyze(context.Background(), options)
	if err == nil {
		t.Error("Expected error for invalid URL, got nil")
	}

	if reportJSON != "" {
		t.Errorf("Expected empty report on error, got: %s", reportJSON)
	}
}

func TestAnalyze_RequestError(t *testing.T) {
	// Create a server that will be closed to simulate connection error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	server.Close()

	options := Options{
		Client:  &http.Client{Timeout: 1 * time.Second},
		Retries: 3,
		Depths:  2,
		Delay:   1 * time.Second,
		Url:     server.URL,
	}

	reportJSON, err := Analyze(context.Background(), options)
	if err == nil {
		t.Error("Expected error for failed request, got nil")
	}

	if reportJSON != "" {
		t.Errorf("Expected empty report on error, got: %s", reportJSON)
	}
}

func TestAnalyze_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	options := Options{
		Client:  &http.Client{Timeout: 10 * time.Second},
		Retries: 3,
		Depths:  2,
		Delay:   1 * time.Second,
		Url:     server.URL,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	reportJSON, err := Analyze(ctx, options)
	if err == nil {
		t.Error("Expected error for cancelled context, got nil")
	}

	if reportJSON != "" {
		t.Errorf("Expected empty report on error, got: %s", reportJSON)
	}
}

func TestAnalyze_ContextTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	options := Options{
		Client:  &http.Client{Timeout: 10 * time.Second},
		Retries: 3,
		Depths:  2,
		Delay:   1 * time.Second,
		Url:     server.URL,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	reportJSON, err := Analyze(ctx, options)
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}

	if reportJSON != "" {
		t.Errorf("Expected empty report on error, got: %s", reportJSON)
	}
}

func TestAnalyze_HTTPErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 page not found"))
	}))
	defer server.Close()

	options := Options{
		Client:  &http.Client{Timeout: 10 * time.Second},
		Retries: 3,
		Depths:  2,
		Delay:   1 * time.Second,
		Url:     server.URL,
	}

	reportJSON, err := Analyze(context.Background(), options)
	if err != nil {
		t.Errorf("Expected no error for HTTP error status, got: %v", err)
	}

	if reportJSON == "" {
		t.Error("Expected non-empty report even for error status")
	}

	// Parse and verify the report contains the error page
	var report Report
	if err := json.Unmarshal([]byte(reportJSON), &report); err != nil {
		t.Errorf("Expected valid JSON, got error: %v", err)
	}

	if len(report.Pages) > 0 && report.Pages[0].HttpStatus != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", report.Pages[0].HttpStatus)
	}
}

func TestAnalyze_WithDifferentDepths(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	testCases := []struct {
		name   string
		depths int
	}{
		{"Depth 0", 0},
		{"Depth 1", 1},
		{"Depth 3", 3},
		{"Depth 5", 5},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			options := Options{
				Client:  &http.Client{Timeout: 10 * time.Second},
				Retries: 3,
				Depths:  tc.depths,
				Delay:   1 * time.Second,
				Url:     server.URL,
			}

			reportJSON, err := Analyze(context.Background(), options)
			if err != nil {
				t.Errorf("Expected no error for depth %d, got: %v", tc.depths, err)
			}

			if reportJSON == "" {
				t.Errorf("Expected non-empty report for depth %d", tc.depths)
			}

			// Verify depth in report
			var report Report
			if err := json.Unmarshal([]byte(reportJSON), &report); err == nil {
				if report.Depth != tc.depths {
					t.Errorf("Expected report depth %d, got %d", tc.depths, report.Depth)
				}
			}
		})
	}
}

func TestAnalyze_ClientTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	options := Options{
		Client:  &http.Client{Timeout: 1 * time.Second}, // Short timeout
		Retries: 3,
		Depths:  2,
		Delay:   1 * time.Second,
		Url:     server.URL,
	}

	reportJSON, err := Analyze(context.Background(), options)
	if err == nil {
		t.Error("Expected timeout error from client, got nil")
	}

	if reportJSON != "" {
		t.Errorf("Expected empty report on timeout, got: %s", reportJSON)
	}
}

func TestAnalyze_NilClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	options := Options{
		Client:  nil, // Nil client will cause panic
		Retries: 3,
		Depths:  2,
		Delay:   1 * time.Second,
		Url:     server.URL,
	}

	// This should panic or cause nil pointer dereference
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for nil client, but no panic occurred")
		}
	}()

	_, _ = Analyze(context.Background(), options)
}

func TestAnalyze_EmptyURL(t *testing.T) {
	options := Options{
		Client:  &http.Client{},
		Retries: 3,
		Depths:  2,
		Delay:   1 * time.Second,
		Url:     "",
	}

	reportJSON, err := Analyze(context.Background(), options)
	if err == nil {
		t.Error("Expected error for empty URL, got nil")
	}

	if reportJSON != "" {
		t.Errorf("Expected empty report on error, got: %s", reportJSON)
	}
}

// Integration test that verifies the entire flow including report creation
func TestAnalyze_Integration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body><a href="/page1">Link</a></body></html>`))
	}))
	defer server.Close()

	options := Options{
		Client:  &http.Client{Timeout: 10 * time.Second},
		Retries: 2,
		Depths:  1,
		Delay:   100 * time.Millisecond,
		Url:     server.URL,
	}

	reportJSON, err := Analyze(context.Background(), options)
	if err != nil {
		t.Errorf("Integration test failed: %v", err)
	}

	if reportJSON == "" {
		t.Error("Expected non-empty report JSON")
	}

	// Verify JSON structure
	var report Report
	if err := json.Unmarshal([]byte(reportJSON), &report); err != nil {
		t.Errorf("Failed to parse report JSON: %v", err)
	}

	// Check required fields
	if report.RootUrl == "" {
		t.Error("RootUrl should not be empty")
	}

	if report.GeneratedAt.IsZero() {
		t.Error("GeneratedAt should be set")
	}
}

// Test with custom RoundTripper to simulate various scenarios
type errorRoundTripper struct {
	err error
}

func (e errorRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, e.err
}

func TestAnalyze_TransportError(t *testing.T) {
	expectedErr := errors.New("transport error")
	client := &http.Client{
		Transport: errorRoundTripper{err: expectedErr},
	}

	options := Options{
		Client:  client,
		Retries: 3,
		Depths:  2,
		Delay:   1 * time.Second,
		Url:     "http://example.com",
	}

	reportJSON, err := Analyze(context.Background(), options)
	if err == nil {
		t.Error("Expected transport error, got nil")
	}

	if reportJSON != "" {
		t.Errorf("Expected empty report on error, got: %s", reportJSON)
	}
}

// Table-driven test for various HTTP status codes
func TestAnalyze_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		statusCode  int
		expectError bool
	}{
		{"OK Status", http.StatusOK, false},
		{"Created Status", http.StatusCreated, false},
		{"Bad Request", http.StatusBadRequest, false},
		{"Internal Error", http.StatusInternalServerError, false},
		{"Gateway Timeout", http.StatusGatewayTimeout, false},
		{"Forbidden", http.StatusForbidden, false},
		{"Unauthorized", http.StatusUnauthorized, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			options := Options{
				Client:  &http.Client{Timeout: 10 * time.Second},
				Retries: 3,
				Depths:  2,
				Delay:   1 * time.Second,
				Url:     server.URL,
			}

			reportJSON, err := Analyze(context.Background(), options)
			if (err != nil) != tt.expectError {
				t.Errorf("Analyze() error = %v, expectError %v", err, tt.expectError)
			}

			if !tt.expectError && reportJSON == "" {
				t.Error("Expected non-empty report for successful request")
			}
		})
	}
}

// Test that the returned JSON contains all expected fields
func TestAnalyze_ReportJSONStructure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	options := Options{
		Client:  &http.Client{Timeout: 10 * time.Second},
		Retries: 3,
		Depths:  2,
		Delay:   1 * time.Second,
		Url:     server.URL,
	}

	reportJSON, err := Analyze(context.Background(), options)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	var report Report
	if err := json.Unmarshal([]byte(reportJSON), &report); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Verify all fields
	if report.RootUrl != server.URL {
		t.Errorf("RootUrl mismatch: got %s, want %s", report.RootUrl, server.URL)
	}

	if report.Depth != options.Depths {
		t.Errorf("Depth mismatch: got %d, want %d", report.Depth, options.Depths)
	}

	if report.GeneratedAt.IsZero() {
		t.Error("GeneratedAt is zero")
	}

	if len(report.Pages) == 0 {
		t.Error("Pages should not be empty")
	}

	// Verify page fields
	page := report.Pages[0]
	if page.Url != server.URL {
		t.Errorf("Page URL mismatch: got %s, want %s", page.Url, server.URL)
	}

	if page.HttpStatus != http.StatusOK {
		t.Errorf("Page HTTP status mismatch: got %d, want %d", page.HttpStatus, http.StatusOK)
	}

	if page.Status == "" {
		t.Error("Page status should not be empty")
	}
}

// Test that JSON is properly formatted (not minified, with indentation)
func TestAnalyze_JSONFormatting(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	options := Options{
		Client:  &http.Client{Timeout: 10 * time.Second},
		Retries: 3,
		Depths:  2,
		Delay:   1 * time.Second,
		Url:     server.URL,
	}

	reportJSON, err := Analyze(context.Background(), options)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check that JSON contains newlines (pretty printed)
	if !strings.Contains(reportJSON, "\n") {
		t.Error("Expected pretty-printed JSON with newlines")
	}

	// Check that JSON contains indentation
	if !strings.Contains(reportJSON, "  ") && !strings.Contains(reportJSON, "\t") {
		t.Error("Expected JSON with indentation")
	}
}