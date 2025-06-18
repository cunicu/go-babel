# SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
{ buildGoModule, libpcap }:
buildGoModule {
  name = "go-babel";
  src = ./.;
  vendorHash = "sha256-bhbLSnn3a4OTh2cMJyjiLzAM3tJwVpQCPoaF09iIkx8=";
  buildInputs = [ libpcap ];
  doCheck = false;
}
