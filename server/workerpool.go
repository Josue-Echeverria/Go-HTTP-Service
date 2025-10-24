package server

import (
	"log"
	"sync"
)

// WorkerPool maneja un pool de workers para procesar tareas
type WorkerPool struct {
	size       int
	workers    []*Worker
	wg         sync.WaitGroup
	stopCh     chan struct{}
	stoppedMux sync.Mutex
	stopped    bool
}

// Worker representa un worker individual
type Worker struct {
	id   int
	pool *WorkerPool
}

// TaskProcessor es una función que procesa una tarea
type TaskProcessor func(task interface{})

// NewWorkerPool crea un nuevo pool de workers
func NewWorkerPool(size int) *WorkerPool {
	if size <= 0 {
		size = 10
	}

	pool := &WorkerPool{
		size:    size,
		workers: make([]*Worker, size),
		stopCh:  make(chan struct{}),
		stopped: false,
	}

	for i := 0; i < size; i++ {
		pool.workers[i] = &Worker{
			id:   i,
			pool: pool,
		}
	}

	return pool
}

// Start inicia todos los workers del pool
func (wp *WorkerPool) Start(queue *TaskQueue, processor TaskProcessor) {
	wp.stoppedMux.Lock()
	if wp.stopped {
		wp.stoppedMux.Unlock()
		return
	}
	wp.stoppedMux.Unlock()

	for _, worker := range wp.workers {
		wp.wg.Add(1)
		go worker.start(queue, processor, &wp.wg, wp.stopCh)
	}

	log.Printf("Worker pool iniciado con %d workers", wp.size)
}

// Stop detiene todos los workers del pool
func (wp *WorkerPool) Stop() {
	wp.stoppedMux.Lock()
	if wp.stopped {
		wp.stoppedMux.Unlock()
		return
	}
	wp.stopped = true
	wp.stoppedMux.Unlock()

	close(wp.stopCh)
	wp.wg.Wait()
	log.Println("Worker pool detenido")
}

// start inicia el loop de procesamiento del worker
func (w *Worker) start(queue *TaskQueue, processor TaskProcessor, wg *sync.WaitGroup, stopCh chan struct{}) {
	defer wg.Done()

	for {
		select {
		case <-stopCh:
			return
		default:
			task, ok := queue.Dequeue()
			if !ok {
				// Cola vacía, esperar por señal
				select {
				case <-stopCh:
					return
				case <-queue.NotEmpty():
					continue
				}
			}

			// Procesar tarea
			processor(task)
		}
	}
}

// Size retorna el tamaño del pool
func (wp *WorkerPool) Size() int {
	return wp.size
}
