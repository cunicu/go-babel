// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package net

import (
	"io"
	"net"
)

type PacketConnWriter struct {
	net.PacketConn
	Dest net.Addr
}

var _ = (io.Writer)(&PacketConnWriter{})

func (w *PacketConnWriter) Write(p []byte) (int, error) {
	return w.WriteTo(p, w.Dest)
}
