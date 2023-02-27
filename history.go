// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package babel

import (
	"math/bits"

	"github.com/stv0g/go-babel/proto"
)

type HistoryVector uint64

func distance(a, b proto.SequenceNumber) uint16 {
	if a > b {
		return a - b
	} else {
		return b - a
	}
}

// A.1. Maintaining Hello History
// See: https://datatracker.ietf.org/doc/html/rfc8966#section-a.1
func (v *HistoryVector) Update(nr, ne proto.SequenceNumber) {
	if distance(nr, ne) > 64 {
		*v = 0
	} else if nr < ne {
		*v >>= ne - nr
	} else if nr > ne {
		*v <<= nr - ne
	}

	*v <<= 1
	*v |= 1
}

// A.2.1. k-out-of-j
// See: https://datatracker.ietf.org/doc/html/rfc8966#section-a.2.1
func (v *HistoryVector) OutOf(k, j, C uint16) uint16 {
	if bits.OnesCount64(uint64(*v)&((1<<j)-1)) >= int(k) {
		return C
	} else {
		return 0
	}
}
