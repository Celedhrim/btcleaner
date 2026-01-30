# BTCleaner - Quick Start Guide

## Installation

### Option 1: Binary (Recommended for Quick Start)

```bash
# The binary is already built and ready to use
./btcleaner --help
```

### Option 2: Build from Source

```bash
make build
```

### Option 3: Docker

```bash
# Build the image
docker build -t btcleaner .

# Or use docker-compose
docker-compose up -d
```

## First Run

### 1. Test Connection (Dry-Run)

Replace the URL, username, and password with your Transmission credentials:

```bash
./btcleaner \
  --transmission-url http://YOUR_TRANSMISSION_IP:9091/transmission/rpc \
  --transmission-user YOUR_USERNAME \
  --transmission-pass YOUR_PASSWORD \
  --dry-run \
  --log-level debug
```

This will show you what would be deleted without actually deleting anything.

### 2. Run Once (One-Shot Mode)

If the dry-run looks good, remove the `--dry-run` flag:

```bash
./btcleaner \
  -u http://YOUR_TRANSMISSION_IP:9091/transmission/rpc \
  -U YOUR_USERNAME \
  -P YOUR_PASSWORD
```

### 3. Run as Daemon (Continuous Monitoring)

For continuous monitoring, add the `-d` flag:

```bash
./btcleaner \
  -u http://YOUR_TRANSMISSION_IP:9091/transmission/rpc \
  -U YOUR_USERNAME \
  -P YOUR_PASSWORD \
  -d \
  -i 5m
```

This will check every 5 minutes and clean up if needed.

## Configuration File Method

Instead of using CLI flags every time, create a `config.yaml` file:

```bash
# Copy the example config
cp config.example.yaml config.yaml

# Edit with your settings
nano config.yaml
```

Then simply run:

```bash
./btcleaner -c config.yaml
```

## Docker Quick Start

Edit `docker-compose.yml` with your Transmission credentials, then:

```bash
docker-compose up -d
docker-compose logs -f
```

## Common Commands

```bash
# Show help
./btcleaner --help

# Run tests
make test

# Clean rebuild
make clean build

# View all make commands
make help
```

## Default Settings

- **Minimum free space:** 100 GB
- **Minimum torrents per tracker:** 2
- **Check interval (daemon):** 1 minute
- **Log level:** info

## Need Help?

- Read [README.md](README.md) for detailed documentation
- See [EXAMPLES.md](EXAMPLES.md) for more usage examples
- Check [P0_IMPLEMENTATION.md](P0_IMPLEMENTATION.md) for technical details

## Support

If you encounter issues:

1. Check Transmission is accessible
2. Verify credentials are correct
3. Run with `--log-level debug` for detailed output
4. Check [README.md](README.md) troubleshooting section

---

**Enjoy using BTCleaner! 🧹**
