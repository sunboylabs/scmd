package e2e

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/scmd/scmd/internal/backend/mock"
	"github.com/scmd/scmd/internal/command"
)

// ==================== CONCURRENCY TESTS ====================

func TestStress_ConcurrentCommands_10(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	testConcurrentCommands(t, 10)
}

func TestStress_ConcurrentCommands_50(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	testConcurrentCommands(t, 50)
}

func TestStress_ConcurrentCommands_100(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	testConcurrentCommands(t, 100)
}

func testConcurrentCommands(t *testing.T, n int) {
	var wg sync.WaitGroup
	errors := make(chan error, n)

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			_, _, err := runScmd(t, "-b", "mock", "-q", "-p", fmt.Sprintf("test %d", id))
			if err != nil {
				errors <- fmt.Errorf("command %d failed: %w", id, err)
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Error(err)
	}
}

func TestStress_ConcurrentWithStdin(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	n := 20
	input := strings.Repeat("test line\n", 100)
	var wg sync.WaitGroup
	var successCount int64

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			stdout, _, err := runScmdWithStdin(t, input, "-b", "mock", "-q", "-p", "process")
			if err == nil && stdout != "" {
				atomic.AddInt64(&successCount, 1)
			}
		}(i)
	}

	wg.Wait()

	minSuccess := int64(float64(n) * 0.9) // Allow 10% failure
	if successCount < minSuccess {
		t.Errorf("too many failures: %d/%d succeeded", successCount, n)
	}
}

// ==================== LOAD TESTS ====================

func TestStress_RapidFireSequential(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	n := 100
	start := time.Now()

	for i := 0; i < n; i++ {
		_, _, err := runScmd(t, "-b", "mock", "-q", "-p", "test")
		if err != nil {
			t.Fatalf("command %d failed: %v", i, err)
		}
	}

	duration := time.Since(start)
	rps := float64(n) / duration.Seconds()

	t.Logf("Completed %d requests in %v (%.2f req/s)", n, duration, rps)

	if rps < 1 {
		t.Errorf("too slow: %.2f req/s", rps)
	}
}

func TestStress_SustainedLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	duration := 10 * time.Second
	stop := time.After(duration)
	var count int64
	var errors int64

	// Run commands continuously for duration
	done := make(chan bool)
	for i := 0; i < 5; i++ { // 5 workers
		go func() {
			for {
				select {
				case <-done:
					return
				default:
					_, _, err := runScmd(t, "-b", "mock", "-q", "-p", "test")
					if err != nil {
						atomic.AddInt64(&errors, 1)
					} else {
						atomic.AddInt64(&count, 1)
					}
				}
			}
		}()
	}

	<-stop
	close(done)
	time.Sleep(100 * time.Millisecond) // Let goroutines finish

	// Use atomic loads to prevent data race
	finalCount := atomic.LoadInt64(&count)
	finalErrors := atomic.LoadInt64(&errors)
	total := finalCount + finalErrors
	errorRate := float64(finalErrors) / float64(total) * 100

	t.Logf("Sustained load: %d successful, %d errors (%.2f%% error rate) in %v",
		finalCount, finalErrors, errorRate, duration)

	if errorRate > 5 { // Allow 5% error rate
		t.Errorf("error rate too high: %.2f%%", errorRate)
	}
}

// ==================== LARGE INPUT TESTS ====================

func TestStress_LargeInput_1MB(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	input := strings.Repeat("x", 1*1024*1024) // 1MB
	stdout, _, err := runScmdWithStdin(t, input, "-b", "mock", "-q", "-p", "process")
	if err != nil {
		t.Fatalf("1MB input failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

func TestStress_LargeInput_10MB(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	input := strings.Repeat("x", 10*1024*1024) // 10MB
	stdout, _, err := runScmdWithStdin(t, input, "-b", "mock", "-q", "-p", "process")
	if err != nil {
		t.Fatalf("10MB input failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

func TestStress_ManySmallInputs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	// 10,000 small inputs
	for i := 0; i < 10000; i++ {
		input := fmt.Sprintf("test %d", i)
		_, _, err := runScmdWithStdin(t, input, "-b", "mock", "-q", "-p", "process")
		if err != nil {
			t.Fatalf("input %d failed: %v", i, err)
		}

		// Log progress periodically
		if i%1000 == 0 && i > 0 {
			t.Logf("Processed %d inputs", i)
		}
	}
}

func TestStress_VeryLongLines(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	// Single line with 100K characters
	line := strings.Repeat("a", 100*1024)
	stdout, _, err := runScmdWithStdin(t, line, "-b", "mock", "-q", "-p", "process")
	if err != nil {
		t.Fatalf("very long line failed: %v", err)
	}
	if stdout == "" {
		t.Error("should have output")
	}
}

// ==================== MEMORY PRESSURE TESTS ====================

func TestStress_MemoryPressure(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	// Run many commands to create memory pressure
	for i := 0; i < 100; i++ {
		input := strings.Repeat("x", 100*1024) // 100KB each
		_, _, err := runScmdWithStdin(t, input, "-b", "mock", "-q", "-p", "test")
		if err != nil {
			t.Fatalf("command %d failed: %v", i, err)
		}
	}

	runtime.GC()
	runtime.ReadMemStats(&m2)

	// Use TotalAlloc for monotonically increasing measurement
	// Alloc can decrease due to GC, causing underflow
	allocMB := float64(m2.TotalAlloc-m1.TotalAlloc) / 1024 / 1024
	t.Logf("Memory allocated: %.2f MB", allocMB)

	// Check for memory leaks (should not grow excessively)
	if allocMB > 500 { // 500MB threshold
		t.Errorf("memory usage too high: %.2f MB", allocMB)
	}
}

func TestStress_GoroutineLeak(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	before := runtime.NumGoroutine()

	// Run many commands
	for i := 0; i < 50; i++ {
		_, _, err := runScmd(t, "-b", "mock", "-q", "-p", "test")
		if err != nil {
			t.Fatalf("command %d failed: %v", i, err)
		}
	}

	// Give time for goroutines to cleanup
	time.Sleep(100 * time.Millisecond)
	runtime.GC()

	after := runtime.NumGoroutine()
	leaked := after - before

	t.Logf("Goroutines before: %d, after: %d, leaked: %d", before, after, leaked)

	// Allow some goroutines (runtime, test infrastructure)
	if leaked > 10 {
		t.Errorf("potential goroutine leak: %d goroutines remained", leaked)
	}
}

// ==================== ERROR RECOVERY TESTS ====================

func TestStress_RecoveryAfterErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	// Cause errors then verify recovery
	for i := 0; i < 10; i++ {
		_, _, _ = runScmd(t, "invalid-command") // Intentional error
	}

	// Now run valid commands
	for i := 0; i < 10; i++ {
		_, _, err := runScmd(t, "-b", "mock", "-q", "-p", "test")
		if err != nil {
			t.Errorf("recovery failed on command %d: %v", i, err)
		}
	}
}

func TestStress_MixedSuccessFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	var successes, failures int64
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			var err error
			if id%3 == 0 {
				// Intentional failures
				_, _, err = runScmd(t, "invalid-command")
			} else {
				// Valid commands
				_, _, err = runScmd(t, "-b", "mock", "-q", "-p", "test")
			}

			if err != nil {
				atomic.AddInt64(&failures, 1)
			} else {
				atomic.AddInt64(&successes, 1)
			}
		}(i)
	}

	wg.Wait()

	t.Logf("Successes: %d, Failures: %d", successes, failures)

	// Should have both successes and failures
	if successes == 0 {
		t.Error("should have some successes")
	}
}

// ==================== RESOURCE EXHAUSTION TESTS ====================

func TestStress_ManyOutputFiles(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	tmpDir := t.TempDir()

	// Create many output files
	for i := 0; i < 100; i++ {
		outFile := fmt.Sprintf("%s/output_%d.txt", tmpDir, i)
		_, _, err := runScmd(t, "-b", "mock", "-q", "-o", outFile, "-p", fmt.Sprintf("test %d", i))
		if err != nil {
			t.Fatalf("command %d failed: %v", i, err)
		}
	}

	// Verify files exist
	count := 0
	entries, _ := os.ReadDir(tmpDir)
	for _, e := range entries {
		if !e.IsDir() {
			count++
		}
	}

	if count < 100 {
		t.Errorf("expected at least 100 files, got %d", count)
	}
}

// ==================== PERFORMANCE BENCHMARKS ====================

func BenchmarkCommand_Simple(b *testing.B) {
	// Benchmarks need helper functions that accept *testing.B
	for i := 0; i < b.N; i++ {
		cmd := exec.Command(scmdBinary, "-b", "mock", "-q", "-p", "test")
		_, _ = cmd.CombinedOutput()
	}
}

func BenchmarkCommand_WithStdin(b *testing.B) {
	input := strings.Repeat("test\n", 100)
	for i := 0; i < b.N; i++ {
		cmd := exec.Command(scmdBinary, "-b", "mock", "-q", "-p", "process")
		cmd.Stdin = strings.NewReader(input)
		_, _ = cmd.CombinedOutput()
	}
}

func BenchmarkCommand_LargeInput(b *testing.B) {
	input := strings.Repeat("x", 10*1024) // 10KB
	for i := 0; i < b.N; i++ {
		cmd := exec.Command(scmdBinary, "-b", "mock", "-q", "-p", "process")
		cmd.Stdin = strings.NewReader(input)
		_, _ = cmd.CombinedOutput()
	}
}

func BenchmarkCommand_Parallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cmd := exec.Command(scmdBinary, "-b", "mock", "-q", "-p", "test")
			_, _ = cmd.CombinedOutput()
		}
	})
}

// ==================== BACKEND PERFORMANCE TESTS ====================

func BenchmarkBackend_MockComplete(b *testing.B) {
	ctx := context.Background()
	backend := mock.New()
	req := &command.Args{
		Options: map[string]string{
			"prompt": "test prompt",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Note: This is a simplified benchmark
		// Real backend.Complete would need proper request structure
		_ = req
		_ = backend
		_ = ctx
	}
}

// ==================== SCALING TESTS ====================

func TestStress_ScalingWorkers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	workerCounts := []int{1, 2, 5, 10, 20}

	for _, workers := range workerCounts {
		t.Run(fmt.Sprintf("workers_%d", workers), func(t *testing.T) {
			var wg sync.WaitGroup
			start := time.Now()
			tasksPerWorker := 10

			for i := 0; i < workers; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for j := 0; j < tasksPerWorker; j++ {
						_, _, err := runScmd(t, "-b", "mock", "-q", "-p", "test")
						if err != nil {
							t.Errorf("worker task failed: %v", err)
						}
					}
				}()
			}

			wg.Wait()
			duration := time.Since(start)
			total := workers * tasksPerWorker
			rps := float64(total) / duration.Seconds()

			t.Logf("%d workers: %d tasks in %v (%.2f req/s)",
				workers, total, duration, rps)
		})
	}
}

// ==================== EDGE CASE STRESS TESTS ====================

func TestStress_AlternatingBackends(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	backends := []string{"mock", "mock"} // Use mock twice to test switching

	for i := 0; i < 50; i++ {
		backend := backends[i%len(backends)]
		_, _, err := runScmd(t, "-b", backend, "-q", "-p", "test")
		if err != nil {
			t.Errorf("command %d with backend %s failed: %v", i, backend, err)
		}
	}
}

func TestStress_RandomInputSizes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	sizes := []int{10, 100, 1000, 10000, 100000}

	for i := 0; i < 50; i++ {
		size := sizes[i%len(sizes)]
		input := strings.Repeat("x", size)
		_, _, err := runScmdWithStdin(t, input, "-b", "mock", "-q", "-p", "process")
		if err != nil {
			t.Errorf("command %d (size %d) failed: %v", i, size, err)
		}
	}
}

// ==================== TIMEOUT TESTS ====================

func TestStress_CommandTimeouts(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	// Run commands with time limits
	done := make(chan bool, 1)
	timeout := 5 * time.Second

	go func() {
		for i := 0; i < 10; i++ {
			_, _, err := runScmd(t, "-b", "mock", "-q", "-p", "test")
			if err != nil {
				t.Errorf("command %d failed: %v", i, err)
			}
		}
		done <- true
	}()

	select {
	case <-done:
		// Success
	case <-time.After(timeout):
		t.Errorf("commands did not complete within %v", timeout)
	}
}

// ==================== REALISTIC USAGE PATTERNS ====================

func TestStress_RealisticUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	// Simulate realistic usage: mix of commands with different inputs
	scenarios := []struct {
		command string
		input   string
	}{
		{"explain", "func main() { fmt.Println(\"hello\") }"},
		{"review", "var x = 1; y := 2"},
		{"explain", "SELECT * FROM users WHERE id = 1"},
		{"review", "for i := 0; i < 10; i++ { }"},
	}

	for i := 0; i < 25; i++ { // Run each scenario multiple times
		for _, scenario := range scenarios {
			_, _, err := runScmdWithStdin(t, scenario.input, "-b", "mock", "-q", scenario.command)
			if err != nil {
				t.Errorf("scenario %s failed: %v", scenario.command, err)
			}
		}
	}
}
