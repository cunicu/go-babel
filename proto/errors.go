// SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package proto

import "errors"

var (
	ErrInvalidLength                = errors.New("invalid TLV length")
	ErrInvalidHeader                = errors.New("invalid TLV header")
	ErrInvalidMagic                 = errors.New("invalid packet magic")
	ErrUnsupportedVersion           = errors.New("unsupported version")
	ErrCompressionNotAllowed        = errors.New("address encoding does not allow compression")
	ErrInvalidAddress               = errors.New("invalid address")
	ErrMissingDefaultPrefix         = errors.New("received update TLV with non-zero omitted value but no previous default prefix")
	ErrInvalidRouterID              = errors.New("invalid router-id")
	ErrInvalidValueType             = errors.New("invalid value type")
	ErrTooShort                     = errors.New("buffer is too short")
	ErrTooLong                      = errors.New("buffer is too long")
	ErrUnsupportedValue             = errors.New("value is not supported")
	ErrUnsupportedButMandatoryValue = errors.New("value is not supported but mandatory")
	ErrInvalidValueForTrailer       = errors.New("value is not supported in packet trailer")
)
