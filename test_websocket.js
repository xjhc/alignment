#!/usr/bin/env node

const WebSocket = require('ws');

// Test WebSocket connection with the credentials from lobby creation
const gameId = '06e0a829-4423-45cc-b2ff-793ea0cd4a24';
const playerId = 'guest:test123';
const sessionToken = 'f86103a56643df63d889b40c0093cd31e29ae706a89db443ce5b0f84f0d24e9e';

const wsUrl = `ws://172.17.0.3:8080/ws?gameId=${gameId}&playerId=${playerId}&sessionToken=${sessionToken}`;

console.log('Testing WebSocket connection to:', wsUrl);

const ws = new WebSocket(wsUrl);

let connected = false;
let timeout = setTimeout(() => {
    if (!connected) {
        console.log('❌ Connection timeout after 10 seconds');
        process.exit(1);
    }
}, 10000);

ws.on('open', function open() {
    connected = true;
    clearTimeout(timeout);
    console.log('✅ WebSocket connected successfully');
    
    // Send a test action
    const testAction = {
        type: 'PING',
        gameId: gameId,
        playerId: playerId,
        payload: {}
    };
    
    ws.send(JSON.stringify(testAction));
    console.log('📤 Sent test action:', testAction.type);
});

ws.on('message', function message(data) {
    try {
        const event = JSON.parse(data);
        console.log('📥 Received event:', event.type, event.payload);
    } catch (err) {
        console.log('📥 Received raw message:', data.toString());
    }
});

ws.on('error', function error(err) {
    console.log('❌ WebSocket error:', err.message);
    clearTimeout(timeout);
    process.exit(1);
});

ws.on('close', function close(code, reason) {
    console.log('🔌 WebSocket closed:', code, reason.toString());
    clearTimeout(timeout);
    process.exit(0);
});

// Clean exit after 5 seconds
setTimeout(() => {
    console.log('✅ Test completed, closing connection');
    ws.close();
}, 5000);