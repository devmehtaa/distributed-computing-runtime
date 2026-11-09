package membership

import (
	"testing"
	"time"
)

func TestRegisterWorker(t *testing.T) {
	registry := NewRegistry(5*time.Second, 10*time.Second)

	id := registry.Register("machine-a", "127.0.0.1", 8, 16384)

	workers := registry.ListWorkers()

	if len(workers) != 1 {
		t.Fatalf("expected 1 worker, got %d", len(workers))
	}

	if workers[0].ID != id {
		t.Fatal("worker ID does not match")
	}

	if workers[0].Status != WorkerAlive {
		t.Fatal("new worker should be ALIVE")
	}
}

func TestHeartbeat(t *testing.T) {
	registry := NewRegistry(5*time.Second, 10*time.Second)

	id := registry.Register("machine-a", "127.0.0.1", 8, 16384)

	// A registered worker should successfully send a heartbeat.
	if !registry.Heartbeat(id) {
		t.Fatal("heartbeat should succeed")
	}

	// An unknown worker should be rejected.
	if registry.Heartbeat("invalid-id") {
		t.Fatal("heartbeat for unknown worker should fail")
	}
}

func TestDeregisterWorker(t *testing.T) {
	registry := NewRegistry(5*time.Second, 10*time.Second)

	id := registry.Register("machine-a", "127.0.0.1", 8, 16384)

	if !registry.Deregister(id) {
		t.Fatal("deregistration should succeed")
	}

	if len(registry.ListWorkers()) != 0 {
		t.Fatal("registry should be empty")
	}
}

func TestFailureDetection(t *testing.T) {
	registry := NewRegistry(20*time.Millisecond, 50*time.Millisecond)

	id := registry.Register("machine-a", "127.0.0.1", 8, 16384)

	time.Sleep(30 * time.Millisecond)
	registry.DetectFailures()

	workers := registry.ListWorkers()

	if workers[0].ID != id || workers[0].Status != WorkerSuspect {
		t.Fatal("worker should be SUSPECT")
	}

	time.Sleep(30 * time.Millisecond)
	registry.DetectFailures()

	workers = registry.ListWorkers()

	if workers[0].Status != WorkerDead {
		t.Fatal("worker should be DEAD")
	}
}
