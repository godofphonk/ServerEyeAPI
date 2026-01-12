const WebSocket = require('ws');

// Connect to WebSocket
const ws = new WebSocket('ws://localhost:8080/ws');

ws.on('open', function open() {
    console.log('Connected to WebSocket');
    
    // Authenticate as agent
    ws.send(JSON.stringify({
        type: "auth",
        server_id: "srv_ab676f04f3cb81c7",
        server_key: "key_b1cdafa453b9c70945d81a5be6372421"
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
    
    const testServerId = 'srv_ab676f04f3cb81c7';
    const headers = {
        'Authorization': 'Bearer srv_ab676f04f3cb81c7:key_b1cdafa453b9c70945d81a5be6372421'
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
