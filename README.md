# Go-HTTP-Service

Servidor HTTP construido desde cero en Go, utilizando únicamente `net` y goroutines, sin librerías HTTP de alto nivel.

## Características

- Servidor HTTP completo sin usar `net/http`
- Pool de workers con goroutines para manejo concurrente
- Colas thread-safe con primitivas de sincronización (mutex, canales)
- Sistema de enrutamiento modular
- Contadores atómicos para estadísticas
- Graceful shutdown
- Cobertura de tests ≥ 90%

## Arquitectura

### Componentes Principales

```
Go-HTTP-Service/
├── server/
│   ├── server.go       # Servidor HTTP principal
│   ├── workerpool.go   # Pool de workers
│   ├── queue.go        # Cola thread-safe
│   ├── counter.go      # Contador atómico
│   ├── router.go       # Sistema de rutas
│   └── *_test.go       # Tests unitarios
├── handlers/
│   ├── handlers.go     # Handlers HTTP
│   └── handlers_test.go
└── main.go             # Punto de entrada
```

### Diseño de Concurrencia

**Worker Pool:**
- Pool de N workers (goroutines) que procesan conexiones
- Cada worker escucha de una cola compartida
- Sin `sleep()` para sincronización, usa canales y select

**Cola Thread-Safe:**
- Implementada con mutex para proteger acceso concurrente
- Canal `notEmpty` para señalizar cuando hay elementos
- Capacidad configurable para control de backpressure

**Contadores Atómicos:**
- Usan `sync/atomic` para operaciones thread-safe
- Rastrean conexiones totales y activas

**Sincronización:**
- `sync.WaitGroup` para coordinar shutdown
- Canales para señales de stop
- `sync.RWMutex` en router para lecturas concurrentes

## Compilación

```bash
# Compilar el proyecto
go build -o server.exe main.go

# O compilar con optimizaciones
go build -ldflags="-s -w" -o server.exe main.go
```

## Ejecución

```bash
# Ejecutar directamente
go run main.go

# O ejecutar el binario compilado
./server.exe
```

El servidor escuchará en `http://localhost:8080`

### Endpoints Disponibles

- `GET /` - Página de bienvenida
- `GET /status` - Estadísticas del servidor (JSON)
- `GET /echo` - Echo de la petición
- `POST /echo` - Echo con body
- `GET /ping` - Health check simple
- `GET /time` - Hora actual en múltiples formatos

## Testing

### Ejecutar todos los tests

```bash
go test ./...
```

### Cobertura de tests

```bash
# Ejecutar tests con cobertura
go test -cover ./...

# Generar reporte detallado de cobertura
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Ver cobertura por función
go tool cover -func=coverage.out
```

### Benchmarks

```bash
go test -bench=. ./...
```

### Tests específicos

```bash
# Server tests
go test -v ./server

# Handler tests
go test -v ./handlers

# Test con verbosidad
go test -v -cover ./...
```

## Configuración

En `main.go` puedes modificar:

```go
addr := ":8080"        // Puerto del servidor
poolSize := 20         // Número de workers
```

En `server/server.go`:

```go
maxHeaderBytes: 1 << 20        // Tamaño máximo de headers (1MB)
readTimeout:    30 * time.Second   // Timeout de lectura
writeTimeout:   30 * time.Second   // Timeout de escritura
```

## Características Técnicas

### Manejo de Conexiones

1. **Accept Loop**: Goroutine dedicada acepta conexiones
2. **Enqueue**: Conexión se encola en cola thread-safe
3. **Worker**: Worker disponible toma tarea de la cola
4. **Process**: Parsea HTTP, ejecuta handler, envía respuesta
5. **Close**: Cierra conexión y decrementa contador

### Primitivas de Sincronización

- **Mutex**: Protege cola y router
- **Atomic**: Contadores de conexiones
- **Canales**: 
  - `shutdownCh`: Señal de shutdown
  - `notEmpty`: Señal de cola no vacía
  - `stopCh`: Detener workers
- **WaitGroup**: Esperar terminación de goroutines

### Parsing HTTP

Parser manual que:
- Lee línea de request (método, path, versión)
- Parsea headers línea por línea
- Lee body según Content-Length
- Maneja errores y límites de tamaño

## Restricciones Cumplidas

✅ **Sin `net/http`**: Solo usa `net` y `bufio`  
✅ **Código desde cero**: Parser HTTP manual  
✅ **Pools y colas thread-safe**: `WorkerPool` y `TaskQueue`  
✅ **Sin `sleep()`**: Usa canales y select  
✅ **Recursos seguros**: Mutex y atomic  
✅ **Modular**: Separación clara de responsabilidades  
✅ **Cobertura ≥90%**: Tests exhaustivos

## Ejemplos de Uso

### Ejemplo con curl

```bash
# GET simple
curl http://localhost:8080/

# Status del servidor
curl http://localhost:8080/status

# Echo con headers
curl -H "X-Custom: value" http://localhost:8080/echo

# POST con body
curl -X POST -d '{"test":"data"}' http://localhost:8080/echo

# Ping
curl http://localhost:8080/ping

# Hora actual
curl http://localhost:8080/time
```

### Agregar un nuevo handler

```go
// En handlers/handlers.go
func MyHandler(req *server.HTTPRequest) *server.HTTPResponse {
    return &server.HTTPResponse{
        StatusCode: 200,
        StatusText: "OK",
        Body:       "Mi respuesta",
        Headers: map[string]string{
            "Content-Type": "text/plain",
        },
    }
}

// En main.go
srv.HandleFunc("GET", "/mypath", handlers.MyHandler)
```

## Métricas y Estadísticas

El endpoint `/status` retorna:

```json
{
  "status": "ok",
  "time": "2025-10-07T...",
  "stats": {
    "total_connections": 1234,
    "active_connections": 5,
    "queue_size": 2
  }
}
```

## Docker

```bash
# Construir imagen
docker build -t go-http-service .

# Ejecutar contenedor
docker run -p 8080:8080 go-http-service
```

## Conventional Commits

Este proyecto utiliza **Conventional Commits** para versionado automático. Los commits deben seguir este formato:

```
<type>[optional scope]: <description>
```

### Tipos principales:
- `feat:` - Nueva funcionalidad (MINOR 1.0.0 → 1.1.0)
- `fix:` - Corrección de bugs (PATCH 1.0.0 → 1.0.1)
- `feat!:` o `BREAKING CHANGE:` - Cambios incompatibles (MAJOR 1.0.0 → 2.0.0)
- `docs:`, `chore:`, `refactor:`, `test:` - No incrementan versión

### Ejemplos:
```bash
git commit -m "feat: agregar endpoint /health"
git commit -m "fix: corregir validación de request"
git commit -m "docs: actualizar README"
```

## Troubleshooting

**Puerto ocupado:**
```bash
# Cambiar puerto en main.go
addr := ":9090"
```

**Timeout en shutdown:**
```bash
# Aumentar timeout en main.go
ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
```

**Queue llena:**
```bash
# Aumentar capacidad de cola en server.go
taskQueue: NewTaskQueue(2000),
```

## Licencia

MIT