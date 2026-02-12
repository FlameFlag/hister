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
    environment.systemPackages = [ config.services.hister.package ];

    services.hister.package = lib.mkDefault (pkgs.callPackage ./package.nix { });

    assertions = [
      {
        assertion = !(config.services.hister.configPath != null && config.services.hister.config != null);
        message = "Only one of services.hister.configPath and services.hister.config can be set";
      }
    ];

    services.hister.user = lib.mkDefault config.system.primaryUser;
    services.hister.dataDir = lib.mkDefault "${config.system.primaryUserHome}/Library/Application Support/hister";

    system.activationScripts.extraActivation.text = ''
      ${lib.getExe' pkgs.coreutils "install"} -d -o ${lib.escapeShellArg config.services.hister.user} -g staff ${lib.escapeShellArg config.services.hister.dataDir}
    '';

    launchd.user.agents.hister = {
      serviceConfig = {
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
      managedBy = "services.hister.enable";
    };
  };
}
