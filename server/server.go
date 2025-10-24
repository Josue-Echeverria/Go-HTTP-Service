package server

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

// HTTPRequest representa una solicitud HTTP parseada
type HTTPRequest struct {
	Method  string
	Path    string
	Version string
	Headers map[string]string
	Body    string
}

// HTTPResponse representa una respuesta HTTP
type HTTPResponse struct {
	StatusCode int
	StatusText string
	Headers    map[string]string
	Body       string
}

// ConnectionTask representa una tarea de conexión a procesar
type ConnectionTask struct {
	Conn net.Conn
	ID   int64
}

// Server representa el servidor HTTP
type Server struct {
	addr           string
	listener       net.Listener
	workerPool     *WorkerPool
	taskQueue      *TaskQueue
	connCounter    *Counter
	activeConns    *Counter
	router         *Router
	shutdownCh     chan struct{}
	wg             sync.WaitGroup
	maxHeaderBytes int
	readTimeout    time.Duration
	writeTimeout   time.Duration
}

// NewServer crea una nueva instancia del servidor
func NewServer(addr string, poolSize int) *Server {
	return &Server{
		addr:           addr,
		workerPool:     NewWorkerPool(poolSize),
		taskQueue:      NewTaskQueue(1000),
		connCounter:    NewCounter(),
		activeConns:    NewCounter(),
		router:         NewRouter(),
		shutdownCh:     make(chan struct{}),
		maxHeaderBytes: 1 << 20, // 1MB
		readTimeout:    30 * time.Second,
		writeTimeout:   30 * time.Second,
	}
}

// Start inicia el servidor
func (s *Server) Start() error {
	var err error
	s.listener, err = net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("error al iniciar listener: %w", err)
	}

	log.Printf("Servidor iniciado en %s", s.addr)

	// Iniciar worker pool
	s.workerPool.Start(s.taskQueue, s.processConnection)

	// Aceptar conexiones
	s.wg.Add(1)
	go s.acceptConnections()

	return nil
}

// acceptConnections acepta nuevas conexiones entrantes
func (s *Server) acceptConnections() {
	defer s.wg.Done()

	for {
		select {
		case <-s.shutdownCh:
			return
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				select {
				case <-s.shutdownCh:
					return
				default:
					log.Printf("Error aceptando conexión: %v", err)
					continue
				}
			}

			connID := s.connCounter.Increment()
			s.activeConns.Increment()

			task := ConnectionTask{
				Conn: conn,
				ID:   connID,
			}

			// Encolar tarea
			if !s.taskQueue.Enqueue(task) {
				log.Printf("Cola llena, rechazando conexión %d", connID)
				conn.Close()
				s.activeConns.Decrement()
			}
		}
	}
}

// processConnection procesa una conexión individual
func (s *Server) processConnection(task interface{}) {
	connTask := task.(ConnectionTask)
	conn := connTask.Conn
	defer func() {
		conn.Close()
		s.activeConns.Decrement()
	}()

	// Configurar timeouts
	conn.SetReadDeadline(time.Now().Add(s.readTimeout))
	conn.SetWriteDeadline(time.Now().Add(s.writeTimeout))

	// Parsear request
	req, err := s.parseRequest(conn)
	if err != nil {
		log.Printf("Error parseando request [conn:%d]: %v", connTask.ID, err)
		s.sendErrorResponse(conn, 400, "Bad Request")
		return
	}

	log.Printf("[%d] %s %s", connTask.ID, req.Method, req.Path)

	// Procesar request con el router
	response := s.router.Handle(req)

	// Enviar respuesta
	if err := s.sendResponse(conn, response); err != nil {
		log.Printf("Error enviando respuesta [conn:%d]: %v", connTask.ID, err)
	}
}

// parseRequest parsea una solicitud HTTP
func (s *Server) parseRequest(conn net.Conn) (*HTTPRequest, error) {
	reader := bufio.NewReader(conn)

	// Leer línea de request
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("error leyendo request line: %w", err)
	}

	parts := strings.Fields(requestLine)
	if len(parts) != 3 {
		return nil, fmt.Errorf("request line inválida")
	}

	req := &HTTPRequest{
		Method:  parts[0],
		Path:    parts[1],
		Version: parts[2],
		Headers: make(map[string]string),
	}

	// Leer headers
	bytesRead := len(requestLine)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("error leyendo headers: %w", err)
		}

		bytesRead += len(line)
		if bytesRead > s.maxHeaderBytes {
			return nil, fmt.Errorf("headers demasiado grandes")
		}

		line = strings.TrimSpace(line)
		if line == "" {
			break
		}

		headerParts := strings.SplitN(line, ":", 2)
		if len(headerParts) == 2 {
			key := strings.TrimSpace(headerParts[0])
			value := strings.TrimSpace(headerParts[1])
			req.Headers[key] = value
		}
	}

	// Leer body si existe Content-Length
	if contentLength, ok := req.Headers["Content-Length"]; ok {
		var length int
		fmt.Sscanf(contentLength, "%d", &length)
		if length > 0 && length < s.maxHeaderBytes {
			bodyBytes := make([]byte, length)
			_, err := reader.Read(bodyBytes)
			if err != nil {
				return nil, fmt.Errorf("error leyendo body: %w", err)
			}
			req.Body = string(bodyBytes)
		}
	}

	return req, nil
}

// sendResponse envía una respuesta HTTP
func (s *Server) sendResponse(conn net.Conn, resp *HTTPResponse) error {
	var response strings.Builder

	// Status line
	response.WriteString(fmt.Sprintf("HTTP/1.1 %d %s\r\n", resp.StatusCode, resp.StatusText))

	// Headers
	if resp.Headers == nil {
		resp.Headers = make(map[string]string)
	}
	resp.Headers["Content-Length"] = fmt.Sprintf("%d", len(resp.Body))
	resp.Headers["Connection"] = "close"
	resp.Headers["Server"] = "CustomHTTPServer/1.0"

	for key, value := range resp.Headers {
		response.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}

	response.WriteString("\r\n")
	response.WriteString(resp.Body)

	_, err := conn.Write([]byte(response.String()))
	return err
}

// sendErrorResponse envía una respuesta de error
func (s *Server) sendErrorResponse(conn net.Conn, code int, message string) {
	resp := &HTTPResponse{
		StatusCode: code,
		StatusText: message,
		Body:       fmt.Sprintf("<html><body><h1>%d %s</h1></body></html>", code, message),
		Headers: map[string]string{
			"Content-Type": "text/html",
		},
	}
	s.sendResponse(conn, resp)
}

// HandleFunc registra un handler para un path específico
func (s *Server) HandleFunc(method, path string, handler HandlerFunc) {
	s.router.Register(method, path, handler)
}

// Shutdown detiene el servidor gracefully
func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Iniciando shutdown...")
	close(s.shutdownCh)

	// Cerrar listener
	if s.listener != nil {
		s.listener.Close()
	}

	// Detener worker pool
	s.workerPool.Stop()

	// Esperar a que terminen las goroutines
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("Shutdown completado")
		return nil
	case <-ctx.Done():
		return fmt.Errorf("timeout en shutdown")
	}
}

// GetStats retorna estadísticas del servidor
func (s *Server) GetStats() map[string]int64 {
	return map[string]int64{
		"total_connections":  s.connCounter.Get(),
		"active_connections": s.activeConns.Get(),
		"queue_size":         s.taskQueue.Size(),
	}
}
