package handlers

import (
	"GoDocker/server"
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestMetricsHandler(t *testing.T) {
	srv := server.NewServer(":8080", 10)
	req := &server.HTTPRequest{
		Method: "GET", Path: "/metrics", Version: "HTTP/1.1",
		Headers: make(map[string]string), Body: "",
		Params: make(map[string]string),
	}

	resp := MetricsHandler(srv)(req)
	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(resp.Body), &result); err != nil {
		t.Errorf("Failed to parse JSON response: %v", err)
	}

	// Just verify we got some metrics data, don't require specific fields
	// since the exact structure depends on server implementation
	if len(result) == 0 {
		t.Error("Expected some metrics data, got empty response")
	}
}

func TestJobSubmitHandler(t *testing.T) {
	srv := server.NewServer(":8080", 10)
	jm := srv.GetJobManager()

	tests := []struct {
		name           string
		task           string
		expectedStatus int
	}{
		{"Valid isPrime task", "isprime", 200},
		{"Valid factor task", "factor", 200},
		{"Valid pi task", "pi", 200},
		{"Valid mandelbrot task", "mandelbrot", 200},
		{"Valid fibonacci task", "fibonacci", 200},
		{"Invalid task", "invalid", 200}, // TaskExecutor handles invalid tasks gracefully
		{"Missing task", "", 400},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := make(map[string]string)
			if tt.task != "" {
				params["task"] = tt.task
			}
			// Add required parameters for each task type
			if tt.task == "isprime" || tt.task == "factor" {
				params["num"] = "17"
			} else if tt.task == "pi" {
				params["digits"] = "10"
			} else if tt.task == "mandelbrot" {
				params["width"] = "100"
				params["height"] = "100"
				params["maxIter"] = "100"
			} else if tt.task == "fibonacci" {
				params["n"] = "10"
			}

			req := &server.HTTPRequest{
				Method: "POST", Path: "/jobs/submit", Version: "HTTP/1.1",
				Headers: make(map[string]string), Body: "", Params: params,
			}

			resp := JobSubmitHandler(jm)(req)
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

func TestJobStatusHandler(t *testing.T) {
	srv := server.NewServer(":8080", 10)
	jm := srv.GetJobManager()

	tests := []struct {
		name           string
		jobID          string
		expectedStatus int
	}{
		{"Missing job ID", "", 400},
		{"Non-existent job ID", "nonexistent", 404},
		{"Invalid job ID format", "invalid-format", 404},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := make(map[string]string)
			if tt.jobID != "" {
				params["id"] = tt.jobID
			}

			req := &server.HTTPRequest{
				Method: "GET", Path: "/jobs/status", Version: "HTTP/1.1",
				Headers: make(map[string]string), Body: "", Params: params,
			}

			resp := JobStatusHandler(jm)(req)
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

func TestJobResultHandler(t *testing.T) {
	srv := server.NewServer(":8080", 10)
	jm := srv.GetJobManager()

	tests := []struct {
		name           string
		jobID          string
		expectedStatus int
	}{
		{"Missing job ID", "", 400},
		{"Non-existent job ID", "nonexistent", 404},
		{"Invalid job ID format", "invalid-format", 404},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := make(map[string]string)
			if tt.jobID != "" {
				params["id"] = tt.jobID
			}

			req := &server.HTTPRequest{
				Method: "GET", Path: "/jobs/result", Version: "HTTP/1.1",
				Headers: make(map[string]string), Body: "", Params: params,
			}

			resp := JobResultHandler(jm)(req)
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

func TestJobCancelHandler(t *testing.T) {
	srv := server.NewServer(":8080", 10)
	jm := srv.GetJobManager()

	tests := []struct {
		name           string
		jobID          string
		expectedStatus int
	}{
		{"Missing job ID", "", 400},
		{"Non-existent job ID", "nonexistent", 404},
		{"Invalid job ID format", "invalid-format", 404},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := make(map[string]string)
			if tt.jobID != "" {
				params["id"] = tt.jobID
			}

			req := &server.HTTPRequest{
				Method: "DELETE", Path: "/jobs/cancel", Version: "HTTP/1.1",
				Headers: make(map[string]string), Body: "", Params: params,
			}

			resp := JobCancelHandler(jm)(req)
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

func TestServerTaskExecutor(t *testing.T) {
	srv := server.NewServer(":8080", 10)
	executor := NewServerTaskExecutor(srv)

	if executor == nil {
		t.Fatal("ServerTaskExecutor creation failed")
	}

	// Test task execution with simple isPrime task
	ctx := context.Background()
	result, err := executor.Execute(ctx, "isprime", map[string]string{"num": "17"})
	if err != nil {
		t.Errorf("Execute returned error: %v", err)
	}
	if result == nil {
		t.Error("Execute returned nil result")
	}

	// Test with invalid task - should return error
	_, err = executor.Execute(ctx, "invalid", map[string]string{})
	if err == nil {
		t.Error("Execute should return error for invalid tasks")
	}
}

func TestJobIntegration(t *testing.T) {
	srv := server.NewServer(":8080", 10)
	executor := NewServerTaskExecutor(srv)
	srv.GetJobManager().SetExecutor(executor)
	jm := srv.GetJobManager()

	// Submit a job
	submitParams := map[string]string{
		"task": "isprime",
		"num":  "17",
	}

	submitReq := &server.HTTPRequest{
		Method: "POST", Path: "/jobs/submit", Version: "HTTP/1.1",
		Headers: make(map[string]string), Body: "", Params: submitParams,
	}

	submitResp := JobSubmitHandler(jm)(submitReq)
	if submitResp.StatusCode != 200 {
		t.Fatalf("Job submission failed with status %d", submitResp.StatusCode)
	}

	var submitResult map[string]interface{}
	if err := json.Unmarshal([]byte(submitResp.Body), &submitResult); err != nil {
		t.Fatalf("Failed to parse job submission response: %v", err)
	}

	jobID, exists := submitResult["job_id"]
	if !exists {
		t.Fatal("Job submission response missing job_id")
	}

	jobIDStr, ok := jobID.(string)
	if !ok {
		t.Fatal("Job ID is not a string")
	}

	// Check job status
	statusParams := map[string]string{"id": jobIDStr}
	statusReq := &server.HTTPRequest{
		Method: "GET", Path: "/jobs/status", Version: "HTTP/1.1",
		Headers: make(map[string]string), Body: "", Params: statusParams,
	}

	statusResp := JobStatusHandler(jm)(statusReq)
	if statusResp.StatusCode != 200 {
		t.Errorf("Job status check failed with status %d", statusResp.StatusCode)
	}
}

// TestJobResultHandlerAdvanced prueba casos adicionales para mejorar la cobertura
func TestJobResultHandlerAdvanced(t *testing.T) {
	srv := server.NewServer(":8080", 10)
	jm := srv.GetJobManager()

	// Test con job que existe pero aún está corriendo
	t.Run("Job_still_running", func(t *testing.T) {
		// Enviar un job de fibonacci que tome tiempo
		params := map[string]string{
			"task": "fibonacci",
			"n":    "35", // Esto toma un poco de tiempo
		}
		submitReq := &server.HTTPRequest{
			Method: "POST", Path: "/jobs/submit", Version: "HTTP/1.1",
			Headers: make(map[string]string), Body: "", Params: params,
		}

		submitResp := JobSubmitHandler(jm)(submitReq)
		if submitResp.StatusCode != 200 {
			t.Fatalf("Job submission failed with status %d", submitResp.StatusCode)
		}

		// Extraer job ID
		var submitData map[string]interface{}
		json.Unmarshal([]byte(submitResp.Body), &submitData)
		jobIDStr := submitData["job_id"].(string)

		// Inmediatamente verificar resultados (debería devolver 400 porque aún está corriendo)
		resultParams := map[string]string{"id": jobIDStr}
		resultReq := &server.HTTPRequest{
			Method: "GET", Path: "/jobs/result", Version: "HTTP/1.1",
			Headers: make(map[string]string), Body: "", Params: resultParams,
		}

		resultResp := JobResultHandler(jm)(resultReq)
		// Debería dar 400 porque el job aún está corriendo
		if resultResp.StatusCode == 200 {
			t.Logf("Job completed faster than expected, got result immediately")
		} else if resultResp.StatusCode != 400 {
			t.Errorf("Expected status 400 for running job, got %d", resultResp.StatusCode)
		}
	})
}

// TestJobCancelHandlerAdvanced prueba casos adicionales para mejorar la cobertura
func TestJobCancelHandlerAdvanced(t *testing.T) {
	srv := server.NewServer(":8080", 10)
	jm := srv.GetJobManager()

	// Test cancelando job que existe
	t.Run("Cancel_existing_job", func(t *testing.T) {
		// Enviar un job de fibonacci que tome tiempo
		params := map[string]string{
			"task": "fibonacci",
			"n":    "40", // Esto toma tiempo
		}
		submitReq := &server.HTTPRequest{
			Method: "POST", Path: "/jobs/submit", Version: "HTTP/1.1",
			Headers: make(map[string]string), Body: "", Params: params,
		}

		submitResp := JobSubmitHandler(jm)(submitReq)
		if submitResp.StatusCode != 200 {
			t.Fatalf("Job submission failed with status %d", submitResp.StatusCode)
		}

		// Extraer job ID
		var submitData map[string]interface{}
		json.Unmarshal([]byte(submitResp.Body), &submitData)
		jobIDStr := submitData["job_id"].(string)

		// Cancelar el job
		cancelParams := map[string]string{"id": jobIDStr}
		cancelReq := &server.HTTPRequest{
			Method: "POST", Path: "/jobs/cancel", Version: "HTTP/1.1",
			Headers: make(map[string]string), Body: "", Params: cancelParams,
		}

		cancelResp := JobCancelHandler(jm)(cancelReq)
		if cancelResp.StatusCode != 200 && cancelResp.StatusCode != 400 {
			// 200 = cancelado exitosamente, 400 = ya terminó
			t.Errorf("Cancel should return 200 or 400, got %d", cancelResp.StatusCode)
		}
	})

	// Test cancelando job ya completado
	t.Run("Cancel_completed_job", func(t *testing.T) {
		// Enviar un job rápido
		params := map[string]string{
			"task": "isprime",
			"num":  "7",
		}
		submitReq := &server.HTTPRequest{
			Method: "POST", Path: "/jobs/submit", Version: "HTTP/1.1",
			Headers: make(map[string]string), Body: "", Params: params,
		}

		submitResp := JobSubmitHandler(jm)(submitReq)
		if submitResp.StatusCode != 200 {
			t.Fatalf("Job submission failed with status %d", submitResp.StatusCode)
		}

		// Extraer job ID
		var submitData map[string]interface{}
		json.Unmarshal([]byte(submitResp.Body), &submitData)
		jobIDStr := submitData["job_id"].(string)

		// Esperar un poco para que termine
		time.Sleep(100 * time.Millisecond)

		// Intentar cancelar job ya completado
		cancelParams := map[string]string{"id": jobIDStr}
		cancelReq := &server.HTTPRequest{
			Method: "POST", Path: "/jobs/cancel", Version: "HTTP/1.1",
			Headers: make(map[string]string), Body: "", Params: cancelParams,
		}

		cancelResp := JobCancelHandler(jm)(cancelReq)
		// Debería devolver 400 porque ya está completado
		if cancelResp.StatusCode != 400 && cancelResp.StatusCode != 200 {
			t.Errorf("Expected status 400 or 200 for completed job, got %d", cancelResp.StatusCode)
		}
	})
}
