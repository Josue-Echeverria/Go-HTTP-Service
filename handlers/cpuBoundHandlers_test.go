package handlers

import (
	"GoDocker/server"
	"encoding/json"
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
	expectedAlgorithm := "Machin"
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

	if result["algorithm"] != expectedAlgorithm {
		t.Errorf("Expected algorithm '%s', got %v", expectedAlgorithm, result["algorithm"])
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
	expectedMessage := "Mandelbrot set generated"
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

	if !strings.Contains(resp.Body, expectedMessage) {
		t.Errorf("Response should contain '%s'", expectedMessage)
	}
}

func TestMatrixMulHandler(t *testing.T) {
	expectedStatus := 200
	expectedErrorStatus := 400
	expectedSize := 3
	expectedSeed := "42"
	expectedFields := []string{"size", "seed", "matrixA", "matrixB", "result", "message"}

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
