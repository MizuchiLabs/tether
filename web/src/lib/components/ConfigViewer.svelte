<script lang="ts">
	import { onMount } from 'svelte';
	import { createHighlighter } from 'shiki';
	import YAML from 'yaml';
	import { RefreshCw, Code, FileBraces } from '@lucide/svelte';
	import { authState } from '$lib/auth.svelte';
	import { Button } from '$lib/ui/button';
	import * as Tabs from '$lib/ui/tabs';

	let { env }: { env: string } = $props();

	let lang: 'json' | 'yaml' = $state('yaml');
	let rawData = $state('');
	let isLoading = $state(false);
	let error = $state('');
	let highlighter: Awaited<ReturnType<typeof createHighlighter>> | null = $state(null);

	onMount(async () => {
		highlighter = await createHighlighter({
			themes: ['catppuccin-latte', 'catppuccin-macchiato'],
			langs: ['json', 'yaml']
		});
		fetchConfig();
	});

	async function fetchConfig() {
		if (!env) return;
		isLoading = true;
		error = '';
		try {
			const res = await fetch(`/config?env=${env}&format=json`, {
				headers: authState.token ? { Authorization: `Bearer ${authState.token}` } : {}
			});
			if (res.ok) {
				const data = await res.json();
				rawData = JSON.stringify(data);
			} else {
				error = 'Failed to fetch config';
				rawData = '';
			}
		} catch (err: any) {
			error = err.message;
			rawData = '';
		} finally {
			isLoading = false;
		}
	}

	const formatted = $derived.by(() => {
		if (!rawData) return '';
		try {
			const obj = JSON.parse(rawData);
			return lang === 'json'
				? JSON.stringify(obj, null, 2)
				: YAML.stringify(obj, {
						indent: 2,
						lineWidth: 0,
						collectionStyle: 'block'
					});
		} catch (e) {
			return rawData;
		}
	});

	const codeHtml = $derived.by(() => {
		if (!highlighter || !formatted) return '';
		return highlighter.codeToHtml(formatted, {
			lang,
			themes: {
				light: 'catppuccin-latte',
				dark: 'catppuccin-macchiato'
			}
		});
	});

	$effect(() => {
		if (env) fetchConfig();
	});
</script>

<div class="flex flex-col gap-4 flex-1">
	<div class="flex items-center justify-between">
		<Tabs.Root value="yaml">
			<Tabs.List>
				<Tabs.Trigger value="yaml" onclick={() => (lang = 'yaml')}>
					<Code />
					YAML
				</Tabs.Trigger>
				<Tabs.Trigger value="json" onclick={() => (lang = 'json')}>
					<FileBraces />
					JSON
				</Tabs.Trigger>
			</Tabs.List>
		</Tabs.Root>

		<Button variant="outline" onclick={fetchConfig} disabled={isLoading}>
			<RefreshCw class={isLoading ? 'animate-spin' : ''} />
			Refresh
		</Button>
	</div>

	{#if error}
		<div
			class="rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-600 dark:border-red-900/50 dark:bg-red-900/20 dark:text-red-400"
		>
			{error}
		</div>
	{/if}

	<div class="relative flex-1 overflow-hidden rounded-xl border shadow-sm bg-card">
		{#if formatted && codeHtml}
			<div class="h-full overflow-auto p-4 text-sm leading-relaxed shiki-container">
				{@html codeHtml}
			</div>
		{:else if isLoading}
			<div class="flex h-64 items-center justify-center">
				<RefreshCw class="animate-spin" />
				Loading configuration...
			</div>
		{:else}
			<div class="flex h-64 items-center justify-center">
				No configuration available for "{env}".
			</div>
		{/if}
	</div>
</div>

<style>
	:global(.shiki-container pre) {
		margin: 0 !important;
		background: transparent !important;
	}
	:global(.shiki-container code) {
		font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono',
			'Courier New', monospace;
	}
	:global(.dark .shiki),
	:global(.dark .shiki span) {
		color: var(--shiki-dark) !important;
		background-color: transparent !important;
		font-style: var(--shiki-dark-font-style) !important;
		font-weight: var(--shiki-dark-font-weight) !important;
		text-decoration: var(--shiki-dark-text-decoration) !important;
	}
</style>
