// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package proto

import (
	"net/netip"

	"golang.org/x/exp/slog"
)

// 4.6.8. Next Hop
// https://datatracker.ietf.org/doc/html/rfc8966#section-4.6.8
type NextHop struct {
	NextHop netip.Addr // The next-hop address advertised by subsequent Update TLVs for this address family.
}

func (n *NextHop) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("nh", n.NextHop))
}
