# BTCleaner

🧹 Automatic torrent cleanup tool for Transmission seedbox. Maintains minimum free disk space by intelligently removing old torrents while respecting tracker quotas.

## Features

- 🔄 **Automatic cleanup**: Monitors disk space and removes oldest torrents when needed
- 📊 **Smart selection**: Respects minimum torrents per tracker (configurable)
- 🚀 **Multiple modes**: One-shot or continuous daemon mode
- 🧪 **Dry-run mode**: Test without actually deleting torrents
- ⚙️ **Flexible configuration**: CLI flags, environment variables, or config file
- 🐳 **Docker ready**: Includes Dockerfile and docker-compose setup
- 🌐 **Multi-tracker support**: Groups public torrents automatically
- 🖥️ **Web UI**: Dashboard for monitoring and manual management (NEW in P1)
- 📡 **Real-time logs**: WebSocket streaming for live log viewing (NEW in P1)
- 🔧 **REST API**: Full API for integration and automation (NEW in P1)

## Quick Start

### Binary Usage

```bash
# One-shot mode (run once and exit)
./btcleaner -u http://localhost:9091/transmission/rpc -U username -P password

# Daemon mode (continuous monitoring)
./btcleaner -u http://localhost:9091/transmission/rpc -U username -P password -d

# Dry-run mode (simulate only)
./btcleaner -u http://localhost:9091/transmission/rpc -U username -P password -n

# With custom settings
./btcleaner \
  -u http://localhost:9091/transmission/rpc \
  -U username -P password \
  -s 50 \              # 50GB minimum free space
  -m 3 \               # Keep minimum 3 torrents per tracker
  -d \                 # Daemon mode
  -i 5m                # Check every 5 minutes

# With Web UI
./btcleaner \
  -u http://localhost:9091/transmission/rpc \
  -U username -P password \
  -w \                 # Enable Web UI
  -p 8888 \            # Web UI port
  -d                   # Daemon mode
# Access at: http://localhost:8888
```

### Docker Usage

```bash
# Pull from GitHub Container Registry
docker pull ghcr.io/celedhrim/btcleaner:latest

# Or build locally
docker build -t ghcr.io/celedhrim/btcleaner .

# Run with environment variables
docker run -d \
  --name btcleaner \
  -e BTCLEANER_TRANSMISSION_URL=http://transmission:9091/transmission/rpc \
  -e BTCLEANER_TRANSMISSION_USERNAME=user \
  -e BTCLEANER_TRANSMISSION_PASSWORD=pass \
  -e BTCLEANER_DAEMON_ENABLED=true \
  ghcr.io/celedhrim/btcleaner:latest

# Or use docker-compose
docker-compose up -d
```

## Configuration

### CLI Flags

| Flag (short) | Flag (long) | Description | Default |
|-------------|-------------|-------------|---------|
| `-u` | `--transmission-url` | Transmission RPC URL | Required |
| `-U` | `--transmission-user` | Transmission username | - |
| `-P` | `--transmission-pass` | Transmission password | - |
| `-s` | `--min-free-space` | Minimum free space (GB) | 100 |
| `-m` | `--min-torrents` | Min torrents per tracker | 2 |
| `-d` | `--daemon` | Enable daemon mode | false |
| `-i` | `--check-interval` | Check interval (daemon) | 1m |
| `-n` | `--dry-run` | Dry run mode | false |
| `-w` | `--web-ui` | Enable Web UI | false |
| `-p` | `--web-port` | Web UI port | 8888 |
| `-r` | `--web-root` | Web UI root path (for reverse proxy) | / |
| `-c` | `--config` | Config file path | - |
| `-l` | `--log-level` | Log level (debug/info/warn/error) | info |

### Environment Variables

All configuration can be set via environment variables with the `BTCLEANER_` prefix:

```bash
export BTCLEANER_TRANSMISSION_URL="http://localhost:9091/transmission/rpc"
export BTCLEANER_TRANSMISSION_USERNAME="user"
export BTCLEANER_TRANSMISSION_PASSWORD="pass"
export BTCLEANER_CLEANER_MIN_FREE_SPACE="107374182400"  # 100GB in bytes
export BTCLEANER_CLEANER_MIN_TORRENTS_PER_TRACKER="2"
export BTCLEANER_DAEMON_ENABLED="true"
export BTCLEANER_DAEMON_CHECK_INTERVAL="1m"
export BTCLEANER_DRY_RUN="false"
export BTCLEANER_LOG_LEVEL="info"
```

### Configuration File

BTCleaner looks for configuration files in the following locations (in priority order):

1. `./btcleaner.yaml` (current directory - highest priority)
2. `./config.yaml` (current directory)
3. `~/.config/btcleaner.yaml` (user configuration)
4. `/etc/btcleaner.yaml` (system configuration)

Or specify a custom path with `-c` flag.

Example configuration (see [config.example.yaml](config.example.yaml)):

```yaml
transmission:
  url: "http://localhost:9091/transmission/rpc"
  username: "user"
  password: "pass"

cleaner:
  # Can use units: GB, MB, KB, or raw bytes
  min_free_space: "100GB"  # or 107374182400
  min_torrents_per_tracker: 2

daemon:
  enabled: true
  check_interval: "1m"

dry_run: false
log_level: "info"
```

Then run:
```bash
./btcleaner  # Automatically finds config file
# Or specify custom location:
./btcleaner -c /path/to/config.yaml
```

### Configuration Priority

1. CLI flags (highest priority)
2. Environment variables
3. Config file
4. Default values (lowest priority)

## How It Works

1. **Check disk space**: Queries Transmission API for available free space
2. **Compare threshold**: If free space < minimum, cleanup is triggered
3. **Group by tracker**: Torrents are grouped by their tracker domain
4. **Sort by age**: Torrents sorted oldest to newest
5. **Smart selection**: Selects oldest torrents while maintaining minimum per tracker
6. **Remove**: Deletes selected torrents and their data (or simulates in dry-run mode)

### Tracker Normalization

- **Single tracker**: Uses the tracker domain (e.g., `tracker.example.com`)
- **Multiple trackers**: Grouped as `public-tracker`
- **Unknown**: Labeled as `unknown`

### Minimum Torrents Constraint

The tool **strictly respects** the minimum torrents per tracker setting. If all trackers have only the minimum number of torrents (or fewer), no cleanup will occur even if disk space is critically low.

Example:
- Min torrents per tracker: 2
- Tracker A: 2 torrents
- Tracker B: 2 torrents
- Tracker C: 2 torrents
→ **Nothing will be deleted** even if space is low

## Building from Source

```bash
# Clone the repository
git clone https://github.com/celedhrim/btcleaner.git
cd btcleaner

# Download dependencies
go mod download

# Build
go build -o btcleaner ./cmd/btcleaner

# Run
./btcleaner --help
```

## Use Cases

### One-Shot with Cron

Run cleanup daily at 3 AM:

```bash
0 3 * * * /path/to/btcleaner -u http://transmission:9091/transmission/rpc -U user -P pass
```

### Daemon Mode in Docker

Continuous monitoring with docker-compose (see [docker-compose.yml](docker-compose.yml))

### Dry-Run Testing

Test what would be deleted:

```bash
./btcleaner -u http://transmission:9091/transmission/rpc -U user -P pass -n -l debug
```

## Logging

Logs are output to stdout/stderr in human-readable format:

```
INFO[0000] BTCleaner starting...
INFO[0000] Transmission URL: http://localhost:9091/transmission/rpc
INFO[0000] Min free space: 100.00 GB
INFO[0000] Min torrents per tracker: 2
INFO[0000] Testing connection to Transmission...
INFO[0001] Successfully connected to Transmission
INFO[0001] Running in one-shot mode
INFO[0001] Current free space: 45.23 GB
INFO[0001] Minimum required: 100.00 GB
WARN[0001] Need to free up 54.77 GB
INFO[0001] Found 156 torrents
INFO[0002] Selected 12 torrents to remove (will free 58.34 GB)
INFO[0002] Removing torrent: [tracker.example.com] Ubuntu 20.04 LTS (4.52 GB)
...
```

Set log level with `-l debug` for detailed information.

## Requirements

- Go 1.25+ (for building)
- Transmission 2.x or 3.x
- Network access to Transmission RPC

## License

MIT License - see LICENSE file for details

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## Author

Created by celed - 2026
