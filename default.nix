# SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
{ buildGo123Module, libpcap }:
buildGo123Module {
  name = "go-babel";
  src = ./.;
  vendorHash = "sha256-f/afNk6l8NbdwjLxVUOzmp6TFzv9AsVjEEW9RcLVmEY=";
  buildInputs = [ libpcap ];
  doCheck = false;
}
