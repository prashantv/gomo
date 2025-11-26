package shardval

import (
	"math/rand"
	"runtime"
	"sync"
	"testing"
)

func TestShardVal_With(t *testing.T) {
	const goroutines = 5

	var total int64
	goIncrements := make([]int, goroutines)
	for i := range goIncrements {
		goIncrements[i] = rand.Intn(2000)
		total += int64(goIncrements[i])
	}

	sharded := NewFixed[int64](3)
	// var sharded Value[int64]

	var wg sync.WaitGroup
	for _, increments := range goIncrements {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for range increments {
				sharded.With(func(i *int64) {
					*i++
				})
				runtime.GC()
			}
		}()
	}

	wg.Wait()

	var count int64
	sharded.ForEach(func(i *int64) {
		count += *i
	})
	if count != total {
		t.Fatalf("mismatch, want %v, got %v", total, count)
	}

	shards := sharded.Shards()
	goMaxProcs := runtime.GOMAXPROCS(0)
	t.Logf("%d shards with %d GOMAXPROCS, ratio = %.2f", shards, goMaxProcs, float64(shards)/float64(goMaxProcs))
}

func TestShardVal_With_Unique(t *testing.T) {
	// Verify concurrent calls to With result in exclusive accesses to the shard.
	const (
		goroutines      = 100
		perGoIncrements = 100
	)

	var sharded Value[sync.Mutex]

	var wg sync.WaitGroup
	for range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for range perGoIncrements {
				sharded.With(func(mu *sync.Mutex) {
					if !mu.TryLock() {
						t.Error("concurrent use of shard in With")
					}
					mu.Unlock()
				})
				runtime.GC()
			}
		}()
	}

	wg.Wait()
}
