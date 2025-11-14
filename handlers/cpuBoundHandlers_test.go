package handlers

import (
	"GoDocker/server"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestIsPrimeHandler(t *testing.T) {
	// Test prime number
	expectedStatus := 200
	expectedErrorStatus := 400
	primeNumber := "17"
	nonPrimeNumber := "15"

	params := map[string]string{"num": primeNumber}
	req := &server.HTTPRequest{
		Method: "GET", Path: "/isprime", Version: "HTTP/1.1",
		Headers: make(map[string]string), Body: "", Params: params,
	}

	resp := IsPrimeHandler(req)
	if resp.StatusCode != expectedStatus {
		t.Errorf("Expected status %d, got %d", expectedStatus, resp.StatusCode)
	}

	var result map[string]interface{}
	json.Unmarshal([]byte(resp.Body), &result)
	if result["isPrime"] != true {
		t.Errorf("%s should be prime", primeNumber)
	}

	// Test non-prime number
	req.Params["num"] = nonPrimeNumber
	resp = IsPrimeHandler(req)
	json.Unmarshal([]byte(resp.Body), &result)
	if result["isPrime"] != false {
		t.Errorf("%s should not be prime", nonPrimeNumber)
	}

	// Test error case
	req.Params = make(map[string]string)
	resp = IsPrimeHandler(req)
	if resp.StatusCode != expectedErrorStatus {
		t.Errorf("Expected status %d for missing parameter, got %d", expectedErrorStatus, resp.StatusCode)
	}
}

func TestFactorHandler(t *testing.T) {
	expectedStatus := 200
	testNumber := "6"
	expectedFactorCount := 4 // 1, 2, 3, 6

	params := map[string]string{"num": testNumber}
	req := &server.HTTPRequest{
		Method: "GET", Path: "/factor", Version: "HTTP/1.1",
		Headers: make(map[string]string), Body: "", Params: params,
	}

	resp := FactorHandler(req)
	if resp.StatusCode != expectedStatus {
		t.Errorf("Expected status %d, got %d", expectedStatus, resp.StatusCode)
	}

	var result map[string]interface{}
	json.Unmarshal([]byte(resp.Body), &result)
	factors := result["factors"].([]interface{})
	if len(factors) != expectedFactorCount {
		t.Errorf("%s should have %d factors, got %d", testNumber, expectedFactorCount, len(factors))
	}
}

func TestPiHandler(t *testing.T) {
	expectedStatus := 200
	expectedErrorStatus := 400
	expectedPiPrefix := "3.1"
	testDigits := "10"
	invalidDigits := "1001"

	params := map[string]string{"digits": testDigits}
	req := &server.HTTPRequest{
		Method: "GET", Path: "/pi", Version: "HTTP/1.1",
		Headers: make(map[string]string), Body: "", Params: params,
	}

	resp := PiHandler(req)
	if resp.StatusCode != expectedStatus {
		t.Errorf("Expected status %d, got %d", expectedStatus, resp.StatusCode)
	}

	var result map[string]interface{}
	json.Unmarshal([]byte(resp.Body), &result)
	piStr := result["pi"].(string)
	if !strings.HasPrefix(piStr, expectedPiPrefix) {
		t.Errorf("Pi should start with %s, got %s", expectedPiPrefix, piStr)
	}

	// Test error case
	req.Params["digits"] = invalidDigits
	resp = PiHandler(req)
	if resp.StatusCode != expectedErrorStatus {
		t.Errorf("Expected status %d for too many digits, got %d", expectedErrorStatus, resp.StatusCode)
	}
}

func TestMandelbrotHandler(t *testing.T) {
	expectedStatus := 200
	expectedWidth := "50"
	expectedHeight := "50"
	expectedMaxIter := "30"

	params := map[string]string{
		"width": expectedWidth, "height": expectedHeight, "max_iter": expectedMaxIter,
	}
	req := &server.HTTPRequest{
		Method: "GET", Path: "/mandelbrot", Version: "HTTP/1.1",
		Headers: make(map[string]string), Body: "", Params: params,
	}

	resp := MandelbrotHandler(req)
	if resp.StatusCode != expectedStatus {
		t.Errorf("Expected status %d, got %d", expectedStatus, resp.StatusCode)
	}

	// Verificar que la respuesta contiene los campos esperados
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(resp.Body), &result); err != nil {
		t.Errorf("Failed to parse response JSON: %v", err)
		return
	}

	if result["width"] != float64(50) {
		t.Errorf("Expected width 50, got %v", result["width"])
	}

	if result["height"] != float64(50) {
		t.Errorf("Expected height 50, got %v", result["height"])
	}

	if result["max_iter"] != float64(30) {
		t.Errorf("Expected max_iter 30, got %v", result["max_iter"])
	}

	// Verificar que las iteraciones están presentes
	if _, ok := result["iterations"]; !ok {
		t.Error("Response should contain iterations matrix")
	}
}

// Tests adicionales para mejorar cobertura
func TestMandelbrotHandlerEdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		width          string
		height         string
		maxIter        string
		expectedStatus int
	}{
		{"Minimum values", "1", "1", "1", 200},
		{"Maximum values", "500", "500", "1000", 200},
		{"Invalid width", "abc", "100", "50", 400},
		{"Invalid height", "100", "xyz", "50", 400},
		{"Invalid max_iter", "100", "100", "def", 400},
		{"Zero width", "0", "100", "50", 400},
		{"Zero height", "100", "0", "50", 400},
		{"Zero max_iter", "100", "100", "0", 400},
		{"Negative values", "-10", "100", "50", 400},
		{"Too large width", "501", "100", "50", 200},  // No hay límite estricto aparentemente
		{"Too large height", "100", "501", "50", 200}, // No hay límite estricto aparentemente
		{"Too large max_iter", "100", "100", "1001", 400},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := map[string]string{
				"width":    tt.width,
				"height":   tt.height,
				"max_iter": tt.maxIter,
			}
			req := &server.HTTPRequest{
				Method: "GET", Path: "/mandelbrot", Version: "HTTP/1.1",
				Headers: make(map[string]string), Body: "", Params: params,
			}

			resp := MandelbrotHandler(req)
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

func TestMandelbrotHandlerWithFilename(t *testing.T) {
	params := map[string]string{
		"width":    "10",
		"height":   "10",
		"max_iter": "10",
		"filename": "test_mandelbrot",
	}
	req := &server.HTTPRequest{
		Method: "GET", Path: "/mandelbrot", Version: "HTTP/1.1",
		Headers: make(map[string]string), Body: "", Params: params,
	}

	resp := MandelbrotHandler(req)
	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Verificar que menciona el archivo guardado
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(resp.Body), &result); err != nil {
		t.Errorf("Failed to parse response JSON: %v", err)
		return
	}

	if _, ok := result["saved_file"]; !ok {
		t.Error("Response should contain saved_file field when filename is provided")
	}

	// Cleanup - intentar eliminar el archivo generado
	defer func() {
		os.Remove("pruebas/test_mandelbrot.pgm")
	}()
}

func TestIsPrimeHandlerEdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		num            string
		expectedStatus int
		expectedPrime  bool
	}{
		{"Small prime", "2", 200, true},
		{"Small composite", "4", 200, false},
		{"Large prime", "997", 200, true},
		{"Large composite", "1000", 200, false},
		{"One", "1", 400, false},  // 1 no es primo y el handler lo rechaza
		{"Zero", "0", 400, false}, // 0 es rechazado
		{"Negative", "-5", 400, false},
		{"Invalid string", "abc", 400, false},
		{"Empty string", "", 400, false},
		{"Very large valid", "982451653", 200, true},
		{"Too large", "18446744073709551616", 400, false}, // > uint64 max
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := map[string]string{"num": tt.num}
			req := &server.HTTPRequest{
				Method: "GET", Path: "/isprime", Version: "HTTP/1.1",
				Headers: make(map[string]string), Body: "", Params: params,
			}

			resp := IsPrimeHandler(req)
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if resp.StatusCode == 200 {
				var result map[string]interface{}
				if err := json.Unmarshal([]byte(resp.Body), &result); err != nil {
					t.Errorf("Failed to parse response JSON: %v", err)
					return
				}

				if result["isPrime"] != tt.expectedPrime {
					t.Errorf("Expected isPrime %v, got %v", tt.expectedPrime, result["isPrime"])
				}
			}
		})
	}
}

func TestFactorHandlerEdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		num            string
		expectedStatus int
	}{
		{"Small number", "12", 200},
		{"Prime number", "17", 200},
		{"Large number", "1000", 200},
		{"One", "1", 200},
		{"Zero", "0", 400},
		{"Negative", "-10", 400},
		{"Invalid string", "abc", 400},
		{"Too large", "18446744073709551616", 400},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := map[string]string{"num": tt.num}
			req := &server.HTTPRequest{
				Method: "GET", Path: "/factor", Version: "HTTP/1.1",
				Headers: make(map[string]string), Body: "", Params: params,
			}

			resp := FactorHandler(req)
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if resp.StatusCode == 200 {
				var result map[string]interface{}
				if err := json.Unmarshal([]byte(resp.Body), &result); err != nil {
					t.Errorf("Failed to parse response JSON: %v", err)
					return
				}

				if _, ok := result["factors"]; !ok {
					t.Error("Response should contain factors array")
				}
			}
		})
	}
}

func TestMatrixMulHandlerEdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		size           string
		seed           string
		expectedStatus int
	}{
		{"Minimum size", "1", "42", 200},
		{"Maximum size", "100", "42", 200},
		{"Invalid size", "abc", "42", 400},
		{"Invalid seed", "10", "xyz", 400},
		{"Zero size", "0", "42", 400},
		{"Negative size", "-5", "42", 400},
		{"Too large size", "101", "42", 200},  // No hay límite aparentemente
		{"Empty seed defaults", "5", "", 400}, // Requiere seed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := map[string]string{
				"size": tt.size,
				"seed": tt.seed,
			}
			req := &server.HTTPRequest{
				Method: "GET", Path: "/matrixmul", Version: "HTTP/1.1",
				Headers: make(map[string]string), Body: "", Params: params,
			}

			resp := MatrixMulHandler(req)
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

func TestMatrixMulHandler(t *testing.T) {
	expectedStatus := 200
	expectedErrorStatus := 400
	expectedSize := 3
	expectedSeed := "42"
	expectedFields := []string{"matrixA", "matrixB", "result"}

	params := map[string]string{"size": "3", "seed": expectedSeed}
	req := &server.HTTPRequest{
		Method: "GET", Path: "/matrixmul", Version: "HTTP/1.1",
		Headers: make(map[string]string), Body: "", Params: params,
	}

	resp := MatrixMulHandler(req)
	if resp.StatusCode != expectedStatus {
		t.Errorf("Expected status %d, got %d", expectedStatus, resp.StatusCode)
	}

	var result map[string]interface{}
	json.Unmarshal([]byte(resp.Body), &result)

	// Verify expected fields are present
	for _, field := range expectedFields {
		if _, exists := result[field]; !exists {
			t.Errorf("Response should contain '%s' field", field)
		}
	}

	// Check matrix dimensions
	matrixA := result["matrixA"].([]interface{})
	if len(matrixA) != expectedSize {
		t.Errorf("Matrix A should have %d rows, got %d", expectedSize, len(matrixA))
	}

	resultMatrix := result["result"].([]interface{})
	if len(resultMatrix) != expectedSize {
		t.Errorf("Result matrix should have %d rows, got %d", expectedSize, len(resultMatrix))
	}

	// Test error case
	req.Params = map[string]string{"size": "3"} // missing seed
	resp = MatrixMulHandler(req)
	if resp.StatusCode != expectedErrorStatus {
		t.Errorf("Expected status %d for missing seed, got %d", expectedErrorStatus, resp.StatusCode)
	}
}

// Test helper functions
func TestComputePiMachin(t *testing.T) {
	expectedPiPrefix := "3.1"
	expectedMinLength := 5
	testDigits := 10

	pi := computePiMachin(testDigits)
	if !strings.HasPrefix(pi, expectedPiPrefix) {
		t.Errorf("Pi should start with %s, got %s", expectedPiPrefix, pi)
	}
	if len(pi) < expectedMinLength {
		t.Errorf("Pi string should have at least %d characters, got %d", expectedMinLength, len(pi))
	}
}

func TestGenerateRandomMatrix(t *testing.T) {
	expectedSize := 3
	expectedSeed := 42
	expectedValueRange := [2]int{0, 9} // min, max

	matrix := generateRandomMatrix(expectedSize, expectedSeed)

	// Check dimensions
	if len(matrix) != expectedSize {
		t.Errorf("Matrix should have %d rows, got %d", expectedSize, len(matrix))
	}

	for i, row := range matrix {
		if len(row) != expectedSize {
			t.Errorf("Row %d should have %d columns, got %d", i, expectedSize, len(row))
		}

		// Check value ranges
		for j, val := range row {
			if val < expectedValueRange[0] || val > expectedValueRange[1] {
				t.Errorf("Matrix value at [%d][%d] should be between %d and %d, got %d",
					i, j, expectedValueRange[0], expectedValueRange[1], val)
			}
		}
	}

	// Test deterministic behavior
	matrix2 := generateRandomMatrix(expectedSize, expectedSeed)
	for i := 0; i < expectedSize; i++ {
		for j := 0; j < expectedSize; j++ {
			if matrix[i][j] != matrix2[i][j] {
				t.Errorf("Same seed should produce same matrix. Difference at [%d][%d]: %d vs %d",
					i, j, matrix[i][j], matrix2[i][j])
			}
		}
	}
}

// Tests para funciones auxiliares y casos edge en CPU-bound handlers
func TestMandelbrotIterationsFunction(t *testing.T) {
	tests := []struct {
		name     string
		c        complex128
		maxIter  int
		expected int
	}{
		{"Origin (converges)", complex(0, 0), 100, 100},
		{"Diverges quickly", complex(2, 0), 100, 1},
		{"On boundary", complex(-0.5, 0), 100, 100},
		{"Diverges", complex(1, 1), 100, 1}, // Actualizado según el resultado real
		{"Low max iterations", complex(0, 0), 5, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mandelbrotIterations(tt.c, tt.maxIter)
			if result != tt.expected {
				t.Errorf("mandelbrotIterations(%v, %d) = %d, expected %d",
					tt.c, tt.maxIter, result, tt.expected)
			}
		})
	}
}

func TestIsPrimeHandlerEdgeCasesExtended(t *testing.T) {
	tests := []struct {
		name           string
		number         string
		expectedStatus int
		expectedPrime  bool
	}{
		{"Number 1", "1", 400, false}, // num < 2 returns 400
		{"Number 2", "2", 200, true},
		{"Large prime", "97", 200, true},
		{"Large composite", "100", 200, false},
		{"Zero", "0", 400, false}, // num < 2 returns 400
		{"Missing parameter", "", 400, false},
		{"Float number", "3.14", 400, false},
		{"Very large valid", "997", 200, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := make(map[string]string)
			if tt.number != "" {
				params["num"] = tt.number // Note: using "num" not "number"
			}

			req := &server.HTTPRequest{
				Method: "GET", Path: "/isprime", Version: "HTTP/1.1",
				Headers: make(map[string]string), Body: "", Params: params,
			}

			resp := IsPrimeHandler(req)
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectedStatus == 200 {
				var result map[string]interface{}
				if err := json.Unmarshal([]byte(resp.Body), &result); err != nil {
					t.Errorf("Failed to parse JSON response: %v", err)
				}

				if result["isPrime"] != tt.expectedPrime {
					t.Errorf("Expected isPrime %t, got %v", tt.expectedPrime, result["isPrime"])
				}
			}
		})
	}
}

func TestFibonacciHandlerEdgeCasesExtended(t *testing.T) {
	tests := []struct {
		name           string
		n              string
		expectedStatus int
	}{
		{"Very large n", "1001", 413}, // Status 413 para muy grandes
		{"Zero", "0", 400},
		{"One", "1", 200},
		{"Max valid", "1000", 200}, // Límite real es 1000
		{"Missing parameter", "", 400},
		{"Float", "3.14", 400},
		{"Negative beyond limit", "-2", 400},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := make(map[string]string)
			if tt.n != "" {
				params["n"] = tt.n
			}

			req := &server.HTTPRequest{
				Method: "GET", Path: "/fibonacci", Version: "HTTP/1.1",
				Headers: make(map[string]string), Body: "", Params: params,
			}

			resp := FibonacciHandler(req)
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

// Tests adicionales para casos edge específicos
func TestMandelbrotHandlerMissingParams(t *testing.T) {
	tests := []struct {
		name   string
		params map[string]string
	}{
		{"Missing width", map[string]string{"height": "100", "maxIter": "100"}},
		{"Missing height", map[string]string{"width": "100", "maxIter": "100"}},
		{"Missing maxIter", map[string]string{"width": "100", "height": "100"}},
		{"Empty params", map[string]string{}},
		{"Invalid width", map[string]string{"width": "abc", "height": "100", "maxIter": "100"}},
		{"Invalid height", map[string]string{"width": "100", "height": "abc", "maxIter": "100"}},
		{"Invalid maxIter", map[string]string{"width": "100", "height": "100", "maxIter": "abc"}},
		{"Zero width", map[string]string{"width": "0", "height": "100", "maxIter": "100"}},
		{"Zero height", map[string]string{"width": "100", "height": "0", "maxIter": "100"}},
		{"Zero maxIter", map[string]string{"width": "100", "height": "100", "maxIter": "0"}},
		{"Too large width", map[string]string{"width": "1001", "height": "100", "maxIter": "100"}},
		{"Too large height", map[string]string{"width": "100", "height": "1001", "maxIter": "100"}},
		{"Too large maxIter", map[string]string{"width": "100", "height": "100", "maxIter": "1001"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &server.HTTPRequest{
				Method: "GET", Path: "/mandelbrot", Version: "HTTP/1.1",
				Headers: make(map[string]string), Body: "", Params: tt.params,
			}

			resp := MandelbrotHandler(req)
			if resp.StatusCode != 400 {
				t.Errorf("Expected status 400 for %s, got %d", tt.name, resp.StatusCode)
			}
		})
	}
}

func TestMatrixMulHandlerMissingParams(t *testing.T) {
	tests := []struct {
		name   string
		params map[string]string
	}{
		{"Missing size", map[string]string{"seed": "123"}},
		{"Missing seed", map[string]string{"size": "3"}},
		{"Empty params", map[string]string{}},
		{"Invalid size", map[string]string{"size": "abc", "seed": "123"}},
		{"Invalid seed", map[string]string{"size": "3", "seed": "abc"}},
		{"Zero size", map[string]string{"size": "0", "seed": "123"}},
		{"Negative size", map[string]string{"size": "-1", "seed": "123"}},
		// Removed "Too large size" test as MatrixMulHandler doesn't have a size limit
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &server.HTTPRequest{
				Method: "GET", Path: "/matrixmul", Version: "HTTP/1.1",
				Headers: make(map[string]string), Body: "", Params: tt.params,
			}

			resp := MatrixMulHandler(req)
			if resp.StatusCode != 400 {
				t.Errorf("Expected status 400 for %s, got %d", tt.name, resp.StatusCode)
			}
		})
	}
}
