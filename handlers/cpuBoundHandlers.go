package handlers

import (
	"GoDocker/server"
	"encoding/json"
	"math/big"
	"strconv"
)

// /isprime?num=N
func IsPrimeHandler(req *server.HTTPRequest) *server.HTTPResponse {
	numStr, ok := req.Params["num"]
	if !ok || numStr == "" {
		return &server.HTTPResponse{
			StatusCode: 400,
			StatusText: "Bad Request",
			Body:       `{"error":"missing required query parameter 'num'"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}
	}

	num, err := strconv.Atoi(numStr)
	if err != nil || num < 2 {
		return &server.HTTPResponse{
			StatusCode: 400,
			StatusText: "Bad Request",
			Body:       `{"error":"invalid number; must be an integer >= 2"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}
	}

	isPrime := true
	for i := 2; i*i <= num; i++ {
		if num%i == 0 {
			isPrime = false
			break
		}
	}

	result := map[string]interface{}{
		"number":  num,
		"isPrime": isPrime,
	}
	body, _ := json.Marshal(result)

	return &server.HTTPResponse{
		StatusCode: 200,
		StatusText: "OK",
		Body:       string(body),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}
}

// /factor?num=N
func FactorHandler(req *server.HTTPRequest) *server.HTTPResponse {
	numStr, ok := req.Params["num"]
	if !ok || numStr == "" {
		return &server.HTTPResponse{
			StatusCode: 400,
			StatusText: "Bad Request",
			Body:       `{"error":"missing required query parameter 'num'"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}
	}

	num, err := strconv.Atoi(numStr)
	if err != nil || num < 1 {
		return &server.HTTPResponse{
			StatusCode: 400,
			StatusText: "Bad Request",
			Body:       `{"error":"invalid number; must be an integer >= 1"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}
	}

	factors := []int{}
	for i := 1; i <= num; i++ {
		if num%i == 0 {
			factors = append(factors, i)
		}
	}

	result := map[string]interface{}{
		"number":  num,
		"factors": factors,
	}
	body, _ := json.Marshal(result)

	return &server.HTTPResponse{
		StatusCode: 200,
		StatusText: "OK",
		Body:       string(body),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}
}

// computePiMachin calcula π usando la fórmula de Machin: π/4 = 4*arctan(1/5) - arctan(1/239)
func computePiMachin(digits int) string {
	// Configurar precisión
	precision := uint(digits*4 + 100)

	// Calcular arctan(1/5) y arctan(1/239) usando series de Taylor
	arctan1_5 := arctanSeries(5, precision, digits*2)
	arctan1_239 := arctanSeries(239, precision, digits*2)

	// π/4 = 4*arctan(1/5) - arctan(1/239)
	piQuarter := big.NewFloat(0.0)
	piQuarter.SetPrec(precision)

	// 4 * arctan(1/5)
	four := big.NewFloat(4.0)
	four.SetPrec(precision)
	fourArctan := big.NewFloat(0.0)
	fourArctan.SetPrec(precision)
	fourArctan.Mul(four, arctan1_5)

	// 4*arctan(1/5) - arctan(1/239)
	piQuarter.Sub(fourArctan, arctan1_239)

	// π = 4 * (π/4)
	pi := big.NewFloat(0.0)
	pi.SetPrec(precision)
	pi.Mul(four, piQuarter)

	return pi.Text('f', digits)
}

// arctanSeries calcula arctan(1/x) usando la serie de Taylor
func arctanSeries(x int, precision uint, terms int) *big.Float {
	result := big.NewFloat(0.0)
	result.SetPrec(precision)

	// 1/x
	invX := big.NewFloat(1.0)
	invX.SetPrec(precision)
	invX.Quo(invX, big.NewFloat(float64(x)).SetPrec(precision))

	// x^2
	xSquared := big.NewFloat(float64(x * x))
	xSquared.SetPrec(precision)

	// Término actual de la serie
	term := big.NewFloat(0.0)
	term.SetPrec(precision)
	term.Set(invX)

	// Potencia actual de (1/x)
	power := big.NewFloat(0.0)
	power.SetPrec(precision)
	power.Set(invX)

	sign := 1
	for n := 0; n < terms; n++ {
		// Calcular término: sign * power / (2*n + 1)
		denominator := big.NewFloat(float64(2*n + 1))
		denominator.SetPrec(precision)

		currentTerm := big.NewFloat(0.0)
		currentTerm.SetPrec(precision)
		currentTerm.Quo(power, denominator)

		if sign > 0 {
			result.Add(result, currentTerm)
		} else {
			result.Sub(result, currentTerm)
		}

		// Preparar siguiente iteración
		power.Quo(power, xSquared)
		sign *= -1
	}

	return result
}

// /pi?digits=N
func PiHandler(req *server.HTTPRequest) *server.HTTPResponse {
	numStr, ok := req.Params["digits"]
	if !ok || numStr == "" {
		return &server.HTTPResponse{
			StatusCode: 400,
			StatusText: "Bad Request",
			Body:       `{"error":"missing required query parameter 'digits'"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}
	}

	num, err := strconv.Atoi(numStr)
	if err != nil || num < 1 || num > 1000 {
		return &server.HTTPResponse{
			StatusCode: 400,
			StatusText: "Bad Request",
			Body:       `{"error":"invalid digits; must be an integer between 1 and 1000"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}
	}

	// Calcular π usando la fórmula de Machin
	pi := computePiMachin(num)

	result := map[string]interface{}{
		"digits":    num,
		"pi":        pi,
	}

	body, _ := json.MarshalIndent(result, "", "  ")

	return &server.HTTPResponse{
		StatusCode: 200,
		StatusText: "OK",
		Body:       string(body),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}
}

// /mandelbrot?width=W&height=H&max_iter=I
func MandelbrotHandler(req *server.HTTPRequest) *server.HTTPResponse {
	// TODO : Implementar handler para generar conjunto de Mandelbrot
	return &server.HTTPResponse{
		StatusCode: 400,
		StatusText: "OK",
		Body:       `{"message":"Mandelbrot set generated"}`,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}
}

// /matrixmul?size=N&seed=S
func MatrixMulHandler(req *server.HTTPRequest) *server.HTTPResponse {
	sizeStr, sizeOk := req.Params["size"]
	seedStr, seedOk := req.Params["seed"]

	if !sizeOk || !seedOk {
		return &server.HTTPResponse{
			StatusCode: 400,
			StatusText: "Bad Request",
			Body:       `{"error":"missing required query parameters"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}
	}

	size, err := strconv.Atoi(sizeStr)
	if err != nil || size < 1 {
		return &server.HTTPResponse{
			StatusCode: 400,
			StatusText: "Bad Request",
			Body:       `{"error":"invalid size parameter"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}
	}

	seed, err := strconv.Atoi(seedStr)
	if err != nil {
		return &server.HTTPResponse{
			StatusCode: 400,
			StatusText: "Bad Request",
			Body:       `{"error":"invalid seed parameter"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}
	}

	matrixA := generateRandomMatrix(size, seed)
	matrixB := generateRandomMatrix(size, seed+1)
	result := make([][]int, size)
	for i := range result {
		result[i] = make([]int, size)
	}

	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			sum := 0
			for k := 0; k < size; k++ {
				sum += matrixA[i][k] * matrixB[k][j]
			}
			result[i][j] = sum
		}
	}

	// Crear respuesta con matrices y resultado
	response := map[string]interface{}{
		"matrixA": matrixA,
		"matrixB": matrixB,
		"result":  result,
	}

	// Devolver resultado
	body, _ := json.MarshalIndent(response, "", "  ")

	return &server.HTTPResponse{
		StatusCode: 200,
		StatusText: "OK",
		Body:       string(body),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}
}

func generateRandomMatrix(size, seed int) [][]int {
	matrix := make([][]int, size)

	for i := 0; i < size; i++ {
		matrix[i] = make([]int, size)
		for j := 0; j < size; j++ {
			matrix[i][j] = int(nextRandom(uint64(seed)) % 10) // Números entre 0 y 9
		}
	}

	return matrix
}
