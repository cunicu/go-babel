# SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
{ buildGoModule, libpcap }:
buildGoModule {
  name = "go-babel";
  src = ./.;
  vendorHash = "sha256-1ntfNzHLv46ni4wOLMoX1fSaZe+kfvIUzeJnrAnglZQ=";
  buildInputs = [ libpcap ];
  doCheck = false;
}
