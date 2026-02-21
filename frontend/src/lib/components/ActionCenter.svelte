<script lang="ts">
	import { onMount } from 'svelte';
	import { fade, slide } from 'svelte/transition';
	import { canPerform } from '$lib/utils/auth';
	import { page } from '$app/stores';

	// Props passed from +page.server.ts
	export let alerts: any[] = [];
	export let totalCount: number = 0;
	export let filters = {
		severity: '',
		isResolved: false,
		limit: 10,
		offset: 0
	};

	let loading = false;

	// 🛡️ SLA: Functional Pagination
	$: totalPages = Math.ceil(totalCount / filters.limit);
	$: currentPage = Math.floor(filters.offset / filters.limit) + 1;

	async function toggleResolve(alertId: string) {
		if (!canPerform($page.data.user.permissions, 'alerts:write')) return;
		
		loading = true;
		const res = await fetch(`/api/v1/alerts/${alertId}/resolve`, { method: 'POST' });
		if (res.ok) {
			// Optimistic UI update or trigger a full invalidate
			alerts = alerts.filter(a => a.id !== alertId);
		}
		loading = false;
	}

	function getSeverityClass(severity: string) {
		switch (severity) {
			case 'critical': return 'bg-red-500/10 text-red-500 border-red-500/20';
			case 'warning': return 'bg-amber-500/10 text-amber-500 border-amber-500/20';
			default: return 'bg-blue-500/10 text-blue-500 border-blue-500/20';
		}
	}
</script>

<section class="space-y-4">
	<div class="flex items-center justify-between">
		<h2 class="text-xl font-bold">Action Center ({totalCount})</h2>
		
		<div class="flex gap-2">
			<select bind:value={filters.severity} class="select-sm bg-base-200 rounded border-none">
				<option value="">All Severities</option>
				<option value="critical">Critical</option>
				<option value="warning">Warning</option>
			</select>
		</div>
	</div>

	<div class="grid gap-3">
		{#each alerts as alert (alert.id)}
			<div 
				transition:slide|local
				class="p-4 border rounded-lg flex items-start justify-between {getSeverityClass(alert.severity)}"
			>
				<div class="space-y-1">
					<div class="flex items-center gap-2">
						<span class="uppercase text-[10px] font-bold tracking-widest">{alert.category}</span>
						<span class="text-xs opacity-60">
							{new Date(alert.created_at).toLocaleString()}
						</span>
					</div>
					<p class="font-medium">{alert.message}</p>
					
					{#if alert.metadata?.trace_id}
						<code class="text-[10px] opacity-70">TRACE: {alert.metadata.trace_id}</code>
					{/if}
				</div>

				{#if !alert.is_resolved}
					<button 
						disabled={loading}
						on:click={() => toggleResolve(alert.id)}
						class="btn-xs btn-ghost border border-current hover:bg-current hover:text-white transition-all"
					>
						Resolve
					</button>
				{/if}
			</div>
		{:else}
			<div class="py-20 text-center opacity-40 italic">
				No active alerts. System is healthy.
			</div>
		{/each}
	</div>

	{#if totalPages > 1}
		<div class="flex items-center justify-center gap-4 pt-4 border-t border-base-content/10">
			<button 
				disabled={currentPage === 1}
				on:click={() => filters.offset -= filters.limit}
				class="px-3 py-1 bg-base-300 rounded disabled:opacity-30"
			>
				Prev
			</button>
			<span class="text-sm font-mono">{currentPage} / {totalPages}</span>
			<button 
				disabled={currentPage === totalPages}
				on:click={() => filters.offset += filters.limit}
				class="px-3 py-1 bg-base-300 rounded disabled:opacity-30"
			>
				Next
			</button>
		</div>
	{/if}
</section>
