package types

import "sync"

type AtomicNumeric struct {
	mu sync.Mutex
	v  interface{}
}

func (an *AtomicNumeric) Set(v interface{}) {
	an.mu.Lock()
	an.v = v
	an.mu.Unlock()
}

func (an *AtomicNumeric) Get() interface{} {
	an.mu.Lock()
	v := an.v
	an.mu.Unlock()
	return v
}
