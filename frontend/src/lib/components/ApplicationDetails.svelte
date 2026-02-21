<script lang="ts">
    import { onMount, onDestroy } from 'svelte';
    import { fade, slide } from 'svelte/transition';
    import { 
        Shield, Globe, Cpu, Database, Terminal, 
        RefreshCw, Settings2, ExternalLink, Lock, Activity,
        AlertCircle
    } from 'lucide-svelte';
    import { createEventDispatcher } from 'svelte';

    // 🛡️ Props & Events
    export let app: any;
    const dispatch = createEventDispatcher();

    // 🛡️ State Management
    let isRedeploying = false;
    let metrics = { cpu: 0, ram: 0, status: 'stable' };
    let metricInterval: ReturnType<typeof setInterval>;

    // 🛡️ SLA: Reactive Metrics Polling
    onMount(() => {
        // In 2026, we fetch real-time jail usage directly from the Muscle via the Brain
        metricInterval = setInterval(async () => {
            try {
                const res = await fetch(`/api/v1/apps/${app.id}/metrics`);
                if (res.ok) metrics = await res.json();
            } catch (e) {
                metrics.status = 'stale';
            }
        }, 5000);
    });

    async function handleRedeploy() {
        if (isRedeploying) return;
        isRedeploying = true;

        try {
            // 📡 Triggering the Go Brain Orchestrator
            const response = await fetch(`/api/v1/apps/${app.id}/deploy`, { method: 'POST' });
            const result = await response.json();

            if (response.ok) {
                // ✅ SOLID: Hand off to the terminal orchestrator
                dispatch('initiate-logs', { traceId: result.trace_id });
            }
        } catch (err) {
            console.error("Redeploy failed:", err);
        } finally {
            isRedeploying = false;
        }
    }

    onDestroy(() => clearInterval(metricInterval));
</script>

<div class="space-y-6" in:fade={{ duration: 200 }}>
    <header class="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-white p-6 rounded-xl border border-kari-warm-gray/10 shadow-sm">
        <div class="flex items-center gap-4">
            <div class="p-3 bg-kari-teal/10 text-kari-teal rounded-xl">
                <Shield size={32} strokeWidth={1.5} />
            </div>
            <div>
                <div class="flex items-center gap-2">
                    <h2 class="text-2xl font-bold text-kari-text">{app.name}</h2>
                    <span class="px-2 py-0.5 rounded-full {metrics.status === 'stable' ? 'bg-emerald-50 text-emerald-600 border-emerald-100' : 'bg-amber-50 text-amber-600 border-amber-100'} text-[10px] font-bold uppercase border">
                        {metrics.status === 'stable' ? 'Isolated & Healthy' : 'Telemetry Lag'}
                    </span>
                </div>
                <p class="text-sm text-kari-warm-gray font-mono">ID: {app.id.split('-')[0]} • {app.environment}</p>
            </div>
        </div>

        <div class="flex items-center gap-3">
            <button 
                on:click={handleRedeploy}
                disabled={isRedeploying}
                class="flex items-center gap-2 px-4 py-2 bg-kari-teal text-white rounded-lg text-sm font-bold shadow-lg shadow-kari-teal/20 hover:bg-[#158e87] transition-all disabled:opacity-50"
            >
                <RefreshCw size={16} class={isRedeploying ? 'animate-spin' : ''} />
                {isRedeploying ? 'Orchestrating...' : 'Force Redeploy'}
            </button>
            <a href="https://{app.domain}" target="_blank" class="p-2 border border-kari-warm-gray/20 rounded-lg text-kari-warm-gray hover:text-kari-teal hover:bg-kari-teal/5 transition-all">
                <ExternalLink size={20} />
            </a>
        </div>
    </header>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div class="lg:col-span-2 space-y-6">
            <div class="bg-white p-6 rounded-xl border border-kari-warm-gray/10 shadow-sm">
                <div class="flex items-center justify-between mb-8">
                    <h3 class="text-sm font-bold text-kari-text uppercase tracking-widest flex items-center gap-2">
                        <Activity size={16} class="text-kari-teal" />
                        Live Jail Telemetry
                    </h3>
                    {#if metrics.status === 'stale'}
                        <span class="flex items-center gap-1 text-[10px] text-amber-600 font-bold uppercase">
                            <AlertCircle size={12} /> Connection Weak
                        </span>
                    {/if}
                </div>

                <div class="space-y-8">
                    <div>
                        <div class="flex justify-between text-[10px] font-black mb-2 uppercase tracking-tighter">
                            <span class="text-kari-warm-gray flex items-center gap-1"><Cpu size={12}/> CPU Usage</span>
                            <span class="text-kari-text">{metrics.cpu}% / {app.cpu_limit}%</span>
                        </div>
                        <div class="h-1.5 w-full bg-gray-100 rounded-full overflow-hidden">
                            <div class="h-full bg-kari-teal transition-all duration-1000" style="width: {(metrics.cpu / app.cpu_limit) * 100}%"></div>
                        </div>
                    </div>

                    <div>
                        <div class="flex justify-between text-[10px] font-black mb-2 uppercase tracking-tighter">
                            <span class="text-kari-warm-gray flex items-center gap-1"><Database size={12}/> Resident Memory</span>
                            <span class="text-kari-text">{metrics.ram}MB / {app.mem_limit}MB</span>
                        </div>
                        <div class="h-1.5 w-full bg-gray-100 rounded-full overflow-hidden">
                            <div class="h-full bg-indigo-500 transition-all duration-1000" style="width: {(metrics.ram / app.mem_limit) * 100}%"></div>
                        </div>
                    </div>
                </div>
            </div>
            
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div class="bg-white p-5 rounded-xl border border-kari-warm-gray/10">
                    <span class="text-[10px] font-bold text-kari-warm-gray uppercase mb-3 block">Network Ingress</span>
                    <div class="font-mono text-xs text-kari-text bg-gray-50 p-3 rounded-lg border border-gray-100">
                        {app.domain} ➜ :{app.target_port}
                    </div>
                </div>
                <div class="bg-white p-5 rounded-xl border border-kari-warm-gray/10">
                    <span class="text-[10px] font-bold text-kari-warm-gray uppercase mb-3 block">Process Strategy</span>
                    <div class="font-mono text-xs text-kari-text bg-gray-50 p-3 rounded-lg border border-gray-100">
                        {app.runtime === 'docker' ? 'OCI Container' : 'Systemd Jail'}
                    </div>
                </div>
            </div>
        </div>

        <div class="space-y-6">
            <div class="bg-kari-text text-white p-6 rounded-xl shadow-2xl relative overflow-hidden">
                <div class="relative z-10 space-y-4">
                    <h3 class="text-xs font-bold text-kari-warm-gray uppercase tracking-widest flex items-center gap-2">
                        <Lock size={14} class="text-kari-teal" /> Security Enforcement
                    </h3>
                    <div class="space-y-3">
                        <div class="text-[11px] leading-relaxed text-white/70">
                            <strong class="text-white">Cgroup v2:</strong> Resources are strictly throttled. The Muscle Agent will SIGKILL if memory exceeds limits.
                        </div>
                        <div class="text-[11px] leading-relaxed text-white/70">
                            <strong class="text-white">TLS 1.3:</strong> Edge-encrypted via Let's Encrypt.
                        </div>
                    </div>
                </div>
                <div class="absolute -right-4 -bottom-4 opacity-5 text-white">
                    <Shield size={120} />
                </div>
            </div>

            <button class="group w-full flex items-center justify-between px-5 py-3 bg-white border border-kari-warm-gray/20 rounded-xl text-sm font-bold text-kari-text hover:border-kari-teal transition-all">
                <div class="flex items-center gap-3">
                    <Settings2 size={18} class="text-kari-warm-gray group-hover:text-kari-teal" />
                    Environment
                </div>
                <span class="text-[10px] bg-gray-100 px-2 py-0.5 rounded uppercase">Encrypted</span>
            </button>
        </div>
    </div>
</div>
