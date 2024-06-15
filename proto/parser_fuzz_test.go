// SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package proto_test

import (
	"testing"

	"cunicu.li/go-babel/proto"
)

func FuzzParser(f *testing.F) {
	f.Fuzz(func(t *testing.T, b []byte) {
		p := proto.NewParser()
		p.Packet(b) //nolint:errcheck
	})
}
