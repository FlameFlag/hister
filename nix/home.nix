{
  config,
  lib,
  pkgs,
  ...
}:
{
  imports = [
    ./options.nix
  ];

  config = lib.mkIf config.services.hister.enable {
    home.packages = [ config.services.hister.package ];

    services.hister.package = lib.mkDefault (pkgs.callPackage ./package.nix { });
    services.hister.dataDir = lib.mkDefault "${config.home.homeDirectory}/.local/share/hister";

    assertions = [
      {
        assertion = !(config.services.hister.configPath != null && config.services.hister.config != null);
        message = "Only one of services.hister.configPath and services.hister.config can be set";
      }
    ];

    home.file.".local/share/hister/.keep".text = "";

    systemd.user.services.hister = {
      Unit = {
        Description = "Hister web history service";
        After = [ "network.target" ];
      };

      Service = {
        ExecStart = "${lib.getExe config.services.hister.package} listen";
        Restart = "on-failure";
        WorkingDirectory = config.services.hister.dataDir;

        Environment = lib.mapAttrsToList (name: value: "${name}=${value}") (
          {
            HISTER_DATA_DIR = config.services.hister.dataDir;
            HISTER_PORT = builtins.toString config.services.hister.port;
          }
          // lib.optionalAttrs (config.services.hister.configPath != null) {
            HISTER_CONFIG = builtins.toString config.services.hister.configPath;
          }
          // lib.optionalAttrs (config.services.hister.config != null) {
            HISTER_CONFIG = "${(pkgs.formats.yaml { }).generate "hister-config.yml"
              config.services.hister.config
            }";
          }
        );
      };

      Install = {
        WantedBy = [ "default.target" ];
      };
    };

    launchd.agents.hister = {
      enable = true;
      config = {
        ProgramArguments = [
          (lib.getExe config.services.hister.package)
          "listen"
        ];
        KeepAlive = true;
        WorkingDirectory = config.services.hister.dataDir;

        EnvironmentVariables = {
          HISTER_DATA_DIR = config.services.hister.dataDir;
          HISTER_PORT = builtins.toString config.services.hister.port;
        }
        // lib.optionalAttrs (config.services.hister.configPath != null) {
          HISTER_CONFIG = builtins.toString config.services.hister.configPath;
        }
        // lib.optionalAttrs (config.services.hister.config != null) {
          HISTER_CONFIG = "${(pkgs.formats.yaml { }).generate "hister-config.yml"
            config.services.hister.config
          }";
        };
      };
    };
  };
}
