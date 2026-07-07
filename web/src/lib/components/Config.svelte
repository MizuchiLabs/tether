<script lang="ts">
	import { api } from '$lib/api';
	import { Button } from '$lib/components/ui/button';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
	import * as Empty from '$lib/components/ui/empty';
	import * as InputGroup from '$lib/components/ui/input-group';
	import * as ScrollArea from '$lib/components/ui/scroll-area';
	import { Spinner } from '$lib/components/ui/spinner';
	import * as Tabs from '$lib/components/ui/tabs';
	import { UseClipboard } from '$lib/hooks/use-clipboard.svelte';
	import { lang } from '$lib/store.svelte';
	import {
		Bug,
		CheckIcon,
		CopyIcon,
		DeleteIcon,
		DownloadIcon,
		FileBraces,
		FunnelIcon,
		GlobeIcon,
		Layers2Icon,
		LockIcon,
		RefreshCw,
		RouteIcon,
		SearchIcon,
		WorkflowIcon
	} from '@lucide/svelte';
	import { dump as TOML } from 'js-toml';
	import { createHighlighter } from 'shiki';
	import YAML from 'yaml';

	let { env }: { env: string } = $props();

	let config = $state.raw<any>(null);
	let isLoading = $state(false);
	let error = $state('');
	let highlighter = $state<Awaited<ReturnType<typeof createHighlighter>> | null>(null);
	let search = $state('');
	let filter = $state('all');

	const clipboard = new UseClipboard();

	$effect(() => {
		createHighlighter({
			themes: ['catppuccin-latte', 'catppuccin-macchiato'],
			langs: ['json', 'yaml', 'toml']
		}).then((h) => {
			highlighter = h;
		});
	});

	async function fetchConfig() {
		if (!env) return;
		isLoading = true;
		error = '';

		try {
			config = await api.config(env);
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
				config = JSON.parse(event.data);
				error = '';
			} catch (err) {
				console.error('Failed to parse SSE data', err);
			}
		};

		return () => {
			eventSource.close();
		};
	});

	const filteredConfig = $derived.by(() => {
		if (!config) return null;

		let result = JSON.parse(JSON.stringify(config));

		const query = search.trim().toLowerCase();
		const protocols = ['http', 'tcp', 'udp'];

		protocols.forEach((proto) => {
			if (!result[proto]) return;

			if (filter !== 'all') {
				if (filter === 'tls') {
					protocols.forEach((p) => delete result[p]);
				} else {
					delete result.tls;

					protocols.forEach((proto) => {
						if (result[proto]) {
							const keep = result[proto][filter];
							result[proto] = keep ? { [filter]: keep } : {};
						}
					});
				}
			}

			if (query) {
				protocols.forEach((proto) => {
					if (!result[proto]) return;

					const categories = ['routers', 'middlewares', 'services'];
					categories.forEach((category) => {
						if (result[proto][category]) {
							const filteredEntries = Object.entries(result[proto][category]).filter(([key]) =>
								key.toLowerCase().includes(query)
							);

							if (filteredEntries.length > 0) {
								result[proto][category] = Object.fromEntries(filteredEntries);
							} else {
								delete result[proto][category];
							}
						}
					});

					if (Object.keys(result[proto]).length === 0) {
						delete result[proto];
					}
				});

				if (result.tls) {
					const tlsString = JSON.stringify(result.tls).toLowerCase();
					if (!tlsString.includes(query)) {
						delete result.tls;
					}
				}

				const hasMatches = protocols.some((p) => result[p]) || result.tls;
				if (!hasMatches) {
					return {};
				}
			}
		});

		return result;
	});

	const formatted = $derived.by(() => {
		if (!filteredConfig) return '';
		try {
			switch (lang.current) {
				case 'json':
					return JSON.stringify(filteredConfig, null, 2);
				case 'yaml':
					return YAML.stringify(filteredConfig, {
						indent: 2,
						lineWidth: 0,
						collectionStyle: 'block'
					});
				case 'toml':
					return TOML(filteredConfig);
			}
		} catch {
			return 'Error formatting configuration data.';
		}
	});

	const isEmpty = $derived(
		filteredConfig && typeof filteredConfig === 'object' && Object.keys(filteredConfig).length === 0
	);

	const codeHtml = $derived.by(() => {
		if (!highlighter || !formatted) return '';
		if (formatted === 'Error formatting configuration data.') return formatted;

		return highlighter.codeToHtml(formatted, {
			lang: lang.current,
			themes: { light: 'catppuccin-latte', dark: 'catppuccin-macchiato' }
		});
	});

	function handleDownload() {
		if (!formatted) return;

		let mimeType = 'application/json';
		switch (lang.current) {
			case 'yaml':
				mimeType = 'application/yaml';
				break;
			case 'toml':
				mimeType = 'application/toml';
				break;
		}
		const blob = new Blob([formatted], { type: mimeType });
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = `${env}.${lang.current}`;
		document.body.appendChild(a);
		a.click();
		document.body.removeChild(a);
		URL.revokeObjectURL(url);
	}
</script>

<div class="flex items-center justify-between gap-2">
	<div class="flex items-center gap-2">
		{#if config && Object.keys(config).length > 0}
			<InputGroup.Root>
				<InputGroup.Input bind:value={search} placeholder="Search..." />
				<InputGroup.Addon align="inline-end" class="pr-4">
					{#if search}
						<InputGroup.Button variant="ghost" size="icon-xs" onclick={() => (search = '')}>
							<DeleteIcon data-icon="inline-start" />
						</InputGroup.Button>
					{/if}
				</InputGroup.Addon>
				<InputGroup.Addon>
					<SearchIcon data-icon="inline-start" />
				</InputGroup.Addon>
			</InputGroup.Root>

			<DropdownMenu.Root>
				<DropdownMenu.Trigger>
					{#snippet child({ props })}
						<Button {...props} variant="outline">
							<FunnelIcon data-icon="inline-start" />
							<span class="capitalize">{filter}</span>
						</Button>
					{/snippet}
				</DropdownMenu.Trigger>
				<DropdownMenu.Content>
					<DropdownMenu.Group>
						<DropdownMenu.RadioGroup bind:value={filter}>
							<DropdownMenu.RadioItem value="all">
								<GlobeIcon data-icon="inline-start" />
								All
							</DropdownMenu.RadioItem>
							<DropdownMenu.RadioItem value="routers">
								<RouteIcon data-icon="inline-start" />
								Routers
							</DropdownMenu.RadioItem>
							<DropdownMenu.RadioItem value="services">
								<WorkflowIcon data-icon="inline-start" />
								Services
							</DropdownMenu.RadioItem>
							<DropdownMenu.RadioItem value="middlewares">
								<Layers2Icon data-icon="inline-start" />
								Middlewares
							</DropdownMenu.RadioItem>
							<DropdownMenu.RadioItem value="tls">
								<LockIcon data-icon="inline-start" />
								TLS
							</DropdownMenu.RadioItem>
						</DropdownMenu.RadioGroup>
					</DropdownMenu.Group>
				</DropdownMenu.Content>
			</DropdownMenu.Root>
		{/if}
	</div>

	<Button variant="outline" onclick={fetchConfig} disabled={isLoading}>
		<RefreshCw class={{ 'animate-spin': isLoading }} data-icon="inline-start" />
		Refresh
	</Button>
</div>

{#if isLoading}
	<Empty.Root class="border border-dashed">
		<Empty.Header>
			<Empty.Media variant="icon">
				<Spinner />
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
{:else if isEmpty}
	<Empty.Root class="border border-dashed">
		<Empty.Header>
			<Empty.Media variant="icon">
				<FileBraces />
			</Empty.Media>
			<Empty.Title
				>{config && Object.keys(config).length > 0
					? 'No matches found'
					: 'Configuration is empty'}</Empty.Title
			>
			<Empty.Description>
				{config && Object.keys(config).length > 0
					? `No results found for "${search}" in ${filter}.`
					: 'Tether returned an empty ruleset for this environment.'}
			</Empty.Description>
		</Empty.Header>
	</Empty.Root>
{:else if formatted && codeHtml}
	<div
		class="group relative flex flex-1 flex-col overflow-hidden rounded-xl border bg-card shadow-sm"
	>
		<div class="flex h-10 items-center justify-between border-b border-muted px-4">
			<div class="flex items-center gap-1.5">
				<div
					class="size-3 rounded-full border border-black/10 bg-red-500/80 dark:border-white/10"
				></div>
				<div
					class="size-3 rounded-full border border-black/10 bg-yellow-500/80 dark:border-white/10"
				></div>
				<div
					class="size-3 rounded-full border border-black/10 bg-green-500/80 dark:border-white/10"
				></div>
			</div>

			<Tabs.Root value={lang.current} onValueChange={(v) => (lang.current = v)}>
				<Tabs.List>
					{#each ['yaml', 'json', 'toml'] as language (language)}
						<Tabs.Trigger value={language} class="font-mono text-xs">{env}.{language}</Tabs.Trigger>
					{/each}
				</Tabs.List>
			</Tabs.Root>

			<div class="flex items-center gap-1 text-muted-foreground">
				<Button
					aria-label="Copy"
					title="Copy"
					variant="ghost"
					size="icon-xs"
					onclick={() => clipboard.copy(formatted)}
				>
					{#if clipboard.copied}
						<CheckIcon data-icon="inline-start" />
					{:else}
						<CopyIcon data-icon="inline-start" />
					{/if}
				</Button>
				<Button
					aria-label="Download"
					title="Download"
					variant="ghost"
					size="icon-xs"
					onclick={handleDownload}
				>
					<DownloadIcon data-icon="inline-start" />
				</Button>
			</div>
		</div>

		<ScrollArea.Root class="h-full w-full">
			<div class="shiki-container px-4 py-2 text-left text-sm leading-relaxed">
				{@html codeHtml}
			</div>
		</ScrollArea.Root>
	</div>
{/if}

<style>
	:global(.shiki-container pre) {
		margin: 0 !important;
		background: transparent !important;
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
