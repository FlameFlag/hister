---
date: '2026-03-11T00:00:00+00:00'
draft: false
title: 'Running as a Service'
---

To keep Hister running in the background and start it automatically on boot, you can set it up as a system service. Below are instructions for Linux (systemd) and macOS (launchctl).

## Linux (systemd)

### 1. Create the service file

Create `~/.config/systemd/user/hister.service`:

```ini
[Unit]
Description=Hister Search Engine
After=network.target

[Service]
Type=simple
ExecStart=/path/to/hister listen
Restart=on-failure
RestartSec=5

# Optional: override config values
# Environment=HISTER__APP__LOG_LEVEL=error

# Optional: use a specific config file
# ExecStart=/path/to/hister --config /path/to/config.yml listen

[Install]
WantedBy=default.target
```

Replace `/path/to/hister` with the actual path to the hister binary.

### 2. Enable and start the service

```bash
# Reload systemd to pick up the new service file
systemctl --user daemon-reload

# Enable the service to start on login
systemctl --user enable hister

# Start the service now
systemctl --user start hister
```

To have user services start at boot (even before logging in):

```bash
sudo loginctl enable-linger $USER
```

### 3. Managing the service

```bash
# Check status
systemctl --user status hister

# View logs
journalctl --user -u hister -f

# Restart (e.g. after updating the binary or config)
systemctl --user restart hister

# Stop
systemctl --user stop hister

# Disable (stop starting on login)
systemctl --user disable hister
```

### Updating the binary

After building or installing a new version of hister, restart the service:

```bash
systemctl --user restart hister
```

---

## macOS (launchctl)

### 1. Create the plist file

Create `~/Library/LaunchAgents/hister.plist`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple Computer//DTD PLIST 1.0//EN"
"http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
  <dict>
    <key>Label</key>
    <string>org.hister.search</string>
    <key>ServiceDescription</key>
    <string>Hister Search Engine</string>
    <key>RunAtLoad</key>
    <true/>
    <key>ProgramArguments</key>
    <array>
      <string>/path/to/hister</string>
      <string>listen</string>
    </array>
    <key>StandardOutPath</key>
    <string>/tmp/hister.log</string>
    <key>StandardErrorPath</key>
    <string>/tmp/hister.log</string>
  </dict>
</plist>
```

Replace `/path/to/hister` with the actual path to the hister binary (e.g. `~/.local/bin/hister` or `/usr/local/bin/hister`).

You can optionally override configuration values using environment variables in the plist:

```xml
    <key>EnvironmentVariables</key>
    <dict>
      <key>HISTER__APP__LOG_LEVEL</key>
      <string>error</string>
    </dict>
```

**Note**: Environment variables in the plist take precedence over values in your config file.

### 2. Load and start the service

```bash
launchctl bootstrap gui/$(id -u) ~/Library/LaunchAgents/hister.plist
```

The service will start immediately and also launch on login.

### 3. Managing the service

```bash
# Restart the service (e.g. after updating the binary or config)
launchctl kickstart -k gui/$(id -u)/org.hister.search

# Stop the service
launchctl bootout gui/$(id -u)/org.hister.search

# Check if it's running
launchctl list | grep hister

# View logs
tail -f /tmp/hister.log
```

### Updating the binary

After building or installing a new version of hister, restart the service to pick up the changes:

```bash
launchctl kickstart -k gui/$(id -u)/org.hister.search
```

