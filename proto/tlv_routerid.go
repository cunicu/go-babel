// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package proto

import (
	"encoding/hex"
	"log/slog"
)

// 4.6.7. Router-Id
// https://datatracker.ietf.org/doc/html/rfc8966#section-4.6.7
type RouterIDValue struct {
	RouterID RouterID // The router-id for routes advertised in subsequent Update TLVs. This MUST NOT consist of all zeroes or all ones.
}

func (r *RouterIDValue) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("rid", hex.EncodeToString(r.RouterID[:])))
}
