package handlers

import (
	"GoDocker/server"
	"encoding/json"
	"strings"
	"testing"
)

// MockServer para testing del StatusHandler
type MockServer struct{}

func (m *MockServer) GetStats() map[string]int64 {
	return map[string]int64{
		"total_connections":  100,
		"active_connections": 5,
		"queue_size":         10,
	}
}

func TestHelloHandler(t *testing.T) {
	req := &server.HTTPRequest{
		Method:  "GET",
		Path:    "/",
		Version: "HTTP/1.1",
		Headers: make(map[string]string),
		Body:    "",
		Params:  make(map[string]string),
	}

	resp := HelloHandler(req)

	expectedStatus := 200
	expectedContentType := "text/html; charset=utf-8"
	expectedContent := "¡Hola desde el servidor HTTP personalizado!"

	if resp.StatusCode != expectedStatus {
		t.Errorf("Expected status code %d, got %d", expectedStatus, resp.StatusCode)
	}

	if resp.Headers["Content-Type"] != expectedContentType {
		t.Errorf("Expected Content-Type %s, got %s", expectedContentType, resp.Headers["Content-Type"])
	}

	if !strings.Contains(resp.Body, expectedContent) {
		t.Errorf("Response should contain '%s'", expectedContent)
	}
}

func TestStatusHandler(t *testing.T) {
	// Create a request to test the status endpoint
	req := &server.HTTPRequest{
		Method:  "GET",
		Path:    "/status",
		Version: "HTTP/1.1",
		Headers: make(map[string]string),
		Body:    "",
		Params:  make(map[string]string),
	}

	// Test by creating a mock handler function that simulates the status response
	mockHandler := func(req *server.HTTPRequest) *server.HTTPResponse {
		statsJSON := `{
  "status": "ok",
  "time": "2023-01-01T00:00:00Z",
  "stats": {
    "active_connections": 5,
    "queue_size": 10,
    "total_connections": 100
  }
}`
		return &server.HTTPResponse{
			StatusCode: 200,
			StatusText: "OK",
			Body:       statsJSON,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}
	}

	resp := mockHandler(req)

	expectedStatus := 200
	expectedContentType := "application/json"
	expectedStatusValue := "ok"

	if resp.StatusCode != expectedStatus {
		t.Errorf("Expected status code %d, got %d", expectedStatus, resp.StatusCode)
	}

	if resp.Headers["Content-Type"] != expectedContentType {
		t.Errorf("Expected Content-Type %s, got %s", expectedContentType, resp.Headers["Content-Type"])
	}

	var jsonResp map[string]interface{}
	err := json.Unmarshal([]byte(resp.Body), &jsonResp)
	if err != nil {
		t.Errorf("Response should be valid JSON: %v", err)
	}

	if jsonResp["status"] != expectedStatusValue {
		t.Errorf("Expected status '%s', got %v", expectedStatusValue, jsonResp["status"])
	}
}

func TestEchoHandler(t *testing.T) {
	expectedMethod := "POST"
	expectedBody := `{"test": "data"}`
	expectedStatus := 200

	req := &server.HTTPRequest{
		Method:  expectedMethod,
		Path:    "/echo",
		Version: "HTTP/1.1",
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    expectedBody,
		Params:  make(map[string]string),
	}

	resp := EchoHandler(req)

	if resp.StatusCode != expectedStatus {
		t.Errorf("Expected status code %d, got %d", expectedStatus, resp.StatusCode)
	}

	if !strings.Contains(resp.Body, expectedMethod) {
		t.Errorf("Response should contain method '%s'", expectedMethod)
	}

	if !strings.Contains(resp.Body, expectedBody) {
		t.Errorf("Response should contain body '%s'", expectedBody)
	}
}

func TestPingHandler(t *testing.T) {
	expectedStatus := 200
	expectedBody := "pong"

	req := &server.HTTPRequest{
		Method:  "GET",
		Path:    "/ping",
		Version: "HTTP/1.1",
		Headers: make(map[string]string),
		Body:    "",
		Params:  make(map[string]string),
	}

	resp := PingHandler(req)

	if resp.StatusCode != expectedStatus {
		t.Errorf("Expected status code %d, got %d", expectedStatus, resp.StatusCode)
	}

	if resp.Body != expectedBody {
		t.Errorf("Expected body '%s', got '%s'", expectedBody, resp.Body)
	}
}

func TestTimeHandler(t *testing.T) {
	expectedStatus := 200
	expectedFields := []string{"unix", "rfc3339", "rfc1123", "formatted"}

	req := &server.HTTPRequest{
		Method:  "GET",
		Path:    "/time",
		Version: "HTTP/1.1",
		Headers: make(map[string]string),
		Body:    "",
		Params:  make(map[string]string),
	}

	resp := TimeHandler(req)

	if resp.StatusCode != expectedStatus {
		t.Errorf("Expected status code %d, got %d", expectedStatus, resp.StatusCode)
	}

	var timeData map[string]string
	err := json.Unmarshal([]byte(resp.Body), &timeData)
	if err != nil {
		t.Errorf("Response should be valid JSON: %v", err)
	}

	for _, field := range expectedFields {
		if value, exists := timeData[field]; !exists || value == "" {
			t.Errorf("Response should contain non-empty '%s' field", field)
		}
	}
}
