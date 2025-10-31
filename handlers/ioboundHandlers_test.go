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
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// setupTestFiles crea archivos de prueba para los tests
func setupTestFiles(t *testing.T) {
	// Crear directorio de pruebas si no existe
	err := os.MkdirAll(filesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Crear archivo de números para pruebas de ordenamiento
	numbersFile := filepath.Join(filesDir, "numbers.txt")
	f, err := os.Create(numbersFile)
	if err != nil {
		t.Fatalf("Failed to create numbers test file: %v", err)
	}
	defer f.Close()

	numbers := []int{42, 17, 89, 3, 156, 7, 91, 23, 8, 65}
	for _, num := range numbers {
		fmt.Fprintln(f, num)
	}

	// Crear archivo de texto para pruebas de wordcount y grep
	textFile := filepath.Join(filesDir, "sample.txt")
	f2, err := os.Create(textFile)
	if err != nil {
		t.Fatalf("Failed to create text test file: %v", err)
	}
	defer f2.Close()

	sampleText := `Hello world
This is a test file
It contains multiple lines
With different words and patterns
Testing grep functionality
Pattern matching test
Another line here
Final line of the test file`

	fmt.Fprint(f2, sampleText)

	// Crear archivo pequeño para pruebas de hash
	hashFile := filepath.Join(filesDir, "hash_test.txt")
	f3, err := os.Create(hashFile)
	if err != nil {
		t.Fatalf("Failed to create hash test file: %v", err)
	}
	defer f3.Close()

	fmt.Fprint(f3, "Hello, World!")

	// Crear archivo con números grandes para pruebas de rendimiento
	largeFile := filepath.Join(filesDir, "large_numbers.txt")
	f4, err := os.Create(largeFile)
	if err != nil {
		t.Fatalf("Failed to create large numbers test file: %v", err)
	}
	defer f4.Close()

	writer := bufio.NewWriter(f4)
	for i := 1000; i >= 1; i-- {
		fmt.Fprintln(writer, i)
	}
	writer.Flush()
}

// cleanupTestFiles limpia archivos de prueba generados
func cleanupTestFiles(t *testing.T) {
	testFiles := []string{
		"numbers.txt", "numbers.txt.sorted",
		"sample.txt", "sample.txt.gz", "sample.txt.xz",
		"hash_test.txt", "large_numbers.txt", "large_numbers.txt.sorted",
	}

	for _, file := range testFiles {
		fullPath := filepath.Join(filesDir, file)
		if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: could not remove test file %s: %v", fullPath, err)
		}
	}
}

func TestSortFileHandler(t *testing.T) {
	setupTestFiles(t)
	defer cleanupTestFiles(t)

	tests := []struct {
		name           string
		filename       string
		algo           string
		expectedStatus int
		expectSorted   bool
	}{
		{
			name:           "Sort with quicksort",
			filename:       "numbers.txt",
			algo:           "quick",
			expectedStatus: 200,
			expectSorted:   true,
		},
		{
			name:           "Sort with mergesort",
			filename:       "numbers.txt",
			algo:           "merge",
			expectedStatus: 200,
			expectSorted:   true,
		},
		{
			name:           "Sort with default algorithm",
			filename:       "numbers.txt",
			algo:           "",
			expectedStatus: 200,
			expectSorted:   true,
		},
		{
			name:           "Missing filename",
			filename:       "",
			algo:           "quick",
			expectedStatus: 400,
			expectSorted:   false,
		},
		{
			name:           "Non-existent file",
			filename:       "nonexistent.txt",
			algo:           "quick",
			expectedStatus: 404,
			expectSorted:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := map[string]string{
				"name": tt.filename,
				"algo": tt.algo,
			}
			req := &server.HTTPRequest{
				Method:  "GET",
				Path:    "/sortfile",
				Version: "HTTP/1.1",
				Headers: make(map[string]string),
				Body:    "",
				Params:  params,
			}

			resp := SortFileHandler(req)

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectSorted && resp.StatusCode == 200 {
				var result map[string]interface{}
				if err := json.Unmarshal([]byte(resp.Body), &result); err != nil {
					t.Errorf("Failed to parse response JSON: %v", err)
					return
				}

				// Verificar que el archivo de salida existe
				outputFile, ok := result["output"].(string)
				if !ok {
					t.Error("Response should contain output filename")
					return
				}

				// Verificar que el archivo está ordenado
				f, err := os.Open(outputFile)
				if err != nil {
					t.Errorf("Failed to open sorted file: %v", err)
					return
				}
				defer f.Close()

				scanner := bufio.NewScanner(f)
				var prev int = -1
				for scanner.Scan() {
					line := strings.TrimSpace(scanner.Text())
					if line == "" {
						continue
					}
					current, err := strconv.Atoi(line)
					if err != nil {
						t.Errorf("Invalid number in sorted file: %s", line)
						continue
					}
					if prev != -1 && current < prev {
						t.Errorf("File is not sorted: %d comes after %d", current, prev)
					}
					prev = current
				}
			}
		})
	}
}

func TestWordCountHandler(t *testing.T) {
	setupTestFiles(t)
	defer cleanupTestFiles(t)

	tests := []struct {
		name           string
		filename       string
		expectedStatus int
		expectedLines  int64
		expectedWords  int64
	}{
		{
			name:           "Count words in sample file",
			filename:       "sample.txt",
			expectedStatus: 200,
			expectedLines:  8,
			expectedWords:  29,
		},
		{
			name:           "Missing filename",
			filename:       "",
			expectedStatus: 400,
		},
		{
			name:           "Non-existent file",
			filename:       "nonexistent.txt",
			expectedStatus: 404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := map[string]string{
				"name": tt.filename,
			}
			req := &server.HTTPRequest{
				Method:  "GET",
				Path:    "/wordcount",
				Version: "HTTP/1.1",
				Headers: make(map[string]string),
				Body:    "",
				Params:  params,
			}

			resp := WordCountHandler(req)

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if resp.StatusCode == 200 {
				var result map[string]interface{}
				if err := json.Unmarshal([]byte(resp.Body), &result); err != nil {
					t.Errorf("Failed to parse response JSON: %v", err)
					return
				}

				lines, ok := result["lines"].(float64)
				if !ok {
					t.Error("Response should contain lines count")
					return
				}

				words, ok := result["words"].(float64)
				if !ok {
					t.Error("Response should contain words count")
					return
				}

				if int64(lines) != tt.expectedLines {
					t.Errorf("Expected %d lines, got %d", tt.expectedLines, int64(lines))
				}

				if int64(words) != tt.expectedWords {
					t.Errorf("Expected %d words, got %d", tt.expectedWords, int64(words))
				}

				// Verificar que también está presente el conteo de bytes
				if _, ok := result["bytes"]; !ok {
					t.Error("Response should contain bytes count")
				}
			}
		})
	}
}

func TestGrepHandler(t *testing.T) {
	setupTestFiles(t)
	defer cleanupTestFiles(t)

	tests := []struct {
		name            string
		filename        string
		pattern         string
		expectedStatus  int
		expectedMatches int
	}{
		{
			name:            "Search for 'test' pattern",
			filename:        "sample.txt",
			pattern:         "test",
			expectedStatus:  200,
			expectedMatches: 3,
		},
		{
			name:            "Search for 'Pattern' pattern",
			filename:        "sample.txt",
			pattern:         "Pattern",
			expectedStatus:  200,
			expectedMatches: 1,
		},
		{
			name:            "Search with regex pattern",
			filename:        "sample.txt",
			pattern:         "^This.*",
			expectedStatus:  200,
			expectedMatches: 1,
		},
		{
			name:           "Missing filename",
			filename:       "",
			pattern:        "test",
			expectedStatus: 400,
		},
		{
			name:           "Missing pattern",
			filename:       "sample.txt",
			pattern:        "",
			expectedStatus: 400,
		},
		{
			name:           "Invalid regex",
			filename:       "sample.txt",
			pattern:        "[",
			expectedStatus: 400,
		},
		{
			name:           "Non-existent file",
			filename:       "nonexistent.txt",
			pattern:        "test",
			expectedStatus: 404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := map[string]string{
				"name":    tt.filename,
				"pattern": tt.pattern,
			}
			req := &server.HTTPRequest{
				Method:  "GET",
				Path:    "/grep",
				Version: "HTTP/1.1",
				Headers: make(map[string]string),
				Body:    "",
				Params:  params,
			}

			resp := GrepHandler(req)

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if resp.StatusCode == 200 {
				var result map[string]interface{}
				if err := json.Unmarshal([]byte(resp.Body), &result); err != nil {
					t.Errorf("Failed to parse response JSON: %v", err)
					return
				}

				matches, ok := result["matches"].(float64)
				if !ok {
					t.Error("Response should contain matches count")
					return
				}

				if int(matches) != tt.expectedMatches {
					t.Errorf("Expected %d matches, got %d", tt.expectedMatches, int(matches))
				}

				// Verificar que están presentes las primeras líneas
				if _, ok := result["first_lines"]; !ok {
					t.Error("Response should contain first_lines array")
				}
			}
		})
	}
}

func TestCompressHandler(t *testing.T) {
	setupTestFiles(t)
	defer cleanupTestFiles(t)

	tests := []struct {
		name           string
		filename       string
		codec          string
		expectedStatus int
		expectOutput   bool
	}{
		{
			name:           "Compress with gzip",
			filename:       "sample.txt",
			codec:          "gzip",
			expectedStatus: 200,
			expectOutput:   true,
		},
		{
			name:           "Unsupported codec",
			filename:       "sample.txt",
			codec:          "bzip2",
			expectedStatus: 400,
			expectOutput:   false,
		},
		{
			name:           "Missing filename",
			filename:       "",
			codec:          "gzip",
			expectedStatus: 400,
			expectOutput:   false,
		},
		{
			name:           "Missing codec",
			filename:       "sample.txt",
			codec:          "",
			expectedStatus: 400,
			expectOutput:   false,
		},
		{
			name:           "Non-existent file",
			filename:       "nonexistent.txt",
			codec:          "gzip",
			expectedStatus: 404,
			expectOutput:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := map[string]string{
				"name":  tt.filename,
				"codec": tt.codec,
			}
			req := &server.HTTPRequest{
				Method:  "GET",
				Path:    "/compress",
				Version: "HTTP/1.1",
				Headers: make(map[string]string),
				Body:    "",
				Params:  params,
			}

			resp := CompressHandler(req)

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectOutput && resp.StatusCode == 200 {
				var result map[string]interface{}
				if err := json.Unmarshal([]byte(resp.Body), &result); err != nil {
					t.Errorf("Failed to parse response JSON: %v", err)
					return
				}

				// Verificar que el archivo de salida existe
				outputFile, ok := result["output"].(string)
				if !ok {
					t.Error("Response should contain output filename")
					return
				}

				// Verificar que el archivo comprimido existe y tiene contenido
				if _, err := os.Stat(outputFile); os.IsNotExist(err) {
					t.Errorf("Compressed file should exist: %s", outputFile)
				}

				// Para gzip, verificar que el archivo se puede descomprimir
				if tt.codec == "gzip" {
					f, err := os.Open(outputFile)
					if err != nil {
						t.Errorf("Failed to open compressed file: %v", err)
						return
					}
					defer f.Close()

					gzReader, err := gzip.NewReader(f)
					if err != nil {
						t.Errorf("Failed to create gzip reader: %v", err)
						return
					}
					defer gzReader.Close()

					// Leer contenido descomprimido
					content, err := io.ReadAll(gzReader)
					if err != nil {
						t.Errorf("Failed to read decompressed content: %v", err)
						return
					}

					if len(content) == 0 {
						t.Error("Decompressed content should not be empty")
					}
				}
			}
		})
	}
}

func TestHashFileHandler(t *testing.T) {
	setupTestFiles(t)
	defer cleanupTestFiles(t)

	// Pre-calcular el hash esperado para el archivo de prueba
	expectedContent := "Hello, World!"
	h := sha256.New()
	h.Write([]byte(expectedContent))
	expectedHash := hex.EncodeToString(h.Sum(nil))

	tests := []struct {
		name           string
		filename       string
		algo           string
		expectedStatus int
		expectedHash   string
	}{
		{
			name:           "Hash with sha256",
			filename:       "hash_test.txt",
			algo:           "sha256",
			expectedStatus: 200,
			expectedHash:   expectedHash,
		},
		{
			name:           "Unsupported algorithm",
			filename:       "hash_test.txt",
			algo:           "md5",
			expectedStatus: 400,
		},
		{
			name:           "Missing filename",
			filename:       "",
			algo:           "sha256",
			expectedStatus: 400,
		},
		{
			name:           "Missing algorithm",
			filename:       "hash_test.txt",
			algo:           "",
			expectedStatus: 400,
		},
		{
			name:           "Non-existent file",
			filename:       "nonexistent.txt",
			algo:           "sha256",
			expectedStatus: 404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := map[string]string{
				"name": tt.filename,
				"algo": tt.algo,
			}
			req := &server.HTTPRequest{
				Method:  "GET",
				Path:    "/hashfile",
				Version: "HTTP/1.1",
				Headers: make(map[string]string),
				Body:    "",
				Params:  params,
			}

			resp := HashFileHandler(req)

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if resp.StatusCode == 200 {
				var result map[string]interface{}
				if err := json.Unmarshal([]byte(resp.Body), &result); err != nil {
					t.Errorf("Failed to parse response JSON: %v", err)
					return
				}

				algo, ok := result["algo"].(string)
				if !ok {
					t.Error("Response should contain algo field")
					return
				}

				if algo != tt.algo {
					t.Errorf("Expected algo %s, got %s", tt.algo, algo)
				}

				hashValue, ok := result["hex"].(string)
				if !ok {
					t.Error("Response should contain hex hash value")
					return
				}

				if tt.expectedHash != "" && hashValue != tt.expectedHash {
					t.Errorf("Expected hash %s, got %s", tt.expectedHash, hashValue)
				}
			}
		})
	}
}

// TestMergeSortFunction prueba la función mergeSort directamente
func TestMergeSortFunction(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []int
	}{
		{
			name:     "Empty slice",
			input:    []int{},
			expected: []int{},
		},
		{
			name:     "Single element",
			input:    []int{42},
			expected: []int{42},
		},
		{
			name:     "Already sorted",
			input:    []int{1, 2, 3, 4, 5},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "Reverse sorted",
			input:    []int{5, 4, 3, 2, 1},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "Random order",
			input:    []int{42, 17, 89, 3, 156, 7},
			expected: []int{3, 7, 17, 42, 89, 156},
		},
		{
			name:     "Duplicates",
			input:    []int{3, 1, 4, 1, 5, 9, 2, 6, 5},
			expected: []int{1, 1, 2, 3, 4, 5, 5, 6, 9},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Hacer copia para no modificar input
			inputCopy := make([]int, len(tt.input))
			copy(inputCopy, tt.input)

			result := mergeSort(inputCopy)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected length %d, got %d", len(tt.expected), len(result))
				return
			}

			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("At index %d: expected %d, got %d", i, tt.expected[i], v)
				}
			}
		})
	}
}

// TestQuickSortFunction prueba la función quickSort directamente
func TestQuickSortFunction(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []int
	}{
		{
			name:     "Empty slice",
			input:    []int{},
			expected: []int{},
		},
		{
			name:     "Single element",
			input:    []int{42},
			expected: []int{42},
		},
		{
			name:     "Already sorted",
			input:    []int{1, 2, 3, 4, 5},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "Reverse sorted",
			input:    []int{5, 4, 3, 2, 1},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "Random order",
			input:    []int{42, 17, 89, 3, 156, 7},
			expected: []int{3, 7, 17, 42, 89, 156},
		},
		{
			name:     "Duplicates",
			input:    []int{3, 1, 4, 1, 5, 9, 2, 6, 5},
			expected: []int{1, 1, 2, 3, 4, 5, 5, 6, 9},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Hacer copia para no modificar input
			inputCopy := make([]int, len(tt.input))
			copy(inputCopy, tt.input)

			quickSort(inputCopy)

			if len(inputCopy) != len(tt.expected) {
				t.Errorf("Expected length %d, got %d", len(tt.expected), len(inputCopy))
				return
			}

			for i, v := range inputCopy {
				if v != tt.expected[i] {
					t.Errorf("At index %d: expected %d, got %d", i, tt.expected[i], v)
				}
			}
		})
	}
}

// BenchmarkMergeSort benchmarks para comparar rendimiento
func BenchmarkMergeSort(b *testing.B) {
	data := make([]int, 1000)
	for i := range data {
		data[i] = 1000 - i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testData := make([]int, len(data))
		copy(testData, data)
		mergeSort(testData)
	}
}

// BenchmarkQuickSort
func BenchmarkQuickSort(b *testing.B) {
	data := make([]int, 1000)
	for i := range data {
		data[i] = 1000 - i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testData := make([]int, len(data))
		copy(testData, data)
		quickSort(testData)
	}
}
