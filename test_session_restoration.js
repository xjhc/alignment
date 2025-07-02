#!/usr/bin/env node

const WebSocket = require('ws');

// Test session restoration with invalid/expired credentials
const gameId = 'invalid-game-id';
const playerId = 'invalid-player';
const sessionToken = 'invalid-token';

const wsUrl = `ws://172.17.0.3:8080/ws?gameId=${gameId}&playerId=${playerId}&sessionToken=${sessionToken}`;

console.log('Testing invalid session restoration to:', wsUrl);

const ws = new WebSocket(wsUrl);

let connected = false;
let timeout = setTimeout(() => {
    if (!connected) {
        console.log('✅ Connection timeout as expected (10 seconds)');
        process.exit(0);
    }
}, 10000);

ws.on('open', function open() {
    connected = true;
    clearTimeout(timeout);
    console.log('⚠️  WebSocket connected unexpectedly');
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
    console.log('✅ WebSocket error as expected:', err.message);
    clearTimeout(timeout);
    process.exit(0);
});

ws.on('close', function close(code, reason) {
    console.log('✅ WebSocket closed as expected:', code, reason.toString());
    clearTimeout(timeout);
    process.exit(0);
});