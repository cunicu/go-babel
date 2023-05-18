// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package proto

import (
	"net/netip"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func mustParsePrefixPointer(s string) *netip.Prefix {
	p := netip.MustParsePrefix(s)
	return &p
}

var _ = Context("Update", func() {
	rid := RouterID{0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef}

	It("equal updates are not less", func() {
		u := &Update{}
		Expect(u.Less(u)).To(BeFalse())
	})

	DescribeTable("Updates are correctly ordered",
		func(a, b *Update) {
			Expect(a.Less(b)).To(BeTrue())
			Expect(b.Less(a)).To(BeFalse())
		},
		Entry("router ID",
			&Update{
				RouterID: RouterIDUnspecified,
			},
			&Update{
				RouterID: RouterIDAllOnes,
			},
		),
		Entry("v4 mapped less",
			&Update{
				Prefix: netip.MustParsePrefix("::FFFF:1.0.0.1/128"),
			},
			&Update{
				Prefix: netip.MustParsePrefix("::FFFF:1.0.0.2/128"),
			},
		),
		Entry("has router ID",
			&Update{
				RouterID: rid,
				Prefix:   netip.MustParsePrefix("fe80::1234:5678:90ab:cdef/128"),
			},
			&Update{
				RouterID: rid,
				Prefix:   netip.MustParsePrefix("fe80::1/128"),
			},
		),
		Entry("has router ID but wrong prefix",
			&Update{
				RouterID: rid,
				Prefix:   netip.MustParsePrefix("fe80::1/128"),
			},
			&Update{
				RouterID: rid,
				Prefix:   netip.MustParsePrefix("fe80::1234:5678:90ab:cdef/127"),
			},
		),
		Entry("prefix len",
			&Update{
				Prefix: netip.MustParsePrefix("fe80::1/128"),
			},
			&Update{
				Prefix: netip.MustParsePrefix("fe80::1/127"),
			},
		),
		Entry("prefix",
			&Update{
				Prefix: netip.MustParsePrefix("fe80::1/128"),
			},
			&Update{
				Prefix: netip.MustParsePrefix("fe80::2/128"),
			},
		),
		Entry("source prefix len",
			&Update{
				SourcePrefix: mustParsePrefixPointer("fe80::1/128"),
			},
			&Update{
				SourcePrefix: mustParsePrefixPointer("fe80::1/127"),
			},
		),
		Entry("source prefix",
			&Update{
				SourcePrefix: mustParsePrefixPointer("fe80::1/128"),
			},
			&Update{
				SourcePrefix: mustParsePrefixPointer("fe80::2/128"),
			},
		),
	)
})
