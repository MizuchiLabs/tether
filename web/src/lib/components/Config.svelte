<script lang="ts">
	import { api } from '$lib/api';
	import { Button } from '$lib/components/ui/button';
	import * as Empty from '$lib/components/ui/empty';
	import * as Tabs from '$lib/components/ui/tabs';
	import { lang } from '$lib/store.svelte';
	import { Bug, Check, Code, Copy, FileBraces, RefreshCw } from '@lucide/svelte';
	import { createHighlighter } from 'shiki';
	import { onMount } from 'svelte';
	import YAML from 'yaml';

	let { env }: { env: string } = $props();

	let configData = $state<any>(null);
	let isLoading = $state(false);
	let error = $state('');
	let copied = $state(false);
	let highlighter = $state<Awaited<ReturnType<typeof createHighlighter>> | null>(null);

	onMount(async () => {
		highlighter = await createHighlighter({
			themes: ['catppuccin-latte', 'catppuccin-macchiato'],
			langs: ['json', 'yaml']
		});
	});

	async function fetchConfig() {
		if (!env) return;
		isLoading = true;
		error = '';

		try {
			configData = await api.config(env);
		} catch (err: any) {
			error = err.message || 'Failed to fetch configuration.';
		} finally {
			isLoading = false;
		}
	}

	$effect(() => {
		if (!env) return;
		fetchConfig();

		const eventSource = api.events(env);
		eventSource.onmessage = (event) => {
			try {
				configData = JSON.parse(event.data);
				error = '';
			} catch (err) {
				console.error('Failed to parse SSE data', err);
			}
		};

		return () => {
			eventSource.close();
		};
	});

	const formattedText = $derived.by(() => {
		if (!configData) return '';
		try {
			return lang.current === 'json'
				? JSON.stringify(configData, null, 2)
				: YAML.stringify(configData, { indent: 2, lineWidth: 0, collectionStyle: 'block' });
		} catch {
			return 'Error formatting configuration data.';
		}
	});

	const isEmptyConfig = $derived.by(() => {
		return configData && typeof configData === 'object' && Object.keys(configData).length === 0;
	});

	const codeHtml = $derived.by(() => {
		if (!highlighter || !formattedText) return '';
		if (formattedText === 'Error formatting configuration data.') return formattedText;

		return highlighter.codeToHtml(formattedText, {
			lang: lang.current,
			themes: { light: 'catppuccin-latte', dark: 'catppuccin-macchiato' }
		});
	});

	async function handleCopy() {
		if (!formattedText) return;
		await navigator.clipboard.writeText(formattedText);
		copied = true;
		setTimeout(() => (copied = false), 2000);
	}
</script>

<div class="flex flex-col gap-4 flex-1">
	<div class="flex items-center justify-between">
		<Tabs.Root bind:value={lang.current}>
			<Tabs.List>
				<Tabs.Trigger value="yaml" class="gap-2">
					<Code class="size-4" /> YAML
				</Tabs.Trigger>
				<Tabs.Trigger value="json" class="gap-2">
					<FileBraces class="size-4" /> JSON
				</Tabs.Trigger>
			</Tabs.List>
		</Tabs.Root>

		<Button variant="outline" onclick={fetchConfig} disabled={isLoading}>
			<RefreshCw class={isLoading ? 'animate-spin' : ''} />
			Refresh
		</Button>
	</div>

	{#if isLoading && !configData}
		<Empty.Root class="border border-dashed">
			<Empty.Header>
				<Empty.Media variant="icon">
					<RefreshCw class="animate-spin" />
				</Empty.Media>
				<Empty.Title>Loading configuration...</Empty.Title>
				<Empty.Description>Fetching configuration from Tether.</Empty.Description>
			</Empty.Header>
		</Empty.Root>
	{:else if error}
		<Empty.Root class="border border-dashed">
			<Empty.Header>
				<Empty.Media variant="icon" class="bg-destructive/30">
					<Bug class="text-destructive" />
				</Empty.Media>
				<Empty.Title class="text-destructive">Error</Empty.Title>
				<Empty.Description class="text-destructive/75">{error}</Empty.Description>
			</Empty.Header>
		</Empty.Root>
	{:else if isEmptyConfig}
		<Empty.Root class="border border-dashed">
			<Empty.Header>
				<Empty.Media variant="icon">
					<FileBraces />
				</Empty.Media>
				<Empty.Title>Configuration is empty</Empty.Title>
				<Empty.Description>
					Tether returned an empty ruleset for this environment.
				</Empty.Description>
			</Empty.Header>
		</Empty.Root>
	{:else if formattedText && codeHtml}
		<div class="relative flex-1 overflow-hidden rounded-xl border shadow-sm bg-card group">
			<Button
				variant="ghost"
				size="icon"
				class="absolute right-3 top-3 z-10 opacity-0 transition-opacity group-hover:opacity-100"
				onclick={handleCopy}
				title="Copy to clipboard"
			>
				{#if copied}
					<Check class="text-green-500" />
				{:else}
					<Copy />
				{/if}
			</Button>

			<div class="h-full overflow-auto p-4 text-sm leading-relaxed shiki-container">
				{@html codeHtml}
			</div>
		</div>
	{/if}
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
