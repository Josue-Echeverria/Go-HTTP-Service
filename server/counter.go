package server

import (
	"sync/atomic"
)

// Counter es un contador thread-safe usando operaciones atómicas
type Counter struct {
	value int64
}

// NewCounter crea un nuevo contador
func NewCounter() *Counter {
	return &Counter{value: 0}
}

// Increment incrementa el contador y retorna el nuevo valor
func (c *Counter) Increment() int64 {
	return atomic.AddInt64(&c.value, 1)
}

// Decrement decrementa el contador y retorna el nuevo valor
func (c *Counter) Decrement() int64 {
	return atomic.AddInt64(&c.value, -1)
}

// Get obtiene el valor actual del contador
func (c *Counter) Get() int64 {
	return atomic.LoadInt64(&c.value)
}

// Set establece un valor específico
func (c *Counter) Set(val int64) {
	atomic.StoreInt64(&c.value, val)
}

// Reset resetea el contador a 0
func (c *Counter) Reset() {
	atomic.StoreInt64(&c.value, 0)
}

// Add suma un delta al contador
func (c *Counter) Add(delta int64) int64 {
	return atomic.AddInt64(&c.value, delta)
}
