+++
date = '2026-02-13T10:59:19+01:00'
draft = false
title = 'Getting Started'
layout = 'documentation'
+++

## Installation

### Option 1: Pre-built Binary

1. Download the latest binary for your platform from the [releases page](https://github.com/asciimoo/hister/releases/latest)
2. Make the binary executable:
   ```bash
   chmod +x hister
   ```
3. Optionally, move it to your PATH:
   ```bash
   sudo mv hister /usr/local/bin/
   ```

### Option 2: Build from Source

**Requirements**: Go 1.16 or later

1. Clone the repository:
   ```bash
   git clone https://github.com/asciimoo/hister.git
   cd hister
   ```

2. Build the binary:
   ```bash
   go build
   ```
3. The `hister` binary will be created in the current directory

### Option 3: [Nix](#nix)

### Option 4: [Docker](https://github.com/asciimoo/hister/blob/master/Dockerfile)


## Browser Extension Setup

To automatically index your browsing history, install the browser extension:

- **Chrome**: [Install from Chrome Web Store](https://chromewebstore.google.com/detail/hister/cciilamhchpmbdnniabclekddabkifhb)
- **Firefox**: [Install from Firefox Add-ons](https://addons.mozilla.org/en-US/firefox/addon/hister/)

After installing the extension, configure it to point to your Hister server (default: `http://127.0.0.1:4433`).

## First Run

Check available commands:
   ```bash
   ./hister help
   ```


1. Start the Hister server:
   ```bash
   ./hister listen
   ```

2. Open your browser and navigate to `http://127.0.0.1:4433`

3. You should see the Hister web interface

## Configuration

Hister can be configured using a YAML configuration file located at `~/.config/hister/config.yml`.

### Generate Default Configuration

To create a configuration file with default values:

```bash
./hister create-config ~/.config/hister/config.yml
```

**Important**: Restart the Hister server after modifying the configuration file.

## Importing Existing Browser History

You can import your existing browser history from Firefox or Chrome:

### Firefox

```bash
./hister import firefox
```

This will automatically locate and import your Firefox history.

### Chrome

```bash
./hister import chrome
```

This will automatically locate and import your Chrome history.

## Command Line Usage

View all available commands:
```bash
./hister help
```

### Index a URL Manually

To manually index a specific URL:
```bash
./hister index https://example.com
```

## Using Hister

Once set up:

1. **Browse the web** with the extension installed - pages are automatically indexed
2. **Search your history** by visiting the Hister web interface
3. **Use advanced queries** with the [Bleve query syntax](https://blevesearch.com/docs/Query-String-Query/)
4. **Create keyword aliases** for frequently searched topics
5. **Configure blacklists** to exclude unwanted content

## Next Steps

- Explore the [advanced search syntax](https://blevesearch.com/docs/Query-String-Query/)
- Configure blacklist, hotkeys, sensitive data patterns and priority rules in your config file
- Set up keyword aliases for efficient searching
- Import your existing browser history

## Troubleshooting

### Server won't start

- Check if port 4433 is already in use
- Verify the configuration file syntax

### Extension not connecting

- Ensure the Hister server is running
- Verify the extension is configured with the correct server URL
- Check browser console for errors

### Import fails

- Ensure your server isn't running during import

## Nix

### Quick usage

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

### NixOS

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

### Add to system packages

If you don't want to use the system module, you can add the package directly to `environment.systemPackages` in your `configuration.nix`:

**NixOS & Darwin (macOS):**

```nix
{ inputs, ... }: {
  environment.systemPackages = [ inputs.hister.packages.${pkgs.system}.default ];
}
```

### Add to user packages (Home-Manager)

If you don't want to use the Home-Manager module, you can add the package directly to `home.packages` in your `home.nix`:

```nix
{ inputs, ... }: {
  home.packages = [ inputs.hister.packages.${pkgs.system}.default ];
}
```

### Home-Manager

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

### Darwin (macOS)

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
