package shardval

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// Value is a sharded value implemented using a [sync.Pool] for per-P locality.
// It is used to represent a value that's sharded into many pieces
// which are updated in `With`, and then assembled in `ForEach`.
//
// The number of shards should scale with O(GOMAXPROCS), though there is no hard limit.
type Value[T any] struct {
	pool sync.Pool

	all  []*T
	free chan *shard
}

// shard is an indirection used to detects elements dropped from the pool (via finalizers).
type shard struct {
	index int
}

func NewFixed[T any](size int) *Value[T] {
	v := &Value[T]{
		all:  make([]*T, size),
		free: make(chan *shard, size),
	}
	for i := range v.all {
		v.all[i] = new(T)
		s := &shard{i}
		runtime.SetFinalizer(s, v.finalize)
		v.free <- s
	}
	return v
}

// With runs the given function with exclusive access to the shard.
//
// Many callers may call With concurrently.
//
// If called concurrently with [ForEach], the same value may be processed by both
// functions so addititional synchronization within T is required.
func (val *Value[T]) With(fn func(*T)) {
	s := val.get()
	defer val.pool.Put(s)
	defer fmt.Println("put", s.index)

	fn(val.all[s.index])
}

// ForEach iterates over all shards.
//
// If called concurretnly with [With], the same value may be processed by both functions
// so addititional synchronization within T is required.
func (val *Value[T]) ForEach(fn func(*T)) {
	for i := range val.all {
		fn(val.all[i])
	}
}

// Shards returns the number of shards.
func (val *Value[T]) Shards() int {
	return len(val.all)
}

func (val *Value[T]) get() *shard {
	// Keep trying till we get a free one..
	for {
		s, ok := val.pool.Get().(*shard)
		if ok {
			fmt.Println("from pool", s.index)
			return s
		}

		select {
		case s := <-val.free:
			fmt.Println("from free", s.index)
			return s
		default:
			// fmt.Println("no free, retry")
			// try in a bit
			time.Sleep(time.Millisecond)
		}
	}
}

func (val *Value[T]) finalize(s *shard) {
	fmt.Println("finalizer", s.index)
	select {
	case val.free <- s:
		runtime.SetFinalizer(s, val.finalize)
	default:
		panic("finalize culdn't put free chan")
	}
}
