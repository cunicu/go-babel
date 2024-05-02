# SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
{
  description = "A implementation of the Babel routing protocol in Go";

  inputs = {
    flake-utils.url = "github:numtide/flake-utils";
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
  };

  outputs = {
    self,
    flake-utils,
    nixpkgs,
  }:
    flake-utils.lib.eachDefaultSystem
    (
      system: let
        pkgs = nixpkgs.legacyPackages.${system};
      in rec {
        packages = rec {
          default = go-babel;
          go-babel = pkgs.buildGoModule {
            name = "go-babel";
            src = ./.;
            vendorHash = "sha256-qp1wvKvHodeY0OO4hODQbqo+29RtO++Z4YgitY46b6o=";
            buildInputs = with pkgs; [
              libpcap
            ];
            doCheck = false;
          };
        };

        devShell = let
          ginkgo =
            pkgs.runCommand "ginkgo" {
              HOME = "/build";
              GOPATH = "/build";
              GO111MODULE = "off";
            } ''
              ln -s ${packages.go-babel.goModules} /build/src
              ${pkgs.go}/bin/go build -o $out/bin/ginkgo github.com/onsi/ginkgo/v2/ginkgo
            '';
        in
          pkgs.mkShell {
            packages = with pkgs; [
              golangci-lint
              reuse
              ginkgo
            ];

            inputsFrom = [
              packages.go-babel
            ];
          };

        formatter = nixpkgs.nixfmt-rfc-style;
      }
    );
}
