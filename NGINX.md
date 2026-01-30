# BTCleaner - Nginx Reverse Proxy Configuration

## Basic Configuration

```nginx
server {
    listen 80;
    server_name btcleaner.example.com;

    location / {
        proxy_pass http://localhost:8888/;
        proxy_http_version 1.1;
        
        # WebSocket support
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        
        # Standard proxy headers
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Timeouts
        proxy_connect_timeout 7d;
        proxy_send_timeout 7d;
        proxy_read_timeout 7d;
    }
}
```

## With Subpath (e.g., /btcleaner)

### BTCleaner Configuration

```bash
./btcleaner \
  -u http://transmission:9091/transmission/rpc \
  -U user -P pass \
  -w -p 8888 -r /btcleaner \
  -d
```

Or in config.yaml:
```yaml
server:
  enabled: true
  port: 8888
  webroot: "/btcleaner"
```

### Nginx Configuration

```nginx
server {
    listen 80;
    server_name example.com;

    # Other services
    location / {
        # Your main site
    }

    # BTCleaner at /btcleaner
    location /btcleaner/ {
        proxy_pass http://localhost:8888/btcleaner/;
        proxy_http_version 1.1;
        
        # WebSocket support
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        
        # Standard proxy headers
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Timeouts for WebSocket
        proxy_connect_timeout 7d;
        proxy_send_timeout 7d;
        proxy_read_timeout 7d;
    }
}
```

## With SSL (HTTPS)

```nginx
server {
    listen 443 ssl http2;
    server_name btcleaner.example.com;

    ssl_certificate /etc/letsencrypt/live/btcleaner.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/btcleaner.example.com/privkey.pem;
    
    # SSL configuration
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;

    location / {
        proxy_pass http://localhost:8888/;
        proxy_http_version 1.1;
        
        # WebSocket support
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        
        # Standard proxy headers
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Timeouts
        proxy_connect_timeout 7d;
        proxy_send_timeout 7d;
        proxy_read_timeout 7d;
    }
}

# Redirect HTTP to HTTPS
server {
    listen 80;
    server_name btcleaner.example.com;
    return 301 https://$server_name$request_uri;
}
```

## With Basic Authentication

```nginx
server {
    listen 80;
    server_name btcleaner.example.com;

    # Basic auth
    auth_basic "BTCleaner Access";
    auth_basic_user_file /etc/nginx/.htpasswd;

    location / {
        proxy_pass http://localhost:8888/;
        proxy_http_version 1.1;
        
        # WebSocket support
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        
        # Standard proxy headers
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Timeouts
        proxy_connect_timeout 7d;
        proxy_send_timeout 7d;
        proxy_read_timeout 7d;
    }
}
```

Create htpasswd file:
```bash
sudo apt install apache2-utils
sudo htpasswd -c /etc/nginx/.htpasswd admin
```

## Multiple BTCleaner Instances

```nginx
server {
    listen 80;
    server_name dashboard.example.com;

    # Seedbox 1
    location /seedbox1/ {
        proxy_pass http://localhost:8881/seedbox1/;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
    }

    # Seedbox 2
    location /seedbox2/ {
        proxy_pass http://localhost:8882/seedbox2/;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
    }
}
```

BTCleaner configurations:
```bash
# Instance 1
./btcleaner -u http://seedbox1:9091/transmission/rpc -w -p 8881 -r /seedbox1 -d

# Instance 2
./btcleaner -u http://seedbox2:9091/transmission/rpc -w -p 8882 -r /seedbox2 -d
```

## Testing Configuration

```bash
# Test nginx configuration
sudo nginx -t

# Reload nginx
sudo nginx -s reload

# Test BTCleaner API
curl http://localhost/btcleaner/api/stats

# Test with authentication
curl -u admin:password http://localhost/btcleaner/api/stats
```

## Docker Compose with Nginx

```yaml
version: '3.8'

services:
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/conf.d/default.conf:ro
      - ./ssl:/etc/nginx/ssl:ro
    depends_on:
      - btcleaner
    restart: unless-stopped

  btcleaner:
    image: btcleaner:latest
    environment:
      BTCLEANER_TRANSMISSION_URL: "http://transmission:9091/transmission/rpc"
      BTCLEANER_SERVER_ENABLED: "true"
      BTCLEANER_SERVER_PORT: "8888"
      BTCLEANER_SERVER_WEBROOT: "/"
      BTCLEANER_DAEMON_ENABLED: "true"
    restart: unless-stopped

  transmission:
    image: linuxserver/transmission:latest
    environment:
      - PUID=1000
      - PGID=1000
    volumes:
      - ./transmission/config:/config
      - ./transmission/downloads:/downloads
    restart: unless-stopped
```

## Troubleshooting

### WebSocket Not Working

Make sure you have these headers:
```nginx
proxy_http_version 1.1;
proxy_set_header Upgrade $http_upgrade;
proxy_set_header Connection "upgrade";
```

### 502 Bad Gateway

Check BTCleaner is running:
```bash
curl http://localhost:8888/api/stats
```

Check nginx error logs:
```bash
sudo tail -f /var/log/nginx/error.log
```

### Subpath Not Working

Make sure:
1. BTCleaner is configured with `-r /subpath`
2. Nginx location matches: `location /subpath/`
3. Nginx proxy_pass includes the subpath: `proxy_pass http://localhost:8888/subpath/;`

### Connection Timeout

Increase timeouts in nginx:
```nginx
proxy_connect_timeout 7d;
proxy_send_timeout 7d;
proxy_read_timeout 7d;
```

## Security Best Practices

1. **Use HTTPS** in production
2. **Enable authentication** (basic auth, OAuth, etc.)
3. **Restrict IP access** if possible:
   ```nginx
   allow 192.168.1.0/24;
   deny all;
   ```
4. **Use firewall** to limit access to BTCleaner port
5. **Keep nginx updated**
6. **Monitor logs** regularly

## Additional Resources

- [Nginx WebSocket Proxying](https://nginx.org/en/docs/http/websocket.html)
- [Nginx Reverse Proxy Guide](https://docs.nginx.com/nginx/admin-guide/web-server/reverse-proxy/)
- [Let's Encrypt SSL](https://letsencrypt.org/getting-started/)
