// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package proto

import (
	"log/slog"
	"net/netip"
)

// 4.6.9. Update
// https://datatracker.ietf.org/doc/html/rfc8966#section-4.6.9
type Update struct {
	Flags    uint8          // The individual bits of this field specify special handling of this TLV (see FlagUpdate*).
	Interval Interval       // An upper bound, expressed in centiseconds, on the time after which the sending node will send a new update for this prefix. This MUST NOT be 0. The receiving node will use this value to compute a hold time for the route table entry. The value FFFF hexadecimal (infinity) expresses that this announcement will not be repeated unless a request is received (Section 3.8.2.3).
	Seqno    SequenceNumber // The originator's sequence number for this update.
	Metric   Metric         // The sender's metric for this route. The value FFFF hexadecimal (infinity) means that this is a route retraction.
	Prefix   netip.Prefix   // The prefix being advertised. This field's size is (Plen/8 - Omitted) rounded upwards.

	// The following fields are not actually encoded in an update TLV.
	// Instead are derived from the parser state / preceding TLVs.
	RouterID RouterID // Taken from a previous RouterID TLV
	NextHop  Address  // Taken from a previous NextHop TLV

	// Sub-TLVs
	SourcePrefix *Prefix
}

func (u *Update) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.Any("flags", u.Flags),
		slog.Duration("intv", u.Interval),
		slog.Any("seqno", u.Seqno),
		slog.Any("metric", u.Metric),
		slog.Any("pfx", u.Prefix),
		slog.Any("rid", u.RouterID),
		slog.Any("nh", u.NextHop),
	}

	if u.SourcePrefix != nil {
		attrs = append(attrs,
			slog.Any("src_prefix", *u.SourcePrefix))
	}

	return slog.GroupValue(attrs...)
}
