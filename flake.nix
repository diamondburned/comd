{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
    flakelight.url = "github:nix-community/flakelight";
  };

  outputs = (
    { flakelight, ... }:

    flakelight ./. {
      license = "ISC";

      devShell.packages =
        pkgs: with pkgs; [
          go
          gopls
          gotools
          nixfmt-rfc-style
        ];

      # package = 
    }
  );
}
