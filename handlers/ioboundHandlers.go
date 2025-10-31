package handlers

import (
	"GoDocker/server"
	"bufio"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// mergeSort implementa el algoritmo de ordenamiento merge sort
func mergeSort(arr []int) []int {
	if len(arr) <= 1 {
		return arr
	}

	mid := len(arr) / 2
	left := mergeSort(arr[:mid])
	right := mergeSort(arr[mid:])

	return merge(left, right)
}

// merge combina dos slices ordenados en uno solo ordenado
func merge(left, right []int) []int {
	result := make([]int, 0, len(left)+len(right))
	i, j := 0, 0

	for i < len(left) && j < len(right) {
		if left[i] <= right[j] {
			result = append(result, left[i])
			i++
		} else {
			result = append(result, right[j])
			j++
		}
	}

	// Agregar elementos restantes
	result = append(result, left[i:]...)
	result = append(result, right[j:]...)

	return result
}

// quickSort implementa el algoritmo de ordenamiento quick sort
func quickSort(arr []int) {
	if len(arr) < 2 {
		return
	}
	quickSortHelper(arr, 0, len(arr)-1)
}

// quickSortHelper es la función recursiva que implementa quick sort
func quickSortHelper(arr []int, low, high int) {
	if low < high {
		pivotIndex := partition(arr, low, high)
		quickSortHelper(arr, low, pivotIndex-1)
		quickSortHelper(arr, pivotIndex+1, high)
	}
}

// partition reorganiza el slice usando el último elemento como pivote
func partition(arr []int, low, high int) int {
	pivot := arr[high]
	i := low - 1

	for j := low; j < high; j++ {
		if arr[j] < pivot {
			i++
			arr[i], arr[j] = arr[j], arr[i]
		}
	}

	arr[i+1], arr[high] = arr[high], arr[i+1]
	return i + 1
}

// SortFileHandler ordena números enteros en el archivo indicado.
// Query params: name=FILE, algo=merge|quick (quick = in-memory, merge -> also in-memory here but placeholder)
func SortFileHandler(req *server.HTTPRequest) *server.HTTPResponse {
	filename, _ := req.Params["name"]
	algo, _ := req.Params["algo"]
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
		nums = mergeSort(nums)
	} else if algo == "quick" {
		quickSort(nums)
	} else {
		// Por defecto usar quicksort
		quickSort(nums)
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
	filename, _ := req.Params["name"]
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
	filename, _ := req.Params["name"]
	pattern, _ := req.Params["pattern"]
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
	filename, _ := req.Params["name"]
	codec, _ := req.Params["codec"]
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
	filename, _ := req.Params["name"]
	algo, _ := req.Params["algo"]
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
