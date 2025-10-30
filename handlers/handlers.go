package handlers

import (
	"bufio"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
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
			"time":   time.Now().Format(time.RFC3339),
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

// SortFileHandler ordena números enteros en el archivo indicado.
// Query params: name=FILE, algo=merge|quick (quick = in-memory, merge -> also in-memory here but placeholder)
func SortFileHandler(req *server.HTTPRequest) *server.HTTPResponse {
	filename := req.Query.Get("name")
	algo := req.Query.Get("algo")
	if filename == "" {
		return &server.HTTPResponse{StatusCode: 400, StatusText: "Bad Request", Body: `{"error":"missing name parameter"}`, Headers: map[string]string{"Content-Type": "application/json"}}
	}

	name := getFilePath(filename)
	start := time.Now()
	f, err := os.Open(name)
	if err != nil {
		return &server.HTTPResponse{StatusCode: 404, StatusText: "Not Found", Body: fmt.Sprintf(`{"error":"%s"}`, err.Error()), Headers: map[string]string{"Content-Type": "application/json"}}
	}
	defer f.Close()

	// Leer todos los números (línea por línea) - eficiente para archivos moderados (>=50MB should be OK on modern machines)
	scanner := bufio.NewScanner(f)
	// aumentar buffer a 4MB por línea si hace falta
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 4*1024*1024)

	nums := make([]int, 0, 100000)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		v, err := strconv.Atoi(line)
		if err != nil {
			// ignorar líneas no numéricas
			continue
		}
		nums = append(nums, v)
	}
	if err := scanner.Err(); err != nil {
		return &server.HTTPResponse{StatusCode: 500, StatusText: "Internal Error", Body: fmt.Sprintf(`{"error":"%s"}`, err.Error()), Headers: map[string]string{"Content-Type": "application/json"}}
	}

	// Elegir algoritmo
	if algo == "merge" {
		// Implementación simple: usar sort.Ints (merge external not implemented here)
		sort.Ints(nums)
	} else {
		sort.Ints(nums)
	}

	outName := getFilePath(filename + ".sorted")
	of, err := os.Create(outName)
	if err != nil {
		return &server.HTTPResponse{StatusCode: 500, StatusText: "Internal Error", Body: fmt.Sprintf(`{"error":"%s"}`, err.Error()), Headers: map[string]string{"Content-Type": "application/json"}}
	}
	defer of.Close()

	w := bufio.NewWriter(of)
	for _, v := range nums {
		fmt.Fprintln(w, v)
	}
	w.Flush()

	elapsed := time.Since(start)
	stats := map[string]interface{}{"output": outName, "count": len(nums), "duration_ms": elapsed.Milliseconds()}
	data, _ := json.MarshalIndent(stats, "", "  ")
	return &server.HTTPResponse{StatusCode: 200, StatusText: "OK", Body: string(data), Headers: map[string]string{"Content-Type": "application/json"}}
}

// WordCountHandler cuenta líneas, palabras y bytes (wc-like)
func WordCountHandler(req *server.HTTPRequest) *server.HTTPResponse {
	filename := req.Query.Get("name")
	if filename == "" {
		return &server.HTTPResponse{StatusCode: 400, StatusText: "Bad Request", Body: `{"error":"missing name parameter"}`, Headers: map[string]string{"Content-Type": "application/json"}}
	}
	name := getFilePath(filename)
	f, err := os.Open(name)
	if err != nil {
		return &server.HTTPResponse{StatusCode: 404, StatusText: "Not Found", Body: fmt.Sprintf(`{"error":"%s"}`, err.Error()), Headers: map[string]string{"Content-Type": "application/json"}}
	}
	defer f.Close()

	var lines, words, bytesCount int64
	r := bufio.NewReader(f)
	buf := make([]byte, 32*1024)
	inWord := false
	for {
		n, err := r.Read(buf)
		if n > 0 {
			bytesCount += int64(n)
			for i := 0; i < n; i++ {
				b := buf[i]
				if b == '\n' {
					lines++
				}
				if b == ' ' || b == '\n' || b == '\t' || b == '\r' {
					if inWord {
						words++
						inWord = false
					}
				} else {
					inWord = true
				}
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return &server.HTTPResponse{StatusCode: 500, StatusText: "Internal Error", Body: fmt.Sprintf(`{"error":"%s"}`, err.Error()), Headers: map[string]string{"Content-Type": "application/json"}}
		}
	}
	if inWord {
		words++
	}

	result := map[string]interface{}{"lines": lines, "words": words, "bytes": bytesCount}
	data, _ := json.MarshalIndent(result, "", "  ")
	return &server.HTTPResponse{StatusCode: 200, StatusText: "OK", Body: string(data), Headers: map[string]string{"Content-Type": "application/json"}}
}

// GrepHandler busca un patrón regex en el archivo y devuelve número de coincidencias y primeras 10 líneas coincidentes
func GrepHandler(req *server.HTTPRequest) *server.HTTPResponse {
	filename := req.Query.Get("name")
	pattern := req.Query.Get("pattern")
	if filename == "" || pattern == "" {
		return &server.HTTPResponse{StatusCode: 400, StatusText: "Bad Request", Body: `{"error":"missing name or pattern parameter"}`, Headers: map[string]string{"Content-Type": "application/json"}}
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return &server.HTTPResponse{StatusCode: 400, StatusText: "Bad Request", Body: fmt.Sprintf(`{"error":"invalid regex: %s"}`, err.Error()), Headers: map[string]string{"Content-Type": "application/json"}}
	}

	name := getFilePath(filename)
	f, err := os.Open(name)
	if err != nil {
		return &server.HTTPResponse{StatusCode: 404, StatusText: "Not Found", Body: fmt.Sprintf(`{"error":"%s"}`, err.Error()), Headers: map[string]string{"Content-Type": "application/json"}}
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 4*1024*1024)

	matches := 0
	firstLines := make([]string, 0, 10)
	for scanner.Scan() {
		line := scanner.Text()
		if re.MatchString(line) {
			matches++
			if len(firstLines) < 10 {
				firstLines = append(firstLines, line)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return &server.HTTPResponse{StatusCode: 500, StatusText: "Internal Error", Body: fmt.Sprintf(`{"error":"%s"}`, err.Error()), Headers: map[string]string{"Content-Type": "application/json"}}
	}

	res := map[string]interface{}{"matches": matches, "first_lines": firstLines}
	data, _ := json.MarshalIndent(res, "", "  ")
	return &server.HTTPResponse{StatusCode: 200, StatusText: "OK", Body: string(data), Headers: map[string]string{"Content-Type": "application/json"}}
}

// CompressHandler comprime un archivo usando gzip o xz (si xz está disponible)
func CompressHandler(req *server.HTTPRequest) *server.HTTPResponse {
	filename := req.Query.Get("name")
	codec := req.Query.Get("codec")
	if filename == "" || codec == "" {
		return &server.HTTPResponse{StatusCode: 400, StatusText: "Bad Request", Body: `{"error":"missing name or codec parameter"}`, Headers: map[string]string{"Content-Type": "application/json"}}
	}
	name := getFilePath(filename)
	if codec == "gzip" {
		in, err := os.Open(name)
		if err != nil {
			return &server.HTTPResponse{StatusCode: 404, StatusText: "Not Found", Body: fmt.Sprintf(`{"error":"%s"}`, err.Error()), Headers: map[string]string{"Content-Type": "application/json"}}
		}
		defer in.Close()
		outName := getFilePath(filename + ".gz")
		out, err := os.Create(outName)
		if err != nil {
			return &server.HTTPResponse{StatusCode: 500, StatusText: "Internal Error", Body: fmt.Sprintf(`{"error":"%s"}`, err.Error()), Headers: map[string]string{"Content-Type": "application/json"}}
		}
		gw := gzip.NewWriter(out)
		if _, err := io.Copy(gw, in); err != nil {
			gw.Close()
			out.Close()
			return &server.HTTPResponse{StatusCode: 500, StatusText: "Internal Error", Body: fmt.Sprintf(`{"error":"%s"}`, err.Error()), Headers: map[string]string{"Content-Type": "application/json"}}
		}
		gw.Close()
		out.Close()
		fi, _ := os.Stat(outName)
		res := map[string]interface{}{"output": outName, "size": fi.Size()}
		data, _ := json.MarshalIndent(res, "", "  ")
		return &server.HTTPResponse{StatusCode: 200, StatusText: "OK", Body: string(data), Headers: map[string]string{"Content-Type": "application/json"}}
	} else if codec == "xz" {
		// Usa xz externo si está disponible: xz -c <file>
		outName := getFilePath(filename + ".xz")
		cmd := exec.Command("xz", "-c", name)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return &server.HTTPResponse{StatusCode: 500, StatusText: "Internal Error", Body: fmt.Sprintf(`{"error":"%s"}`, err.Error()), Headers: map[string]string{"Content-Type": "application/json"}}
		}
		if err := cmd.Start(); err != nil {
			return &server.HTTPResponse{StatusCode: 500, StatusText: "Internal Error", Body: fmt.Sprintf(`{"error":"%s"}`, err.Error()), Headers: map[string]string{"Content-Type": "application/json"}}
		}
		out, err := os.Create(outName)
		if err != nil {
			stdout.Close()
			return &server.HTTPResponse{StatusCode: 500, StatusText: "Internal Error", Body: fmt.Sprintf(`{"error":"%s"}`, err.Error()), Headers: map[string]string{"Content-Type": "application/json"}}
		}
		if _, err := io.Copy(out, stdout); err != nil {
			out.Close()
			return &server.HTTPResponse{StatusCode: 500, StatusText: "Internal Error", Body: fmt.Sprintf(`{"error":"%s"}`, err.Error()), Headers: map[string]string{"Content-Type": "application/json"}}
		}
		out.Close()
		if err := cmd.Wait(); err != nil {
			return &server.HTTPResponse{StatusCode: 500, StatusText: "Internal Error", Body: fmt.Sprintf(`{"error":"%s"}`, err.Error()), Headers: map[string]string{"Content-Type": "application/json"}}
		}
		fi, _ := os.Stat(outName)
		res := map[string]interface{}{"output": outName, "size": fi.Size()}
		data, _ := json.MarshalIndent(res, "", "  ")
		return &server.HTTPResponse{StatusCode: 200, StatusText: "OK", Body: string(data), Headers: map[string]string{"Content-Type": "application/json"}}
	}
	return &server.HTTPResponse{StatusCode: 400, StatusText: "Bad Request", Body: `{"error":"unsupported codec"}`, Headers: map[string]string{"Content-Type": "application/json"}}
}

// HashFileHandler calcula hash (sha256) de un archivo
func HashFileHandler(req *server.HTTPRequest) *server.HTTPResponse {
	filename := req.Query.Get("name")
	algo := req.Query.Get("algo")
	if filename == "" || algo == "" {
		return &server.HTTPResponse{StatusCode: 400, StatusText: "Bad Request", Body: `{"error":"missing name or algo parameter"}`, Headers: map[string]string{"Content-Type": "application/json"}}
	}
	if algo != "sha256" {
		return &server.HTTPResponse{StatusCode: 400, StatusText: "Bad Request", Body: `{"error":"unsupported algo"}`, Headers: map[string]string{"Content-Type": "application/json"}}
	}
	name := getFilePath(filename)
	f, err := os.Open(name)
	if err != nil {
		return &server.HTTPResponse{StatusCode: 404, StatusText: "Not Found", Body: fmt.Sprintf(`{"error":"%s"}`, err.Error()), Headers: map[string]string{"Content-Type": "application/json"}}
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return &server.HTTPResponse{StatusCode: 500, StatusText: "Internal Error", Body: fmt.Sprintf(`{"error":"%s"}`, err.Error()), Headers: map[string]string{"Content-Type": "application/json"}}
	}
	sum := hex.EncodeToString(h.Sum(nil))
	res := map[string]interface{}{"algo": "sha256", "hex": sum}
	data, _ := json.MarshalIndent(res, "", "  ")
	return &server.HTTPResponse{StatusCode: 200, StatusText: "OK", Body: string(data), Headers: map[string]string{"Content-Type": "application/json"}}
}
