{
  inputs = {
    flake-utils.url = "github:numtide/flake-utils";
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs {
          inherit system;
        };
        otel-tui = pkgs.callPackage ./otel-tui.nix {
          src = self;
          version = "v0.7.0-dev";
        };
      in
      {
        defaultPackage = otel-tui;
        packages = {
          otel-tui = otel-tui;
        };
        devShells = {
          default = pkgs.mkShell {
            packages = [
              otel-tui
              pkgs.go
              pkgs.gopls
              pkgs.gotools
              pkgs.go-tools
            ];
          };
          otel-tui = pkgs.mkShell {
            packages = [ otel-tui ];
          };
        };
      }
    );
}
