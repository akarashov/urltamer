package pool

import "sync"

// Resettable is an interface that types must implement to be used with the Pool.
type Resettable interface {
	Reset()
}

// Pool is a generic pool for any type that implements the Resettable interface.
type Pool[T Resettable] struct {
	pool sync.Pool
}

// New creates a new Pool for the given type T.
// The New provided function fn is used to create new instances of T when needed.
func New[T Resettable](fn func() T) *Pool[T] {
	return &Pool[T]{
		pool: sync.Pool{
			New: func() any {
				return fn()
			},
		},
	}
}

// Get retrieves an instance of T from the pool.
// If the pool is empty, it creates a new instance using the provided function.
func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

// Put returns an instance of T back to the pool after resetting it.
// The Reset() method is called on the instance before putting it back to ensure that it is in a clean state for the next user.
func (p *Pool[T]) Put(x T) {
	x.Reset()
	p.pool.Put(x)
}
