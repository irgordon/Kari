<script lang="ts">
    import { fade, slide } from 'svelte/transition';
    import { 
        Shield, 
        Globe, 
        Cpu, 
        Database, 
        Terminal, 
        RefreshCw, 
        Settings2,
        ExternalLink,
        Lock,
        Activity
    } from 'lucide-svelte';

    // 🛡️ Props: Expecting a full Application domain object
    export let app: any;
    
    // Internal State
    let isRedeploying = false;

    async function triggerRedeploy() {
        isRedeploying = true;
        // Logic to hit /api/v1/apps/{id}/deploy will go here
        setTimeout(() => isRedeploying = false, 2000); 
    }
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
                    <span class="px-2 py-0.5 rounded-full bg-emerald-50 text-emerald-600 text-[10px] font-bold uppercase border border-emerald-100">
                        Isolated
                    </span>
                </div>
                <p class="text-sm text-kari-warm-gray font-mono">Jail ID: {app.id.split('-')[0]}</p>
            </div>
        </div>

        <div class="flex items-center gap-3">
            <button 
                on:click={triggerRedeploy}
                disabled={isRedeploying}
                class="flex items-center gap-2 px-4 py-2 bg-kari-teal text-white rounded-lg text-sm font-bold shadow-lg shadow-kari-teal/20 hover:bg-[#158e87] transition-all disabled:opacity-50"
            >
                <RefreshCw size={16} class={isRedeploying ? 'animate-spin' : ''} />
                Force Redeploy
            </button>
            <a 
                href="https://{app.domain}" 
                target="_blank"
                class="p-2 border border-kari-warm-gray/20 rounded-lg text-kari-warm-gray hover:text-kari-teal hover:bg-kari-teal/5 transition-all"
            >
                <ExternalLink size={20} />
            </a>
        </div>
    </header>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div class="lg:col-span-2 space-y-6">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div class="bg-white p-5 rounded-xl border border-kari-warm-gray/10 shadow-sm">
                    <div class="flex items-center gap-3 mb-4 text-kari-warm-gray">
                        <Globe size={18} />
                        <span class="text-xs font-bold uppercase tracking-widest">Proxy Config</span>
                    </div>
                    <div class="space-y-2">
                        <div class="flex justify-between text-sm">
                            <span class="text-kari-warm-gray">Public Domain</span>
                            <span class="font-mono text-kari-text">{app.domain}</span>
                        </div>
                        <div class="flex justify-between text-sm">
                            <span class="text-kari-warm-gray">Internal Port</span>
                            <span class="font-mono text-kari-text">{app.target_port}</span>
                        </div>
                    </div>
                </div>

                <div class="bg-white p-5 rounded-xl border border-kari-warm-gray/10 shadow-sm">
                    <div class="flex items-center gap-3 mb-4 text-kari-warm-gray">
                        <Terminal size={18} />
                        <span class="text-xs font-bold uppercase tracking-widest">Git Source</span>
                    </div>
                    <div class="space-y-2">
                        <div class="flex justify-between text-sm">
                            <span class="text-kari-warm-gray">Branch</span>
                            <span class="px-2 py-0.5 bg-gray-100 rounded text-[10px] font-mono">{app.branch}</span>
                        </div>
                        <div class="flex justify-between text-sm">
                            <span class="text-kari-warm-gray">Last Commit</span>
                            <span class="font-mono text-kari-text">#a1b2c3d</span>
                        </div>
                    </div>
                </div>
            </div>

            <div class="bg-white p-6 rounded-xl border border-kari-warm-gray/10 shadow-sm">
                <h3 class="text-sm font-bold text-kari-text uppercase tracking-widest mb-6 flex items-center gap-2">
                    <Activity size={16} class="text-kari-teal" />
                    Jail Performance
                </h3>
                <div class="space-y-6">
                    <div>
                        <div class="flex justify-between text-xs font-bold mb-2">
                            <span class="text-kari-warm-gray uppercase">CPU Allocation</span>
                            <span class="text-kari-text">12% / 100%</span>
                        </div>
                        <div class="h-2 w-full bg-gray-100 rounded-full overflow-hidden">
                            <div class="h-full bg-kari-teal transition-all duration-1000" style="width: 12%"></div>
                        </div>
                    </div>
                    <div>
                        <div class="flex justify-between text-xs font-bold mb-2">
                            <span class="text-kari-warm-gray uppercase">Memory Resident</span>
                            <span class="text-kari-text">256MB / 1GB</span>
                        </div>
                        <div class="h-2 w-full bg-gray-100 rounded-full overflow-hidden">
                            <div class="h-full bg-indigo-500 transition-all duration-1000" style="width: 25%"></div>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <div class="space-y-6">
            <div class="bg-kari-text text-white p-6 rounded-xl shadow-xl">
                <h3 class="text-xs font-bold text-kari-warm-gray uppercase tracking-widest mb-4 flex items-center gap-2">
                    <Lock size={14} />
                    Security Posture
                </h3>
                <div class="space-y-4">
                    <div class="p-3 bg-white/5 border border-white/10 rounded-lg">
                        <p class="text-[10px] font-bold text-kari-teal uppercase mb-1">Cgroup Policy</p>
                        <p class="text-xs text-white/80 leading-relaxed">Hard-limit enforced by Rust Muscle Agent. No escape vectors detected.</p>
                    </div>
                    <div class="p-3 bg-white/5 border border-white/10 rounded-lg">
                        <p class="text-[10px] font-bold text-emerald-400 uppercase mb-1">SSL/TLS 1.3</p>
                        <p class="text-xs text-white/80 leading-relaxed">Auto-renewing certificate active via Let's Encrypt.</p>
                    </div>
                </div>
            </div>

            <button class="w-full flex items-center justify-center gap-2 py-3 border border-kari-warm-gray/20 rounded-lg text-sm font-bold text-kari-warm-gray hover:text-kari-text hover:bg-white transition-all">
                <Settings2 size={16} />
                Environment Variables
            </button>
        </div>
    </div>
</div>
