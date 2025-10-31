package handlers

import (
	"GoDocker/server"
	"encoding/json"
	"strconv"
)

// JobSubmitHandler maneja /jobs/submit
func JobSubmitHandler(jm *server.JobManager) server.HandlerFunc {
	return func(req *server.HTTPRequest) *server.HTTPResponse {
		task, _ := req.Params["task"]
		if task == "" {
			return &server.HTTPResponse{
				StatusCode: 400,
				StatusText: "Bad Request",
				Headers:    map[string]string{"Content-Type": "application/json"},
				Body:       `{"error": "missing task parameter"}`,
			}
		}

		// Obtener prioridad (por defecto normal)
		priority := server.PriorityNormal
		if prioStr, ok := req.Params["prio"]; ok {
			switch prioStr {
			case "low":
				priority = server.PriorityLow
			case "high":
				priority = server.PriorityHigh
			}
		}

		// Copiar parámetros (excepto task y prio)
		params := make(map[string]string)
		for k, v := range req.Params {
			if k != "task" && k != "prio" {
				params[k] = v
			}
		}

		// Enviar trabajo
		job, err := jm.Submit(task, params, priority)
		if err != nil {
			if err.Error() == "queue full" {
				retryAfter := 5000 // 5 segundos
				return &server.HTTPResponse{
					StatusCode: 503,
					StatusText: "Service Unavailable",
					Headers: map[string]string{
						"Content-Type": "application/json",
						"Retry-After":  strconv.Itoa(retryAfter / 1000),
					},
					Body: `{"error": "queue full", "retry_after_ms": ` + strconv.Itoa(retryAfter) + `}`,
				}
			}

			return &server.HTTPResponse{
				StatusCode: 500,
				StatusText: "Internal Server Error",
				Headers:    map[string]string{"Content-Type": "application/json"},
				Body:       `{"error": "` + err.Error() + `"}`,
			}
		}

		result := map[string]interface{}{
			"job_id": job.ID,
			"status": job.Status,
		}

		jsonData, _ := json.MarshalIndent(result, "", "  ")

		return &server.HTTPResponse{
			StatusCode: 200,
			StatusText: "OK",
			Headers:    map[string]string{"Content-Type": "application/json"},
			Body:       string(jsonData),
		}
	}
}

// JobStatusHandler maneja /jobs/status
func JobStatusHandler(jm *server.JobManager) server.HandlerFunc {
	return func(req *server.HTTPRequest) *server.HTTPResponse {
		jobID, _ := req.Params["id"]
		if jobID == "" {
			return &server.HTTPResponse{
				StatusCode: 400,
				StatusText: "Bad Request",
				Headers:    map[string]string{"Content-Type": "application/json"},
				Body:       `{"error": "missing id parameter"}`,
			}
		}

		job, err := jm.GetJob(jobID)
		if err != nil {
			return &server.HTTPResponse{
				StatusCode: 404,
				StatusText: "Not Found",
				Headers:    map[string]string{"Content-Type": "application/json"},
				Body:       `{"error": "job not found"}`,
			}
		}

		info := job.GetInfo()

		// Retornar solo status, progress, eta
		result := map[string]interface{}{
			"job_id":   info["job_id"],
			"status":   info["status"],
			"progress": info["progress"],
			"eta_ms":   info["eta_ms"],
		}

		if err, ok := info["error"]; ok {
			result["error"] = err
		}

		jsonData, _ := json.MarshalIndent(result, "", "  ")

		return &server.HTTPResponse{
			StatusCode: 200,
			StatusText: "OK",
			Headers:    map[string]string{"Content-Type": "application/json"},
			Body:       string(jsonData),
		}
	}
}

// JobResultHandler maneja /jobs/result
func JobResultHandler(jm *server.JobManager) server.HandlerFunc {
	return func(req *server.HTTPRequest) *server.HTTPResponse {
		jobID, _ := req.Params["id"]
		if jobID == "" {
			return &server.HTTPResponse{
				StatusCode: 400,
				StatusText: "Bad Request",
				Headers:    map[string]string{"Content-Type": "application/json"},
				Body:       `{"error": "missing id parameter"}`,
			}
		}

		job, err := jm.GetJob(jobID)
		if err != nil {
			return &server.HTTPResponse{
				StatusCode: 404,
				StatusText: "Not Found",
				Headers:    map[string]string{"Content-Type": "application/json"},
				Body:       `{"error": "job not found"}`,
			}
		}

		info := job.GetInfo()

		// Si no está done, retornar estado actual
		if job.Status != server.JobDone {
			result := map[string]interface{}{
				"job_id": info["job_id"],
				"status": info["status"],
			}

			if err, ok := info["error"]; ok {
				result["error"] = err
			}

			jsonData, _ := json.MarshalIndent(result, "", "  ")

			return &server.HTTPResponse{
				StatusCode: 200,
				StatusText: "OK",
				Headers:    map[string]string{"Content-Type": "application/json"},
				Body:       string(jsonData),
			}
		}

		// Retornar resultado completo
		jsonData, _ := json.MarshalIndent(info, "", "  ")

		return &server.HTTPResponse{
			StatusCode: 200,
			StatusText: "OK",
			Headers:    map[string]string{"Content-Type": "application/json"},
			Body:       string(jsonData),
		}
	}
}

// JobCancelHandler maneja /jobs/cancel
func JobCancelHandler(jm *server.JobManager) server.HandlerFunc {
	return func(req *server.HTTPRequest) *server.HTTPResponse {
		jobID, _ := req.Params["id"]
		if jobID == "" {
			return &server.HTTPResponse{
				StatusCode: 400,
				StatusText: "Bad Request",
				Headers:    map[string]string{"Content-Type": "application/json"},
				Body:       `{"error": "missing id parameter"}`,
			}
		}

		canceled, err := jm.CancelJob(jobID)
		if err != nil {
			return &server.HTTPResponse{
				StatusCode: 404,
				StatusText: "Not Found",
				Headers:    map[string]string{"Content-Type": "application/json"},
				Body:       `{"error": "job not found"}`,
			}
		}

		status := "canceled"
		if !canceled {
			status = "not_cancelable"
		}

		result := map[string]interface{}{
			"job_id": jobID,
			"status": status,
		}

		jsonData, _ := json.MarshalIndent(result, "", "  ")

		return &server.HTTPResponse{
			StatusCode: 200,
			StatusText: "OK",
			Headers:    map[string]string{"Content-Type": "application/json"},
			Body:       string(jsonData),
		}
	}
}
