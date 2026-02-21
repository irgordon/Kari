<script lang="ts">
    import { onMount, onDestroy } from 'svelte';
    import { Terminal } from 'xterm';
    import { FitAddon } from 'xterm-addon-fit';
    import { TerminalStreamService, type LogChunk } from '$lib/api/terminalStream';
    
    // xterm.js requires its core CSS to render the canvas/DOM grid correctly
    import 'xterm/css/xterm.css';

    // Props
    export let traceId: string;
    
    // Component State
    let terminalElement: HTMLDivElement;
    let terminal: Terminal;
    let fitAddon: FitAddon;
    let streamService: TerminalStreamService;
    
    let status: 'connecting' | 'streaming' | 'completed' | 'error' = 'connecting';

    onMount(() => {
        // 1. Initialize the xterm.js instance with our Brand Colors
        terminal = new Terminal({
            fontFamily: '"IBM Plex Mono", monospace',
            fontSize: 13,
            lineHeight: 1.4,
            cursorBlink: true,
            disableStdin: true, // This is a read-only log stream
            theme: {
                // Invert the palette for the terminal: Dark background, Light text
                background: '#1A1A1C',       // Brand Text Color as the dark background
                foreground: '#F4F5F6',       // Brand Light Gray as text
                cursor: '#1BA8A0',           // Brand Primary Teal
                selectionBackground: 'rgba(27, 168, 160, 0.3)', // Translucent Teal
                // ANSI Color Overrides to make standard logs look beautiful
                cyan: '#1BA8A0',             // Map ANSI cyan to our Brand Teal
                green: '#10B981',            // Success green
                red: '#EF4444',              // Error red
                yellow: '#F59E0B',           // Warning yellow
            }
        });

        // 2. Attach the FitAddon so the terminal dynamically resizes to our flex container
        fitAddon = new FitAddon();
        terminal.loadAddon(fitAddon);
        terminal.open(terminalElement);
        fitAddon.fit();

        // Handle window resizing
        const resizeObserver = new ResizeObserver(() => {
            fitAddon.fit();
        });
        resizeObserver.observe(terminalElement);

        // 3. Initialize the WebSocket SLA Service
        streamService = new TerminalStreamService(traceId, {
            onOpen: () => {
                status = 'streaming';
                terminal.clear();
                terminal.writeln('\x1b[36m[Karı]\x1b[0m Connection established. Awaiting build logs...');
            },
            onMessage: (chunk: LogChunk) => {
                // Write the raw payload directly to xterm. 
                // xterm natively parses \r\n and ANSI color codes (\x1b[31m).
                terminal.write(chunk.message);
                
                if (chunk.is_eof) {
                    status = 'completed';
                    terminal.writeln('\r\n\x1b[36m[Karı]\x1b[0m Deployment pipeline finished.');
                }
            },
            onError: (err) => {
                status = 'error';
                console.error('Terminal stream error:', err);
            },
            onClose: (code, reason, wasClean) => {
                if (status !== 'completed') {
                    status = 'error';
                }
            }
        });

        // 4. Connect to the Go API
        streamService.connect();

        // Cleanup observer on unmount
        return () => resizeObserver.disconnect();
    });

    onDestroy(() => {
        // Prevent memory leaks by severing the WebSocket and destroying the xterm canvas
        if (streamService) streamService.disconnect();
        if (terminal) terminal.dispose();
    });
</script>

<div class="card flex flex-col h-[600px] w-full bg-white">
    <div class="h-12 border-b border-kari-warm-gray/20 flex items-center justify-between px-4 shrink-0 bg-gray-50/50">
        <div class="flex items-center gap-3">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-kari-warm-gray" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
            <h3 class="font-sans font-semibold text-sm text-kari-text">Deployment Logs</h3>
            <span class="text-xs font-mono text-kari-warm-gray bg-kari-warm-gray/10 px-2 py-0.5 rounded">
                {traceId.split('-')[0]} </span>
        </div>

        <div class="flex items-center gap-2">
            {#if status === 'connecting'}
                <span class="relative flex h-2.5 w-2.5">
                  <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-yellow-400 opacity-75"></span>
                  <span class="relative inline-flex rounded-full h-2.5 w-2.5 bg-yellow-500"></span>
                </span>
                <span class="text-xs font-medium text-kari-warm-gray">Connecting...</span>
            {:else if status === 'streaming'}
                <span class="relative flex h-2.5 w-2.5">
                  <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-kari-teal opacity-75"></span>
                  <span class="relative inline-flex rounded-full h-2.5 w-2.5 bg-kari-teal"></span>
                </span>
                <span class="text-xs font-medium text-kari-teal">Live</span>
            {:else if status === 'completed'}
                <span class="h-2.5 w-2.5 rounded-full bg-green-500"></span>
                <span class="text-xs font-medium text-kari-text">Completed</span>
            {:else if status === 'error'}
                <span class="h-2.5 w-2.5 rounded-full bg-red-500"></span>
                <span class="text-xs font-medium text-red-600">Connection Lost</span>
            {/if}
        </div>
    </div>

    <div class="relative flex-1 bg-[#1A1A1C] p-3 overflow-hidden rounded-b-lg">
        <div bind:this={terminalElement} class="absolute inset-3"></div>
    </div>
</div>

<style>
    /* Optional overrides to make xterm's scrollbar match our brand.
       Since the terminal background is dark, we use a subtle dark scrollbar. 
    */
    :global(.xterm-viewport::-webkit-scrollbar) {
        width: 8px;
    }
    :global(.xterm-viewport::-webkit-scrollbar-track) {
        background: #1A1A1C;
    }
    :global(.xterm-viewport::-webkit-scrollbar-thumb) {
        background: #8E8F93; /* Warm Gray */
        border-radius: 4px;
    }
    :global(.xterm-viewport::-webkit-scrollbar-thumb:hover) {
        background: #1BA8A0; /* Primary Teal */
    }
</style>
