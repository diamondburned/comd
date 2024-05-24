{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
    flakelight.url = "github:nix-community/flakelight";
  };

  outputs = (
    { self, flakelight, ... }:

    flakelight ./. {
      license = "ISC";

      devShell.packages =
        pkgs: with pkgs; [
          go
          gopls
          gotools
          nixfmt-rfc-style
        ];

      package =
        { buildGoModule, defaultMeta }:

        buildGoModule {
          pname = "comd";
          version = self.rev or "unknown";
          vendorHash = "sha256-q90/jmMctvkKraw40BLkgaTF7YQkj8xbWlALNB73Cdw=";
          src = ./.;
          meta = defaultMeta;
        };
    }
  );
}
