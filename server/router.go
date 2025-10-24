package server

import (
	"fmt"
	"sync"
)

// HandlerFunc es una función que maneja una petición HTTP
type HandlerFunc func(*HTTPRequest) *HTTPResponse

// Route representa una ruta registrada
type Route struct {
	Method  string
	Path    string
	Handler HandlerFunc
}

// Router maneja el enrutamiento de peticiones
type Router struct {
	routes map[string]map[string]HandlerFunc // method -> path -> handler
	mu     sync.RWMutex
}

// NewRouter crea un nuevo router
func NewRouter() *Router {
	return &Router{
		routes: make(map[string]map[string]HandlerFunc),
	}
}

// Register registra un nuevo handler para un método y path
func (r *Router) Register(method, path string, handler HandlerFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.routes[method] == nil {
		r.routes[method] = make(map[string]HandlerFunc)
	}

	r.routes[method][path] = handler
}

// Handle procesa una petición y retorna una respuesta
func (r *Router) Handle(req *HTTPRequest) *HTTPResponse {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Buscar handler para el método y path
	if methodRoutes, ok := r.routes[req.Method]; ok {
		if handler, ok := methodRoutes[req.Path]; ok {
			return handler(req)
		}
	}

	// No encontrado
	return &HTTPResponse{
		StatusCode: 404,
		StatusText: "Not Found",
		Body:       fmt.Sprintf("<html><body><h1>404 Not Found</h1><p>%s %s</p></body></html>", req.Method, req.Path),
		Headers: map[string]string{
			"Content-Type": "text/html",
		},
	}
}

// GetRoutes retorna todas las rutas registradas
func (r *Router) GetRoutes() []Route {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var routes []Route
	for method, paths := range r.routes {
		for path := range paths {
			routes = append(routes, Route{
				Method: method,
				Path:   path,
			})
		}
	}
	return routes
}
