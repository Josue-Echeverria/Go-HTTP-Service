package server

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
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
	Params  map[string]string
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
	Conn        net.Conn
	ID          int64
	EnqueueTime time.Time
}

// Server representa el servidor HTTP
type Server struct {
	addr           string
	listener       net.Listener
	workerPool     *WorkerPool
	taskQueue      *TaskQueue
	connCounter    *Counter
	activeConns    *Counter
	busyWorkers    *Counter
	router         *Router
	metricsManager *MetricsManager
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
		busyWorkers:    NewCounter(),
		router:         NewRouter(),
		metricsManager: NewMetricsManager(),
		shutdownCh:     make(chan struct{}),
		maxHeaderBytes: 1 << 20,
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
				Conn:        conn,
				ID:          connID,
				EnqueueTime: time.Now(),
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

// parseRequest parsea una solicitud HTTP
func (s *Server) parseRequest(conn net.Conn) (*HTTPRequest, error) {
	// Configurar timeout para lectura
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	reader := bufio.NewReader(conn)

	// Leer request line con manejo de EOF
	requestLineBytes, _, err := reader.ReadLine()
	if err != nil {
		if err == io.EOF {
			// Conexión cerrada antes de enviar datos - no es un error grave
			return nil, fmt.Errorf("connection closed by client")
		}
		return nil, fmt.Errorf("error reading request line: %v", err)
	}

	// Verificar que la línea no esté vacía
	line := strings.TrimSpace(string(requestLineBytes))
	if line == "" {
		return nil, fmt.Errorf("empty request line")
	}

	// Parsear method, path, version
	parts := strings.Fields(line)
	if len(parts) != 3 {
		return nil, fmt.Errorf("malformed request line: %s", line)
	}

	paramsIndex := strings.Index(parts[1], "?")

	// Safely determine the path portion (avoid slicing with -1)
	path := parts[1]
	if paramsIndex != -1 {
		path = parts[1][:paramsIndex]
	}

	req := &HTTPRequest{
		Method:  parts[0],
		Path:    path,
		Version: parts[2],
		Headers: make(map[string]string),
		Params:  make(map[string]string),
	}

	if paramsIndex != -1 {
		// Parsear query string
		queryString := parts[1][paramsIndex+1:]
		for _, param := range strings.Split(queryString, "&") {
			kv := strings.SplitN(param, "=", 2)
			if len(kv) == 2 {
				key := strings.TrimSpace(kv[0])
				value := strings.TrimSpace(kv[1])
				req.Params[key] = value
			}
		}
	}

	// Leer headers
	for {
		headerLineBytes, _, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break // Fin de headers
			}
			return nil, fmt.Errorf("error reading headers: %v", err)
		}

		line := strings.TrimSpace(string(headerLineBytes))
		if line == "" {
			break // Fin de headers
		}

		// Parsear header
		headerParts := strings.SplitN(line, ":", 2)
		if len(headerParts) == 2 {
			key := strings.TrimSpace(headerParts[0])
			value := strings.TrimSpace(headerParts[1])
			req.Headers[key] = value
		}
	}

	// Leer body si existe Content-Length
	body := ""
	if contentLengthStr, ok := req.Headers["Content-Length"]; ok {
		contentLength, err := strconv.Atoi(contentLengthStr)
		if err == nil && contentLength > 0 {
			bodyBytes := make([]byte, contentLength)
			_, err := io.ReadFull(reader, bodyBytes)
			if err != nil && err != io.EOF {
				return nil, fmt.Errorf("error reading body: %v", err)
			}
			body = string(bodyBytes)
		}
	}

	req.Body = body
	return req, nil
}

// processConnection procesa una conexión individual con mejor error handling
func (s *Server) processConnection(task interface{}) {
	connTask := task.(ConnectionTask)
	conn := connTask.Conn
	connID := connTask.ID

	// Incrementar workers ocupados
	s.busyWorkers.Increment()
	defer s.busyWorkers.Decrement()

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic in connection %d: %v", connID, r)
		}
		conn.Close()
		s.activeConns.Decrement()
	}()

	// Calcular tiempo de espera en cola
	dequeueTime := time.Now()
	waitTime := dequeueTime.Sub(connTask.EnqueueTime)

	// Configurar timeouts
	conn.SetReadDeadline(time.Now().Add(s.readTimeout))
	conn.SetWriteDeadline(time.Now().Add(s.writeTimeout))

	// Parsear request con manejo de errores mejorado
	req, err := s.parseRequest(conn)
	if err != nil {
		// Distinguir entre errores normales (EOF) y errores reales
		if strings.Contains(err.Error(), "connection closed by client") {
			// Log silencioso para conexiones cerradas por cliente
			log.Printf("Connection %d: client disconnected", connID)
		} else if strings.Contains(err.Error(), "EOF") {
			// Log silencioso para EOF
			log.Printf("Connection %d: premature EOF", connID)
		} else {
			// Log para errores reales
			log.Printf("Error parsing request [conn:%d]: %v", connID, err)
		}

		// Intentar enviar respuesta de error si la conexión sigue activa
		s.sendErrorResponse(conn, 400, "Bad Request")
		return
	}

	// Log para ver qué se está solicitando
	log.Printf("Connection %d: %s %s", connID, req.Method, req.Path)

	// Obtener métricas para este endpoint
	endpoint := fmt.Sprintf("%s %s", req.Method, req.Path)
	metrics := s.metricsManager.GetOrCreate(endpoint)

	// Registrar tiempo de espera en cola
	metrics.RecordWaitTime(waitTime)
	metrics.IncrementActive()

	// Medir tiempo de ejecución del handler
	execStart := time.Now()
	response := s.router.Handle(req)
	execDuration := time.Since(execStart)

	// Registrar métricas
	metrics.RecordExecTime(execDuration)
	metrics.DecrementActive()

	// Enviar respuesta
	err = s.sendResponse(conn, response)
	if err != nil {
		log.Printf("Error sending response [conn:%d]: %v", connID, err)
	}
}

// sendResponse con mejor error handling
func (s *Server) sendResponse(conn net.Conn, resp *HTTPResponse) error {
	// Construir response HTTP
	var response strings.Builder

	response.WriteString(fmt.Sprintf("HTTP/1.1 %d %s\r\n", resp.StatusCode, resp.StatusText))

	// Agregar headers
	for key, value := range resp.Headers {
		response.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}

	// Content-Length
	response.WriteString(fmt.Sprintf("Content-Length: %d\r\n", len(resp.Body)))

	// Línea vacía para separar headers del body
	response.WriteString("\r\n")

	// Body
	response.WriteString(resp.Body)

	// Enviar respuesta con timeout
	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err := conn.Write([]byte(response.String()))

	return err
}

// sendErrorResponse envía una respuesta de error simple
func (s *Server) sendErrorResponse(conn net.Conn, statusCode int, statusText string) {
	errorBody := fmt.Sprintf(`{"error":"%s"}`, statusText)

	response := &HTTPResponse{
		StatusCode: statusCode,
		StatusText: statusText,
		Body:       errorBody,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	s.sendResponse(conn, response)
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

// GetMetrics retorna métricas detalladas del servidor
func (s *Server) GetMetrics() map[string]interface{} {
	stats := s.metricsManager.GetAllStats()

	// Agregar métricas del worker pool
	stats["worker_pool"] = map[string]interface{}{
		"size":         s.workerPool.Size(),
		"busy_workers": s.busyWorkers.Get(),
		"idle_workers": int64(s.workerPool.Size()) - s.busyWorkers.Get(),
	}

	// Agregar métricas globales
	stats["global"] = map[string]interface{}{
		"total_connections":  s.connCounter.Get(),
		"active_connections": s.activeConns.Get(),
		"queue_size":         s.taskQueue.Size(),
		"queue_capacity":     s.taskQueue.Capacity(),
	}

	return stats
}
