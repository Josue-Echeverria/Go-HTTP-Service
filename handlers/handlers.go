package handlers

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"GoDocker/server"
)

// Directorio base para archivos de prueba
const filesDir = "pruebas"

// getFilePath construye la ruta completa al archivo en el directorio de pruebas
func getFilePath(filename string) string {
	return filepath.Join(filesDir, filename)
}

// HelloHandler maneja peticiones a /
func HelloHandler(req *server.HTTPRequest) *server.HTTPResponse {
	body := `<!DOCTYPE html>
	<html>
	<head>
		<title>Custom HTTP Server</title>
	</head>
	<body>
		<h1>¡Hola desde el servidor HTTP personalizado!</h1>
		<p>Este servidor fue construido desde cero usando solo net y goroutines.</p>
		<p>Hora del servidor: ` + time.Now().Format(time.RFC1123) + `</p>
	</body>
	</html>`

	return &server.HTTPResponse{
		StatusCode: 200,
		StatusText: "OK",
		Body:       body,
		Headers: map[string]string{
			"Content-Type": "text/html; charset=utf-8",
		},
	}
}

// StatusHandler maneja peticiones a /status
func StatusHandler(srv *server.Server) server.HandlerFunc {
	return func(req *server.HTTPRequest) *server.HTTPResponse {
		stats := srv.GetStats()

		statsJSON, _ := json.MarshalIndent(map[string]interface{}{
			"status": "ok",
			"stats":  stats,
		}, "", "  ")

		return &server.HTTPResponse{
			StatusCode: 200,
			StatusText: "OK",
			Body:       string(statsJSON),
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}
	}
}

// MetricsHandler maneja peticiones a /metrics con métricas detalladas
func MetricsHandler(srv *server.Server) server.HandlerFunc {
	return func(req *server.HTTPRequest) *server.HTTPResponse {
		metrics := srv.GetMetrics()

		metricsJSON, err := json.MarshalIndent(metrics, "", "  ")
		if err != nil {
			return &server.HTTPResponse{
				StatusCode: 500,
				StatusText: "Internal Server Error",
				Body:       `{"error": "failed to marshal metrics"}`,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
			}
		}

		return &server.HTTPResponse{
			StatusCode: 200,
			StatusText: "OK",
			Body:       string(metricsJSON),
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}
	}
}

// EchoHandler maneja peticiones a /echo
func EchoHandler(req *server.HTTPRequest) *server.HTTPResponse {
	response := fmt.Sprintf(`<!DOCTYPE html>
	<html>
	<head>
		<title>Echo</title>
	</head>
	<body>
    <h1>Echo Request</h1>
    <h2>Method: %s</h2>
    <h2>Path: %s</h2>
    <h2>Version: %s</h2>
    <h3>Headers:</h3>
    <ul>`, req.Method, req.Path, req.Version)

	for key, value := range req.Headers {
		response += fmt.Sprintf("<li><strong>%s:</strong> %s</li>", key, value)
	}

	response += "</ul>"

	if req.Body != "" {
		response += fmt.Sprintf("<h3>Body:</h3><pre>%s</pre>", req.Body)
	}

	response += "</body></html>"

	return &server.HTTPResponse{
		StatusCode: 200,
		StatusText: "OK",
		Body:       response,
		Headers: map[string]string{
			"Content-Type": "text/html; charset=utf-8",
		},
	}
}

// PingHandler maneja peticiones a /ping
func PingHandler(req *server.HTTPRequest) *server.HTTPResponse {
	return &server.HTTPResponse{
		StatusCode: 200,
		StatusText: "OK",
		Body:       "pong",
		Headers: map[string]string{
			"Content-Type": "text/plain",
		},
	}
}

// FaviconHandler maneja peticiones a /favicon.ico
func FaviconHandler(req *server.HTTPRequest) *server.HTTPResponse {
	// Retornar 204 No Content para favicon (sin ícono)
	return &server.HTTPResponse{
		StatusCode: 204,
		StatusText: "No Content",
		Body:       "",
		Headers:    map[string]string{},
	}
}

// TimeHandler maneja peticiones a /time
func TimeHandler(req *server.HTTPRequest) *server.HTTPResponse {
	now := time.Now()
	timeData := map[string]string{
		"unix":      fmt.Sprintf("%d", now.Unix()),
		"rfc3339":   now.Format(time.RFC3339),
		"rfc1123":   now.Format(time.RFC1123),
		"formatted": now.Format("2006-01-02 15:04:05 MST"),
	}

	jsonData, _ := json.MarshalIndent(timeData, "", "  ")

	return &server.HTTPResponse{
		StatusCode: 200,
		StatusText: "OK",
		Body:       string(jsonData),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}
}
