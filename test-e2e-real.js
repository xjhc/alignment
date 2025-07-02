#!/usr/bin/env node

/**
 * Real E2E Test - Tests the complete flow against live servers
 * Run this when both frontend and backend are running
 */

const WebSocket = require('ws');

const SERVER_URL = 'http://172.17.0.3:8080';
const WS_URL = 'ws://172.17.0.3:8080/ws';

async function runE2ETest() {
    console.log('🧪 Starting Real E2E Test');
    console.log('📡 Server URL:', SERVER_URL);
    console.log('🔌 WebSocket URL:', WS_URL);
    
    let testResults = {
        authCheck: false,
        lobbyCreation: false,
        lobbyListing: false,
        websocketConnection: false,
        lobbyStateReceived: false,
    };
    
    try {
        // Test 1: Authentication Check
        console.log('\n1️⃣ Testing authentication check...');
        const authResponse = await fetch(`${SERVER_URL}/api/me`);
        if (authResponse.ok) {
            const authData = await authResponse.json();
            console.log('✅ Auth check passed:', authData);
            testResults.authCheck = true;
        } else {
            throw new Error(`Auth check failed: ${authResponse.status}`);
        }
        
        // Test 2: Lobby Creation
        console.log('\n2️⃣ Testing lobby creation...');
        const createRequest = {
            user_id: 'guest:e2e-test-' + Date.now(),
            player_name: 'E2E Test Player',
            lobby_name: 'E2E Test Lobby',
            player_avatar: '🤖',
        };
        
        const createResponse = await fetch(`${SERVER_URL}/api/games`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(createRequest),
        });
        
        if (!createResponse.ok) {
            throw new Error(`Lobby creation failed: ${createResponse.status}`);
        }
        
        const createData = await createResponse.json();
        console.log('✅ Lobby created:', createData);
        testResults.lobbyCreation = true;
        
        // Test 3: Lobby Listing
        console.log('\n3️⃣ Testing lobby listing...');
        const listResponse = await fetch(`${SERVER_URL}/api/games`);
        if (listResponse.ok) {
            const listData = await listResponse.json();
            console.log('✅ Lobby listing passed. Found', listData.lobbies.length, 'lobbies');
            testResults.lobbyListing = true;
        } else {
            throw new Error(`Lobby listing failed: ${listResponse.status}`);
        }
        
        // Test 4: WebSocket Connection
        console.log('\n4️⃣ Testing WebSocket connection...');
        const wsUrl = `${WS_URL}?gameId=${createData.game_id}&playerId=${createData.player_id}&sessionToken=${createData.session_token}`;
        
        await new Promise((resolve, reject) => {
            const ws = new WebSocket(wsUrl);
            let timeout = setTimeout(() => {
                reject(new Error('WebSocket connection timeout'));
            }, 10000);
            
            ws.on('open', () => {
                console.log('✅ WebSocket connected successfully');
                clearTimeout(timeout);
                testResults.websocketConnection = true;
                
                // Wait for lobby state update
                timeout = setTimeout(() => {
                    ws.close();
                    resolve();
                }, 2000);
            });
            
            ws.on('message', (data) => {
                try {
                    const event = JSON.parse(data);
                    console.log('📥 Received event:', event.type);
                    
                    if (event.type === 'LOBBY_STATE_UPDATE') {
                        console.log('✅ Lobby state received:', event.payload);
                        testResults.lobbyStateReceived = true;
                    }
                } catch (err) {
                    console.log('📥 Received raw message:', data.toString());
                }
            });
            
            ws.on('error', (error) => {
                clearTimeout(timeout);
                reject(error);
            });
            
            ws.on('close', () => {
                clearTimeout(timeout);
                resolve();
            });
        });
        
    } catch (error) {
        console.error('❌ E2E Test failed:', error.message);
        process.exit(1);
    }
    
    // Print Results
    console.log('\n📊 Test Results:');
    console.log('─'.repeat(40));
    Object.entries(testResults).forEach(([test, passed]) => {
        console.log(`${passed ? '✅' : '❌'} ${test}: ${passed ? 'PASSED' : 'FAILED'}`);
    });
    
    const passedTests = Object.values(testResults).filter(Boolean).length;
    const totalTests = Object.keys(testResults).length;
    
    console.log('─'.repeat(40));
    console.log(`🎯 Overall: ${passedTests}/${totalTests} tests passed`);
    
    if (passedTests === totalTests) {
        console.log('🎉 All E2E tests passed!');
        process.exit(0);
    } else {
        console.log('💥 Some E2E tests failed!');
        process.exit(1);
    }
}

// Run the test
runE2ETest();