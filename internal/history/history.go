// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

// Package history implements a history vector keep track of received Hellos
package history

import (
	"math/bits"

	"github.com/stv0g/go-babel/proto"
)

type HelloHistory struct {
	expectedSeqno proto.SequenceNumber
	vector        uint16
}

// Reset resets the history vector
func (h *HelloHistory) Reset() {
	h.vector = 0
}

func (h *HelloHistory) Empty() bool {
	return h.vector == 0
}

func (h *HelloHistory) Missed() (reset bool) {
	h.vector <<= 1
	h.expectedSeqno++

	return h.Empty()
}

// A.1. Maintaining Hello History
// See: https://datatracker.ietf.org/doc/html/rfc8966#section-a.1
func (h *HelloHistory) Update(seqno proto.SequenceNumber) (reset bool) {
	if proto.SeqnoAbsDistance(seqno, h.expectedSeqno) > 16 {
		reset = true
		h.Reset()
	} else if proto.SeqnoLess(seqno, h.expectedSeqno) {
		h.vector >>= h.expectedSeqno - seqno
	} else if proto.SeqnoLess(h.expectedSeqno, seqno) {
		h.vector <<= seqno - h.expectedSeqno
	}

	// Append a new bit
	h.vector <<= 1
	h.vector |= 1

	h.expectedSeqno = seqno + 1

	return
}

// A.2.1. k-out-of-j
// See: https://datatracker.ietf.org/doc/html/rfc8966#section-a.2.1
func (h *HelloHistory) OutOf(k, j int) bool {
	if bits.OnesCount16(h.vector<<(16-j)) >= int(k) {
		return true
	} else {
		return false
	}
}
