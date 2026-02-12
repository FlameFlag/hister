# Hister

**Web history on steroids**

Hister is a web history management tool that provides blazing fast, content-based search for visited websites. Unlike traditional browser history that only searches URLs and titles, Hister indexes the full content of web pages you visit, enabling deep and meaningful search across your browsing history.

![hister screenshot](assets/screenshot.png)

![hister screencast](assets/demo.gif)


## Features

- **Privacy-focused**: Keep your browsing history indexed locally - don't use remote search engines if it isn't necessary
- **Full-text indexing**: Search through the actual content of web pages you've visited
- **Advanced search capabilities**: Utilize a powerful [query language](https://blevesearch.com/docs/Query-String-Query/) for precise results
- **Efficient retrieval**: Use keyword aliases to quickly find content
- **Flexible content management**: Configure blacklist and priority rules for better control

### Command line actions
- **Index**: download and index any URL
- **Import**: import and index existing Firefox/Chrome browser history

## Setup & run

### Install the extension

Available for [Chrome](https://chromewebstore.google.com/detail/hister/cciilamhchpmbdnniabclekddabkifhb) and [Firefox](https://addons.mozilla.org/en-US/firefox/addon/hister/)

### Download pre-built binary

Grab a pre-built binary from the [latest release](https://github.com/asciimoo/hister/releases/latest). (Don't forget to `chmod +x`)

Execute `./hister` to see all available commands.

### Build for yourself

 - Clone the repository
 - Build with `go build`
 - Run `./hister help` to list the available commands
 - Execute `./hister listen` to start the web application

## Configuration

Settings can be configured in `~/.config/hister/config.yml` config file - don't forget to restart webapp after updating.

Execute `./hister create-config config.yml` to generate a configuration file with the default configuration values.


### Nix

#### Quick usage

Run directly from the repository:

```bash
nix run github:asciimoo/hister
```

Add to your current shell session:

```bash
nix shell github:asciimoo/hister
```

Install permanently to your user profile:

```bash
nix profile install github:asciimoo/hister
```

#### NixOS

Add the following to your `flake.nix`:

```nix
{
  inputs.hister.url = "github:asciimoo/hister";

  outputs = { self, nixpkgs, hister, ... }: {
    nixosConfigurations.yourHostname = nixpkgs.lib.nixosSystem {
      modules = [
        ./configuration.nix
        hister.nixosModules.default
      ];
    };
  };
}
```

Then enable the service:

```nix
services.hister = {
  enable = true;
  port = 4433;
  dataDir = "/var/lib/hister";
  configPath = /path/to/config.yml; # optional, use existing YAML file
  config = {  # optional, or use Nix attrset (automatically converted to YAML)
    app = {
      directory = "~/.config/hister/";
      search_url = "https://google.com/search?q={query}";
    };
    server = {
      address = "127.0.0.1:4433";
    };
  };
};
```

**Note**: Only one of `configPath` or `config` can be set at a time.

#### Add to system packages

If you don't want to use the system module, you can add the package directly to `environment.systemPackages` in your `configuration.nix`:

**NixOS & Darwin (macOS):**

```nix
{ inputs, ... }: {
  environment.systemPackages = [ inputs.hister.packages.${pkgs.system}.default ];
}
```

#### Add to user packages (Home-Manager)

If you don't want to use the Home-Manager module, you can add the package directly to `home.packages` in your `home.nix`:

```nix
{ inputs, ... }: {
  home.packages = [ inputs.hister.packages.${pkgs.system}.default ];
}
```

#### Home-Manager

Add the following to your `flake.nix`:

```nix
{
  inputs.hister.url = "github:asciimoo/hister";

  outputs = { self, nixpkgs, home-manager, hister, ... }: {
    homeConfigurations."yourUsername" = home-manager.lib.homeManagerConfiguration {
      modules = [
        ./home.nix
        hister.homeModules.default
      ];
    };
  };
}
```

Then enable the service:

```nix
services.hister = {
  enable = true;
  port = 4433;
  dataDir = "/home/yourUsername/.local/share/hister";
  configPath = /path/to/config.yml; # optional, use existing YAML file
  config = {  # optional, or use Nix attrset (automatically converted to YAML)
    app = {
      directory = "~/.config/hister/";
      search_url = "https://google.com/search?q={query}";
    };
    server = {
      address = "127.0.0.1:4433";
    };
  };
};
```

**Note**: Only one of `configPath` or `config` can be set at a time.

#### Darwin (macOS)

Add the following to your `flake.nix`:

```nix
{
  inputs.hister.url = "github:asciimoo/hister";

  outputs = { self, darwin, hister, ... }: {
    darwinConfigurations."yourHostname" = darwin.lib.darwinSystem {
      modules = [
        ./configuration.nix
        hister.darwinModules.default
      ];
    };
  };
}
```

Then enable the service:

```nix
services.hister = {
  enable = true;
  port = 4433;
  dataDir = "/Users/yourUsername/Library/Application Support/hister";
  configPath = /path/to/config.yml; # optional
  config = {  # optional, or use Nix attrset (automatically converted to YAML)
    app = {
      directory = "~/.config/hister/";
      search_url = "https://google.com/search?q={query}";
    };
    server = {
      address = "127.0.0.1:4433";
    };
  };
};
```

## Bugs

Bugs or suggestions? Visit the [issue tracker](https://github.com/asciimoo/hister/issues).


## License

AGPLv3
