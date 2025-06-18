import { useEffect, useRef, useState } from 'react';
import { Terminal as XTerm } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import { WebLinksAddon } from '@xterm/addon-web-links';
import { io, Socket } from 'socket.io-client';
import '@xterm/xterm/css/xterm.css';

interface TerminalProps {
  repoPath?: string;
  localPath?: string;
  onClose: () => void;
}

export function Terminal({ repoPath, localPath, onClose }: TerminalProps) {
  const terminalRef = useRef<HTMLDivElement>(null);
  const terminalInstance = useRef<XTerm | null>(null);
  const fitAddon = useRef<FitAddon | null>(null);
  const socketRef = useRef<Socket | null>(null);
  const [isConnected, setIsConnected] = useState(false);

  useEffect(() => {
    if (!terminalRef.current) return;

    // Create terminal instance
    const term = new XTerm({
      cursorBlink: true,
      theme: {
        background: '#1e1e1e',
        foreground: '#ffffff',
        cursor: '#ffffff',
        cursorAccent: '#000000',
        selectionBackground: '#3a3a3a',
        black: '#000000',
        red: '#cd3131',
        green: '#0dbc79',
        yellow: '#e5e510',
        blue: '#2472c8',
        magenta: '#bc3fbc',
        cyan: '#11a8cd',
        white: '#e5e5e5',
        brightBlack: '#666666',
        brightRed: '#f14c4c',
        brightGreen: '#23d18b',
        brightYellow: '#f5f543',
        brightBlue: '#3b8eea',
        brightMagenta: '#d670d6',
        brightCyan: '#29b8db',
        brightWhite: '#e5e5e5'
      },
      fontSize: 14,
      fontFamily: 'Monaco, Menlo, "Ubuntu Mono", "Courier New", monospace',
      rows: 24,
      cols: 80
    });

    // Add addons
    const fit = new FitAddon();
    const webLinks = new WebLinksAddon();
    
    term.loadAddon(fit);
    term.loadAddon(webLinks);

    // Open terminal
    term.open(terminalRef.current);
    fit.fit();

    terminalInstance.current = term;
    fitAddon.current = fit;

    // Connect to WebSocket
    const socket = io('http://localhost:8080');
    socketRef.current = socket;

    socket.on('connect', () => {
      console.log('Connected to terminal server');
      setIsConnected(true);
      
      // Initialize terminal session
      socket.emit('terminal_init', {
        workingDirectory: localPath
      });
    });

    socket.on('terminal_ready', () => {
      console.log('Terminal session ready');
    });

    socket.on('terminal_output', (data) => {
      term.write(data.data);
    });

    socket.on('terminal_error', (data) => {
      term.writeln(`\r\nError: ${data.message}`);
    });

    socket.on('terminal_closed', (data) => {
      term.writeln(`\r\nTerminal closed with code: ${data.code}`);
      setIsConnected(false);
    });

    socket.on('disconnect', () => {
      console.log('Disconnected from terminal server');
      setIsConnected(false);
    });

    // Handle input
    term.onData((data) => {
      if (socket.connected) {
        socket.emit('terminal_command', { command: data });
      }
    });

    // Handle resize
    const handleResize = () => {
      if (fitAddon.current) {
        fitAddon.current.fit();
        if (socket.connected) {
          const { cols, rows } = term;
          socket.emit('terminal_resize', { cols, rows });
        }
      }
    };
    window.addEventListener('resize', handleResize);

    return () => {
      window.removeEventListener('resize', handleResize);
      if (socket) {
        socket.disconnect();
      }
      term.dispose();
      terminalInstance.current = null;
      fitAddon.current = null;
      socketRef.current = null;
    };
  }, [localPath]);

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-xl w-11/12 max-w-4xl h-3/4 max-h-[600px] flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-gray-200">
          <div className="flex items-center space-x-2">
            <div className="w-3 h-3 bg-red-500 rounded-full"></div>
            <div className="w-3 h-3 bg-yellow-500 rounded-full"></div>
            <div className="w-3 h-3 bg-green-500 rounded-full"></div>
            <span className="ml-4 font-medium text-gray-700">
              Terminal {(localPath || repoPath) && `- ${(localPath || repoPath)?.split('/').pop()}`}
            </span>
          </div>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600 text-xl font-bold"
          >
            ×
          </button>
        </div>

        {/* Terminal content */}
        <div className="flex-1 p-4 bg-gray-900">
          <div 
            ref={terminalRef} 
            className="w-full h-full"
            style={{ height: '100%' }}
          />
        </div>

        {/* Status bar */}
        <div className="px-4 py-2 bg-gray-100 border-t border-gray-200 text-sm text-gray-600">
          <div className="flex items-center justify-between">
            <span>
              Status: {isConnected ? 'Connected' : 'Connecting...'}
            </span>
            {(localPath || repoPath) && (
              <span>Path: {localPath || repoPath}</span>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}