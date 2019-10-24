// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package ratewr

import (
	"context"
	"errors"
	"sync"
	"time"
)

// Limiter defines interface for limiter objects for wrapper
type Limiter interface {
	Allow() bool
	Wait(ctx context.Context) (err error)
}

// NaiveBucket is a naive implementation of a Limiter. It's based on
// https://www.calhoun.io/rate-limiting-api-calls-in-go/
type NaiveBucket struct {
	Capacity     int
	DripInterval time.Duration
	PerDrip      int
	consumed     int
	started      bool
	kill         chan bool
	m            sync.Mutex
}

// Start starts a naive bucket
func (b *NaiveBucket) Start() error {
	if b.started {
		return errors.New("bucket was already started")
	}
	ticker := time.NewTicker(b.DripInterval)
	b.started = true
	b.kill = make(chan bool, 1)

	go func() {
		for {
			select {
			case <-ticker.C:
				b.m.Lock()
				b.consumed -= b.PerDrip
				if b.consumed < 0 {
					b.consumed = 0
				}
				b.m.Unlock()
			case <-b.kill:
				return
			}
		}
	}()
	return nil
}

// Shutdown bucket
func (b *NaiveBucket) Shutdown() error {
	if !b.started {
		return errors.New("bucket was never started")
	}
	b.kill <- true
	return nil
}

// Consume consumes a token from the bucket
func (b *NaiveBucket) Consume(amt int) error {
	b.m.Lock()
	defer b.m.Unlock()

	if b.Capacity-b.consumed < amt {
		return errors.New("not enough capacity")
	}
	b.consumed += amt
	return nil
}

// Allow implements Limiter interface
func (b *NaiveBucket) Allow() bool {
	err := b.Consume(1)
	return (err == nil)
}

// Wait implements Limiter inferface
func (b *NaiveBucket) Wait(ctx context.Context) (err error) {
	//don't waits, its naive ;)
	return b.Consume(1)
}
