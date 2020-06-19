package main

func NewLimiter(limit int) *Limiter {
	return &Limiter{
		guardChannel: make(chan struct{}, limit),
	}
}

type Limiter struct {
	guardChannel chan struct{}
}

func (l *Limiter) Acquire() {

	l.guardChannel <- struct{}{}
}

func (l *Limiter) Release() {
	<-l.guardChannel
}
