<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { Terminal } from 'xterm';
  import { FitAddon } from 'xterm-addon-fit';
  import 'xterm/css/xterm.css';

  // 🛡️ SLA: Strict Props for Traceability
  export let socketUrl: string;
  export let traceId: string;

  let terminalElement: HTMLDivElement;
  let term: Terminal;
  let fitAddon: FitAddon;
  let socket: WebSocket;

  // Standard ANSI colors for Kari's branding
  const theme = {
    background: '#1e1e1e',
    foreground: '#d4d4d4',
    cursor: '#ffffff',
    selection: '#3a3d41',
  };

  onMount(() => {
    // 1. Initialize Xterm.js
    term = new Terminal({
      cursorBlink: true,
      fontSize: 14,
      fontFamily: '"Cascadia Code", Menlo, monospace',
      theme: theme,
      convertEol: true, // 🛡️ SLA: Handles Rust's \n as \r\n automatically
    });

    fitAddon = new FitAddon();
    term.loadAddon(fitAddon);
    term.open(terminalElement);
    fitAddon.fit();

    // 2. Establish WebSocket Connection (The Pipe)
    socket = new WebSocket(socketUrl);

    socket.onopen = () => {
      term.writeln(`\x1b[1;32m🚀 Connected to Build Stream [${traceId}]\x1b[0m`);
      term.writeln(`\x1b[33mWaiting for logs...\x1b[0m\r\n`);
    };

    socket.onmessage = (event) => {
      // 🛡️ Handle Special Build Signals from Go Brain
      if (event.data === '🏗️ BUILD_COMPLETE') {
        term.writeln(`\r\n\x1b[1;32m✅ Build Successful. Service Restarted.\x1b[0m`);
        return;
      }
      
      // Write raw ANSI output from Rust Agent
      term.write(event.data);
    };

    socket.onerror = (error) => {
      term.writeln(`\r\n\x1b[1;31m🚨 Connection Error: Unable to reach Go Brain\x1b[0m`);
    };

    socket.onclose = () => {
      term.writeln(`\r\n\x1b[1;33m🔌 Stream Closed.\x1b[0m`);
    };

    // Handle responsiveness
    window.addEventListener('resize', handleResize);
  });

  const handleResize = () => {
    fitAddon?.fit();
  };

  onDestroy(() => {
    window.removeEventListener('resize', handleResize);
    socket?.close();
    term?.dispose();
  });
</script>

<div class="terminal-container">
  <div bind:this={terminalElement} class="terminal-view"></div>
</div>

<style>
  .terminal-container {
    width: 100%;
    height: 400px;
    background: #1e1e1e;
    border-radius: 8px;
    padding: 10px;
    box-shadow: 0 4px 6px -1px rgb(0 0 0 / 0.1);
    overflow: hidden;
  }

  .terminal-view {
    height: 100%;
    width: 100%;
  }

  /* 🛡️ Customizing xterm scrollbar for Kari UX */
  :global(.xterm-viewport::-webkit-scrollbar) {
    width: 8px;
  }
  :global(.xterm-viewport::-webkit-scrollbar-thumb) {
    background: #3a3d41;
    border-radius: 4px;
  }
</style>
