{
  description = "Hister - Web history on steroids";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
    flake-parts.inputs.nixpkgs-lib.follows = "nixpkgs";
  };

  outputs =
    inputs:
    inputs.flake-parts.lib.mkFlake { inherit inputs; } {
      systems = [
        "x86_64-linux"
        "x86_64-darwin"
        "aarch64-linux"
        "aarch64-darwin"
      ];

      perSystem =
        {
          config,
          self',
          inputs',
          pkgs,
          system,
          ...
        }:
        {
          packages.default = pkgs.callPackage ./nix/package.nix { histerRev = inputs.self.rev or "unknown"; };
          packages.hister = pkgs.callPackage ./nix/package.nix { histerRev = inputs.self.rev or "unknown"; };

          devShells.default = pkgs.mkShell {
            packages = builtins.attrValues {
              inherit (pkgs) go gopls gotools;
            };
          };
        };

      flake = {
        nixosModules.default = inputs.self.nixosModules.hister;
        nixosModules.hister = import ./nix/nixos.nix;

        homeModules.default = inputs.self.homeModules.hister;
        homeModules.hister = import ./nix/home.nix;

        darwinModules.default = inputs.self.darwinModules.hister;
        darwinModules.hister = import ./nix/darwin.nix;
      };
    };
}
