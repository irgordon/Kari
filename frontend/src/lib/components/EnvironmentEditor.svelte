<script lang="ts">
    import { API } from '$lib/api/client';

    // Props
    export let appId: string;
    export let initialEnvVars: Record<string, string> = {};

    // Local State
    let isSaving = false;
    let saveMessage: { type: 'success' | 'error', text: string } | null = null;

    // Convert the dictionary into a reactive array of objects for easier DOM iteration
    // We generate a random local ID so Svelte's #each block can track DOM nodes efficiently
    let envList = Object.entries(initialEnvVars || {}).map(([key, value]) => ({
        id: Math.random().toString(36).substr(2, 9),
        key,
        value,
        show: false // Toggle for password visibility
    }));

    // ==============================================================================
    // Actions
    // ==============================================================================

    function addVariable() {
        envList = [...envList, { 
            id: Math.random().toString(36).substr(2, 9), 
            key: '', 
            value: '', 
            show: false 
        }];
    }

    function removeVariable(idToRemove: string) {
        envList = envList.filter(env => env.id !== idToRemove);
    }

    function toggleVisibility(idToToggle: string) {
        envList = envList.map(env => 
            env.id === idToToggle ? { ...env, show: !env.show } : env
        );
    }

    async function saveEnvironment() {
        if (isSaving) return;
        
        isSaving = true;
        saveMessage = null;

        // 1. Reconstruct the dictionary, stripping out empty keys to keep the DB clean
        const payload: Record<string, string> = {};
        for (const item of envList) {
            const cleanKey = item.key.trim();
            if (cleanKey !== '') {
                payload[cleanKey] = item.value;
            }
        }

        try {
            // 2. Transmit to the Go API. 
            // The API handles the AES-256-GCM encryption before saving to Postgres.
            await API.put(`/api/v1/applications/${appId}/env`, { env_vars: payload });
            
            saveMessage = { type: 'success', text: 'Environment variables securely updated and encrypted.' };
            
            // Auto-hide the success message after 3 seconds
            setTimeout(() => { saveMessage = null; }, 3000);
        } catch (error: any) {
            console.error('Failed to save env vars:', error);
            saveMessage = { type: 'error', text: error.message || 'Failed to save environment variables.' };
        } finally {
            isSaving = false;
        }
    }

    // Helper to enforce standard uppercase/underscore formatting for keys
    function handleKeyInput(event: Event, index: number) {
        const input = event.target as HTMLInputElement;
        // Strip spaces, dashes, and special chars, convert to uppercase
        envList[index].key = input.value.toUpperCase().replace(/[^A-Z0-9_]/g, '');
    }
</script>

<div class="card bg-white overflow-hidden flex flex-col h-full">
    <div class="px-6 py-5 border-b border-kari-warm-gray/20 bg-gray-50/50 flex items-center justify-between shrink-0">
        <div>
            <h3 class="text-base font-sans font-semibold text-kari-text">Environment Variables</h3>
            <p class="mt-1 text-xs text-kari-warm-gray">Keys are injected securely during the deployment build step and runtime.</p>
        </div>
        <button 
            type="button" 
            on:click={addVariable}
            class="inline-flex items-center px-3 py-1.5 border border-kari-warm-gray/30 shadow-sm text-xs font-medium rounded text-kari-text bg-white hover:bg-kari-light-gray focus:outline-none transition-colors"
        >
            <svg class="-ml-0.5 mr-1 h-4 w-4 text-kari-teal" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
            </svg>
            Add Variable
        </button>
    </div>

    <div class="p-6 flex-1 overflow-y-auto bg-kari-light-gray/20">
        {#if envList.length === 0}
            <div class="text-center py-8">
                <svg class="mx-auto h-10 w-10 text-kari-warm-gray/50" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                </svg>
                <p class="mt-2 text-sm text-kari-text font-medium">No environment variables defined</p>
                <p class="mt-1 text-xs text-kari-warm-gray">Click "Add Variable" to inject secrets into your app.</p>
            </div>
        {:else}
            <div class="space-y-3">
                {#each envList as item, i (item.id)}
                    <div class="flex items-start gap-3 animate-fade-in-up">
                        <div class="w-1/3">
                            <input 
                                type="text" 
                                placeholder="DATABASE_URL" 
                                bind:value={item.key}
                                on:input={(e) => handleKeyInput(e, i)}
                                class="w-full font-mono text-sm px-3 py-2 border border-kari-warm-gray/30 rounded shadow-sm focus:ring-1 focus:ring-kari-teal focus:border-kari-teal text-kari-text placeholder-kari-warm-gray/50"
                            >
                        </div>
                        
                        <div class="flex-1 relative">
                            <input 
                                type={item.show ? "text" : "password"} 
                                placeholder="postgres://user:pass@localhost..." 
                                bind:value={item.value}
                                class="w-full font-mono text-sm px-3 py-2 pr-10 border border-kari-warm-gray/30 rounded shadow-sm focus:ring-1 focus:ring-kari-teal focus:border-kari-teal text-kari-text placeholder-kari-warm-gray/50"
                            >
                            <button 
                                type="button" 
                                on:click={() => toggleVisibility(item.id)}
                                class="absolute inset-y-0 right-0 px-3 flex items-center text-kari-warm-gray hover:text-kari-teal transition-colors focus:outline-none"
                                title={item.show ? "Hide value" : "Show value"}
                            >
                                {#if item.show}
                                    <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.29 3.29m0 0a10.01 10.01 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243m-4.242-4.242L3 3m18 18l-3.29-3.29M12 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" /></svg>
                                {:else}
                                    <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" /></svg>
                                {/if}
                            </button>
                        </div>

                        <button 
                            type="button" 
                            on:click={() => removeVariable(item.id)}
                            class="p-2 text-kari-warm-gray hover:text-red-500 hover:bg-red-50 rounded transition-colors focus:outline-none mt-0.5"
                            title="Remove variable"
                        >
                            <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
                        </button>
                    </div>
                {/each}
            </div>
        {/if}
    </div>

    <div class="px-6 py-4 border-t border-kari-warm-gray/20 bg-gray-50/50 shrink-0 flex items-center justify-between">
        
        <div class="flex-1 mr-4 min-h-[1.5rem]">
            {#if saveMessage}
                <p class={`text-sm font-medium ${saveMessage.type === 'error' ? 'text-red-600' : 'text-kari-teal'}`}>
                    {saveMessage.text}
                </p>
            {/if}
        </div>

        <button 
            type="button" 
            on:click={saveEnvironment}
            disabled={isSaving}
            class="inline-flex items-center justify-center px-4 py-2 rounded shadow-sm text-sm font-sans font-medium text-white bg-kari-text hover:bg-black focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-kari-text transition-colors disabled:opacity-70"
        >
            {#if isSaving}
                <svg class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg>
                Encrypting & Saving...
            {:else}
                Save Changes
            {/if}
        </button>
    </div>
</div>

<style>
    .animate-fade-in-up {
        animation: fadeInUp 0.2s ease-out forwards;
    }
    @keyframes fadeInUp {
        from { opacity: 0; transform: translateY(4px); }
        to { opacity: 1; transform: translateY(0); }
    }
</style>
