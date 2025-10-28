package handlers

import (
	"GoDocker/server"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestFibonacciHandler(t *testing.T) {
	// Test valid request
	expectedStatus := 200
	expectedErrorStatus := 400

	params := map[string]string{"n": "5"}
	req := &server.HTTPRequest{
		Method: "GET", Path: "/fibonacci", Version: "HTTP/1.1",
		Headers: make(map[string]string), Body: "", Params: params,
	}

	resp := FibonacciHandler(req)
	if resp.StatusCode != expectedStatus {
		t.Errorf("Expected status %d, got %d", expectedStatus, resp.StatusCode)
	}

	// Test missing parameter
	req.Params = make(map[string]string)
	resp = FibonacciHandler(req)
	if resp.StatusCode != expectedErrorStatus {
		t.Errorf("Expected status %d for missing parameter, got %d", expectedErrorStatus, resp.StatusCode)
	}
}

func TestReverseHandler(t *testing.T) {
	expectedStatus := 200
	expectedReversed := "olleh"

	params := map[string]string{"text": "hello"}
	req := &server.HTTPRequest{
		Method: "GET", Path: "/reverse", Version: "HTTP/1.1",
		Headers: make(map[string]string), Body: "", Params: params,
	}

	resp := ReverseHandler(req)
	if resp.StatusCode != expectedStatus {
		t.Errorf("Expected status %d, got %d", expectedStatus, resp.StatusCode)
	}

	var result map[string]string
	json.Unmarshal([]byte(resp.Body), &result)
	if result["reversed"] != expectedReversed {
		t.Errorf("Expected '%s', got '%s'", expectedReversed, result["reversed"])
	}
}

func TestToUpperHandler(t *testing.T) {
	expectedStatus := 200
	expectedUpper := "HELLO WORLD"

	params := map[string]string{"text": "hello world"}
	req := &server.HTTPRequest{
		Method: "GET", Path: "/toupper", Version: "HTTP/1.1",
		Headers: make(map[string]string), Body: "", Params: params,
	}

	resp := ToUpperHandler(req)
	if resp.StatusCode != expectedStatus {
		t.Errorf("Expected status %d, got %d", expectedStatus, resp.StatusCode)
	}

	var result map[string]string
	json.Unmarshal([]byte(resp.Body), &result)
	if result["upper"] != expectedUpper {
		t.Errorf("Expected '%s', got '%s'", expectedUpper, result["upper"])
	}
}

func TestRandomNumberHandler(t *testing.T) {
	// Test valid request
	expectedStatus := 200
	expectedErrorStatus := 400
	expectedMin := 1
	expectedMax := 10

	params := map[string]string{"min": "1", "max": "10"}
	req := &server.HTTPRequest{
		Method: "GET", Path: "/random", Version: "HTTP/1.1",
		Headers: make(map[string]string), Body: "", Params: params,
	}

	resp := RandomNumberHandler(req)
	if resp.StatusCode != expectedStatus {
		t.Errorf("Expected status %d, got %d", expectedStatus, resp.StatusCode)
	}

	// Verify random number is in expected range
	var result map[string]interface{}
	json.Unmarshal([]byte(resp.Body), &result)
	if randomNum, ok := result["random"].(float64); ok {
		if int(randomNum) < expectedMin || int(randomNum) > expectedMax {
			t.Errorf("Random number %d should be between %d and %d", int(randomNum), expectedMin, expectedMax)
		}
	}

	// Test missing parameters
	req.Params = map[string]string{"min": "1"}
	resp = RandomNumberHandler(req)
	if resp.StatusCode != expectedErrorStatus {
		t.Errorf("Expected status %d for missing max parameter, got %d", expectedErrorStatus, resp.StatusCode)
	}
}

func TestHashHandler(t *testing.T) {
	expectedStatus := 200
	expectedErrorStatus := 400
	expectedText := "hello"

	params := map[string]string{"text": expectedText}
	req := &server.HTTPRequest{
		Method: "GET", Path: "/hash", Version: "HTTP/1.1",
		Headers: make(map[string]string), Body: "", Params: params,
	}

	resp := HashHandler(req)
	if resp.StatusCode != expectedStatus {
		t.Errorf("Expected status %d, got %d", expectedStatus, resp.StatusCode)
	}

	var result map[string]interface{}
	err := json.Unmarshal([]byte(resp.Body), &result)
	if err != nil {
		t.Errorf("Failed to parse JSON: %v", err)
		return
	}

	if result["text"] != expectedText {
		t.Errorf("Expected text '%s', got %v", expectedText, result["text"])
	}

	// Test missing parameter
	req.Params = make(map[string]string)
	resp = HashHandler(req)
	if resp.StatusCode != expectedErrorStatus {
		t.Errorf("Expected status %d for missing parameter, got %d", expectedErrorStatus, resp.StatusCode)
	}
}

func TestCreateFileHandler(t *testing.T) {
	expectedStatus := 201
	expectedFileName := "test_file.txt"

	params := map[string]string{
		"name": expectedFileName, "content": "test", "repeat": "1",
	}
	req := &server.HTTPRequest{
		Method: "POST", Path: "/createFile", Version: "HTTP/1.1",
		Headers: make(map[string]string), Body: "", Params: params,
	}

	resp := CreateFileHandler(req)
	if resp.StatusCode != expectedStatus {
		t.Errorf("Expected status %d, got %d", expectedStatus, resp.StatusCode)
	}

	// Cleanup
	os.Remove(expectedFileName)
}

func TestDeleteFileHandler(t *testing.T) {
	// Create test file
	expectedStatus := 200
	expectedFileName := "test_delete.txt"

	file, _ := os.Create(expectedFileName)
	file.Close()

	params := map[string]string{"name": expectedFileName}
	req := &server.HTTPRequest{
		Method: "DELETE", Path: "/deleteFile", Version: "HTTP/1.1",
		Headers: make(map[string]string), Body: "", Params: params,
	}

	resp := DeleteFileHandler(req)
	if resp.StatusCode != expectedStatus {
		t.Errorf("Expected status %d, got %d", expectedStatus, resp.StatusCode)
	}
}

func TestSimulateHandler(t *testing.T) {
	expectedStatus := 200
	expectedTask := "cpu"
	expectedDuration := "1"

	params := map[string]string{"seconds": expectedDuration, "task": expectedTask}
	req := &server.HTTPRequest{
		Method: "POST", Path: "/simulate", Version: "HTTP/1.1",
		Headers: make(map[string]string), Body: "", Params: params,
	}

	resp := SimulateHandler(req)
	if resp.StatusCode != expectedStatus {
		t.Errorf("Expected status %d, got %d", expectedStatus, resp.StatusCode)
	}

	var result map[string]string
	json.Unmarshal([]byte(resp.Body), &result)
	if result["task"] != expectedTask {
		t.Errorf("Expected task '%s', got '%s'", expectedTask, result["task"])
	}
}

func TestSleepHandler(t *testing.T) {
	expectedStatus := 200

	params := map[string]string{"seconds": "1"}
	req := &server.HTTPRequest{
		Method: "POST", Path: "/sleep", Version: "HTTP/1.1",
		Headers: make(map[string]string), Body: "", Params: params,
	}

	resp := SleepHandler(req)
	if resp.StatusCode != expectedStatus {
		t.Errorf("Expected status %d, got %d", expectedStatus, resp.StatusCode)
	}
}

func TestLoadTestHandler(t *testing.T) {
	expectedStatus := 200
	expectedTasks := 2
	expectedSleep := 1

	params := map[string]string{"tasks": "2", "sleep": "1"}
	req := &server.HTTPRequest{
		Method: "POST", Path: "/loadtest", Version: "HTTP/1.1",
		Headers: make(map[string]string), Body: "", Params: params,
	}

	resp := LoadTestHandler(req)
	if resp.StatusCode != expectedStatus {
		t.Errorf("Expected status %d, got %d", expectedStatus, resp.StatusCode)
	}

	var result map[string]interface{}
	json.Unmarshal([]byte(resp.Body), &result)
	if tasks, ok := result["tasks"].(float64); ok {
		if int(tasks) != expectedTasks {
			t.Errorf("Expected tasks %d, got %d", expectedTasks, int(tasks))
		}
	}
	if sleep, ok := result["sleep"].(float64); ok {
		if int(sleep) != expectedSleep {
			t.Errorf("Expected sleep %d, got %d", expectedSleep, int(sleep))
		}
	}
}

func TestHelpHandler(t *testing.T) {
	expectedStatus := 200
	expectedContent := "API Help"
	expectedContentType := "text/html"

	req := &server.HTTPRequest{
		Method: "GET", Path: "/help", Version: "HTTP/1.1",
		Headers: make(map[string]string), Body: "", Params: make(map[string]string),
	}

	resp := HelpHandler(req)
	if resp.StatusCode != expectedStatus {
		t.Errorf("Expected status %d, got %d", expectedStatus, resp.StatusCode)
	}

	if resp.Headers["Content-Type"] != expectedContentType {
		t.Errorf("Expected Content-Type %s, got %s", expectedContentType, resp.Headers["Content-Type"])
	}

	if !strings.Contains(resp.Body, expectedContent) {
		t.Errorf("Response should contain '%s'", expectedContent)
	}
}
