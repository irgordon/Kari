<script lang="ts">
	import { onMount, afterUpdate } from 'svelte';
	import { fade } from 'svelte/transition';

	export let logs: { content: string; trace_id: string }[] = [];
	export let activeTraceId: string;

	let viewport: HTMLElement;
	let autoscroll = true;

	// 🛡️ SLA: Filter logs by the active trace only to prevent cross-deployment contamination
	$: filteredLogs = logs.filter(l => l.trace_id === activeTraceId);

	// 🛡️ Logic to detect if we should autoscroll
	function handleScroll() {
		const { scrollTop, scrollHeight, clientHeight } = viewport;
		const isAtBottom = scrollHeight - scrollTop - clientHeight < 50;
		autoscroll = isAtBottom;
	}

	afterUpdate(() => {
		if (autoscroll && viewport) {
			viewport.scrollTo({ top: viewport.scrollHeight, behavior: 'smooth' });
		}
	});

	function formatLine(content: string) {
		if (content.startsWith('[ERR]')) {
			return { text: content.replace('[ERR] ', ''), class: 'text-error font-bold' };
		}
		if (content.startsWith('[OUT]')) {
			return { text: content.replace('[OUT] ', ''), class: 'text-base-content/80' };
		}
		return { text: content, class: 'text-primary italic' }; // System messages
	}
</script>

<div class="relative w-full bg-neutral text-neutral-content rounded-lg border border-base-content/10 shadow-2xl">
	<div class="flex items-center justify-between px-4 py-2 bg-base-300 rounded-t-lg border-b border-base-content/5">
		<div class="flex gap-1.5">
			<div class="w-3 h-3 rounded-full bg-error/50"></div>
			<div class="w-3 h-3 rounded-full bg-warning/50"></div>
			<div class="w-3 h-3 rounded-full bg-success/50"></div>
		</div>
		<span class="text-[10px] font-mono opacity-50 uppercase tracking-widest">Build Telemetry: {activeTraceId.slice(0, 8)}</span>
	</div>

	<div 
		bind:this={viewport}
		on:scroll={handleScroll}
		class="h-[400px] overflow-y-auto p-4 font-mono text-xs leading-relaxed selection:bg-primary selection:text-primary-content"
	>
		{#each filteredLogs as log}
			{@const formatted = formatLine(log.content)}
			<div transition:fade={{ duration: 100 }} class="flex gap-3 border-l-2 border-transparent hover:border-primary/20 hover:bg-base-content/5 px-2">
				<span class="opacity-30 select-none">[{new Date().toLocaleTimeString()}]</span>
				<span class={formatted.class}>{formatted.text}</span>
			</div>
		{:else}
			<div class="flex items-center justify-center h-full opacity-20 italic">
				Waiting for Muscle to initialize jail...
			</div>
		{/each}
	</div>

	{#if !autoscroll && filteredLogs.length > 0}
		<button 
			on:click={() => { autoscroll = true; viewport.scrollTo({ top: viewport.scrollHeight, behavior: 'smooth' }); }}
			class="absolute bottom-4 right-4 btn btn-xs btn-primary shadow-lg"
			transition:fade
		>
			⬇ Follow Tail
		</button>
	{/if}
</div>
