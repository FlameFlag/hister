---
date: '2026-03-06T19:45:22-05:00'
draft: false
title: 'Troubleshooting'
---

## Common Issues

### Server won't start

- Check if port 4433 (or whatever was configured instead) is already in use
- Verify the configuration file syntax

### Extension not connecting

- Ensure your Hister server is running
- Verify the extension is configured with the correct server URL
- Check browser console for errors (also, see below for debugging the extension itself)
- Check firewall settings

### Browser import fails

- Ensure your Hister server is running

## Debugging the Web Extension

The Web extension's logs will not be visible in the default browser console.
Instead:

### Firefox

1. Go to `about:debugging#/runtime/this-firefox`
2. Press the "Inspect" button to the right of "Hister".
