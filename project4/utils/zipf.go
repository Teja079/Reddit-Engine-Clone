package utils

import (
	"math/rand"
	"time"
)

func NewZipfGenerator(s, v float64, imax uint64) *rand.Zipf {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return rand.NewZipf(r, s, v, imax)
}
