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

// Test para casos edge y mejorar cobertura
func TestFibonacciHandlerEdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		n              string
		expectedStatus int
	}{
		{"Zero", "0", 400}, // Fibonacci handler rechaza 0
		{"One", "1", 200},
		{"Large valid", "50", 200},
		{"Too large", "1001", 413}, // Request Entity Too Large
		{"Negative", "-5", 400},
		{"Invalid string", "abc", 400},
		{"Empty string", "", 400},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := map[string]string{"n": tt.n}
			req := &server.HTTPRequest{
				Method: "GET", Path: "/fibonacci", Version: "HTTP/1.1",
				Headers: make(map[string]string), Body: "", Params: params,
			}

			resp := FibonacciHandler(req)
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

func TestRandomNumberHandlerEdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		min            string
		max            string
		expectedStatus int
	}{
		{"Equal min max", "5", "5", 200},
		{"Reversed range", "10", "1", 200}, // Should swap internally
		{"Large range", "1", "1000000", 200},
		{"Invalid min", "abc", "10", 400},
		{"Invalid max", "1", "xyz", 400},
		{"Missing min", "", "10", 400},
		{"Missing max", "1", "", 400},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := map[string]string{"min": tt.min, "max": tt.max}
			req := &server.HTTPRequest{
				Method: "GET", Path: "/random", Version: "HTTP/1.1",
				Headers: make(map[string]string), Body: "", Params: params,
			}

			resp := RandomNumberHandler(req)
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

func TestCreateFileHandlerEdgeCases(t *testing.T) {
	defer cleanupBasicTestFiles(t)

	tests := []struct {
		name           string
		fileName       string
		content        string
		repeat         string
		expectedStatus int
	}{
		{"Normal file", "test.txt", "content", "1", 201}, // Created status
		{"High repeat", "test2.txt", "x", "1000", 201},
		{"Zero repeat", "test3.txt", "content", "0", 400}, // Invalid repeat
		{"Invalid repeat", "test4.txt", "content", "abc", 400},
		{"Missing name", "", "content", "1", 500},                  // Internal error when name missing
		{"Missing content", "test5.txt", "", "1", 201},             // Allows empty content
		{"Empty repeat defaults", "test6.txt", "content", "", 400}, // Empty repeat is invalid
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := map[string]string{
				"name":    tt.fileName,
				"content": tt.content,
				"repeat":  tt.repeat,
			}
			req := &server.HTTPRequest{
				Method: "POST", Path: "/file", Version: "HTTP/1.1",
				Headers: make(map[string]string), Body: "", Params: params,
			}

			resp := CreateFileHandler(req)
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

func TestSimulateHandlerEdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		seconds        string
		task           string
		expectedStatus int
	}{
		{"Zero seconds", "0", "test", 400}, // Invalid - no simulation for 0 seconds
		{"Max seconds", "30", "test", 400}, // Simulate handler tiene límites diferentes
		{"Too many seconds", "31", "test", 400},
		{"Invalid seconds", "abc", "test", 400},
		{"Negative seconds", "-1", "test", 400},
		{"Missing task", "5", "", 200},                                  // Permite task vacío
		{"Long task name", "5", "very_long_task_name_for_testing", 200}, // Permite task largo
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := map[string]string{
				"seconds": tt.seconds,
				"task":    tt.task,
			}
			req := &server.HTTPRequest{
				Method: "POST", Path: "/simulate", Version: "HTTP/1.1",
				Headers: make(map[string]string), Body: "", Params: params,
			}

			resp := SimulateHandler(req)
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

// Helper function for cleaning basic test files
func cleanupBasicTestFiles(t *testing.T) {
	testFiles := []string{
		"test.txt", "test2.txt", "test3.txt", "test4.txt",
		"test5.txt", "test6.txt", "testfile.txt",
	}

	for _, file := range testFiles {
		fullPath := getFilePath(file)
		if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: could not remove test file %s: %v", fullPath, err)
		}
	}
}

// Tests para funciones auxiliares y casos edge no cubiertos
func TestNextRandomFunction(t *testing.T) {
	// Test con seed 0 (debería usar timestamp)
	result1 := nextRandom(0)
	result2 := nextRandom(0)

	// Los resultados deberían ser diferentes debido a timestamps diferentes
	if result1 == result2 {
		// Es posible que sean iguales si se ejecutan muy rápido, no es error crítico
		t.Logf("nextRandom(0) results: %d, %d - may be same due to timing", result1, result2)
	}

	// Test con seed específico (debería ser determinístico)
	seed := uint64(12345)
	expected := seed*1103515245 + 12345
	actual := nextRandom(seed)

	if actual != expected {
		t.Errorf("nextRandom(%d) = %d, expected %d", seed, actual, expected)
	}
}

func TestRandomInRangeFunction(t *testing.T) {
	tests := []struct {
		name string
		min  int
		max  int
	}{
		{"Normal range", 1, 10},
		{"Reversed range", 10, 1}, // Should swap internally
		{"Equal values", 5, 5},
		{"Single value range", 0, 0},
		{"Large range", 1, 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := randomInRange(tt.min, tt.max)

			expectedMin := tt.min
			expectedMax := tt.max
			if expectedMin > expectedMax {
				expectedMin, expectedMax = expectedMax, expectedMin
			}

			if result < expectedMin || result > expectedMax {
				t.Errorf("randomInRange(%d, %d) = %d, should be in range [%d, %d]",
					tt.min, tt.max, result, expectedMin, expectedMax)
			}
		})
	}
}

func TestDeleteFileHandlerMissingParam(t *testing.T) {
	// Test sin parámetro name
	req := &server.HTTPRequest{
		Method: "DELETE", Path: "/file", Version: "HTTP/1.1",
		Headers: make(map[string]string), Body: "",
		Params: make(map[string]string), // Sin parámetro name
	}

	resp := DeleteFileHandler(req)
	if resp.StatusCode != 400 {
		t.Errorf("Expected status 400 for missing name param, got %d", resp.StatusCode)
	}
}

func TestSleepHandlerEdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		seconds        string
		expectedStatus int
	}{
		{"Zero seconds", "0", 400},
		{"Negative seconds", "-1", 400},
		{"Invalid seconds", "abc", 400},
		{"Too many seconds", "61", 400},
		{"Valid seconds", "1", 200},
		{"Max valid seconds", "5", 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := map[string]string{"seconds": tt.seconds}
			req := &server.HTTPRequest{
				Method: "POST", Path: "/sleep", Version: "HTTP/1.1",
				Headers: make(map[string]string), Body: "", Params: params,
			}

			resp := SleepHandler(req)
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

func TestLoadTestHandlerEdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		tasks          string
		sleep          string
		expectedStatus int
	}{
		{"Zero tasks", "0", "1", 400},
		{"Too many tasks", "101", "1", 400},
		{"Negative tasks", "-1", "1", 400},
		{"Invalid tasks", "abc", "1", 400},
		{"Invalid sleep", "5", "abc", 400},
		{"Negative sleep", "5", "-1", 400},
		{"Valid minimal", "1", "0", 200},
		{"Valid normal", "10", "1", 200},
		{"Max tasks", "100", "0", 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := map[string]string{
				"tasks": tt.tasks,
				"sleep": tt.sleep,
			}
			req := &server.HTTPRequest{
				Method: "POST", Path: "/loadtest", Version: "HTTP/1.1",
				Headers: make(map[string]string), Body: "", Params: params,
			}

			resp := LoadTestHandler(req)
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

func TestReverseHandlerMissingParam(t *testing.T) {
	// Test sin parámetro text
	req := &server.HTTPRequest{
		Method: "PUT", Path: "/reverse", Version: "HTTP/1.1",
		Headers: make(map[string]string), Body: "",
		Params: make(map[string]string), // Sin parámetro text
	}

	resp := ReverseHandler(req)
	if resp.StatusCode != 400 {
		t.Errorf("Expected status 400 for missing text param, got %d", resp.StatusCode)
	}
}

func TestToUpperHandlerMissingParam(t *testing.T) {
	// Test sin parámetro text
	req := &server.HTTPRequest{
		Method: "PUT", Path: "/toupper", Version: "HTTP/1.1",
		Headers: make(map[string]string), Body: "",
		Params: make(map[string]string), // Sin parámetro text
	}

	resp := ToUpperHandler(req)
	if resp.StatusCode != 400 {
		t.Errorf("Expected status 400 for missing text param, got %d", resp.StatusCode)
	}
}

func TestHashHandlerMissingParam(t *testing.T) {
	// Test sin parámetro text
	req := &server.HTTPRequest{
		Method: "PUT", Path: "/hash", Version: "HTTP/1.1",
		Headers: make(map[string]string), Body: "",
		Params: make(map[string]string), // Sin parámetro text
	}

	resp := HashHandler(req)
	if resp.StatusCode != 400 {
		t.Errorf("Expected status 400 for missing text param, got %d", resp.StatusCode)
	}
}
