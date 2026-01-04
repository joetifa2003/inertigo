package pool

import (
	"sync"
)

type Pool[T any] struct {
	syncPool  *sync.Pool
	beforePut func(p T)
	beforeGet func(p T)
}

type PoolBuilder[T any] interface {
	applyBuilder(pool *Pool[T])
}

type WithPoolBeforePut[T any] func(p T)

func (t WithPoolBeforePut[T]) applyBuilder(pool *Pool[T]) {
	pool.beforePut = t
}

type WithPoolBeforeGet[T any] func(p T)

func (t WithPoolBeforeGet[T]) applyBuilder(pool *Pool[T]) {
	pool.beforeGet = t
}

func NewPool[T any](newFunc func() T, builders ...PoolBuilder[T]) Pool[T] {
	syncPool := &sync.Pool{
		New: func() any {
			return newFunc()
		},
	}

	p := Pool[T]{
		syncPool: syncPool,
	}

	for _, b := range builders {
		b.applyBuilder(&p)
	}

	return p
}

func (p *Pool[T]) Get() T {
	v := p.syncPool.Get().(T)
	if p.beforeGet != nil {
		p.beforeGet(v)
	}
	return v
}

func (p *Pool[T]) Put(v T) {
	if p.beforePut != nil {
		p.beforePut(v)
	}
	p.syncPool.Put(v)
}
