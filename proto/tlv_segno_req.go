// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package proto

import (
	"net/netip"

	"golang.org/x/exp/slog"
)

// 4.6.11. Seqno Request
// https://datatracker.ietf.org/doc/html/rfc8966#section-4.6.11
type SeqnoRequest struct {
	Seqno    SequenceNumber // The sequence number that is being requested.
	HopCount uint8          // The maximum number of times that this TLV may be forwarded, plus 1. This MUST NOT be 0.
	RouterID RouterID       // The Router-Id that is being requested. This MUST NOT consist of all zeroes or all ones.
	Prefix   netip.Prefix   // The prefix being requested. This field's size is Plen/8 rounded upwards.

	// Sub-TLVs
	SourcePrefix *Prefix
}

func (s *SeqnoRequest) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.Any("seqno", s.Seqno),
		slog.Any("hopcnt", s.HopCount),
		slog.Any("rid", s.RouterID),
		slog.Any("pfx", s.Prefix),
	}

	if s.SourcePrefix != nil {
		attrs = append(attrs,
			slog.Any("src_pfx", *s.SourcePrefix))
	}

	return slog.GroupValue(attrs...)
}
