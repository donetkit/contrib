package tracer

import (
	crand "crypto/rand"
	"encoding/binary"
	"math/rand"
	"time"
)

func newSeed() int64 {
	var seed int64
	if err := binary.Read(crand.Reader, binary.BigEndian, &seed); err != nil {
		// fallback to timestamp
		seed = time.Now().UnixNano()
	}
	return seed
}

var seed = newSeed()

func newGlobalRand() *rand.Rand {
	src := rand.NewSource(seed)
	if src64, ok := src.(rand.Source64); ok {
		return rand.New(src64)
	}
	return rand.New(src)
}
