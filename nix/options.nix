{ lib, pkgs, ... }:
{
  options.services.hister = {
    enable = lib.mkEnableOption "Hister web history service";

    package = lib.mkPackageOption pkgs "hister" { };

    dataDir = lib.mkOption {
      type = lib.types.path;
      default = "/var/lib/hister";
      description = "Directory where Hister stores its data.";
    };

    dataDirMode = lib.mkOption {
      type = lib.types.str;
      default = "0755";
      description = "Permissions mode for the data directory.";
    };

    port = lib.mkOption {
      type = lib.types.port;
      default = 4433;
      description = "Port on which Hister listens.";
    };

    configPath = lib.mkOption {
      type = lib.types.nullOr lib.types.path;
      default = null;
      description = "Path to an existing configuration file.";
    };

    config = lib.mkOption {
      type = with lib.types; nullOr attrs;
      default = null;
      description = "Configuration as a Nix attribute set. This will be converted to a YAML file.";
      example = {
        app = {
          directory = "~/.config/hister/";
          search_url = "https://google.com/search?q={query}";
        };
        server = {
          address = "127.0.0.1:4433";
        };
      };
    };

    user = lib.mkOption {
      type = lib.types.str;
      default = "hister";
      description = "User account under which Hister runs.";
    };

    group = lib.mkOption {
      type = lib.types.str;
      default = "hister";
      description = "Group under which Hister runs.";
    };
  };
}
