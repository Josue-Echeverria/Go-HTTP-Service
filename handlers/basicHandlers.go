package handlers

import (
	"GoDocker/server"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"
)

// nextRandom genera el siguiente número pseudoaleatorio usando algoritmo LCG
func nextRandom(seed uint64) uint64 {
	if seed == 0 {
		seed = uint64(time.Now().UnixNano())
	}
	// Constantes del algoritmo LCG (Linear Congruential Generator)
	seed = seed*1103515245 + 12345
	return seed
}

// randomInRange genera un número aleatorio entre min y max (inclusive)
func randomInRange(min, max int) int {
	if min > max {
		min, max = max, min
	}
	if min == max {
		return min
	}

	rangeSize := max - min + 1
	return min + int(nextRandom(0)%uint64(rangeSize))
}

// /fibonacci?n=<n>
func FibonacciHandler(req *server.HTTPRequest) *server.HTTPResponse {
	val, ok := req.Params["n"]
	if !ok || val == "" {
		return &server.HTTPResponse{
			StatusCode: 400,
			StatusText: "Bad Request",
			Body:       `{"error":"missing required query parameter 'n'"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}
	}

	const maxN = 1000
	n, _ := strconv.Atoi(val)
	if n > maxN {
		return &server.HTTPResponse{
			StatusCode: 413,
			StatusText: "Payload Too Large",
			Body:       `{"error":"n too large; maximum allowed is 1000"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}
	}

	if n <= 0 {
		return &server.HTTPResponse{
			StatusCode: 400,
			StatusText: "Bad Request",
			Body:       `{"error":"n must be greater than 0"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}
	}

	fibSeq := make([]int, n)

	if n >= 1 {
		fibSeq[0] = 0
	}
	if n >= 2 {
		fibSeq[1] = 1
	}

	for i := 2; i < n; i++ {
		fibSeq[i] = fibSeq[i-1] + fibSeq[i-2]
	}

	jsonData, _ := json.MarshalIndent(fibSeq, "", "  ")

	return &server.HTTPResponse{
		StatusCode: 200,
		StatusText: "OK",
		Body:       string(jsonData),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}
}

// /createFile?name=filename&content=text&repeat=x
func CreateFileHandler(req *server.HTTPRequest) *server.HTTPResponse {
	name, nameOk := req.Params["name"]
	content, contentOk := req.Params["content"]
	repeatStr, repeatOk := req.Params["repeat"]

	if !nameOk || !contentOk || !repeatOk {
		return &server.HTTPResponse{
			StatusCode: 400,
			StatusText: "Bad Request",
			Body:       `{"error":"missing required query parameters"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}
	}

	repeat, err := strconv.Atoi(repeatStr)
	if err != nil || repeat < 1 {
		return &server.HTTPResponse{
			StatusCode: 400,
			StatusText: "Bad Request",
			Body:       `{"error":"invalid repeat parameter"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}
	}

	// Crear archivo
	file, err := os.Create(name)
	if err != nil {
		return &server.HTTPResponse{
			StatusCode: 500,
			StatusText: "Internal Server Error",
			Body:       `{"error":"failed to create file"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}
	}
	defer file.Close()

	for i := 0; i < repeat; i++ {
		_, err := file.WriteString(content + "\n")
		if err != nil {
			return &server.HTTPResponse{
				StatusCode: 500,
				StatusText: "Internal Server Error",
				Body:       `{"error":"failed to write to file"}`,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
			}
		}
	}

	return &server.HTTPResponse{
		StatusCode: 201,
		StatusText: "Created",
		Body:       `{"message":"file created successfully"}`,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}
}

// /deleteFile?name=filename
func DeleteFileHandler(req *server.HTTPRequest) *server.HTTPResponse {
	name, nameOk := req.Params["name"]
	if !nameOk || name == "" {
		return &server.HTTPResponse{
			StatusCode: 400,
			StatusText: "Bad Request",
			Body:       `{"error":"missing required query parameter 'name'"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}
	}

	err := os.Remove(name)
	if err != nil {
		return &server.HTTPResponse{
			StatusCode: 500,
			StatusText: "Internal Server Error",
			Body:       `{"error":"failed to delete file"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}
	}

	return &server.HTTPResponse{
		StatusCode: 200,
		StatusText: "OK",
		Body:       `{"message":"file deleted successfully"}`,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}
}

// /reverse?text=yourtext
func ReverseHandler(req *server.HTTPRequest) *server.HTTPResponse {
	text, textOk := req.Params["text"]
	if !textOk {
		return &server.HTTPResponse{
			StatusCode: 400,
			StatusText: "Bad Request",
			Body:       `{"error":"missing required query parameter 'text'"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}
	}

	runes := []rune(text)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	jsonData, _ := json.MarshalIndent(map[string]string{"reversed": string(runes)}, "", "  ")

	return &server.HTTPResponse{
		StatusCode: 200,
		StatusText: "OK",
		Body:       string(jsonData),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}
}

// /toupper?text=yourtext
func ToUpperHandler(req *server.HTTPRequest) *server.HTTPResponse {
	text, textOk := req.Params["text"]
	if !textOk {
		return &server.HTTPResponse{
			StatusCode: 400,
			StatusText: "Bad Request",
			Body:       `{"error":"missing required query parameter 'text'"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}
	}

	upper := strings.ToUpper(text)
	jsonData, _ := json.MarshalIndent(map[string]string{"upper": upper}, "", "  ")

	return &server.HTTPResponse{
		StatusCode: 200,
		StatusText: "OK",
		Body:       string(jsonData),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}
}

// /random?min=x&max=y
func RandomNumberHandler(req *server.HTTPRequest) *server.HTTPResponse {
	minStr, minOk := req.Params["min"]
	maxStr, maxOk := req.Params["max"]
	if !minOk || !maxOk {
		return &server.HTTPResponse{
			StatusCode: 400,
			StatusText: "Bad Request",
			Body:       `{"error":"missing required query parameters 'min' and 'max'"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}
	}

	// Parsear los valores min y max
	min, err := strconv.Atoi(minStr)
	if err != nil {
		return &server.HTTPResponse{
			StatusCode: 400,
			StatusText: "Bad Request",
			Body:       `{"error":"invalid 'min' parameter - must be an integer"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}
	}

	max, err := strconv.Atoi(maxStr)
	if err != nil {
		return &server.HTTPResponse{
			StatusCode: 400,
			StatusText: "Bad Request",
			Body:       `{"error":"invalid 'max' parameter - must be an integer"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}
	}

	// Validar rango razonable
	const maxRange = 1000000
	if max-min > maxRange {
		return &server.HTTPResponse{
			StatusCode: 400,
			StatusText: "Bad Request",
			Body:       `{"error":"range too large - maximum allowed range is 1,000,000"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}
	}

	// Generar número aleatorio usando algoritmo LCG simple
	randomNum := randomInRange(min, max)

	// Crear respuesta JSON con información adicional
	response := map[string]interface{}{
		"random":    randomNum,
		"min":       min,
		"max":       max,
		"range":     max - min + 1,
		"algorithm": "LCG (Linear Congruential Generator)",
	}

	jsonData, _ := json.MarshalIndent(response, "", "  ")

	return &server.HTTPResponse{
		StatusCode: 200,
		StatusText: "OK",
		Body:       string(jsonData),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}
}

// /hash?text=yourtext
func HashHandler(req *server.HTTPRequest) *server.HTTPResponse {
	text, textOk := req.Params["text"]
	if !textOk {
		return &server.HTTPResponse{
			StatusCode: 400,
			StatusText: "Bad Request",
			Body:       `{"error":"missing required query parameter 'text'"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}
	}

	// Implementar hash DJB2 simple
	hash := uint32(5381)
	for _, char := range []byte(text) {
		hash = ((hash << 5) + hash) + uint32(char)
	}

	response := map[string]interface{}{
		"hash": hash,
		"text": text,
	}

	jsonData, _ := json.MarshalIndent(response, "", "  ")

	return &server.HTTPResponse{
		StatusCode: 200,
		StatusText: "OK",
		Body:       string(jsonData),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}
}

// /simulate?seconds=s&task=name
func SimulateHandler(req *server.HTTPRequest) *server.HTTPResponse {
	secondsStr, secondsOk := req.Params["seconds"]
	task, taskOk := req.Params["task"]
	if !secondsOk || !taskOk {
		return &server.HTTPResponse{
			StatusCode: 400,
			StatusText: "Bad Request",
			Body:       `{"error":"missing required query parameters 'seconds' and 'task'"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}
	}

	seconds, err := strconv.Atoi(secondsStr)
	if err != nil || seconds < 1 || seconds > 10 {
		return &server.HTTPResponse{
			StatusCode: 400,
			StatusText: "Bad Request",
			Body:       `{"error":"seconds must be between 1 and 10"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}
	}

	startTime := time.Now()
	targetDuration := time.Duration(seconds) * time.Second
	var operations int64

	// Simulación
	for time.Since(startTime) < targetDuration {
		for i := 0; i < 1000; i++ {
			_ = float64(i) * 3.14159
			operations++
		}
	}

	elapsed := time.Since(startTime)
	response := map[string]interface{}{
		"status":     "completed",
		"task":       task,
		"duration":   elapsed.String(),
		"operations": operations,
	}

	jsonData, _ := json.MarshalIndent(response, "", "  ")

	return &server.HTTPResponse{
		StatusCode: 200,
		StatusText: "OK",
		Body:       string(jsonData),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}
}

// /sleep?seconds=s
func SleepHandler(req *server.HTTPRequest) *server.HTTPResponse {
	secondsStr, secondsOk := req.Params["seconds"]
	if !secondsOk {
		return &server.HTTPResponse{
			StatusCode: 400,
			StatusText: "Bad Request",
			Body:       `{"error":"missing required query parameter 'seconds'"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}
	}

	seconds, err := strconv.Atoi(secondsStr)
	if err != nil || seconds < 1 || seconds > 60 {
		return &server.HTTPResponse{
			StatusCode: 400,
			StatusText: "Bad Request",
			Body:       `{"error":"seconds must be between 1 and 60"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}
	}

	startTime := time.Now()
	targetDuration := time.Duration(seconds) * time.Second

	// Simular "espera" usando busy-wait con operaciones ligeras
	var iterations int64
	for time.Since(startTime) < targetDuration {
		// Operaciones muy ligeras para simular espera sin usar sleep
		for i := 0; i < 1000; i++ {
			iterations++
			// Operación mínima para evitar optimización del compilador
			_ = iterations % 2
		}
	}

	elapsed := time.Since(startTime)

	response := map[string]interface{}{
		"status":          "completed",
		"requested_sleep": secondsStr + "s",
		"actual_duration": elapsed.String(),
		"iterations":      iterations,
		"method":          "busy-wait simulation",
		"note":            "Simulated sleep without blocking thread",
	}

	jsonData, _ := json.MarshalIndent(response, "", "  ")

	return &server.HTTPResponse{
		StatusCode: 200,
		StatusText: "OK",
		Body:       string(jsonData),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}
}

// /loadtest?tasks=n&sleep=x
func LoadTestHandler(req *server.HTTPRequest) *server.HTTPResponse {
	tasksStr, tasksOk := req.Params["tasks"]
	sleepStr, sleepOk := req.Params["sleep"]
	if !tasksOk || !sleepOk {
		return &server.HTTPResponse{
			StatusCode: 400,
			StatusText: "Bad Request",
			Body:       `{"error":"missing required query parameters 'tasks' and 'sleep'"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}
	}

	tasks, err := strconv.Atoi(tasksStr)
	if err != nil || tasks < 1 || tasks > 100 {
		return &server.HTTPResponse{
			StatusCode: 400,
			StatusText: "Bad Request",
			Body:       `{"error":"tasks must be between 1 and 100"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}
	}

	sleepMs, err := strconv.Atoi(sleepStr)
	if err != nil || sleepMs < 0 || sleepMs > 1000 {
		return &server.HTTPResponse{
			StatusCode: 400,
			StatusText: "Bad Request",
			Body:       `{"error":"sleep must be between 0 and 1000 milliseconds"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}
	}

	startTime := time.Now()
	var totalOperations int64

	// Ejecutar tareas simples
	for taskID := 1; taskID <= tasks; taskID++ {
		// Trabajo simple por tarea
		for i := 0; i < 1000; i++ {
			_ = float64(i*taskID) * 2.0
			totalOperations++
		}

		// Simular sleep si se especifica
		if sleepMs > 0 {
			sleepStart := time.Now()
			targetSleep := time.Duration(sleepMs) * time.Millisecond
			for time.Since(sleepStart) < targetSleep {
				_ = totalOperations % 2 // Operación mínima
			}
		}
	}

	totalDuration := time.Since(startTime)

	response := map[string]interface{}{
		"status":           "completed",
		"total_tasks":      tasks,
		"sleep_per_task":   sleepStr + "ms",
		"total_duration":   totalDuration.String(),
		"total_operations": totalOperations,
		"tasks_per_second": float64(tasks) / totalDuration.Seconds(),
	}

	jsonData, _ := json.MarshalIndent(response, "", "  ")

	return &server.HTTPResponse{
		StatusCode: 200,
		StatusText: "OK",
		Body:       string(jsonData),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}
}

// /help
func HelpHandler(req *server.HTTPRequest) *server.HTTPResponse {
	helpText := `<!DOCTYPE html>
	<html>
	<head>
		<title>Help</title>
	</head>
	<body>
		<h1>API Help</h1>
		<ul>
			<li><a href="/hash?text=yourtext">/hash?text=yourtext</a></li>
			<li><a href="/simulate?seconds=s&task=name">/simulate?seconds=s&task=name</a></li>
			<li><a href="/sleep?seconds=s">/sleep?seconds=s</a></li>
			<li><a href="/loadtest?tasks=n&sleep=x">/loadtest?tasks=n&sleep=x</a></li>
			<li><a href="/help">/help</a></li>
		</ul>
	</body>
	</html>`

	return &server.HTTPResponse{
		StatusCode: 200,
		StatusText: "OK",
		Body:       helpText,
		Headers: map[string]string{
			"Content-Type": "text/html",
		},
	}
}
