package proto_test

import (
	"testing"

	"github.com/stv0g/go-babel/proto"
)

func FuzzParser(f *testing.F) {
	f.Fuzz(func(t *testing.T, b []byte) {
		p := proto.Parser{}
		p.Packet(b)
	})
}
