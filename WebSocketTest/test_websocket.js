const WebSocket = require('ws');

// Connect to WebSocket
const ws = new WebSocket('ws://localhost:8080/ws');

ws.on('open', function open() {
    console.log('Connected to WebSocket');
    
    // Authenticate as agent
    ws.send(JSON.stringify({
        type: "auth",
        server_id: "srv_46KekYs6AeIm",
        server_key: "key_WKeSw0E8MgUy2GaOiWA4IcQk"
    }));
});

ws.on('message', function message(data) {
    const msg = JSON.parse(data);
    console.log('Received:', msg);
    
    // After auth success, send metrics
    if (msg.type === 'auth_success') {
        console.log('Authentication successful, sending metrics...');
        
        // Send metrics with new strict typing structure
        ws.send(JSON.stringify({
            type: "metrics",
            data: {
                server_id: "srv_46KekYs6AeIm",
                metrics: {
                    cpu: 45.2,
                    memory: 67.8,
                    disk: 23.1,
                    network: 120.5,
                    time: new Date().toISOString()
                },
                system: {
                    os: "Ubuntu 22.04",
                    architecture: "x86_64",
                    kernel: "5.15.0",
                    uptime: 86400,
                    hostname: "test-server"
                }
            }
        }));
        
        // Send heartbeat
        ws.send(JSON.stringify({
            type: "heartbeat",
            timestamp: Math.floor(Date.now() / 1000)
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
    http.get('http://localhost:8080/api/servers/srv_46KekYs6AeIm/metrics', (res) => {
        let data = '';
        res.on('data', chunk => data += chunk);
        res.on('end', () => {
            console.log('Metrics API:', JSON.parse(data));
        });
    });
    
    // Get server status
    http.get('http://localhost:8080/api/servers/srv_46KekYs6AeIm/status', (res) => {
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
        server_id: "srv_46KekYs6AeIm",
        command: {
            restart: true,
            message: "Test command from API"
        }
    });
    
    const req = http.request({
        hostname: 'localhost',
        port: 8080,
        path: '/api/servers/srv_46KekYs6AeIm/command',
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
