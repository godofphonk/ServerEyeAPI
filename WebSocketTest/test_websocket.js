const WebSocket = require('ws');

// Connect to WebSocket
const ws = new WebSocket('ws://localhost:8080/ws');

ws.on('open', function open() {
    console.log('Connected to WebSocket');
    
    // Authenticate as agent
    ws.send(JSON.stringify({
        type: "auth",
        server_id: "srv_a4d02892b695b53c",
        server_key: "key_2f6165986f84cd41b5cd9176003e2d2a"
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
            data: {
                cpu: 45.2,
                memory: 67.8,
                disk: 23.4,
                network: {
                    bytes_sent: 1024000,
                    bytes_recv: 2048000
                },
                timestamp: new Date().toISOString()
            }
        }));
        
        // Send heartbeat
        ws.send(JSON.stringify({
            type: "heartbeat",
            timestamp: Math.floor(Date.now() / 1000)
        }));
        
        // Test API endpoints
        testAPIEndpoints();
    }
});

ws.on('error', function error(err) {
    console.error('WebSocket error:', err);
});

ws.on('close', function close() {
    console.log('WebSocket disconnected');
});

// Test API endpoints
function testAPIEndpoints() {
    console.log('\n=== Testing API endpoints ===');
    
    const testServerId = 'srv_a4d02892b695b53c';
    const headers = {
        'Authorization': 'Bearer srv_a4d02892b695b53c:key_2f6165986f84cd41b5cd9176003e2d2a'
    };
    
    // Test metrics API
    fetch(`http://localhost:8080/api/servers/${testServerId}/metrics`, { headers })
        .then(res => res.json())
        .then(data => console.log('Metrics API:', data))
        .catch(err => console.error('Metrics API error:', err));
    
    // Test servers API
    fetch('http://localhost:8080/api/servers', { headers })
        .then(res => res.json())
        .then(data => console.log('Servers API:', data))
        .catch(err => console.error('Servers API error:', err));
    
    // Test status API
    fetch(`http://localhost:8080/api/servers/${testServerId}/status`, { headers })
        .then(res => res.json())
        .then(data => console.log('Status API:', data))
        .catch(err => console.error('Status API error:', err));
    
    // Test command API
    fetch(`http://localhost:8080/api/servers/${testServerId}/command`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            ...headers
        },
        body: JSON.stringify({
            command: {
                message: 'Test command from API',
                restart: true,
                timestamp: Math.floor(Date.now() / 1000)
            }
        })
    })
        .then(res => res.json())
        .then(data => console.log('Command API:', data))
        .catch(err => console.error('Command API error:', err));
}
