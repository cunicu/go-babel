// SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package proto

import (
	"log/slog"
	"time"
)

// 4.6.3. Acknowledgment Request
// https://datatracker.ietf.org/doc/html/rfc8966#section-4.6.3
type AcknowledgmentRequest struct {
	Opaque   uint16        // An arbitrary value that will be echoed in the receiver's Acknowledgment TLV.
	Interval time.Duration // A time interval after which the sender will assume that this packet has been lost. This MUST NOT be 0. The receiver MUST send an Acknowledgment TLV before this time has elapsed (with a margin allowing for propagation time).
}

func (a *AcknowledgmentRequest) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("opaque", a.Opaque))
}

// 4.6.4. Acknowledgment
// https://datatracker.ietf.org/doc/html/rfc8966#section-4.6.4
type Acknowledgment struct {
	Opaque uint16 // Set to the Opaque value of the Acknowledgment Request that prompted this Acknowledgment.
}

func (a *Acknowledgment) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("opaque", a.Opaque))
}
