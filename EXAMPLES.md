# BTCleaner Usage Examples

## Basic Examples

### One-Shot Mode (Run Once)

```bash
# Basic usage with dry-run to see what would be deleted
./btcleaner \
  --transmission-url http://transmission.example.com:9091/transmission/rpc \
  --transmission-user myuser \
  --transmission-pass mypass \
  --dry-run

# Actually clean up space
./btcleaner \
  -u http://transmission.example.com:9091/transmission/rpc \
  -U myuser \
  -P mypass
```

### Daemon Mode (Continuous Monitoring)

```bash
# Run in background, check every 5 minutes
./btcleaner \
  -u http://transmission.example.com:9091/transmission/rpc \
  -U myuser \
  -P mypass \
  -d \
  -i 5m \
  -l info
```

### Custom Space and Retention Settings

```bash
# Keep minimum 50GB free, keep at least 3 torrents per tracker
./btcleaner \
  -u http://transmission.example.com:9091/transmission/rpc \
  -U myuser \
  -P mypass \
  -s 50 \
  -m 3 \
  -d
```

## Configuration File Examples

### Basic Config

Create `config.yaml`:

```yaml
transmission:
  url: "http://transmission.example.com:9091/transmission/rpc"
  username: "myuser"
  password: "mypass"

cleaner:
  min_free_space: 53687091200  # 50GB
  min_torrents_per_tracker: 2

daemon:
  enabled: true
  check_interval: "5m"

log_level: "info"
```

Run with:
```bash
./btcleaner -c config.yaml
```

### Advanced Config

```yaml
transmission:
  url: "https://seedbox.example.com:9091/transmission/rpc"
  username: "admin"
  password: "secure_password"

cleaner:
  min_free_space: 107374182400  # 100GB
  min_torrents_per_tracker: 3

daemon:
  enabled: true
  check_interval: "2m"

dry_run: false
log_level: "debug"
```

## Environment Variables Examples

### Docker Environment

```bash
docker run -d \
  --name btcleaner \
  --restart unless-stopped \
  -e BTCLEANER_TRANSMISSION_URL=http://transmission:9091/transmission/rpc \
  -e BTCLEANER_TRANSMISSION_USERNAME=user \
  -e BTCLEANER_TRANSMISSION_PASSWORD=pass \
  -e BTCLEANER_CLEANER_MIN_FREE_SPACE=107374182400 \
  -e BTCLEANER_CLEANER_MIN_TORRENTS_PER_TRACKER=2 \
  -e BTCLEANER_DAEMON_ENABLED=true \
  -e BTCLEANER_DAEMON_CHECK_INTERVAL=3m \
  -e BTCLEANER_LOG_LEVEL=info \
  ghcr.io/celedhrim/btcleaner:latest
```

### Shell Script with Environment

```bash
#!/bin/bash
export BTCLEANER_TRANSMISSION_URL="http://localhost:9091/transmission/rpc"
export BTCLEANER_TRANSMISSION_USERNAME="myuser"
export BTCLEANER_TRANSMISSION_PASSWORD="mypass"
export BTCLEANER_CLEANER_MIN_FREE_SPACE="53687091200"  # 50GB
export BTCLEANER_CLEANER_MIN_TORRENTS_PER_TRACKER="3"
export BTCLEANER_DAEMON_ENABLED="true"
export BTCLEANER_DAEMON_CHECK_INTERVAL="10m"
export BTCLEANER_LOG_LEVEL="info"

./btcleaner
```

## Docker Compose Examples

### Standalone BTCleaner

```yaml
version: '3.8'

services:
  btcleaner:
    image: ghcr.io/celedhrim/btcleaner:latest
    container_name: btcleaner
    restart: unless-stopped
    environment:
      BTCLEANER_TRANSMISSION_URL: "http://192.168.1.100:9091/transmission/rpc"
      BTCLEANER_TRANSMISSION_USERNAME: "user"
      BTCLEANER_TRANSMISSION_PASSWORD: "pass"
      BTCLEANER_CLEANER_MIN_FREE_SPACE: "107374182400"
      BTCLEANER_CLEANER_MIN_TORRENTS_PER_TRACKER: "2"
      BTCLEANER_DAEMON_ENABLED: "true"
      BTCLEANER_DAEMON_CHECK_INTERVAL: "5m"
      BTCLEANER_LOG_LEVEL: "info"
```

### With Local Transmission

```yaml
version: '3.8'

services:
  transmission:
    image: linuxserver/transmission:latest
    container_name: transmission
    environment:
      - PUID=1000
      - PGID=1000
      - TZ=Europe/Paris
    volumes:
      - ./transmission/config:/config
      - ./transmission/downloads:/downloads
    ports:
      - "9091:9091"
    restart: unless-stopped

  btcleaner:
    image: ghcr.io/celed/btcleaner:latest
    container_name: btcleaner
    restart: unless-stopped
    depends_on:
      - transmission
    environment:
      BTCLEANER_TRANSMISSION_URL: "http://transmission:9091/transmission/rpc"
      BTCLEANER_TRANSMISSION_USERNAME: ""
      BTCLEANER_TRANSMISSION_PASSWORD: ""
      BTCLEANER_CLEANER_MIN_FREE_SPACE: "107374182400"
      BTCLEANER_CLEANER_MIN_TORRENTS_PER_TRACKER: "2"
      BTCLEANER_DAEMON_ENABLED: "true"
      BTCLEANER_DAEMON_CHECK_INTERVAL: "1m"
      BTCLEANER_LOG_LEVEL: "info"
```

## Cron Job Examples

### Daily Cleanup at 3 AM

```bash
# Add to crontab (crontab -e)
0 3 * * * /usr/local/bin/btcleaner -u http://localhost:9091/transmission/rpc -U user -P pass >> /var/log/btcleaner.log 2>&1
```

### Every 6 Hours

```bash
0 */6 * * * /usr/local/bin/btcleaner -u http://localhost:9091/transmission/rpc -U user -P pass -s 75 -m 3
```

### With Notification (example with curl webhook)

```bash
#!/bin/bash
LOG_FILE="/tmp/btcleaner_$(date +%Y%m%d_%H%M%S).log"

/usr/local/bin/btcleaner \
  -u http://localhost:9091/transmission/rpc \
  -U user \
  -P pass \
  -s 100 \
  -m 2 \
  | tee "$LOG_FILE"

# Send notification if errors occurred
if [ $? -ne 0 ]; then
    curl -X POST https://hooks.example.com/webhook \
         -H 'Content-Type: application/json' \
         -d "{\"text\":\"BTCleaner failed! Check $LOG_FILE\"}"
fi
```

## Systemd Service Example

Create `/etc/systemd/system/btcleaner.service`:

```ini
[Unit]
Description=BTCleaner - Automatic Transmission Cleanup
After=network.target

[Service]
Type=simple
User=btcleaner
Group=btcleaner
Restart=always
RestartSec=10
Environment="BTCLEANER_TRANSMISSION_URL=http://localhost:9091/transmission/rpc"
Environment="BTCLEANER_TRANSMISSION_USERNAME=user"
Environment="BTCLEANER_TRANSMISSION_PASSWORD=pass"
Environment="BTCLEANER_CLEANER_MIN_FREE_SPACE=107374182400"
Environment="BTCLEANER_CLEANER_MIN_TORRENTS_PER_TRACKER=2"
Environment="BTCLEANER_DAEMON_ENABLED=true"
Environment="BTCLEANER_DAEMON_CHECK_INTERVAL=5m"
Environment="BTCLEANER_LOG_LEVEL=info"
ExecStart=/usr/local/bin/btcleaner

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl daemon-reload
sudo systemctl enable btcleaner
sudo systemctl start btcleaner
sudo systemctl status btcleaner
```

View logs:
```bash
sudo journalctl -u btcleaner -f
```

## Testing and Debugging

### Test Connection

```bash
./btcleaner \
  -u http://localhost:9091/transmission/rpc \
  -U user \
  -P pass \
  -n \
  -l debug
```

### See What Would Be Deleted

```bash
./btcleaner \
  -u http://localhost:9091/transmission/rpc \
  -U user \
  -P pass \
  --dry-run \
  --log-level debug \
  --min-free-space 200  # Set high to trigger cleanup
```

### Monitor in Real-Time

```bash
# In one terminal, start daemon
./btcleaner -u http://localhost:9091/transmission/rpc -U user -P pass -d -i 1m -l debug

# In another terminal, watch transmission
watch -n 10 'transmission-remote localhost:9091 -l'
```

## Common Scenarios

### Scenario 1: Aggressive Cleanup

You need to free a lot of space quickly:

```bash
./btcleaner \
  -u http://localhost:9091/transmission/rpc \
  -U user \
  -P pass \
  -s 200 \    # High threshold
  -m 1 \      # Keep only 1 per tracker
  -l info
```

### Scenario 2: Conservative Cleanup

You want to keep more torrents:

```bash
./btcleaner \
  -u http://localhost:9091/transmission/rpc \
  -U user \
  -P pass \
  -s 50 \     # Lower threshold
  -m 5 \      # Keep 5 per tracker
  -d \
  -i 30m      # Check every 30 minutes
```

### Scenario 3: Testing Before Production

```bash
# First, dry-run to see what would happen
./btcleaner -u http://localhost:9091/transmission/rpc -U user -P pass -n -l debug

# If satisfied, run for real
./btcleaner -u http://localhost:9091/transmission/rpc -U user -P pass -l info

# Then set up daemon
./btcleaner -u http://localhost:9091/transmission/rpc -U user -P pass -d -i 5m
```

## Tips and Best Practices

1. **Always test with dry-run first** before actual cleanup
2. **Set appropriate min-torrents** based on your tracker diversity
3. **Monitor logs** regularly to ensure expected behavior
4. **Use daemon mode** in production for continuous monitoring
5. **Start with conservative settings** and adjust based on needs
6. **Keep backups** of important .torrent files if needed
7. **Monitor disk usage** trends to adjust check-interval
