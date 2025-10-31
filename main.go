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
	srv.HandleFunc("GET", "/metrics", handlers.MetricsHandler(srv))
	srv.HandleFunc("GET", "/favicon.ico", handlers.FaviconHandler)
	srv.HandleFunc("GET", "/echo", handlers.EchoHandler)
	srv.HandleFunc("POST", "/echo", handlers.EchoHandler)
	srv.HandleFunc("GET", "/ping", handlers.PingHandler)
	srv.HandleFunc("GET", "/time", handlers.TimeHandler) // /time

	// Rutas basicas
	srv.HandleFunc("GET", "/fibonacci", handlers.FibonacciHandler) // /fibonacci?num=N
	srv.HandleFunc("POST", "/file", handlers.CreateFileHandler)    // /createFile?name=filename&content=text&repeat=x
	srv.HandleFunc("DELETE", "/file", handlers.DeleteFileHandler)  // /deleteFile?name=filename
	srv.HandleFunc("PUT", "/reverse", handlers.ReverseHandler)     // /reverse?text=yourtext
	srv.HandleFunc("PUT", "/toupper", handlers.ToUpperHandler)     // /toupper?text=yourtext
	srv.HandleFunc("GET", "/random", handlers.RandomNumberHandler) // /random?min=x&max=y
	srv.HandleFunc("PUT", "/hash", handlers.HashHandler)           // /hash?text=yourtext
	srv.HandleFunc("POST", "/simulate", handlers.SimulateHandler)  // /simulate?seconds=s&task=name
	srv.HandleFunc("POST", "/sleep", handlers.SleepHandler)        // /sleep?seconds=s
	srv.HandleFunc("POST", "/loadtest", handlers.LoadTestHandler)  // /loadtest?tasks=n&sleep=x
	srv.HandleFunc("GET", "/help", handlers.HelpHandler)           // /help

	// CPU-bound
	srv.HandleFunc("GET", "/isprime", handlers.IsPrimeHandler)       // /isprime?num=N
	srv.HandleFunc("GET", "/factor", handlers.FactorHandler)         // /factor?num=N
	srv.HandleFunc("GET", "/pi", handlers.PiHandler)                 // /pi?digits=N
	srv.HandleFunc("GET", "/mandelbrot", handlers.MandelbrotHandler) // /mandelbrot?width=W&height=H&max_iter=I
	srv.HandleFunc("GET", "/matrixmul", handlers.MatrixMulHandler)   // /matrixmul?size=N&seed=S

	// IO-bound (large file operations)
	srv.HandleFunc("GET", "/sortfile", handlers.SortFileHandler)   // /sortfile?name=FILE&algo=merge|quick
	srv.HandleFunc("GET", "/wordcount", handlers.WordCountHandler) // /wordcount?name=FILE
	srv.HandleFunc("GET", "/grep", handlers.GrepHandler)           // /grep?name=FILE&pattern=REGEX
	srv.HandleFunc("GET", "/compress", handlers.CompressHandler)   // /compress?name=FILE&codec=gzip|xz
	srv.HandleFunc("GET", "/hashfile", handlers.HashFileHandler)   // /hashfile?name=FILE&algo=sha256

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
