// SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package proto

import (
	"log/slog"
	"net/netip"
)

// 4.6.10. Route Request
// https://datatracker.ietf.org/doc/html/rfc8966#section-4.6.10
type RouteRequest struct {
	Prefix netip.Prefix // The prefix being requested. This field's size is Plen/8 rounded upwards.

	// Sub-TLVs
	SourcePrefix *Prefix
}

func (r *RouteRequest) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.Any("pfx", r.Prefix),
	}

	if r.SourcePrefix != nil {
		attrs = append(attrs,
			slog.Any("src_pfx", *r.SourcePrefix))
	}

	return slog.GroupValue(attrs...)
}
