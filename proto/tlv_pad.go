// SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package proto

import "log/slog"

// 4.6.1. Pad1
// https://datatracker.ietf.org/doc/html/rfc8966#section-4.6.1
type Pad1 struct{}

// 4.6.2. PadN
// https://datatracker.ietf.org/doc/html/rfc8966#section-4.6.2
type PadN struct {
	N int
}

func (p *PadN) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("n", p.N))
}
