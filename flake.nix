{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
    flakelight.url = "github:nix-community/flakelight";
  };

  outputs = (
    { self, flakelight, ... }:

    let
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
          meta = defaultMeta // {
            description = "comd implements a Command Daemon that listens for HTTP requests to execute commands.";
            mainProgram = "comd";
          };
        };

      module =
        {
          isHome ? false,
        }:

        {
          system,
          config,
          lib,
          pkgs,
          ...
        }:

        let
          selfConfig = config.services.comd;

          serviceConfig = {
            ExecStart = ''
              ${lib.getExe selfConfig.package} \
                -c ${pkgs.writeText "comd.json" (builtins.toJSON selfConfig.config)} \
                -l ${lib.escapeShellArg selfConfig.listenAddr} \
                ${lib.escapeShellArgs selfConfig.extraFlags}
            '';
            Restart = "always";
            RestartSec = "5";
            NoNewPrivileges = true;
            RestrictSUIDSGID = true;
          };
        in

        with lib;
        with builtins;

        {
          options.services.comd = {
            enable = mkEnableOption "Enable comd";

            listenAddr = mkOption {
              description = "The HTTP address to listen on";
              type = types.str;
            };

            config = mkOption {
              description = "The configuration for comd";
              type = types.submodule {
                options = {
                  base_path = mkOption {
                    description = "The base path for the commands";
                    type = types.str;
                    default = "/";
                    example = "/commands";
                  };
                  execute = mkOption {
                    description = "The options to configure command execution";
                    type = types.submodule {
                      options = {
                        shell = mkOption {
                          description = "The shell to execute the command with";
                          type = types.listOf types.str;
                          default = lib.splitString " " "${pkgs.bash}/bin/bash -c %";
                        };
                        timeout = mkOption {
                          description = "The timeout for command execution";
                          type = types.str;
                          default = "5s";
                        };
                      };
                    };
                    default = { };
                  };
                  commands = mkOption {
                    description = "A map of command names to their commands";
                    type =
                      with types;
                      attrsOf (oneOf [
                        str
                        attrs
                        (listOf str)
                      ]);
                  };
                };
              };
            };

            path = mkOption {
              description = "The list of packages to be added to the PATH";
              type = types.listOf types.str;
              default = [ ];
            };

            extraFlags = mkOption {
              description = "Extra flags to pass to comd";
              type = types.listOf types.str;
              default = [ ];
            };

            package = mkOption {
              description = "The comd package to use";
              type = types.package;
              default = (pkgs.extend self.overlays.default).comd;
              # default = self.packages.${system}.default;
            };
          };

          config = mkIf selfConfig.enable (
            if isHome then
              {
                systemd.user.services.comd = {
                  Unit.Description = selfConfig.package.meta.description;
                  Install.WantedBy = [ "default.target" ];
                  Service = serviceConfig // {
                    Environment = "PATH=${lib.makeBinPath selfConfig.path}";
                  };
                };
              }
            else
              {
                systemd.services.comd = {
                  description = selfConfig.package.meta.description;
                  wantedBy = [ "multi-user.target" ];
                  path = selfConfig.path;
                  inherit serviceConfig;
                };
              }
          );
        };

      checks = {
        vm =
          pkgs:
          pkgs.testers.runNixOSTest {
            name = "comd-vm-check";
            nodes.machine =
              {
                config,
                pkgs,
                lib,
                ...
              }:
              {
                imports = [ self.nixosModules.default ];

                services.comd = {
                  enable = true;
                  listenAddr = ":8081";
                  config = {
                    base_path = "/commands";
                    commands."finish" = "touch /tmp/comd-finished";
                  };
                };

                environment.systemPackages = with pkgs; [ curl ];
              };
            testScript = ''
              import time

              machine.start()
              machine.wait_for_unit("comd.service")
              machine.wait_for_open_port(8081)
              time.sleep(0.5)
              machine.succeed("curl -X POST http://localhost:8081/commands/finish")
              machine.succeed("test -f /tmp/comd-finished")
            '';
          };
      };
    in

    flakelight ./. {
      inherit
        license
        devShell
        package
        checks
        ;
      nixosModule = module { isHome = false; };
      homeModule = module { isHome = true; };
    }
  );
}
