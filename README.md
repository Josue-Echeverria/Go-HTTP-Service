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

## Endpoints IO-bound (procesamiento intensivo)

Se agregan endpoints para probar carga I/O y procesamiento en archivos grandes. Todos aceptan parámetros por query string.

- `GET /sortfile?name=FILE&algo=merge|quick`
  - Ordena un archivo que contiene números enteros (uno por línea). `algo=quick` usa ordenamiento en memoria (rápido para archivos moderados). `algo=merge` está previsto para merge-external, actualmente utiliza también una ordenación eficiente en memoria.
  - Recomendado para archivos >= 50MB; el handler procesa la entrada en streaming y escribe un archivo de salida `FILE.sorted`.
  - Respuesta JSON con nombre del archivo de salida, cantidad de números procesados y duración en ms.

- `GET /wordcount?name=FILE`
  - Cuenta líneas, palabras y bytes (equivalente a `wc`). Diseñado para archivos grandes: lectura por bloques y conteo en streaming.
  - Respuesta JSON: `{ "lines": ..., "words": ..., "bytes": ... }`.

- `GET /grep?name=FILE&pattern=REGEX`
  - Busca un patrón regex en el archivo y devuelve el número total de coincidencias y las primeras 10 líneas que coinciden.
  - Útil para validar patrones en archivos grandes sin cargarlos completamente en memoria.

- `GET /compress?name=FILE&codec=gzip|xz`
  - Comprime el archivo usando `gzip` (usando librería estándar) o `xz` (si `xz` está instalado en el sistema). El resultado se guarda en `FILE.gz` o `FILE.xz`.
  - Respuesta JSON con nombre de salida y tamaño en bytes.
  - Nota: para `xz` el servidor invoca la utilidad externa `xz` (debe estar disponible en PATH).

- `GET /hashfile?name=FILE&algo=sha256`
  - Calcula el hash del archivo con sha256 (lectura en streaming). Retorna el hash en hexadecimal.

Buenas prácticas y limitaciones
- Asegúrate de que el usuario que ejecuta el servidor tenga permisos de lectura/escritura sobre los archivos indicados.
- Para archivos muy grandes (> a varios cientos de MB) la opción `quick` podría consumir bastante RAM; si trabajas con archivos enormes considera implementar una variante de "external merge" (dividir en chunks, ordenar cada uno, y luego hacer k-way merge).
- `xz` debe estar instalado si quieres usar `codec=xz`.

Ejemplos (curl)

```bash
# Ordenar un archivo
curl "http://localhost:8080/sortfile?name=/data/bignums.txt&algo=quick"

# Contar palabras y líneas
curl "http://localhost:8080/wordcount?name=/data/biglog.txt"

# Buscar regex
curl "http://localhost:8080/grep?name=/data/biglog.txt&pattern=Error"

# Comprimir con gzip
curl "http://localhost:8080/compress?name=/data/bigfile.txt&codec=gzip"

# Calcular sha256
curl "http://localhost:8080/hashfile?name=/data/bigfile.txt&algo=sha256"
```

Implementación y ayuda
- Las funciones realizan E/S en streaming y evitan cargar innecesariamente todo el archivo en memoria cuando es posible.
- Si quieres que implemente una versión de "external merge" (ordenamiento por batches y mezcla k-way) puedo hacerlo; dime si quieres que lo haga y cuál es el tamaño típico de los archivos en tu entorno.
