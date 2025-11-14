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
	// Create a server instance
	srv := server.NewServer(":8080", 10)

	// Create a request to test the status endpoint
	req := &server.HTTPRequest{
		Method:  "GET",
		Path:    "/status",
		Version: "HTTP/1.1",
		Headers: make(map[string]string),
		Body:    "",
		Params:  make(map[string]string),
	}

	// Use the real StatusHandler with server
	handler := StatusHandler(srv)
	resp := handler(req)

	if resp.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	if resp.Headers["Content-Type"] != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", resp.Headers["Content-Type"])
	}

	if !strings.Contains(resp.Body, "status") || !strings.Contains(resp.Body, "stats") {
		t.Errorf("Response should contain status and stats information: %s", resp.Body)
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

func TestGetFilePath(t *testing.T) {
	filename := "test.txt"
	expectedPath := "pruebas/test.txt"

	result := getFilePath(filename)

	// Normalizar separadores de path para Windows/Linux
	if !strings.HasSuffix(result, expectedPath) && !strings.HasSuffix(result, "pruebas\\test.txt") {
		t.Errorf("Expected path to end with %s, got %s", expectedPath, result)
	}
}

func TestFaviconHandler(t *testing.T) {
	req := &server.HTTPRequest{
		Method:  "GET",
		Path:    "/favicon.ico",
		Version: "HTTP/1.1",
		Headers: make(map[string]string),
		Body:    "",
		Params:  make(map[string]string),
	}

	resp := FaviconHandler(req)

	if resp.StatusCode != 204 {
		t.Errorf("Expected status code 204, got %d", resp.StatusCode)
	}

	if resp.Body != "" {
		t.Errorf("Expected empty body, got %s", resp.Body)
	}
}

func TestEchoHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		headers        map[string]string
		body           string
		expectedStatus int
	}{
		{
			name:           "GET request",
			method:         "GET",
			headers:        map[string]string{"User-Agent": "test"},
			body:           "",
			expectedStatus: 200,
		},
		{
			name:           "POST request",
			method:         "POST",
			headers:        map[string]string{"Content-Type": "application/json"},
			body:           `{"test": "data"}`,
			expectedStatus: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &server.HTTPRequest{
				Method:  tt.method,
				Path:    "/echo",
				Version: "HTTP/1.1",
				Headers: tt.headers,
				Body:    tt.body,
				Params:  make(map[string]string),
			}

			resp := EchoHandler(req)

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			// Verificar que la respuesta contiene el método
			if !strings.Contains(resp.Body, tt.method) {
				t.Errorf("Response should contain method '%s'", tt.method)
			}

			// Verificar que la respuesta contiene el path
			if !strings.Contains(resp.Body, "/echo") {
				t.Errorf("Response should contain path '/echo'")
			}

			// Verificar Content-Type HTML
			if resp.Headers["Content-Type"] != "text/html; charset=utf-8" {
				t.Errorf("Expected Content-Type 'text/html; charset=utf-8', got %s", resp.Headers["Content-Type"])
			}

			// Si hay body, verificar que aparece en la respuesta
			if tt.body != "" && !strings.Contains(resp.Body, tt.body) {
				t.Errorf("Response should contain body content '%s'", tt.body)
			}
		})
	}
}
