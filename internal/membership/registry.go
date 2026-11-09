package membership

import (
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Registry struct {
	mu      sync.RWMutex
	workers map[string]*Worker

	suspectAfter time.Duration
	deadAfter    time.Duration
}

func NewRegistry(suspectAfter, deadAfter time.Duration) *Registry {
	if suspectAfter <= 0 || deadAfter <= suspectAfter {
		panic("invalid failure detection thresholds")
	}

	return &Registry{
		workers:      make(map[string]*Worker),
		suspectAfter: suspectAfter,
		deadAfter:    deadAfter,
	}
}

// Register adds a new worker to the cluster.
func (r *Registry) Register(
	hostname string,
	ip string,
	cpuCores int32,
	memoryMB int64,
) string {
	workerID := uuid.NewString()

	worker := &Worker{
		ID:            workerID,
		Hostname:      hostname,
		IP:            ip,
		CPUCores:      cpuCores,
		MemoryMB:      memoryMB,
		Status:        WorkerAlive,
		LastHeartbeat: time.Now(),
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.workers[workerID] = worker

	return workerID
}

// Heartbeat updates the last-seen time of a worker.
func (r *Registry) Heartbeat(workerID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	worker, exists := r.workers[workerID]
	if !exists || worker.Status == WorkerDead {
		return false
	}

	worker.LastHeartbeat = time.Now()
	worker.Status = WorkerAlive

	return true
}

// Deregister removes a worker that leaves gracefully.
func (r *Registry) Deregister(workerID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.workers[workerID]
	if !exists {
		return false
	}

	delete(r.workers, workerID)

	return true
}

// DetectFailures identifies workers that stopped sending heartbeats.
func (r *Registry) DetectFailures() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()

	for _, worker := range r.workers {
		elapsed := now.Sub(worker.LastHeartbeat)

		if elapsed >= r.deadAfter {
			worker.Status = WorkerDead
		} else if elapsed >= r.suspectAfter {
			worker.Status = WorkerSuspect
		}
	}
}

// ListWorkers returns a snapshot of registered workers.
func (r *Registry) ListWorkers() []Worker {
	r.mu.RLock()
	defer r.mu.RUnlock()

	workers := make([]Worker, 0, len(r.workers))

	for _, worker := range r.workers {
		workers = append(workers, *worker)
	}

	sort.Slice(workers, func(i, j int) bool {
		return workers[i].ID < workers[j].ID
	})

	return workers
}
