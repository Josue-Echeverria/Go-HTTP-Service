package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"GoDocker/handlers"
	"GoDocker/server"
)

func main() {
	// Configuración
	addr := ":8080"
	poolSize := 20 // Número de workers en el pool

	// Crear servidor
	srv := server.NewServer(addr, poolSize)

	// Registrar handlers
	srv.HandleFunc("GET", "/", handlers.HelloHandler)
	srv.HandleFunc("GET", "/status", handlers.StatusHandler(srv))
	srv.HandleFunc("GET", "/echo", handlers.EchoHandler)
	srv.HandleFunc("POST", "/echo", handlers.EchoHandler)
	srv.HandleFunc("GET", "/ping", handlers.PingHandler)
	srv.HandleFunc("GET", "/time", handlers.TimeHandler)

	// Iniciar servidor
	if err := srv.Start(); err != nil {
		log.Fatalf("Error iniciando servidor: %v", err)
	}

	// Esperar señal de shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	log.Println("\nSeñal de interrupción recibida, cerrando servidor...")

	// Shutdown graceful con timeout de 30 segundos
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Error durante shutdown: %v", err)
		os.Exit(1)
	}

	log.Println("Servidor cerrado exitosamente")
}
