package handlers

import (
	"GoDocker/server"
	"context"
	"fmt"
	"strconv"
)

// ServerTaskExecutor implementa TaskExecutor para ejecutar tareas del servidor
type ServerTaskExecutor struct {
	srv *server.Server
}

// NewServerTaskExecutor crea un nuevo executor
func NewServerTaskExecutor(srv *server.Server) *ServerTaskExecutor {
	return &ServerTaskExecutor{srv: srv}
}

// Execute ejecuta una tarea basándose en su nombre
func (e *ServerTaskExecutor) Execute(ctx context.Context, task string, params map[string]string) (map[string]interface{}, error) {
	// Crear request simulado
	req := &server.HTTPRequest{
		Method: "GET",
		Path:   "/" + task,
		Params: params,
	}

	// Ejecutar handler correspondiente según el task
	var resp *server.HTTPResponse

	switch task {
	// CPU-bound tasks
	case "isprime":
		resp = IsPrimeHandler(req)
	case "factor":
		resp = FactorHandler(req)
	case "pi":
		resp = PiHandler(req)
	case "mandelbrot":
		resp = MandelbrotHandler(req)
	case "matrixmul":
		resp = MatrixMulHandler(req)
	case "fibonacci":
		resp = FibonacciHandler(req)

	// IO-bound tasks
	case "sortfile":
		resp = SortFileHandler(req)
	case "wordcount":
		resp = WordCountHandler(req)
	case "grep":
		resp = GrepHandler(req)
	case "compress":
		resp = CompressHandler(req)
	case "hashfile":
		resp = HashFileHandler(req)

	default:
		return nil, fmt.Errorf("unknown task: %s", task)
	}

	// Verificar si el contexto fue cancelado
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Convertir respuesta a resultado
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("task failed: %s (status %d)", resp.Body, resp.StatusCode)
	}

	result := map[string]interface{}{
		"status_code": resp.StatusCode,
		"body":        resp.Body,
	}

	// Intentar parsear métricas si existen en headers
	if execTime, ok := resp.Headers["X-Exec-Time"]; ok {
		if ms, err := strconv.ParseFloat(execTime, 64); err == nil {
			result["exec_time_ms"] = ms
		}
	}

	return result, nil
}
