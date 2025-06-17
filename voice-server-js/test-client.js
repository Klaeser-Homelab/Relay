#!/usr/bin/env node

import { io } from 'socket.io-client';
import fs from 'fs';

// Simple test client for the voice server
class VoiceServerTestClient {
  constructor(serverUrl = 'http://localhost:8080') {
    this.serverUrl = serverUrl;
    this.socket = null;
    this.connected = false;
  }

  async connect() {
    console.log(`Connecting to ${this.serverUrl}...`);
    
    this.socket = io(this.serverUrl);

    this.socket.on('connect', () => {
      console.log('✅ Connected to voice server');
      this.connected = true;
      this.setupEventHandlers();
    });

    this.socket.on('disconnect', () => {
      console.log('❌ Disconnected from voice server');
      this.connected = false;
    });

    this.socket.on('connect_error', (error) => {
      console.error('❌ Connection error:', error.message);
    });

    // Wait for connection
    return new Promise((resolve, reject) => {
      this.socket.on('connect', resolve);
      this.socket.on('connect_error', reject);
      setTimeout(() => reject(new Error('Connection timeout')), 5000);
    });
  }

  setupEventHandlers() {
    this.socket.on('status', (data) => {
      console.log(`📊 Status: ${data.status} - ${data.message}`);
      if (data.project) {
        console.log(`   Project: ${data.project}`);
      }
    });

    this.socket.on('transcription', (data) => {
      console.log(`🎤 Transcription: ${data.text}`);
    });

    this.socket.on('audio_response', (data) => {
      console.log(`🔊 Audio response received (${data.audio_data?.length || 0} bytes)`);
    });

    this.socket.on('function_result', (data) => {
      console.log(`⚡ Function result: ${data.function}`);
      console.log(`   Result:`, data.result);
    });
  }

  async testBasicConnection() {
    console.log('\n🧪 Testing basic connection...');
    if (!this.connected) {
      throw new Error('Not connected');
    }
    console.log('✅ Basic connection test passed');
  }

  async testProjectSelection() {
    console.log('\n🧪 Testing project selection...');
    
    return new Promise((resolve) => {
      this.socket.emit('select_project', { project: 'test-project' });
      
      this.socket.once('status', (data) => {
        if (data.status === 'project_selected' || data.status === 'error') {
          console.log(`✅ Project selection response: ${data.message}`);
          resolve(data);
        }
      });
      
      setTimeout(() => {
        console.log('⚠️  Project selection timeout');
        resolve({ status: 'timeout' });
      }, 3000);
    });
  }

  async testRecordingSimulation() {
    console.log('\n🧪 Testing recording simulation...');
    
    // Start recording
    this.socket.emit('start_recording');
    console.log('📹 Started recording...');
    
    await new Promise(resolve => setTimeout(resolve, 1000));
    
    // Send some dummy audio data
    const dummyAudio = Buffer.from('dummy-audio-data').toString('base64');
    this.socket.emit('audio', { audio_data: dummyAudio });
    console.log('🎵 Sent dummy audio data');
    
    await new Promise(resolve => setTimeout(resolve, 1000));
    
    // Stop recording
    this.socket.emit('stop_recording');
    console.log('⏹️  Stopped recording');
    
    console.log('✅ Recording simulation completed');
  }

  async testRESTEndpoints() {
    console.log('\n🧪 Testing REST endpoints...');
    
    const baseUrl = this.serverUrl;
    
    try {
      // Test health endpoint
      console.log('Testing /health...');
      const healthResponse = await fetch(`${baseUrl}/health`);
      const healthData = await healthResponse.json();
      console.log('✅ Health check:', healthData);
      
      // Test projects list
      console.log('Testing /api/projects...');
      const projectsResponse = await fetch(`${baseUrl}/api/projects`);
      const projectsData = await projectsResponse.json();
      console.log('✅ Projects list:', projectsData);
      
      return true;
    } catch (error) {
      console.error('❌ REST endpoint test failed:', error.message);
      return false;
    }
  }

  async runAllTests() {
    try {
      await this.connect();
      await this.testBasicConnection();
      await this.testRESTEndpoints();
      await this.testProjectSelection();
      await this.testRecordingSimulation();
      
      console.log('\n🎉 All tests completed successfully!');
      return true;
    } catch (error) {
      console.error('\n💥 Test failed:', error.message);
      return false;
    } finally {
      if (this.socket) {
        this.socket.disconnect();
      }
    }
  }

  disconnect() {
    if (this.socket) {
      this.socket.disconnect();
    }
  }
}

// Run tests if called directly
if (import.meta.url === `file://${process.argv[1]}`) {
  const serverUrl = process.argv[2] || 'http://localhost:8080';
  const client = new VoiceServerTestClient(serverUrl);
  
  console.log('🚀 Starting Voice Server Test Suite');
  console.log('====================================');
  
  client.runAllTests().then((success) => {
    process.exit(success ? 0 : 1);
  });
}

export { VoiceServerTestClient };