const WebSocket = require('ws');

// Connect to WebSocket
const ws = new WebSocket('ws://localhost:8080/ws');

ws.on('open', function open() {
    console.log('Connected to WebSocket');
    
    // Authenticate as agent
    ws.send(JSON.stringify({
        type: "auth",
        server_id: "srv_llNfRjLdF7zb",
        server_key: "key_Xp1tVnPhJbD5H9B3vXpRjVnP"
    }));
});

ws.on('message', function message(data) {
    const msg = JSON.parse(data);
    console.log('Received:', msg);
    
    // After auth success, send metrics
    if (msg.type === 'auth_success') {
        console.log('Authentication successful, sending metrics...');
        
        // Send metrics
        ws.send(JSON.stringify({
            type: "metrics",
            server_id: "srv_llNfRjLdF7zb",
            server_key: "key_Xp1tVnPhJbD5H9B3vXpRjVnP",
            data: {
                cpu: 45.2,
                memory: 67.8,
                disk: 23.1
            }
        }));
        
        // Send heartbeat
        ws.send(JSON.stringify({
            type: "heartbeat",
            server_id: "srv_llNfRjLdF7zb",
            server_key: "key_Xp1tVnPhJbD5H9B3vXpRjVnP"
        }));
    }
});

ws.on('error', function error(err) {
    console.error('WebSocket error:', err);
});

ws.on('close', function close() {
    console.log('WebSocket disconnected');
});

// Test API endpoints after a delay
setTimeout(() => {
    console.log('\n=== Testing API endpoints ===');
    
    const http = require('http');
    
    // Get server metrics
    http.get('http://localhost:8080/api/servers/srv_llNfRjLdF7zb/metrics', (res) => {
        let data = '';
        res.on('data', chunk => data += chunk);
        res.on('end', () => {
            console.log('Metrics API:', JSON.parse(data));
        });
    });
    
    // Get server status
    http.get('http://localhost:8080/api/servers/srv_llNfRjLdF7zb/status', (res) => {
        let data = '';
        res.on('data', chunk => data += chunk);
        res.on('end', () => {
            console.log('Status API:', JSON.parse(data));
        });
    });
    
    // List all servers
    http.get('http://localhost:8080/api/servers', (res) => {
        let data = '';
        res.on('data', chunk => data += chunk);
        res.on('end', () => {
            console.log('Servers API:', JSON.parse(data));
        });
    });
    
    // Send command
    const postData = JSON.stringify({
        server_id: "srv_llNfRjLdF7zb",
        command: {
            restart: true,
            message: "Test command from API"
        }
    });
    
    const req = http.request({
        hostname: 'localhost',
        port: 8080,
        path: '/api/servers/srv_llNfRjLdF7zb/command',
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Content-Length': Buffer.byteLength(postData)
        }
    }, (res) => {
        let data = '';
        res.on('data', chunk => data += chunk);
        res.on('end', () => {
            console.log('Command API:', JSON.parse(data));
        });
    });
    
    req.write(postData);
    req.end();
    
}, 3000);
