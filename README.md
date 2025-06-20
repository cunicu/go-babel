<div align="center">
    <img width="300" src="docs/go_babel_logo.svg" >

[![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/cunicu/go-babel/test.yaml?style=flat-square)](https://github.com/cunicu/go-babel/actions)
[![goreportcard](https://goreportcard.com/badge/github.com/cunicu/go-babel?style=flat-square)](https://goreportcard.com/report/github.com/cunicu/go-babel)
[![Codecov branch](https://img.shields.io/codecov/c/github/cunicu/go-babel/main?style=flat-square&token=6XoWouQg6K)](https://app.codecov.io/gh/cunicu/go-babel/tree/main)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue?style=flat-square)](https://github.com/cunicu/go-babel/blob/main/LICENSES/Apache-2.0.txt)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/cunicu/go-babel?style=flat-square)
[![Go Reference](https://pkg.go.dev/badge/github.com/cunicu/go-babel.svg)](https://pkg.go.dev/github.com/cunicu/go-babel)
</div>

go-babel is an implementation of the [Babel routing protocol](https://www.irif.fr/~jch/software/babel/) in the Go programming language.

## RFCs / Status

go-babel aims at implementing the following RFCs and drafts:

### Under implementation

- [**RFC 8966:** The Babel Routing Protocol](https://datatracker.ietf.org/doc/html/rfc8966)

### Planned

- [**RFC 9229:** IPv4 Routes with an IPv6 Next Hop in the Babel Routing Protocol](https://datatracker.ietf.org/doc/rfc9079/)
- [**RFC 9079:** Source-Specific Routing in the Babel Routing Protocol](https://datatracker.ietf.org/doc/rfc9079/)
- [**RFC 8967:** MAC Authentication for the Babel Routing Protocol](https://datatracker.ietf.org/doc/rfc8967/)
- [**RFC 8968:** Babel Routing Protocol over Datagram Transport Layer Security](https://datatracker.ietf.org/doc/rfc8968/)
- [**RFC 9467:** Relaxed Packet Counter Verification for Babel MAC Authentication](https://datatracker.ietf.org/doc/rfc9467/)
- [**RFC 9616:** Delay-based Metric Extension for the Babel Routing Protocol](https://datatracker.ietf.org/doc/html/rfc9616/)

## Limitations

- Link cost calculation is only supported for wired links using the 2-out-of-3 strategy. ETX is not (yet) supported.

## References

- <https://www.irif.fr/~jch/software/babel/>
- <https://www.youtube.com/watch?v=Mflw4BuksHQ>
- <https://www.youtube.com/watch?v=1zMDLVln3XM>

## Contact

Please have a look at the contact page: [cunicu.li/docs/contact](https://cunicu.li/docs/contact).

## License

go-babel is licensed under the [Apache 2.0](./LICENSE) license.

- SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
- SPDX-License-Identifier: Apache-2.0
