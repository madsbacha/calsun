{
  description = "CalSun - Sunrise/sunset calendar subscription service";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            gotools
            go-tools
          ];

          shellHook = ''
            echo "CalSun development environment"
            echo "Go version: $(go version)"
          '';
        };

        packages.default = pkgs.buildGoModule {
          pname = "calsun";
          version = "0.1.0";
          src = ./.;
          vendorHash = null; # Will be set after go.sum is generated
        };
      }
    );
}
