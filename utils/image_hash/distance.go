package image_hash

import (
	"math/bits"
)

func popcnt(x uint64) int { return bits.OnesCount64(x) }
