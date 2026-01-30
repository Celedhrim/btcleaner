package server

import "strings"

// generateHTML generates the embedded HTML for the web interface
func (s *Server) generateHTML(webRoot string) string {
	version := s.version
	if version == "" {
		version = "dev"
	}
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>BTCleaner - Dashboard</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            background: #f5f5f5;
            color: #333;
            line-height: 1.6;
        }

        .container {
            max-width: 1400px;
            margin: 0 auto;
            padding: 20px;
        }

        header {
            background: #2c3e50;
            color: white;
            padding: 20px 0;
            margin-bottom: 30px;
            box-shadow: 0 2px 5px rgba(0,0,0,0.1);
        }

        header h1 {
            font-size: 28px;
            font-weight: 600;
        }

        header .header-content {
            display: flex;
            justify-content: space-between;
            align-items: center;
        }

        header .version {
            font-size: 12px;
            opacity: 0.7;
            font-weight: 400;
        }

        header p {
            opacity: 0.9;
            margin-top: 5px;
        }

        .stats-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }

        .stat-card {
            background: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 5px rgba(0,0,0,0.1);
        }

        .stat-card.wide {
            grid-column: span 2;
        }

        .stat-card h3 {
            font-size: 14px;
            color: #666;
            text-transform: uppercase;
            margin-bottom: 10px;
            font-weight: 500;
        }

        .stat-card .value {
            font-size: 32px;
            font-weight: 700;
            color: #2c3e50;
            word-wrap: break-word;
        }

        .stat-card#status-card .value {
            font-size: 20px;
        }

        .stat-card.warning .value {
            color: #e74c3c;
        }

        .stat-card.success .value {
            color: #27ae60;
        }

        .card {
            background: white;
            border-radius: 8px;
            box-shadow: 0 2px 5px rgba(0,0,0,0.1);
            margin-bottom: 20px;
            overflow: hidden;
        }

        .card-header {
            padding: 15px 20px;
            border-bottom: 1px solid #e0e0e0;
            display: flex;
            justify-content: space-between;
            align-items: center;
            background: #fafafa;
        }

        .card-header h2 {
            font-size: 18px;
            font-weight: 600;
            color: #2c3e50;
        }

        .card-body {
            padding: 20px;
        }

        .table-container {
            overflow-x: auto;
        }

        table {
            width: 100%;
            border-collapse: collapse;
        }

        thead {
            background: #fafafa;
        }

        th {
            padding: 12px;
            text-align: left;
            font-weight: 600;
            color: #666;
            font-size: 14px;
            border-bottom: 2px solid #e0e0e0;
        }

        td {
            padding: 12px;
            border-bottom: 1px solid #f0f0f0;
            font-size: 14px;
        }

        tbody tr:hover {
            background: #f9f9f9;
        }

        tbody tr.candidate {
            background: #fff3cd;
            border-left: 4px solid #f39c12;
        }

        tbody tr.candidate:hover {
            background: #ffe8a1;
        }

        .btn {
            padding: 8px 16px;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-size: 14px;
            font-weight: 500;
            transition: all 0.2s;
        }

        .btn-danger {
            background: #e74c3c;
            color: white;
        }

        .btn-danger:hover {
            background: #c0392b;
        }

        .btn-refresh {
            background: #3498db;
            color: white;
        }

        .btn-refresh:hover {
            background: #2980b9;
        }

        .badge {
            display: inline-block;
            padding: 4px 8px;
            border-radius: 4px;
            font-size: 12px;
            font-weight: 600;
        }

        .badge-success {
            background: #d4edda;
            color: #155724;
        }

        .tracker-list {
            margin-top: 10px;
            max-height: 200px;
            overflow-y: auto;
        }

        .tracker-item {
            display: flex;
            justify-content: space-between;
            padding: 8px 12px;
            background: #f8f9fa;
            border-radius: 4px;
            margin-bottom: 6px;
            font-size: 14px;
        }

        .tracker-item .tracker-name {
            font-weight: 600;
            color: #2c3e50;
        }

        .tracker-item .tracker-stats {
            color: #666;
        }

        .history-item {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 8px 12px;
            background: #f8f9fa;
            border-radius: 4px;
            margin-bottom: 6px;
            font-size: 13px;
        }

        .history-item .history-name {
            font-weight: 500;
            color: #2c3e50;
            flex: 1;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
            margin-right: 10px;
        }

        .history-item .history-info {
            display: flex;
            align-items: center;
            gap: 8px;
            font-size: 12px;
            color: #666;
        }

        .history-badge {
            padding: 2px 6px;
            border-radius: 3px;
            font-size: 11px;
            font-weight: 600;
        }

        .history-badge.auto {
            background: #d1ecf1;
            color: #0c5460;
        }

        .history-badge.manual {
            background: #fff3cd;
            color: #856404;
        }

        .badge-warning {
            background: #fff3cd;
            color: #856404;
        }

        .logs-container {
            background: #1e1e1e;
            color: #d4d4d4;
            padding: 15px;
            border-radius: 4px;
            font-family: 'Courier New', monospace;
            font-size: 13px;
            max-height: 400px;
            overflow-y: auto;
        }

        .log-entry {
            margin-bottom: 4px;
            line-height: 1.5;
        }

        .log-time {
            color: #858585;
        }

        .log-level {
            font-weight: 600;
            margin: 0 8px;
        }

        .log-level.debug { color: #9cdcfe; }
        .log-level.info { color: #4ec9b0; }
        .log-level.warn { color: #dcdcaa; }
        .log-level.error { color: #f48771; }

        .loading {
            text-align: center;
            padding: 40px;
            color: #999;
        }

        .empty {
            text-align: center;
            padding: 40px;
            color: #999;
        }

        .refresh-time {
            color: #999;
            font-size: 12px;
        }

        @media (max-width: 768px) {
            .stats-grid {
                grid-template-columns: 1fr;
            }
            
            .container {
                padding: 10px;
            }

            table {
                font-size: 12px;
            }

            th, td {
                padding: 8px;
            }
        }
    </style>
</head>
<body>
    <header>
        <div class="container">
            <div class="header-content">
                <div>
                    <h1>🧹 BTCleaner Dashboard</h1>
                    <p>Automatic Transmission Seedbox Cleanup</p>
                </div>
                <div class="version">` + version + `</div>
            </div>
        </div>
    </header>

    <div class="container">
        <!-- Statistics -->
        <div class="stats-grid">
            <div class="stat-card" id="free-space-card">
                <h3>Free Space</h3>
                <div class="value" id="free-space">--</div>
            </div>
            <div class="stat-card">
                <h3>Min Required</h3>
                <div class="value" id="min-space">--</div>
            </div>
            <div class="stat-card">
                <h3>Total Torrents</h3>
                <div class="value" id="total-torrents">--</div>
                <div style="font-size: 14px; color: #666; margin-top: 5px;" id="total-space">--</div>
            </div>
            <div class="stat-card" id="status-card">
                <h3>Status</h3>
                <div class="value" id="status">--</div>
            </div>
            <div class="stat-card wide">
                <h3>Trackers</h3>
                <div class="tracker-list" id="tracker-list">
                    <div style="color: #999; text-align: center; padding: 20px;">Loading...</div>
                </div>
            </div>
            <div class="stat-card wide">
                <h3>Cleanup History</h3>
                <div class="tracker-list" id="history-list">
                    <div style="color: #999; text-align: center; padding: 20px;">No recent deletions</div>
                </div>
            </div>
        </div>

        <!-- Torrents Table -->
        <div class="card">
            <div class="card-header">
                <h2>Torrents</h2>
                <div>
                    <span class="refresh-time" id="torrents-update">Never</span>
                    <button class="btn btn-refresh" onclick="loadTorrents()">Refresh</button>
                </div>
            </div>
            <div class="card-body">
                <div class="table-container">
                    <table id="torrents-table">
                        <thead>
                            <tr>
                                <th>Name</th>
                                <th>Tracker</th>
                                <th>Size</th>
                                <th>Added</th>
                                <th>Action</th>
                            </tr>
                        </thead>
                        <tbody id="torrents-body">
                            <tr>
                                <td colspan="5" class="loading">Loading...</td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </div>
        </div>

        <!-- Logs -->
        <div class="card">
            <div class="card-header">
                <h2>Logs</h2>
                <div>
                    <span class="refresh-time" id="logs-status">Connecting...</span>
                    <button class="btn btn-refresh" onclick="loadLogs()">Refresh</button>
                </div>
            </div>
            <div class="card-body">
                <div class="logs-container" id="logs-container">
                    <div class="log-entry">Connecting to log stream...</div>
                </div>
            </div>
        </div>
    </div>

    <script>
        const webRoot = '{{WEBROOT}}';
        const apiBase = webRoot === '/' ? '' : webRoot;
        let ws = null;
        let reconnectInterval = null;
        let candidateIds = [];

        // Load statistics
        async function loadStats() {
            try {
                const response = await fetch(apiBase + '/api/stats');
                const data = await response.json();
                
                document.getElementById('free-space').textContent = data.free_space_gb.toFixed(2) + ' GB';
                document.getElementById('min-space').textContent = data.min_free_space_gb.toFixed(2) + ' GB';
                document.getElementById('total-torrents').textContent = data.total_torrents;
                document.getElementById('total-space').textContent = data.total_space_gb.toFixed(2) + ' GB used';
                
                // Display tracker stats
                const trackerList = document.getElementById('tracker-list');
                if (data.tracker_stats && data.tracker_stats.length > 0) {
                    trackerList.innerHTML = data.tracker_stats
                        .sort((a, b) => b.size_gb - a.size_gb)
                        .map(t => 
                            "<div class='tracker-item'>" +
                                "<span class='tracker-name'>" + escapeHtml(t.name) + "</span>" +
                                "<span class='tracker-stats'>" + t.count + " torrents, " + t.size_gb.toFixed(2) + " GB</span>" +
                            "</div>"
                        ).join('');
                } else {
                    trackerList.innerHTML = "<div style='color: #999; text-align: center; padding: 20px;'>No trackers</div>";
                }
                
                const freeCard = document.getElementById('free-space-card');
                if (data.needs_cleanup) {
                    freeCard.classList.add('warning');
                    freeCard.classList.remove('success');
                    
                    let statusText = '⚠️ Cleanup Needed';
                    if (data.candidates_count > 0) {
                        statusText += ' (' + data.candidates_count + ' torrents, +' + 
                                     data.space_to_recover_gb.toFixed(2) + ' GB)';
                    }
                    document.getElementById('status').textContent = statusText;
                    
                    // Load candidates when cleanup is needed
                    loadCandidates();
                } else {
                    freeCard.classList.add('success');
                    freeCard.classList.remove('warning');
                    document.getElementById('status').textContent = '✓ OK';
                    candidateIds = [];
                }
            } catch (error) {
                console.error('Failed to load stats:', error);
            }
        }

        // Load candidates for deletion
        async function loadCandidates() {
            try {
                const response = await fetch(apiBase + '/api/candidates');
                candidateIds = await response.json();
                // Refresh torrents to apply highlighting
                loadTorrents();
            } catch (error) {
                console.error('Failed to load candidates:', error);
            }
        }

        // Load deletion history
        async function loadHistory() {
            try {
                const response = await fetch(apiBase + '/api/history');
                const history = await response.json();
                
                const historyList = document.getElementById('history-list');
                if (history && history.length > 0) {
                    historyList.innerHTML = history.slice(0, 20).map(h => {
                        const date = new Date(h.deleted_at);
                        const timeAgo = getTimeAgo(date);
                        const reasonBadge = h.reason === 'auto' ? 
                            "<span class='history-badge auto'>AUTO</span>" : 
                            "<span class='history-badge manual'>MANUAL</span>";
                        
                        return "<div class='history-item'>" +
                            "<span class='history-name' title='" + escapeHtml(h.name) + "'>" + 
                            truncate(h.name, 40) + "</span>" +
                            "<div class='history-info'>" +
                            reasonBadge +
                            "<span>" + h.size_gb.toFixed(2) + " GB</span>" +
                            "<span style='color: #999;'>" + timeAgo + "</span>" +
                            "</div>" +
                            "</div>";
                    }).join('');
                } else {
                    historyList.innerHTML = "<div style='color: #999; text-align: center; padding: 20px;'>No recent deletions</div>";
                }
            } catch (error) {
                console.error('Failed to load history:', error);
            }
        }

        // Load torrents
        async function loadTorrents() {
            try {
                const response = await fetch(apiBase + '/api/torrents');
                const torrents = await response.json();
                
                const tbody = document.getElementById('torrents-body');
                
                if (torrents.length === 0) {
                    tbody.innerHTML = '<tr><td colspan="5" class="empty">No torrents found</td></tr>';
                    return;
                }
                
                tbody.innerHTML = torrents.map(t => {
                    const size = (t.totalSize / (1024 * 1024 * 1024)).toFixed(2);
                    const date = new Date(t.addedDate).toLocaleDateString();
                    const nameEscaped = escapeHtml(t.name).replace(/'/g, "&apos;");
                    const isCandidate = candidateIds.includes(t.id);
                    const candidateClass = isCandidate ? " class='candidate'" : "";
                    return "<tr" + candidateClass + ">" +
                        "<td title='" + escapeHtml(t.name) + "'>" + truncate(t.name, 60) + "</td>" +
                        "<td><span class='badge badge-success'>" + t.normalizedTracker + "</span></td>" +
                        "<td>" + size + " GB</td>" +
                        "<td>" + date + "</td>" +
                        "<td><button class='btn btn-danger' onclick='deleteTorrent(" + t.id + ", \"" + nameEscaped + "\")'>Delete</button></td>" +
                        "</tr>";
                }).join('');
                
                document.getElementById('torrents-update').textContent = 'Updated: ' + new Date().toLocaleTimeString();
            } catch (error) {
                console.error('Failed to load torrents:', error);
                document.getElementById('torrents-body').innerHTML = '<tr><td colspan="5" class="empty">Failed to load torrents</td></tr>';
            }
        }

        // Delete torrent
        async function deleteTorrent(id, name) {
            if (!confirm("Are you sure you want to delete \"" + name + "\"?")) {
                return;
            }
            
            try {
                const response = await fetch(apiBase + "/api/delete?id=" + id, {
                    method: 'POST'
                });
                
                if (response.ok) {
                    alert('Torrent deleted successfully');
                    loadTorrents();
                    loadStats();
                    loadHistory();
                } else {
                    alert('Failed to delete torrent');
                }
            } catch (error) {
                console.error('Failed to delete torrent:', error);
                alert('Failed to delete torrent');
            }
        }

        // Load logs (fallback if WebSocket fails)
        async function loadLogs() {
            try {
                const response = await fetch(apiBase + '/api/logs');
                const logs = await response.json();
                displayLogs(logs);
            } catch (error) {
                console.error('Failed to load logs:', error);
            }
        }

        // Display logs
        function displayLogs(logs) {
            const container = document.getElementById('logs-container');
            container.innerHTML = logs.map(log => {
                const time = new Date(log.timestamp).toLocaleTimeString();
                return "<div class='log-entry'>" +
                    "<span class='log-time'>" + time + "</span>" +
                    "<span class='log-level " + log.level + "'>[" + log.level.toUpperCase() + "]</span>" +
                    "<span>" + escapeHtml(log.message) + "</span>" +
                    "</div>";
            }).join('');
            container.scrollTop = container.scrollHeight;
        }

        // WebSocket for real-time logs
        function connectWebSocket() {
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const wsUrl = protocol + "//" + window.location.host + apiBase + "/ws/logs";
            
            ws = new WebSocket(wsUrl);
            
            ws.onopen = () => {
                console.log('WebSocket connected');
                document.getElementById('logs-status').textContent = '🟢 Live';
                if (reconnectInterval) {
                    clearInterval(reconnectInterval);
                    reconnectInterval = null;
                }
            };
            
            ws.onmessage = (event) => {
                const log = JSON.parse(event.data);
                const container = document.getElementById('logs-container');
                const time = new Date(log.timestamp).toLocaleTimeString();
                const entry = "<div class='log-entry'>" +
                    "<span class='log-time'>" + time + "</span>" +
                    "<span class='log-level " + log.level + "'>[" + log.level.toUpperCase() + "]</span>" +
                    "<span>" + escapeHtml(log.message) + "</span>" +
                    "</div>";
                container.innerHTML += entry;
                container.scrollTop = container.scrollHeight;
            };
            
            ws.onerror = (error) => {
                console.error('WebSocket error:', error);
                document.getElementById('logs-status').textContent = '🔴 Disconnected';
            };
            
            ws.onclose = () => {
                console.log('WebSocket disconnected');
                document.getElementById('logs-status').textContent = '🔴 Disconnected';
                
                // Try to reconnect
                if (!reconnectInterval) {
                    reconnectInterval = setInterval(() => {
                        console.log('Attempting to reconnect WebSocket...');
                        connectWebSocket();
                    }, 5000);
                }
            };
        }

        // Utility functions
        function truncate(str, len) {
            return str.length > len ? str.substring(0, len) + '...' : str;
        }

        function escapeHtml(text) {
            const div = document.createElement('div');
            div.textContent = text;
            return div.innerHTML;
        }

        function getTimeAgo(date) {
            const seconds = Math.floor((new Date() - date) / 1000);
            
            if (seconds < 60) return seconds + 's ago';
            const minutes = Math.floor(seconds / 60);
            if (minutes < 60) return minutes + 'm ago';
            const hours = Math.floor(minutes / 60);
            if (hours < 24) return hours + 'h ago';
            const days = Math.floor(hours / 24);
            return days + 'd ago';
        }

        // Initialize
        loadStats();
        loadTorrents();
        loadHistory();
        connectWebSocket();

        // Auto-refresh stats and torrents
        setInterval(loadStats, 10000); // Every 10 seconds
        setInterval(loadTorrents, 30000); // Every 30 seconds
        setInterval(loadHistory, 30000); // Every 30 seconds
    </script>
</body>
</html>`
	
	// Replace the placeholder with the actual webRoot
	return strings.ReplaceAll(html, "{{WEBROOT}}", webRoot)
}
