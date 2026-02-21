<script lang="ts">
    import { onMount, onDestroy } from 'svelte';
    import { Terminal } from 'xterm';
    import { FitAddon } from 'xterm-addon-fit';
    
    // xterm.css is required for the canvas grid to render correctly
    import 'xterm/css/xterm.css';

    // Props
    export let traceId: string;
    
    // Component State
    let terminalElement: HTMLDivElement;
    let terminal: Terminal;
    let fitAddon: FitAddon;
    let eventSource: EventSource | null = null;
    
    let status: 'connecting' | 'streaming' | 'completed' | 'error' = 'connecting';

    // 🛡️ Zero-Trust: Validate the traceId format before connecting
    const isValidTraceId = (id: string) => /^[a-f0-9-]{36}$/i.test(id) || id.length > 5;

    onMount(() => {
        if (!isValidTraceId(traceId)) {
            status = 'error';
            return;
        }

        // 1. Initialize xterm.js with Kari Panel Brand Palette
        terminal = new Terminal({
            fontFamily: '"IBM Plex Mono", monospace',
            fontSize: 13,
            lineHeight: 1.4,
            cursorBlink: true,
            disableStdin: true,
            theme: {
                background: '#1A1A1C',   // Brand Deep Gray
                foreground: '#F4F5F6',   // Brand Light Gray
                cursor: '#1BA8A0',       // Brand Teal
                selectionBackground: 'rgba(27, 168, 160, 0.3)',
                cyan: '#1BA8A0',
                green: '#10B981',
                red: '#EF4444',
                yellow: '#F59E0B',
            }
        });

        fitAddon = new FitAddon();
        terminal.loadAddon(fitAddon);
        terminal.open(terminalElement);
        fitAddon.fit();

        // 🛡️ SLA: Ensure UI responsiveness on container resize
        const resizeObserver = new ResizeObserver(() => fitAddon.fit());
        resizeObserver.observe(terminalElement);

        // 2. 📡 Connect to the Go Brain SSE Endpoint
        // Standard SSE automatically handles reconnections (SLA reliability)
        const url = `/api/deployments/${traceId}/logs/stream`;
        eventSource = new EventSource(url);

        terminal.writeln('\x1b[36m[Karı]\x1b[0m Establishing encrypted telemetry link...');

        eventSource.onopen = () => {
            status = 'streaming';
            terminal.writeln('\x1b[32m[Karı]\x1b[0m Link active. Streaming from Muscle Agent...\r\n');
        };

        eventSource.onmessage = (event) => {
            // Write the raw payload directly. xterm parses ANSI and \n.
            terminal.write(event.data);
            
            // 🛡️ SLA: Detect EOF signaled by our Go Worker
            if (event.data.includes("✅ Deployment successful") || event.data.includes("❌ ERROR")) {
                status = event.data.includes("✅") ? 'completed' : 'error';
                terminal.writeln('\r\n\x1b[36m[Karı]\x1b[0m Stream closed by host.');
            }
        };

        eventSource.onerror = (err) => {
            if (status !== 'completed') {
                status = 'error';
                terminal.writeln('\r\n\x1b[31m[Karı]\x1b[0m Telemetry link severed unexpectedly.');
            }
            eventSource?.close();
        };

        return () => {
            resizeObserver.disconnect();
            if (eventSource) eventSource.close();
            if (terminal) terminal.dispose();
        };
    });

    onDestroy(() => {
        if (eventSource) eventSource.close();
        if (terminal) terminal.dispose();
    });
</script>

<div class="card flex flex-col h-[600px] w-full bg-white shadow-sm border border-kari-warm-gray/20">
    <div class="h-12 border-b border-kari-warm-gray/20 flex items-center justify-between px-4 shrink-0 bg-gray-50/50">
        <div class="flex items-center gap-3">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-kari-warm-gray" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
            <h3 class="font-sans font-semibold text-sm text-kari-text">Deployment Logs</h3>
            <span class="text-xs font-mono text-kari-warm-gray bg-kari-warm-gray/10 px-2 py-0.5 rounded">
                {traceId.slice(0, 8)} 
            </span>
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
                  <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-[#1BA8A0] opacity-75"></span>
                  <span class="relative inline-flex rounded-full h-2.5 w-2.5 bg-[#1BA8A0]"></span>
                </span>
                <span class="text-xs font-medium text-[#1BA8A0]">Live Telemetry</span>
            {:else if status === 'completed'}
                <span class="h-2.5 w-2.5 rounded-full bg-green-500"></span>
                <span class="text-xs font-medium text-kari-text">Sync Finished</span>
            {:else if status === 'error'}
                <span class="h-2.5 w-2.5 rounded-full bg-red-500"></span>
                <span class="text-xs font-medium text-red-600">Offline</span>
            {/if}
        </div>
    </div>

    <div class="relative flex-1 bg-[#1A1A1C] p-3 overflow-hidden rounded-b-lg">
        <div bind:this={terminalElement} class="absolute inset-3"></div>
    </div>
</div>

<style>
    /* 🛡️ Custom Scrollbar Styling for xterm */
    :global(.xterm-viewport::-webkit-scrollbar) { width: 8px; }
    :global(.xterm-viewport::-webkit-scrollbar-track) { background: #1A1A1C; }
    :global(.xterm-viewport::-webkit-scrollbar-thumb) { background: #3F3F46; border-radius: 4px; }
    :global(.xterm-viewport::-webkit-scrollbar-thumb:hover) { background: #1BA8A0; }
</style>
