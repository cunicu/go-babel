// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package proto_test

import (
	"testing"

	"github.com/stv0g/go-babel/proto"
)

func FuzzParser(f *testing.F) {
	f.Fuzz(func(t *testing.T, b []byte) {
		p := proto.NewParser()
		p.Packet(b) //nolint:errcheck
	})
}
