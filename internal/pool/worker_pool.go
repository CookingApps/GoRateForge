package pool

import (
	"context"
	"sync"

	"github.com/CookingApps/GoRateForge/internal/job"
	"github.com/CookingApps/GoRateForge/internal/limiter"
)

type PoolStats struct {
	Submitted int
	Succeeded int
	Failed    int
	Active    int
}

type WorkerPool struct {
	numWorkers int                  // how many worker goroutines to run
	jobChan    chan job.Job         // buffered channel for submitting jobs
	limiter    *limiter.RateLimiter // our rate limiter
	wg         sync.WaitGroup       // to wait for all workers to finish
	mu         sync.Mutex           // protect stats and shutdown state
	stats      PoolStats            // we'll define this struct later
	ctx        context.Context      // for graceful shutdown
	cancel     context.CancelFunc   // to trigger shutdown
	isShutdown bool
}

func NewWorkerPool(numWorkers int, limiter *limiter.RateLimiter) *WorkerPool {

	return &WorkerPool{
		numWorkers: numWorkers,
		limiter:    limiter,
		jobChan:    make(chan job.Job, 200), // buffered channel
		stats:      PoolStats{},
		isShutdown: false,
		// wg and mu will be zero value, which is fine for now
	}
}

// Start begins all the worker goroutines
func (wp *WorkerPool) Start(ctx context.Context) {
	wp.mu.Lock()
	wp.ctx, wp.cancel = context.WithCancel(ctx)
	wp.mu.Unlock()

	for i := 0; i < wp.numWorkers; i++ {
		wp.wg.Add(1)
		go func() {
			defer wp.wg.Done()

			for {
				select {
				case j, ok := <-wp.jobChan:
					if !ok {
						return // channel closed
					}

					// Safely increment active count
					wp.mu.Lock()
					wp.stats.Active++
					wp.mu.Unlock()

					// Respect rate limiting
					if err := wp.limiter.Wait(wp.ctx); err != nil {
						wp.mu.Lock()
						wp.stats.Active--
						wp.mu.Unlock()
						return
					}

					// Execute the job
					err := j.Execute(wp.ctx)

					// Update stats safely
					wp.mu.Lock()
					wp.stats.Active--
					if err != nil {
						wp.stats.Failed++
					} else {
						wp.stats.Succeeded++
					}
					wp.mu.Unlock()

				case <-wp.ctx.Done():
					return
				}
			}
		}()
	}
}

// Submit adds a job to the pool
func (wp *WorkerPool) Submit(j job.Job) bool {
	wp.mu.Lock()
	if wp.isShutdown {
		wp.mu.Unlock()
		return false
	}
	wp.mu.Unlock()

	select {
	case wp.jobChan <- j:
		wp.mu.Lock()
		wp.stats.Submitted++
		wp.mu.Unlock()
		return true
	default:
		// channel is full (backpressure)
		return false
	}
}

// Shutdown gracefully stops the worker pool
func (wp *WorkerPool) Shutdown() {
	wp.mu.Lock()
	if wp.isShutdown {
		wp.mu.Unlock()
		return
	}
	wp.isShutdown = true
	wp.mu.Unlock()

	if wp.cancel != nil {
		wp.cancel()
	}

	close(wp.jobChan) // stop workers from receiving new jobs
	wp.wg.Wait()      // wait for all workers to finish
}

// Stats returns a copy of current statistics
func (wp *WorkerPool) Stats() PoolStats {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	return wp.stats
}
