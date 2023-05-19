// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package proto

import (
	"time"

	"golang.org/x/exp/slog"
)

// 4.6.5. Hello
// https://datatracker.ietf.org/doc/html/rfc8966#section-4.6.5
type Hello struct {
	Flags    uint16         // The individual bits of this field specify special handling of this TLV (see FlagHello*).
	Seqno    SequenceNumber // If the Unicast flag is set, this is the value of the sending node's outgoing Unicast Hello seqno for this neighbour. Otherwise, it is the sending node's outgoing Multicast Hello seqno for this interface.
	Interval time.Duration  // If nonzero, this is an upper bound, on the time after which the sending node will send a new scheduled Hello TLV with the same setting of the Unicast flag. If this is 0, then this Hello represents an unscheduled Hello and doesn't carry any new information about times at which Hellos are sent.

	// Sub-TLVs
	Timestamp *TimestampHello
}

func (h *Hello) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.Any("flags", h.Flags),
		slog.Any("seqno", h.Seqno),
		slog.Any("intv", h.Interval),
	}

	if h.Timestamp != nil {
		attrs = append(attrs, slog.Any("timestamp", *h.Timestamp))
	}

	return slog.GroupValue(attrs...)
}

// 3.1.  Timestamp sub-TLV in Hello TLVs
// https://datatracker.ietf.org/doc/html/draft-ietf-babel-rtt-extension-00#section-3.1
type TimestampHello struct {
	Transmit Timestamp
}

func (t *TimestampHello) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("tx", t.Transmit))
}
