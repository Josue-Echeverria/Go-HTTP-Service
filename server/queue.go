package server

import (
	"sync"
)

// TaskQueue es una cola thread-safe para tareas
type TaskQueue struct {
	items    []interface{}
	capacity int
	mu       sync.Mutex
	notEmpty chan struct{}
	closed   bool
}

// NewTaskQueue crea una nueva cola de tareas
func NewTaskQueue(capacity int) *TaskQueue {
	if capacity <= 0 {
		capacity = 100
	}

	return &TaskQueue{
		items:    make([]interface{}, 0, capacity),
		capacity: capacity,
		notEmpty: make(chan struct{}, 1),
		closed:   false,
	}
}

// Enqueue agrega un elemento a la cola
func (q *TaskQueue) Enqueue(item interface{}) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return false
	}

	if len(q.items) >= q.capacity {
		return false
	}

	q.items = append(q.items, item)

	// Señalizar que hay elementos
	select {
	case q.notEmpty <- struct{}{}:
	default:
	}

	return true
}

// Dequeue remueve y retorna un elemento de la cola
func (q *TaskQueue) Dequeue() (interface{}, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) == 0 {
		return nil, false
	}

	item := q.items[0]
	q.items = q.items[1:]

	return item, true
}

// Size retorna el tamaño actual de la cola
func (q *TaskQueue) Size() int64 {
	q.mu.Lock()
	defer q.mu.Unlock()
	return int64(len(q.items))
}

// NotEmpty retorna un canal que se señaliza cuando hay elementos
func (q *TaskQueue) NotEmpty() <-chan struct{} {
	return q.notEmpty
}

// Close cierra la cola
func (q *TaskQueue) Close() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.closed = true
	close(q.notEmpty)
}

// IsEmpty retorna true si la cola está vacía
func (q *TaskQueue) IsEmpty() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.items) == 0
}

// IsFull retorna true si la cola está llena
func (q *TaskQueue) IsFull() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.items) >= q.capacity
}
