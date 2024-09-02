# SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
{ buildGoModule, libpcap }:
buildGoModule {
  name = "go-babel";
  src = ./.;
  vendorHash = "sha256-olVuy+VTHPN2QwAhhhTWclZw4tx+rSppZCQgylJSoTU";
  buildInputs = [ libpcap ];
  doCheck = false;
}
