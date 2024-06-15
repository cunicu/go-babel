// SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package proto

import "bytes"

// Less compares two update values and returns true if the passed update
// is larger.
// The comparison is implemented in a way such it reduces the size of the
// route update packets when compressed as defined in:
// https://www.rfc-editor.org/rfc/rfc8966.html#section-4.6.9
func (a *Update) Less(b *Update) bool {
	// Router ID
	if rc := bytes.Compare(a.RouterID[:], b.RouterID[:]); rc != 0 {
		return rc < 0
	}

	// IPv4 mapped in IPv6
	aIs4in6 := a.Prefix.Bits() >= 96 && a.Prefix.Addr().Is4In6()
	bIs4in6 := b.Prefix.Bits() >= 96 && b.Prefix.Addr().Is4In6()

	if aIs4in6 != bIs4in6 {
		return bIs4in6
	}

	// Has router ID in prefix
	aHasRID := !aIs4in6 && a.Prefix.Bits() == 128 && bytes.Equal(a.Prefix.Addr().AsSlice()[8:], a.RouterID[:])
	bHasRID := !bIs4in6 && b.Prefix.Bits() == 128 && bytes.Equal(b.Prefix.Addr().AsSlice()[8:], b.RouterID[:])

	if aHasRID != bHasRID {
		return aHasRID
	}

	// Prefix length
	if rc := a.Prefix.Bits() - b.Prefix.Bits(); rc != 0 {
		return rc > 0
	}

	// Prefix
	if rc := a.Prefix.Addr().Compare(b.Prefix.Addr()); rc != 0 {
		return rc < 0
	}

	// Has source prefix
	aHasSrcPfx := a.SourcePrefix != nil
	bHasSrcPfx := b.SourcePrefix != nil

	if (aHasSrcPfx != bHasSrcPfx) || (!aHasSrcPfx && !bHasSrcPfx) {
		return aHasSrcPfx
	}

	// Source prefix length
	if rc := a.SourcePrefix.Bits() - b.SourcePrefix.Bits(); rc != 0 {
		return rc > 0
	}

	// Source prefix
	if rc := a.SourcePrefix.Addr().Compare(b.SourcePrefix.Addr()); rc != 0 {
		return rc < 0
	}

	return false
}
