package server

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// JobStatus representa el estado de un trabajo
type JobStatus string

const (
	JobQueued   JobStatus = "queued"
	JobRunning  JobStatus = "running"
	JobDone     JobStatus = "done"
	JobError    JobStatus = "error"
	JobCanceled JobStatus = "canceled"
	JobTimeout  JobStatus = "timeout"
)

// JobPriority representa la prioridad de un trabajo
type JobPriority int

const (
	PriorityLow    JobPriority = 0
	PriorityNormal JobPriority = 1
	PriorityHigh   JobPriority = 2
)

// Job representa un trabajo asíncrono
type Job struct {
	ID          string                 `json:"job_id"`
	Task        string                 `json:"task"`
	Params      map[string]string      `json:"params"`
	Status      JobStatus              `json:"status"`
	Priority    JobPriority            `json:"priority"`
	Progress    int                    `json:"progress"`
	ETA         int64                  `json:"eta_ms"`
	Result      map[string]interface{} `json:"result,omitempty"`
	Error       string                 `json:"error,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Timeout     time.Duration          `json:"-"`
	CancelFunc  context.CancelFunc     `json:"-"`
	mu          sync.RWMutex           `json:"-"`
}

// UpdateProgress actualiza el progreso del job
func (j *Job) UpdateProgress(progress int) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.Progress = progress
}

// UpdateETA actualiza el tiempo estimado de finalización
func (j *Job) UpdateETA(eta time.Duration) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.ETA = eta.Milliseconds()
}

// SetResult establece el resultado del job
func (j *Job) SetResult(result map[string]interface{}) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.Status = JobDone
	j.Result = result
	j.Progress = 100
	now := time.Now()
	j.CompletedAt = &now
}

// SetError establece un error en el job
func (j *Job) SetError(err error) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.Status = JobError
	j.Error = err.Error()
	now := time.Now()
	j.CompletedAt = &now
}

// Cancel intenta cancelar el job
func (j *Job) Cancel() bool {
	j.mu.Lock()
	defer j.mu.Unlock()

	if j.Status == JobDone || j.Status == JobError || j.Status == JobCanceled {
		return false // No se puede cancelar
	}

	if j.CancelFunc != nil {
		j.CancelFunc()
	}

	j.Status = JobCanceled
	now := time.Now()
	j.CompletedAt = &now
	return true
}

// GetInfo retorna información del job de forma thread-safe
func (j *Job) GetInfo() map[string]interface{} {
	j.mu.RLock()
	defer j.mu.RUnlock()

	info := map[string]interface{}{
		"job_id":     j.ID,
		"task":       j.Task,
		"status":     j.Status,
		"progress":   j.Progress,
		"eta_ms":     j.ETA,
		"created_at": j.CreatedAt.Format(time.RFC3339),
	}

	if j.StartedAt != nil {
		info["started_at"] = j.StartedAt.Format(time.RFC3339)
	}

	if j.CompletedAt != nil {
		info["completed_at"] = j.CompletedAt.Format(time.RFC3339)
	}

	if j.Error != "" {
		info["error"] = j.Error
	}

	if j.Result != nil {
		info["result"] = j.Result
	}

	return info
}

// JobManager gestiona trabajos asincrónicos
type JobManager struct {
	jobs            map[string]*Job
	queues          map[string][]*Job // Cola por tipo de tarea
	mu              sync.RWMutex
	maxQueueSize    int
	maxConcurrent   map[string]int // Límite de concurrencia por tipo
	activeCounts    map[string]int // Trabajos activos por tipo
	cpuTimeout      time.Duration
	ioTimeout       time.Duration
	persistenceFile string
	shutdownCh      chan struct{}
	wg              sync.WaitGroup
	executor        TaskExecutor // Interface para ejecutar tareas
}

// TaskExecutor ejecuta tareas específicas
type TaskExecutor interface {
	Execute(ctx context.Context, task string, params map[string]string) (map[string]interface{}, error)
}

// NewJobManager crea un nuevo gestor de trabajos
func NewJobManager(maxQueueSize int, cpuTimeout, ioTimeout time.Duration, persistenceFile string) *JobManager {
	jm := &JobManager{
		jobs:            make(map[string]*Job),
		queues:          make(map[string][]*Job),
		maxQueueSize:    maxQueueSize,
		maxConcurrent:   make(map[string]int),
		activeCounts:    make(map[string]int),
		cpuTimeout:      cpuTimeout,
		ioTimeout:       ioTimeout,
		persistenceFile: persistenceFile,
		shutdownCh:      make(chan struct{}),
	}

	// Configurar límites de concurrencia por tipo
	jm.maxConcurrent["cpu"] = 4 // Máximo 4 trabajos CPU concurrentes
	jm.maxConcurrent["io"] = 10 // Máximo 10 trabajos IO concurrentes

	// Cargar jobs persistidos
	jm.loadJobs()

	// Iniciar procesador de cola
	jm.wg.Add(1)
	go jm.processQueue()

	return jm
}

// SetExecutor establece el executor de tareas
func (jm *JobManager) SetExecutor(executor TaskExecutor) {
	jm.executor = executor
}

// Submit encola un nuevo trabajo
func (jm *JobManager) Submit(task string, params map[string]string, priority JobPriority) (*Job, error) {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	// Determinar tipo de tarea
	taskType := jm.getTaskType(task)

	// Verificar tamaño de cola
	if len(jm.queues[taskType]) >= jm.maxQueueSize {
		return nil, fmt.Errorf("queue full")
	}

	// Crear job
	jobID := fmt.Sprintf("%s-%d", task, time.Now().UnixNano())

	// Determinar timeout según tipo
	timeout := jm.cpuTimeout
	if taskType == "io" {
		timeout = jm.ioTimeout
	}

	job := &Job{
		ID:        jobID,
		Task:      task,
		Params:    params,
		Status:    JobQueued,
		Priority:  priority,
		Progress:  0,
		CreatedAt: time.Now(),
		Timeout:   timeout,
	}

	// Agregar a jobs y cola
	jm.jobs[jobID] = job
	jm.queues[taskType] = append(jm.queues[taskType], job)

	// Ordenar cola por prioridad
	jm.sortQueue(taskType)

	// Persistir
	jm.saveJobs()

	return job, nil
}

// GetJob obtiene un trabajo por ID
func (jm *JobManager) GetJob(jobID string) (*Job, error) {
	jm.mu.RLock()
	defer jm.mu.RUnlock()

	job, exists := jm.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("job not found")
	}

	return job, nil
}

// CancelJob intenta cancelar un trabajo
func (jm *JobManager) CancelJob(jobID string) (bool, error) {
	jm.mu.RLock()
	job, exists := jm.jobs[jobID]
	jm.mu.RUnlock()

	if !exists {
		return false, fmt.Errorf("job not found")
	}

	canceled := job.Cancel()

	// Persistir cambio
	jm.mu.Lock()
	jm.saveJobs()
	jm.mu.Unlock()

	return canceled, nil
}

// processQueue procesa la cola de trabajos
func (jm *JobManager) processQueue() {
	defer jm.wg.Done()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-jm.shutdownCh:
			return
		case <-ticker.C:
			jm.processNextJob()
		}
	}
}

// processNextJob intenta procesar el siguiente trabajo de mayor prioridad
func (jm *JobManager) processNextJob() {
	jm.mu.Lock()

	// Buscar el siguiente job a procesar (ordenado por prioridad)
	var nextJob *Job
	var taskType string

	for typ, queue := range jm.queues {
		if len(queue) == 0 {
			continue
		}

		// Verificar límite de concurrencia
		if jm.activeCounts[typ] >= jm.maxConcurrent[typ] {
			continue
		}

		// Tomar el primero de la cola (ya ordenado por prioridad)
		for i, job := range queue {
			if job.Status == JobQueued {
				nextJob = job
				taskType = typ
				// Remover de la cola
				jm.queues[typ] = append(queue[:i], queue[i+1:]...)
				break
			}
		}

		if nextJob != nil {
			break
		}
	}

	if nextJob == nil {
		jm.mu.Unlock()
		return
	}

	// Marcar como running
	now := time.Now()
	nextJob.mu.Lock()
	nextJob.Status = JobRunning
	nextJob.StartedAt = &now
	nextJob.mu.Unlock()

	jm.activeCounts[taskType]++
	jm.mu.Unlock()

	// Ejecutar en goroutine separada
	jm.wg.Add(1)
	go jm.executeJob(nextJob, taskType)
}

// executeJob ejecuta un trabajo
func (jm *JobManager) executeJob(job *Job, taskType string) {
	defer jm.wg.Done()
	defer func() {
		jm.mu.Lock()
		jm.activeCounts[taskType]--
		jm.saveJobs()
		jm.mu.Unlock()
	}()

	// Crear contexto con timeout y cancelación
	ctx, cancel := context.WithTimeout(context.Background(), job.Timeout)
	job.CancelFunc = cancel
	defer cancel()

	// Ejecutar tarea
	resultCh := make(chan map[string]interface{}, 1)
	errorCh := make(chan error, 1)

	go func() {
		result, err := jm.executeTask(ctx, job)
		if err != nil {
			errorCh <- err
		} else {
			resultCh <- result
		}
	}()

	// Esperar resultado o timeout
	select {
	case result := <-resultCh:
		job.SetResult(result)
	case err := <-errorCh:
		job.SetError(err)
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			job.mu.Lock()
			job.Status = JobTimeout
			job.Error = "timeout exceeded"
			now := time.Now()
			job.CompletedAt = &now
			job.mu.Unlock()
		}
	}
}

// executeTask ejecuta la tarea específica usando el executor
func (jm *JobManager) executeTask(ctx context.Context, job *Job) (map[string]interface{}, error) {
	if jm.executor == nil {
		// Simulación de progreso para testing
		for i := 0; i <= 100; i += 10 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
				job.UpdateProgress(i)
				time.Sleep(100 * time.Millisecond)
			}
		}

		return map[string]interface{}{
			"task":      job.Task,
			"completed": true,
			"message":   "Job completed successfully (simulated)",
		}, nil
	}

	// Usar el executor real
	return jm.executor.Execute(ctx, job.Task, job.Params)
}

// getTaskType determina el tipo de tarea (cpu o io)
func (jm *JobManager) getTaskType(task string) string {
	cpuTasks := map[string]bool{
		"isprime":    true,
		"factor":     true,
		"pi":         true,
		"mandelbrot": true,
		"matrixmul":  true,
	}

	if cpuTasks[task] {
		return "cpu"
	}
	return "io"
}

// sortQueue ordena la cola por prioridad (high > normal > low)
func (jm *JobManager) sortQueue(taskType string) {
	queue := jm.queues[taskType]

	// Bubble sort simple por prioridad
	for i := 0; i < len(queue); i++ {
		for j := i + 1; j < len(queue); j++ {
			if queue[j].Priority > queue[i].Priority {
				queue[i], queue[j] = queue[j], queue[i]
			}
		}
	}
}

// saveJobs persiste los jobs a disco
func (jm *JobManager) saveJobs() {
	if jm.persistenceFile == "" {
		return
	}

	// Serializar jobs
	type jobPersist struct {
		ID          string                 `json:"job_id"`
		Task        string                 `json:"task"`
		Params      map[string]string      `json:"params"`
		Status      JobStatus              `json:"status"`
		Priority    JobPriority            `json:"priority"`
		Progress    int                    `json:"progress"`
		Result      map[string]interface{} `json:"result,omitempty"`
		Error       string                 `json:"error,omitempty"`
		CreatedAt   time.Time              `json:"created_at"`
		StartedAt   *time.Time             `json:"started_at,omitempty"`
		CompletedAt *time.Time             `json:"completed_at,omitempty"`
	}

	var persist []jobPersist
	for _, job := range jm.jobs {
		job.mu.RLock()
		persist = append(persist, jobPersist{
			ID:          job.ID,
			Task:        job.Task,
			Params:      job.Params,
			Status:      job.Status,
			Priority:    job.Priority,
			Progress:    job.Progress,
			Result:      job.Result,
			Error:       job.Error,
			CreatedAt:   job.CreatedAt,
			StartedAt:   job.StartedAt,
			CompletedAt: job.CompletedAt,
		})
		job.mu.RUnlock()
	}

	data, err := json.MarshalIndent(persist, "", "  ")
	if err != nil {
		return
	}

	os.WriteFile(jm.persistenceFile, data, 0644)
}

// loadJobs carga los jobs desde disco
func (jm *JobManager) loadJobs() {
	if jm.persistenceFile == "" {
		return
	}

	data, err := os.ReadFile(jm.persistenceFile)
	if err != nil {
		return
	}

	type jobPersist struct {
		ID          string                 `json:"job_id"`
		Task        string                 `json:"task"`
		Params      map[string]string      `json:"params"`
		Status      JobStatus              `json:"status"`
		Priority    JobPriority            `json:"priority"`
		Progress    int                    `json:"progress"`
		Result      map[string]interface{} `json:"result,omitempty"`
		Error       string                 `json:"error,omitempty"`
		CreatedAt   time.Time              `json:"created_at"`
		StartedAt   *time.Time             `json:"started_at,omitempty"`
		CompletedAt *time.Time             `json:"completed_at,omitempty"`
	}

	var persist []jobPersist
	if err := json.Unmarshal(data, &persist); err != nil {
		return
	}

	// Restaurar jobs (solo los no completados)
	for _, jp := range persist {
		if jp.Status == JobRunning {
			// Marcar como error (servidor se reinició)
			jp.Status = JobError
			jp.Error = "server restarted"
		}

		job := &Job{
			ID:          jp.ID,
			Task:        jp.Task,
			Params:      jp.Params,
			Status:      jp.Status,
			Priority:    jp.Priority,
			Progress:    jp.Progress,
			Result:      jp.Result,
			Error:       jp.Error,
			CreatedAt:   jp.CreatedAt,
			StartedAt:   jp.StartedAt,
			CompletedAt: jp.CompletedAt,
		}

		jm.jobs[job.ID] = job

		// Re-encolar si estaba queued
		if job.Status == JobQueued {
			taskType := jm.getTaskType(job.Task)
			jm.queues[taskType] = append(jm.queues[taskType], job)
		}
	}
}

// GetQueueStats retorna estadísticas de las colas
func (jm *JobManager) GetQueueStats() map[string]interface{} {
	jm.mu.RLock()
	defer jm.mu.RUnlock()

	stats := make(map[string]interface{})

	for taskType, queue := range jm.queues {
		stats[taskType] = map[string]interface{}{
			"queued":         len(queue),
			"active":         jm.activeCounts[taskType],
			"max_concurrent": jm.maxConcurrent[taskType],
		}
	}

	stats["total_jobs"] = len(jm.jobs)

	return stats
}

// Shutdown detiene el job manager
func (jm *JobManager) Shutdown() {
	close(jm.shutdownCh)
	jm.wg.Wait()

	jm.mu.Lock()
	jm.saveJobs()
	jm.mu.Unlock()
}
