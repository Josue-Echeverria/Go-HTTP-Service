package handlers

import (
	"GoDocker/server"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
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
		"digits": num,
		"pi":     pi,
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

// /mandelbrot?width=W&height=H&max_iter=I&filename=name
func MandelbrotHandler(req *server.HTTPRequest) *server.HTTPResponse {
	widthStr, widthOk := req.Params["width"]
	heightStr, heightOk := req.Params["height"]
	maxIterStr, maxIterOk := req.Params["max_iter"]

	if !widthOk || !heightOk || !maxIterOk {
		return &server.HTTPResponse{
			StatusCode: 400,
			StatusText: "Bad Request",
			Body:       `{"error":"missing required query parameters 'width', 'height', and 'max_iter'"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}
	}

	width, err := strconv.Atoi(widthStr)
	if err != nil || width < 1 || width > 2000 {
		return &server.HTTPResponse{
			StatusCode: 400,
			StatusText: "Bad Request",
			Body:       `{"error":"invalid width; must be an integer between 1 and 2000"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}
	}

	height, err := strconv.Atoi(heightStr)
	if err != nil || height < 1 || height > 2000 {
		return &server.HTTPResponse{
			StatusCode: 400,
			StatusText: "Bad Request",
			Body:       `{"error":"invalid height; must be an integer between 1 and 2000"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}
	}

	maxIter, err := strconv.Atoi(maxIterStr)
	if err != nil || maxIter < 1 || maxIter > 1000 {
		return &server.HTTPResponse{
			StatusCode: 400,
			StatusText: "Bad Request",
			Body:       `{"error":"invalid max_iter; must be an integer between 1 and 1000"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}
	}

	// Parámetros opcionales
	filename := req.Params["filename"]

	// Generar conjunto de Mandelbrot
	iterations := generateMandelbrotSet(width, height, maxIter)

	// Respuesta base
	response := map[string]interface{}{
		"width":      width,
		"height":     height,
		"max_iter":   maxIter,
		"iterations": iterations,
		"stats": map[string]interface{}{
			"total_pixels":      width * height,
			"computed_pixels":   len(iterations) * len(iterations[0]),
			"coordinate_system": "complex plane from -2-2i to 2+2i",
		},
	}

	// Si se especifica filename, guardar imagen PGM
	if filename != "" {
		err := saveMandelbrotPGM(iterations, width, height, maxIter, filename)
		if err != nil {
			response["file_error"] = err.Error()
		} else {
			response["saved_file"] = filename + ".pgm"
			response["file_format"] = "PGM (Portable Gray Map)"
		}
	}

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

// generateMandelbrotSet genera el conjunto de Mandelbrot
func generateMandelbrotSet(width, height, maxIter int) [][]int {
	iterations := make([][]int, height)
	for i := range iterations {
		iterations[i] = make([]int, width)
	}

	// Mapear coordenadas de píxel a plano complejo
	// Rango: -2.5 a 1.5 en x, -2.0 a 2.0 en y
	xMin, xMax := -2.5, 1.5
	yMin, yMax := -2.0, 2.0

	for py := 0; py < height; py++ {
		for px := 0; px < width; px++ {
			// Convertir coordenadas de píxel a coordenadas complejas
			x := xMin + float64(px)*(xMax-xMin)/float64(width)
			y := yMin + float64(py)*(yMax-yMin)/float64(height)

			// Calcular iteraciones para el punto (x + yi)
			iterations[py][px] = mandelbrotIterations(complex(x, y), maxIter)
		}
	}

	return iterations
}

// mandelbrotIterations calcula el número de iteraciones para un punto complejo
func mandelbrotIterations(c complex128, maxIter int) int {
	z := complex(0, 0)

	for iter := 0; iter < maxIter; iter++ {
		// z = z² + c
		z = z*z + c

		// Si |z| > 2, el punto diverge
		if real(z)*real(z)+imag(z)*imag(z) > 4.0 {
			return iter
		}
	}

	// Punto está en el conjunto (no diverge)
	return maxIter
}

// saveMandelbrotPGM guarda el conjunto de Mandelbrot como imagen PGM
func saveMandelbrotPGM(iterations [][]int, width, height, maxIter int, filename string) error {
	file, err := os.Create(filename + ".pgm")
	if err != nil {
		return err
	}
	defer file.Close()

	// Escribir cabecera PGM
	fmt.Fprintf(file, "P2\n")
	fmt.Fprintf(file, "# Mandelbrot Set\n")
	fmt.Fprintf(file, "%d %d\n", width, height)
	fmt.Fprintf(file, "255\n")

	// Escribir datos de píxeles
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Mapear iteraciones a escala de grises (0-255)
			var grayValue int
			if iterations[y][x] == maxIter {
				grayValue = 0 // Negro para puntos en el conjunto
			} else {
				// Gradiente para puntos que divergen
				grayValue = (iterations[y][x] * 255) / maxIter
			}

			fmt.Fprintf(file, "%d ", grayValue)
		}
		fmt.Fprintf(file, "\n")
	}

	return nil
}
