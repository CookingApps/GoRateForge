package job

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// NewsFetcherJob simulates fetching tech news
type NewsFetcherJob struct {
	id string
}

func NewNewsFetcherJob() *NewsFetcherJob {
	return &NewsFetcherJob{id: fmt.Sprintf("news-%d", time.Now().UnixNano())}
}

func (j *NewsFetcherJob) ID() string   { return j.id }
func (j *NewsFetcherJob) Type() string { return "news-fetcher" }

func (j *NewsFetcherJob) Execute(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Simulate work
	time.Sleep(time.Duration(rand.Intn(800)+200) * time.Millisecond)

	if rand.Float32() < 0.15 { // 15% chance of failure
		return fmt.Errorf("failed to fetch news for job %s", j.id)
	}

	fmt.Printf("✅ [Job %s] Tech news fetched successfully!\n", j.id)
	return nil
}

// PriceCheckerJob simulates checking crypto prices
type PriceCheckerJob struct {
	id string
}

func NewPriceChekerJob() *PriceCheckerJob {
	return &PriceCheckerJob{id: fmt.Sprintf("price-%d", time.Now().UnixNano())}
}

func (p *PriceCheckerJob) ID() string {
	return p.id
}

func (p *PriceCheckerJob) Type() string {
	return "price-checker"
}

func (j *PriceCheckerJob) Execute(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	time.Sleep(time.Duration(rand.Intn(600)+300) * time.Millisecond)

	if rand.Float32() < 0.1 {
		return fmt.Errorf("price check failed for %s", j.id)
	}

	fmt.Printf("💰 [Job %s] Crypto price checked!\n", j.id)
	return nil
}
