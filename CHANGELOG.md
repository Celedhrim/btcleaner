# Changelog

## v1.0.0 - Complete Rewrite (2026-01-30)

**Major version** - Complete rewrite with extensive new features and breaking changes.

### 🚀 New Features

#### Core Functionality
- **Smart tracker management**: Respects minimum torrents per tracker
- **Multi-tracker support**: Groups public torrents automatically
- **Disk space monitoring**: Continuous monitoring with configurable thresholds
- **Dry-run mode**: Test cleanup without actual deletion (`-n` flag)
- **Daemon mode**: Continuous background monitoring (`-d` flag)

#### Web UI (NEW)
- **Dashboard**: Real-time torrent monitoring and statistics
- **Live logs**: WebSocket-based real-time log streaming
- **Manual management**: Select and remove torrents via UI
- **Status tracking**: Visual indicators for cleanup status
- **Reverse proxy support**: Configurable webroot for nginx/traefik

#### Configuration System
- **Multiple config locations**: 
  - `./btcleaner.yaml` (priority)
  - `./config.yaml`
  - `~/.config/btcleaner.yaml`
  - `/etc/btcleaner.yaml`
- **Human-readable sizes**: Support for `"100GB"`, `"500MB"`, etc. in config
- **Flexible options**: CLI flags, environment variables, or config file
- **Auto-discovery**: Finds config files automatically

#### Developer Features
- **Comprehensive tests**: Unit and integration tests
- **REST API**: Full API for automation
- **Clean architecture**: Modular, maintainable codebase
- **Docker support**: Dockerfile and docker-compose included

### 🔧 Configuration Changes (Breaking)

**Old config format:**
```yaml
transmission_url: "http://user:password@127.0.0.1:9091/transmission/rpc"
path: "/path/to/download/dir"
free_giga: 100
tracker_keep: 2
exclude: [...]
```

**New config format:**
```yaml
transmission:
  url: "http://localhost:9091/transmission/rpc"
  username: "user"
  password: "password"

cleaner:
  min_free_space: "100GB"  # Can use units!
  min_torrents_per_tracker: 2

server:
  enabled: false
  port: 8888
  webroot: "/"

daemon:
  enabled: false
  check_interval: "1m"

dry_run: false
log_level: "info"
```

### 📝 Command Line Changes

**Old flags:**
- `-c, --config` - Config file path
- `--do` - Commit torrent deletion
- `--cron` - Quiet mode
- `-f, --free_giga` - Target GiB free
- `-k, --tracker_keep` - Torrents to keep per tracker

**New flags:**
- `-c, --config` - Config file path (unchanged)
- `-u, --transmission-url` - Transmission RPC URL
- `-U, --transmission-user` - Username
- `-P, --transmission-pass` - Password
- `-s, --min-free-space` - Minimum free space in GB
- `-m, --min-torrents` - Minimum torrents per tracker
- `-d, --daemon` - Run in daemon mode
- `-i, --check-interval` - Check interval in daemon mode
- `-n, --dry-run` - Dry run mode (replaces `--do` logic)
- `-w, --web-ui` - Enable web UI
- `-p, --web-port` - Web UI port
- `-l, --log-level` - Log level

### 🗑️ Removed Features
- `exclude` list - Will be added in future version if needed
- `--cron` flag - Use daemon mode with appropriate check interval
- `path` configuration - No longer needed, handled by Transmission

### 📦 Dependencies
- Updated to Go 1.25.6
- Added Viper for configuration management
- Added Gorilla WebSocket for real-time features
- Cleaner dependency management

### 🐛 Bug Fixes
- Improved error handling throughout
- Better connection management
- Fixed race conditions in concurrent operations

### 📚 Documentation
- Comprehensive README with examples
- Quick start guide (QUICKSTART.md)
- Implementation details (P0_IMPLEMENTATION.md, P1_IMPLEMENTATION.md)
- Nginx configuration guide (NGINX.md)
- Example configurations

---

## v0.3.0
 * Update mods
 * breaking change on config , now transmission config is put has an url

## v0.2.1
 * Don't display excluded torrent in cron mode

## v0.2.0
 * Add exclude list to never delete some torrents

## v0.1.1
 * Fix config file search path
 * Remove unused error check

## v0.1.0
 * First release!
