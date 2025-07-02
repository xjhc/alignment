// Package admin provides the admin dashboard UI
package admin

import (
	"net/http"
)

// AdminDashboardHTML is a simple HTML dashboard for server administration
const AdminDashboardHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Alignment Server Admin Dashboard</title>
    <style>
        body { 
            font-family: Arial, sans-serif; 
            margin: 20px; 
            background-color: #f5f5f5; 
        }
        .container { 
            max-width: 1200px; 
            margin: 0 auto; 
        }
        .header { 
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white; 
            padding: 20px; 
            border-radius: 8px; 
            margin-bottom: 20px;
        }
        .card { 
            background: white; 
            padding: 20px; 
            margin: 10px 0; 
            border-radius: 8px; 
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .status-healthy { color: #28a745; font-weight: bold; }
        .status-overloaded { color: #dc3545; font-weight: bold; }
        .game-list { display: grid; gap: 10px; }
        .game-item { 
            display: flex; 
            justify-content: space-between; 
            align-items: center; 
            padding: 10px; 
            background: #f8f9fa; 
            border-radius: 4px;
        }
        .btn { 
            background: #007bff; 
            color: white; 
            border: none; 
            padding: 8px 16px; 
            border-radius: 4px; 
            cursor: pointer; 
            text-decoration: none;
            font-size: 12px;
        }
        .btn:hover { background: #0056b3; }
        .btn-danger { background: #dc3545; }
        .btn-danger:hover { background: #c82333; }
        .refresh-btn { 
            position: fixed; 
            top: 20px; 
            right: 20px; 
            background: #28a745;
        }
        .json-view { 
            background: #f8f9fa; 
            padding: 15px; 
            border-radius: 4px; 
            font-family: 'Courier New', monospace; 
            font-size: 12px;
            max-height: 400px; 
            overflow-y: auto;
        }
        .loading {
            text-align: center;
            color: #666;
            font-style: italic;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎯 Alignment Server Admin Dashboard</h1>
            <p>Monitor and manage the Alignment game server</p>
        </div>

        <button class="btn refresh-btn" onclick="refreshDashboard()">🔄 Refresh</button>

        <div class="card">
            <h2>Server Status</h2>
            <div id="server-status" class="loading">Loading server status...</div>
        </div>

        <div class="card">
            <h2>Active Games</h2>
            <div id="games-list" class="loading">Loading games...</div>
        </div>

        <div class="card">
            <h2>Raw Server Stats</h2>
            <div id="raw-stats" class="json-view loading">Loading stats...</div>
        </div>
    </div>

    <script>
        let authHeader = '';

        // Prompt for admin credentials on load
        function authenticate() {
            const username = prompt('Admin Username:');
            const password = prompt('Admin Password:');
            if (username && password) {
                authHeader = 'Basic ' + btoa(username + ':' + password);
                return true;
            }
            return false;
        }

        function makeAuthenticatedRequest(url) {
            return fetch(url, {
                headers: {
                    'Authorization': authHeader
                }
            });
        }

        async function loadServerStatus() {
            try {
                const response = await makeAuthenticatedRequest('/admin/status');
                if (response.status === 401) {
                    if (authenticate()) {
                        return loadServerStatus(); // Retry with new credentials
                    } else {
                        throw new Error('Authentication required');
                    }
                }
                
                if (!response.ok) {
                    throw new Error('Network response was not ok');
                }
                
                const data = await response.json();
                
                // Update server status
                const statusElement = document.getElementById('server-status');
                const healthClass = data.health === 'HEALTHY' ? 'status-healthy' : 'status-overloaded';
                statusElement.innerHTML = '<span class="' + healthClass + '">' + data.health + '</span>';
                
                // Update games list
                const gamesElement = document.getElementById('games-list');
                if (data.active_games && data.active_games.length > 0) {
                    let gamesHTML = '<div class="game-list">';
                    data.active_games.forEach(game => {
                        gamesHTML += '<div class="game-item">';
                        gamesHTML += '<div>';
                        gamesHTML += '<strong>' + game.game_id + '</strong><br>';
                        gamesHTML += 'Players: ' + game.player_count + ' | Status: ' + game.status;
                        gamesHTML += '</div>';
                        gamesHTML += '<div>';
                        gamesHTML += '<a href="#" class="btn" onclick="viewGameState(\'' + game.game_id + '\')">View State</a> ';
                        gamesHTML += '<a href="#" class="btn btn-danger" onclick="terminateGame(\'' + game.game_id + '\')">Terminate</a>';
                        gamesHTML += '</div>';
                        gamesHTML += '</div>';
                    });
                    gamesHTML += '</div>';
                    gamesElement.innerHTML = gamesHTML;
                } else {
                    gamesElement.innerHTML = '<p>No active games</p>';
                }
                
                // Update raw stats
                document.getElementById('raw-stats').textContent = JSON.stringify(data.stats, null, 2);
                
            } catch (error) {
                console.error('Error loading server status:', error);
                document.getElementById('server-status').innerHTML = '<span style="color: red;">Error: ' + error.message + '</span>';
            }
        }

        async function viewGameState(gameId) {
            try {
                const response = await makeAuthenticatedRequest('/admin/games/' + gameId + '/state');
                if (!response.ok) {
                    throw new Error('Failed to fetch game state');
                }
                const state = await response.json();
                
                // Open a new window with the game state
                const newWindow = window.open('', '_blank', 'width=800,height=600');
                newWindow.document.write('<html><head><title>Game State: ' + gameId + '</title></head><body>');
                newWindow.document.write('<h1>Game State: ' + gameId + '</h1>');
                newWindow.document.write('<pre style="font-family: monospace; font-size: 12px; background: #f5f5f5; padding: 15px;">' + JSON.stringify(state, null, 2) + '</pre>');
                newWindow.document.write('</body></html>');
                newWindow.document.close();
            } catch (error) {
                alert('Error fetching game state: ' + error.message);
            }
        }

        async function terminateGame(gameId) {
            if (!confirm('Are you sure you want to terminate game ' + gameId + '? This action cannot be undone.')) {
                return;
            }

            try {
                const response = await fetch('/admin/games/' + gameId, {
                    method: 'DELETE',
                    headers: {
                        'Authorization': authHeader
                    }
                });
                
                if (!response.ok) {
                    throw new Error('Failed to terminate game');
                }
                
                alert('Game ' + gameId + ' has been terminated');
                refreshDashboard();
            } catch (error) {
                alert('Error terminating game: ' + error.message);
            }
        }

        function refreshDashboard() {
            loadServerStatus();
        }

        // Auto-refresh every 10 seconds
        setInterval(refreshDashboard, 10000);

        // Load data on page load
        document.addEventListener('DOMContentLoaded', function() {
            if (authenticate()) {
                loadServerStatus();
            }
        });
    </script>
</body>
</html>
`

// ServeAdminUI serves the admin dashboard HTML
func (h *Handlers) ServeAdminUI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(AdminDashboardHTML))
}