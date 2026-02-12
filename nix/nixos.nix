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

    users.users = lib.mkIf (config.services.hister.user == "hister") {
      hister = {
        description = "Hister web history service";
        group = config.services.hister.group;
        isSystemUser = true;
        home = config.services.hister.dataDir;
      };
    };

    users.groups = lib.mkIf (config.services.hister.group == "hister") {
      hister = { };
    };

    systemd.tmpfiles.rules = [
      "d '${config.services.hister.dataDir}' ${config.services.hister.dataDirMode} ${config.services.hister.user} ${config.services.hister.group} - -"
      "z '${config.services.hister.dataDir}' ${config.services.hister.dataDirMode} ${config.services.hister.user} ${config.services.hister.group} - -"
    ];

    systemd.services.hister = {
      description = "Hister web history service";
      after = [
        "network.target"
        "systemd-tmpfiles-setup.service"
      ];
      requires = [ "systemd-tmpfiles-setup.service" ];
      wantedBy = [ "multi-user.target" ];

      environment = {
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

      serviceConfig = {
        ExecStart = "${lib.getExe config.services.hister.package} listen";
        Restart = "on-failure";
        User = config.services.hister.user;
        Group = config.services.hister.group;
        WorkingDirectory = config.services.hister.dataDir;
      };
    };

    networking.firewall.allowedTCPPorts = lib.mkIf config.services.hister.enable [
      config.services.hister.port
    ];
  };
}
