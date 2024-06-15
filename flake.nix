# SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
{
  description = "A implementation of the Babel routing protocol in Go";

  inputs = {
    flake-utils.url = "github:numtide/flake-utils";
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
  };

  outputs =
    {
      self,
      flake-utils,
      nixpkgs,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs {
          inherit system;
          inherit overlays;
        };

        overlay = final: prev: { go-babel = final.callPackage ./default.nix { }; };
        overlays = [ overlay ];
      in
      {
        inherit overlays;

        packages = {
          default = pkgs.go-babel;
        };

        devShell =
          let
            ginkgo =
              pkgs.runCommand "ginkgo"
                {
                  HOME = "/build";
                  GOPATH = "/build";
                  GO111MODULE = "off";
                }
                ''
                  ln -s ${pkgs.go-babel.goModules} /build/src
                  ${pkgs.go}/bin/go build -o $out/bin/ginkgo github.com/onsi/ginkgo/v2/ginkgo
                '';
          in
          pkgs.mkShell {
            packages = with pkgs; [
              golangci-lint
              reuse
              ginkgo
            ];

            inputsFrom = with pkgs; [ go-babel ];

            hardeningDisable = [ "fortify" ];
          };

        formatter = pkgs.nixfmt-rfc-style;
      }
    );
}
