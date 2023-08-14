// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package proto

import (
	"log/slog"
	"net/netip"
	"time"
)

// 4.6.6. IHU
// https://datatracker.ietf.org/doc/html/rfc8966#section-4.6.6
type IHU struct {
	RxCost   uint16        // The rxcost according to the sending node of the interface whose address is specified in the Address field. The value FFFF hexadecimal (infinity) indicates that this interface is unreachable.
	Interval time.Duration // An upper bound, on the time after which the sending node will send a new IHU; this MUST NOT be 0. The receiving node will use this value in order to compute a hold time for this symmetric association.
	Address  netip.Addr    // The address of the destination node, in the format specified by the AE field. Address compression is not allowed.

	// Sub-TLVs
	Timestamp *TimestampIHU
}

func (i *IHU) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.Any("rxcost", i.RxCost),
		slog.Duration("intv", i.Interval),
		slog.Any("addr", i.Address),
	}

	if i.Timestamp != nil {
		attrs = append(attrs,
			slog.Any("ts", *i.Timestamp))
	}

	return slog.GroupValue(attrs...)
}

// 3.2.  Timestamp sub-TLV in IHU TLVs
// https://datatracker.ietf.org/doc/html/draft-ietf-babel-rtt-extension-00#section-3.2
type TimestampIHU struct {
	Origin  Timestamp
	Receive Timestamp
}

func (t *TimestampIHU) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("orign", t.Origin),
		slog.Any("rx", t.Receive))
}
