# Hister

**Web history on steroids**

Hister is a web history management tool that provides blazing fast, content-based search for visited websites. Unlike traditional browser history that only searches URLs and titles, Hister indexes the full content of web pages you visit.

![hister screenshot](docs/assets/screenshot.png)

![hister screencast](docs/assets/demo.gif)


## Features

- **Privacy-focused**: Keep your browsing history indexed locally - don't use remote search engines if it isn't necessary
- **Full-text indexing**: Search through the actual content of web pages you've visited
- **Advanced search capabilities**: Utilize a powerful [query language](https://blevesearch.com/docs/Query-String-Query/) for precise results
- **Efficient retrieval**: Use keyword aliases to quickly find content
- **Flexible content management**: Configure blacklist and priority rules for better control

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

### Docker

Build and run with Docker Compose:

```bash
# Build and start the container
docker-compose up -d

# View logs
docker-compose logs -f

# Stop the container
docker-compose down

# Stop and remove volumes (delete database)
docker-compose down -v
```

The database and configuration will be stored in a Docker volume and persist across container restarts. Access the web interface at http://localhost:4433/

## Configuration

Settings can be configured in `~/.config/hister/config.yml` config file - don't forget to restart webapp after updating.

Execute `./hister create-config config.yml` to generate a configuration file with the default configuration values.


## Check out our [Documentation](https://hister.org/documentation/) for more details

## Bugs

Bugs or suggestions? Visit the [issue tracker](https://github.com/asciimoo/hister/issues).


## License

AGPLv3
