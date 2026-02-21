<script lang="ts">
    import { onMount, onDestroy } from 'svelte';
    import { fade } from 'svelte/transition';
    import { Terminal } from 'xterm';
    import { FitAddon } from 'xterm-addon-fit';
    
    // Core CSS for canvas rendering
    import 'xterm/css/xterm.css';

    // Props
    export let traceId: string;
    
    // Component State
    let terminalElement: HTMLDivElement;
    let terminal: Terminal;
    let fitAddon: FitAddon;
    let eventSource: EventSource | null = null;
    
    let status: 'connecting' | 'streaming' | 'completed' | 'error' = 'connecting';
    let autoscroll = true;

    // 🛡️ Zero-Trust: Validate traceId before opening the stream
    const isValidTraceId = (id: string) => /^[a-f0-9-]{36}$/i.test(id) || id.length > 5;

    onMount(() => {
        if (!isValidTraceId(traceId)) {
            status = 'error';
            return;
        }

        // 1. Initialize xterm.js with Kari Brand Identity
        terminal = new Terminal({
            fontFamily: '"IBM Plex Mono", monospace',
            fontSize: 13,
            lineHeight: 1.4,
            cursorBlink: true,
            disableStdin: true,
            theme: {
                background: '#1A1A1C',   // Deep Gray
                foreground: '#F4F5F6',   // Off-white
                cursor: '#1BA8A0',       // Kari Teal
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

        // 🛡️ SLA: UI Responsiveness
        const resizeObserver = new ResizeObserver(() => fitAddon.fit());
        resizeObserver.observe(terminalElement);

        // 2. 📡 Connect to the Go Brain SSE Hub
        const url = `/api/deployments/${traceId}/logs/stream`;
        eventSource = new EventSource(url);

        terminal.writeln('\x1b[36m[Karı]\x1b[0m Synchronizing telemetry via Go Brain...');

        eventSource.onopen = () => {
            status = 'streaming';
            terminal.writeln('\x1b[32m[Karı]\x1b[0m Secure link established. Awaiting Muscle logs...\r\n');
        };

        eventSource.onmessage = (event) => {
            // Write payload directly to the canvas
            terminal.write(event.data);
            
            // 🛡️ SLA: Detect terminal conditions to update UI state
            if (event.data.includes("✅ Deployment successful")) {
                status = 'completed';
                terminal.writeln('\r\n\x1b[36m[Karı]\x1b[0m Pipeline finished successfully.');
            } else if (event.data.includes("❌ ERROR")) {
                status = 'error';
                terminal.writeln('\r\n\x1b[31m[Karı]\x1b[0m Pipeline aborted due to error.');
            }

            // Handle autoscroll if user hasn't scrolled up manually
            if (autoscroll) {
                terminal.scrollToBottom();
            }
        };

        eventSource.onerror = (err) => {
            if (status !== 'completed') {
                status = 'error';
                terminal.writeln('\r\n\x1b[31m[Karı]\x1b[0m Telemetry heartbeat lost.');
            }
            eventSource?.close();
        };

        return () => {
            resizeObserver.disconnect();
            if (eventSource) eventSource.close();
            if (terminal) terminal.dispose();
        };
    });

    // 🛡️ Logic to detect if user has scrolled up (manual investigation)
    function handleTerminalScroll() {
        if (!terminal) return;
        const buffer = terminal.buffer.active;
        // If the current viewport is not at the end of the scrollback buffer
        autoscroll = buffer.viewportY >= buffer.baseY;
    }

    onDestroy(() => {
        if (eventSource) eventSource.close();
        if (terminal) terminal.dispose();
    });
</script>

<div class="card flex flex-col h-[600px] w-full bg-white shadow-sm border border-kari-warm-gray/20 rounded-lg overflow-hidden">
    <div class="h-12 border-b border-kari-warm-gray/20 flex items-center justify-between px-4 shrink-0 bg-gray-50/50">
        <div class="flex items-center gap-3">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-kari-warm-gray" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
            <h3 class="font-sans font-semibold text-sm text-kari-text">Deployment Telemetry</h3>
            <span class="text-xs font-mono text-kari-warm-gray bg-kari-warm-gray/10 px-2 py-0.5 rounded">
                {traceId.slice(0, 8)} 
            </span>
        </div>

        <div class="flex items-center gap-4">
            {#if !autoscroll && status === 'streaming'}
                <button 
                    on:click={() => { autoscroll = true; terminal.scrollToBottom(); }}
                    class="text-[10px] font-bold text-[#1BA8A0] animate-bounce uppercase tracking-tighter"
                    transition:fade
                >
                    ⬇ Resync to Tail
                </button>
            {/if}

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
                    <span class="text-xs font-medium text-[#1BA8A0]">Live Feed</span>
                {:else if status === 'completed'}
                    <span class="h-2.5 w-2.5 rounded-full bg-green-500"></span>
                    <span class="text-xs font-medium text-kari-text">Completed</span>
                {:else if status === 'error'}
                    <span class="h-2.5 w-2.5 rounded-full bg-red-500"></span>
                    <span class="text-xs font-medium text-red-600">Sync Lost</span>
                {/if}
            </div>
        </div>
    </div>

    <div class="relative flex-1 bg-[#1A1A1C] p-3 overflow-hidden">
        <div 
            bind:this={terminalElement} 
            on:wheel={handleTerminalScroll}
            class="absolute inset-3"
        ></div>
    </div>
</div>

<style>
    :global(.xterm-viewport::-webkit-scrollbar) { width: 8px; }
    :global(.xterm-viewport::-webkit-scrollbar-track) { background: #1A1A1C; }
    :global(.xterm-viewport::-webkit-scrollbar-thumb) { background: #3F3F46; border-radius: 4px; }
    :global(.xterm-viewport::-webkit-scrollbar-thumb:hover) { background: #1BA8A0; }
</style>
