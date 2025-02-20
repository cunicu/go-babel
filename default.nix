# SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
{ buildGo123Module, libpcap }:
buildGo123Module {
  name = "go-babel";
  src = ./.;
  vendorHash = "sha256-IMEwQMhqMqOQ3+OM/Wp1qAIQYVoSuXLj4nqW5k6xGBY=";
  buildInputs = [ libpcap ];
  doCheck = false;
}
