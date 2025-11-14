package main

import (
	"GoDocker/handlers"
	"GoDocker/server"
	"testing"
)

// Test básico para verificar que main puede configurar el servidor
func TestMainConfiguration(t *testing.T) {
	// Test de configuración básica como lo haría main()
	addr := ":8080"
	poolSize := 20

	// Crear servidor como en main
	srv := server.NewServer(addr, poolSize)
	if srv == nil {
		t.Fatal("Server creation failed")
	}

	// Configurar executor como en main
	executor := handlers.NewServerTaskExecutor(srv)
	if executor == nil {
		t.Fatal("Executor creation failed")
	}

	srv.GetJobManager().SetExecutor(executor)

	// Registrar algunos handlers como en main
	srv.HandleFunc("GET", "/", handlers.HelloHandler)
	srv.HandleFunc("GET", "/status", handlers.StatusHandler(srv))
	srv.HandleFunc("GET", "/ping", handlers.PingHandler)

	// Verificar que el servidor fue configurado correctamente
	if srv.GetJobManager() == nil {
		t.Error("JobManager not configured")
	}
}

// Test para verificar configuración de rutas como en main
func TestRouteConfiguration(t *testing.T) {
	srv := server.NewServer(":8080", 10)

	// Configurar algunas rutas principales como en main
	srv.HandleFunc("GET", "/", handlers.HelloHandler)
	srv.HandleFunc("GET", "/ping", handlers.PingHandler)
	srv.HandleFunc("GET", "/time", handlers.TimeHandler)
	srv.HandleFunc("GET", "/fibonacci", handlers.FibonacciHandler)
	srv.HandleFunc("POST", "/file", handlers.CreateFileHandler)
	srv.HandleFunc("DELETE", "/file", handlers.DeleteFileHandler)
	srv.HandleFunc("PUT", "/reverse", handlers.ReverseHandler)
	srv.HandleFunc("PUT", "/toupper", handlers.ToUpperHandler)
	srv.HandleFunc("GET", "/random", handlers.RandomNumberHandler)
	srv.HandleFunc("PUT", "/hash", handlers.HashHandler)
	srv.HandleFunc("POST", "/simulate", handlers.SimulateHandler)
	srv.HandleFunc("POST", "/sleep", handlers.SleepHandler)
	srv.HandleFunc("POST", "/loadtest", handlers.LoadTestHandler)
	srv.HandleFunc("GET", "/help", handlers.HelpHandler)
	srv.HandleFunc("GET", "/isprime", handlers.IsPrimeHandler)
	srv.HandleFunc("GET", "/factor", handlers.FactorHandler)
	srv.HandleFunc("GET", "/pi", handlers.PiHandler)
	srv.HandleFunc("GET", "/mandelbrot", handlers.MandelbrotHandler)
	srv.HandleFunc("GET", "/matrixmul", handlers.MatrixMulHandler)
	srv.HandleFunc("GET", "/sortfile", handlers.SortFileHandler)
	srv.HandleFunc("GET", "/wordcount", handlers.WordCountHandler)
	srv.HandleFunc("GET", "/grep", handlers.GrepHandler)
	srv.HandleFunc("GET", "/compress", handlers.CompressHandler)
	srv.HandleFunc("GET", "/hashfile", handlers.HashFileHandler)

	// Job management routes (necesitan parámetro JobManager)
	jm := srv.GetJobManager()
	srv.HandleFunc("POST", "/jobs/submit", handlers.JobSubmitHandler(jm))
	srv.HandleFunc("GET", "/jobs/status", handlers.JobStatusHandler(jm))
	srv.HandleFunc("GET", "/jobs/result", handlers.JobResultHandler(jm))
	srv.HandleFunc("DELETE", "/jobs/cancel", handlers.JobCancelHandler(jm))

	// Test passed if no panics occurred during route registration
	t.Log("All routes registered successfully")
}
