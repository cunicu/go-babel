// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package proto

import (
	"net/netip"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("Parser", func() {
	var p *Parser
	BeforeEach(func() {
		p = &Parser{}
	})

	It("uint8", func() {
		var b []byte
		var i int
		var err error
		var n uint8

		vs := []uint8{0xaa, 0xbb, 0xcc}
		for i, v := range vs {
			b = p.appendUint8(b, v)
			Expect(b).To(HaveLen((i + 1) * 1))
		}

		for i = 0; len(b) > 0; i++ {
			b, n, err = p.uint8(b)
			Expect(err).To(Succeed())
			Expect(n).To(Equal(vs[i]))
		}

		Expect(b).To(BeEmpty())
	})

	It("uint16", func() {
		var b []byte
		var i int
		var err error
		var n uint16

		vs := []uint16{0xaaaa, 0xbbbb, 0xcccc}
		for i, v := range vs {
			b = p.appendUint16(b, v)
			Expect(b).To(HaveLen((i + 1) * 2))
		}

		for i = 0; len(b) > 0; i++ {
			b, n, err = p.uint16(b)
			Expect(err).To(Succeed())
			Expect(n).To(Equal(vs[i]))
		}

		Expect(b).To(BeEmpty())
	})

	It("uint64", func() {
		var b []byte
		var i int
		var err error
		var n uint64

		vs := []uint64{0xaaaaaaaa, 0xbbbbbbbb, 0xcccccccc}
		for i, v := range vs {
			b = p.appendUint64(b, v)
			Expect(b).To(HaveLen((i + 1) * 8))
		}

		for i = 0; len(b) > 0; i++ {
			b, n, err = p.uint64(b)
			Expect(err).To(Succeed())
			Expect(n).To(Equal(vs[i]))
		}

		Expect(b).To(BeEmpty())
	})

	It("Interval", func() {
		v1 := 12 * time.Second

		b := p.appendInterval(nil, v1)
		Expect(b).To(HaveLen(2))

		b, v2, err := p.interval(b)
		Expect(err).To(Succeed())
		Expect(v2).To(Equal(v1))
		Expect(b).To(BeEmpty())
	})

	It("RouterID", func() {
		v1 := RouterID(0x12345678)

		b := p.appendRouterID(nil, v1)
		Expect(b).To(HaveLen(8))

		b, v2, err := p.routerID(b)
		Expect(err).To(Succeed())
		Expect(v2).To(Equal(v1))
		Expect(b).To(BeEmpty())
	})

	DescribeTable("Address",
		func(addr string, len int, expAE uint8) {
			v1 := netip.MustParseAddr(addr)

			b, ae := p.appendAddress(nil, v1)
			Expect(b).To(HaveLen(len))
			Expect(ae).To(Equal(expAE))

			b, v2, err := p.address(b, ae)
			Expect(err).To(Succeed())
			Expect(v2).To(Equal(v1))
			Expect(b).To(BeEmpty())
		},
		Entry("AddressEncodingIPv4", "1.1.1.1", 4, AddressEncodingIPv4),
		Entry("AddressEncodingIPv6", "fd3d:bd4f:9738::1036:d55b:fb01:b6d1", 16, AddressEncodingIPv6),
		Entry("AddressEncodingWildcard", "::", 0, AddressEncodingWildcard),
		Entry("AddressEncodingIPv6LinkLocal", "fe80::1234:5678:90AB:CDEF", 8, AddressEncodingIPv6LinkLocal),
	)

	DescribeTable("Prefix",
		func(addr string, len int, expAE, expPlen uint8) {
			v1 := netip.MustParsePrefix(addr)

			b, ae, plen, _ := p.appendPrefix(nil, v1, false)
			Expect(b).To(HaveLen(len))
			Expect(ae).To(Equal(expAE))
			Expect(plen).To(Equal(expPlen))

			b, v2, err := p.prefix(b, ae, plen, 0)
			Expect(err).To(Succeed())
			Expect(v2).To(Equal(v1))
			Expect(b).To(BeEmpty())
		},
		Entry("AddressEncodingIPv4", "1.1.0.0/16", 4, AddressEncodingIPv4, uint8(16)),
		Entry("AddressEncodingIPv6", "fd3d:bd4f:9738::/48", 16, AddressEncodingIPv6, uint8(48)),
		Entry("AddressEncodingWildcard", "::/0", 0, AddressEncodingWildcard, uint8(0)),
		Entry("AddressEncodingIPv6LinkLocal", "fe80::1234:5678:90AB:CDEF/128", 8, AddressEncodingIPv6LinkLocal, uint8(128)),
	)

	pfx := netip.MustParsePrefix("10.0.0.0/24")

	It("Compressed Prefixes", Pending, func() {
		// TODO
	})

	DescribeTable("Values",
		func(typ1 ValueType, v1 Value) {
			b := p.AppendValue(nil, v1)
			Expect(b).To(HaveLen(int(p.ValueLength(v1))))

			p.Reset()

			_, v2, typ2, err := p.value(b)
			Expect(err).To(Succeed())
			Expect(typ2).To(Equal(typ1))
			Expect(v2).To(Equal(v1))
		},
		Entry("Pad1", TypePad1, &Pad1{}),
		Entry("PadN", TypePadN, &PadN{
			N: 111,
		}),
		Entry("AcknowledgmentRequest", TypeAcknowledgmentRequest, &AcknowledgmentRequest{
			Opaque:   0x1234,
			Interval: 4 * time.Second,
		}),
		Entry("Acknowledgment", TypeAcknowledgment, &Acknowledgment{
			Opaque: 0x1234,
		}),
		Entry("Hello", TypeHello, &Hello{
			Flags:    FlagHelloUnicast,
			Seqno:    1233,
			Interval: 33 * time.Second,
		}),
		Entry("Hello with Timestamp", TypeHello, &Hello{
			Flags:    FlagHelloUnicast,
			Seqno:    1233,
			Interval: 33 * time.Second,
			Timestamp: &TimestampHello{
				Transmit: 532235,
			},
		}),
		Entry("IHU", TypeIHU, &IHU{
			RxCost:   0xABCD,
			Interval: 2 * time.Second,
			Address:  netip.MustParseAddr("1.2.3.4"),
		}),
		Entry("IHU with Timestamp", TypeIHU, &IHU{
			RxCost:   0xABCD,
			Interval: 2 * time.Second,
			Address:  netip.MustParseAddr("1.2.3.4"),
			Timestamp: &TimestampIHU{
				Origin:  42394723,
				Receive: 23283423,
			},
		}),
		Entry("RouterID", TypeRouterID, &RouterIDValue{
			RouterID: 0xababcdcd,
		}),
		Entry("NextHop", TypeNextHop, &NextHop{
			NextHop: netip.MustParseAddr("1.2.3.4"),
		}),
		Entry("Update", TypeUpdate, &Update{
			Flags:    FlagUpdateRouterID,
			Interval: 2 * time.Second,
			Seqno:    1233,
			Metric:   100,
			Prefix:   netip.MustParsePrefix("192.168.0.0/16"),
		}),
		Entry("Update with SourcePrefix", TypeUpdate, &Update{
			Flags:        FlagUpdateRouterID,
			Interval:     2 * time.Second,
			Seqno:        1233,
			Metric:       100,
			Prefix:       netip.MustParsePrefix("192.168.0.0/16"),
			SourcePrefix: &pfx,
		}),
		Entry("RouteRequest", TypeRouteRequest, &RouteRequest{
			Prefix: netip.MustParsePrefix("192.168.0.0/16"),
		}),
		Entry("RouteRequest with SourcePrefix", TypeRouteRequest, &RouteRequest{
			Prefix:       netip.MustParsePrefix("192.168.0.0/16"),
			SourcePrefix: &pfx,
		}),
		Entry("SeqnoRequest", TypeSeqnoRequest, &SeqnoRequest{
			Seqno:    1233,
			HopCount: 99,
			RouterID: 0xababcdcd,
			Prefix:   netip.MustParsePrefix("192.168.0.0/16"),
		}),
		Entry("SeqnoRequest with SourcePrefix", TypeSeqnoRequest, &SeqnoRequest{
			Seqno:        1233,
			HopCount:     99,
			RouterID:     0xababcdcd,
			Prefix:       netip.MustParsePrefix("192.168.0.0/16"),
			SourcePrefix: &pfx,
		}),
	)

	It("Ignore unsupported sub-TLVs", Pending, func() {
		// TODO
	})

	It("Fail on mandatory sub-TLVs", Pending, func() {
		// TODO
	})

	It("Ignore unsupported TLVs", Pending, func() {
		// TODO
	})
})
