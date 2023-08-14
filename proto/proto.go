// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package proto

import "log/slog"

const (
	PacketHeaderMagic   = 42
	PacketHeaderVersion = 2
	PacketHeaderLength  = 4
)

// 4.2. Packet Format
// https://datatracker.ietf.org/doc/html/rfc8966#section-4.2
type Packet struct {
	Body    []Value
	Trailer []Value
}

func (p *Packet) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("num_body", len(p.Body)),
		slog.Any("num_trailer", len(p.Trailer)))
}

func IsBabelPacket(buf []byte) bool {
	return len(buf) >= 1 && buf[0] == PacketHeaderMagic
}
