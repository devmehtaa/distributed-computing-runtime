package membership

import "time"

type WorkerStatus string

const (
	WorkerAlive   WorkerStatus = "ALIVE"
	WorkerSuspect WorkerStatus = "SUSPECT"
	WorkerDead    WorkerStatus = "DEAD"
)

type Worker struct {
	ID       string
	Hostname string
	IP       string

	CPUCores int32
	MemoryMB int64

	Status        WorkerStatus
	LastHeartbeat time.Time
}
