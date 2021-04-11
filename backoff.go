/* The implementation is borrowed from https://github.com/lestrrat-go/backoff */

package raft

import (
	"math/rand"
	"time"
)

type Exponential struct {
	minInterval float64
	maxInterval float64
	current     float64
	multiplier  float64
}

var defaultMultiplier = 1.5

func init() {
	rand.Seed(time.Now().UnixNano())
}

func NewExponentialInterval(optMin, optMax, mult float64) *Exponential {
	if optMin > optMax {
		optMin = optMax
	}

	if mult <= 1 {
		mult = defaultMultiplier
	}

	return &Exponential{
		minInterval: optMin,
		maxInterval: optMax,
		multiplier:  mult,
	}
}

func (e *Exponential) Next() time.Duration {
	var next float64
	if e.current == 0 {
		next = e.minInterval
	} else {
		next = e.current * e.multiplier
	}

	if next > e.maxInterval {
		next = e.maxInterval
	}

	if next < e.minInterval {
		next = e.minInterval
	}
	diff := next - e.minInterval
	if diff <= 0 {
		diff = e.maxInterval - e.minInterval
	}
	next = float64(rand.Intn(int(diff)) + int(e.minInterval))
	e.current = next
	return time.Duration(next)

}
