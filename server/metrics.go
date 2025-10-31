package server

import (
	"math"
	"sync"
	"time"
)

// RequestMetrics almacena métricas de requests por endpoint
type RequestMetrics struct {
	mu             sync.RWMutex
	waitTimes      []float64 // Tiempo en cola (ms)
	execTimes      []float64 // Tiempo de ejecución (ms)
	totalRequests  int64
	activeRequests int64
	endpoint       string
	lastUpdateTime time.Time
}

// NewRequestMetrics crea nuevas métricas para un endpoint
func NewRequestMetrics(endpoint string) *RequestMetrics {
	return &RequestMetrics{
		waitTimes:      make([]float64, 0, 1000),
		execTimes:      make([]float64, 0, 1000),
		endpoint:       endpoint,
		lastUpdateTime: time.Now(),
	}
}

// RecordWaitTime registra tiempo de espera en cola
func (rm *RequestMetrics) RecordWaitTime(duration time.Duration) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	ms := float64(duration.Milliseconds())
	rm.waitTimes = append(rm.waitTimes, ms)

	// Mantener solo últimas 1000 mediciones
	if len(rm.waitTimes) > 1000 {
		rm.waitTimes = rm.waitTimes[len(rm.waitTimes)-1000:]
	}
}

// RecordExecTime registra tiempo de ejecución
func (rm *RequestMetrics) RecordExecTime(duration time.Duration) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	ms := float64(duration.Milliseconds())
	rm.execTimes = append(rm.execTimes, ms)

	if len(rm.execTimes) > 1000 {
		rm.execTimes = rm.execTimes[len(rm.execTimes)-1000:]
	}

	rm.totalRequests++
	rm.lastUpdateTime = time.Now()
}

// IncrementActive incrementa requests activos
func (rm *RequestMetrics) IncrementActive() {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.activeRequests++
}

// DecrementActive decrementa requests activos
func (rm *RequestMetrics) DecrementActive() {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.activeRequests--
}

// GetStats retorna estadísticas calculadas
func (rm *RequestMetrics) GetStats() map[string]interface{} {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	return map[string]interface{}{
		"endpoint":        rm.endpoint,
		"total_requests":  rm.totalRequests,
		"active_requests": rm.activeRequests,
		"wait_time": map[string]float64{
			"avg_ms":  calculateAverage(rm.waitTimes),
			"std_dev": calculateStdDev(rm.waitTimes),
			"min_ms":  calculateMin(rm.waitTimes),
			"max_ms":  calculateMax(rm.waitTimes),
			"p50_ms":  calculatePercentile(rm.waitTimes, 0.50),
			"p95_ms":  calculatePercentile(rm.waitTimes, 0.95),
			"p99_ms":  calculatePercentile(rm.waitTimes, 0.99),
		},
		"exec_time": map[string]float64{
			"avg_ms":  calculateAverage(rm.execTimes),
			"std_dev": calculateStdDev(rm.execTimes),
			"min_ms":  calculateMin(rm.execTimes),
			"max_ms":  calculateMax(rm.execTimes),
			"p50_ms":  calculatePercentile(rm.execTimes, 0.50),
			"p95_ms":  calculatePercentile(rm.execTimes, 0.95),
			"p99_ms":  calculatePercentile(rm.execTimes, 0.99),
		},
		"last_update": rm.lastUpdateTime.Format(time.RFC3339),
	}
}

// MetricsManager gestiona métricas de todos los endpoints
type MetricsManager struct {
	mu      sync.RWMutex
	metrics map[string]*RequestMetrics
}

// NewMetricsManager crea un nuevo gestor de métricas
func NewMetricsManager() *MetricsManager {
	return &MetricsManager{
		metrics: make(map[string]*RequestMetrics),
	}
}

// GetOrCreate obtiene o crea métricas para un endpoint
func (mm *MetricsManager) GetOrCreate(endpoint string) *RequestMetrics {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	if metrics, exists := mm.metrics[endpoint]; exists {
		return metrics
	}

	metrics := NewRequestMetrics(endpoint)
	mm.metrics[endpoint] = metrics
	return metrics
}

// GetAllStats retorna estadísticas de todos los endpoints
func (mm *MetricsManager) GetAllStats() map[string]interface{} {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	allStats := make(map[string]interface{})

	for endpoint, metrics := range mm.metrics {
		allStats[endpoint] = metrics.GetStats()
	}

	return allStats
}

// Funciones auxiliares para cálculos estadísticos

func calculateAverage(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return math.Round(sum/float64(len(values))*100) / 100
}

func calculateStdDev(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	avg := calculateAverage(values)
	variance := 0.0
	for _, v := range values {
		diff := v - avg
		variance += diff * diff
	}
	variance /= float64(len(values))
	return math.Round(math.Sqrt(variance)*100) / 100
}

func calculateMin(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	min := values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
	}
	return math.Round(min*100) / 100
}

func calculateMax(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	max := values[0]
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	return math.Round(max*100) / 100
}

func calculatePercentile(values []float64, percentile float64) float64 {
	if len(values) == 0 {
		return 0
	}

	// Copiar y ordenar
	sorted := make([]float64, len(values))
	copy(sorted, values)

	// Bubble sort simple (para arrays pequeños está bien)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	index := int(float64(len(sorted)) * percentile)
	if index >= len(sorted) {
		index = len(sorted) - 1
	}
	return math.Round(sorted[index]*100) / 100
}
