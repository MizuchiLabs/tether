<script lang="ts" module>
	import { createHighlighter } from 'shiki';

	// Shared across mounts, the card remounts on every env switch.
	const highlighter = createHighlighter({
		themes: ['catppuccin-latte', 'catppuccin-macchiato'],
		langs: ['json', 'yaml', 'toml']
	});
</script>

<script lang="ts">
	import type { Sections, TraefikConfig } from '#lib/api.js';
	import { Button } from '#lib/components/ui/button/index.js';
	import * as DropdownMenu from '#lib/components/ui/dropdown-menu/index.js';
	import * as Empty from '#lib/components/ui/empty/index.js';
	import * as InputGroup from '#lib/components/ui/input-group/index.js';
	import * as ScrollArea from '#lib/components/ui/scroll-area/index.js';
	import * as Tabs from '#lib/components/ui/tabs/index.js';
	import { UseClipboard } from '#lib/hooks/use-clipboard.svelte.js';
	import { env, lang } from '#lib/store.svelte.js';
	import {
		CheckIcon,
		CopyIcon,
		DownloadIcon,
		FileBracesIcon,
		FunnelIcon,
		GlobeIcon,
		Layers2Icon,
		LockIcon,
		RouteIcon,
		SearchIcon,
		SearchXIcon,
		WorkflowIcon,
		XIcon
	} from '@lucide/svelte';
	import { dump as toTOML } from 'js-toml';
	import { toast } from 'svelte-sonner';
	import YAML from 'yaml';

	let { config }: { config: TraefikConfig } = $props();

	const FILTERS = [
		{ value: 'all', label: 'All', noun: 'entries', icon: GlobeIcon },
		{ value: 'routers', label: 'Routers', noun: 'routers', icon: RouteIcon },
		{ value: 'services', label: 'Services', noun: 'services', icon: WorkflowIcon },
		{ value: 'middlewares', label: 'Middlewares', noun: 'middlewares', icon: Layers2Icon },
		{ value: 'tls', label: 'TLS', noun: 'TLS settings', icon: LockIcon }
	];

	const FORMATS = [
		{ value: 'yaml', label: 'YAML', mime: 'application/yaml' },
		{ value: 'json', label: 'JSON', mime: 'application/json' },
		{ value: 'toml', label: 'TOML', mime: 'application/toml' }
	];

	let search = $state('');
	let filter = $state('all');
	let shiki = $state.raw<Awaited<typeof highlighter> | null>(null);
	highlighter.then((h) => (shiki = h));

	const clipboard = new UseClipboard({ delay: 1500 });

	const activeFilter = $derived(FILTERS.find((f) => f.value === filter) ?? FILTERS[0]);
	const format = $derived(FORMATS.find((f) => f.value === lang.current) ?? FORMATS[0]);
	const hasConfig = $derived(Object.keys(config ?? {}).length > 0);
	const filtered = $derived(filterConfig(config ?? {}, filter, search));
	const hasMatches = $derived(Object.keys(filtered).length > 0);

	const formatted = $derived.by(() => {
		try {
			switch (format.value) {
				case 'json':
					return JSON.stringify(filtered, null, 2);
				case 'toml':
					return toTOML(filtered);
				default:
					return YAML.stringify(filtered, { indent: 2, lineWidth: 0, collectionStyle: 'block' });
			}
		} catch {
			return null;
		}
	});

	const html = $derived(
		shiki && formatted
			? shiki.codeToHtml(formatted, {
					lang: format.value,
					themes: { light: 'catppuccin-latte', dark: 'catppuccin-macchiato' }
				})
			: null
	);

	// Keeps entries whose name or content contains the query, so hosts and server URLs match too.
	function filterConfig(config: TraefikConfig, filter: string, query: string) {
		const q = query.trim().toLowerCase();
		const matches = (value: unknown) => !q || JSON.stringify(value).toLowerCase().includes(q);

		const result: TraefikConfig = {};
		if (filter !== 'tls') {
			for (const proto of ['http', 'tcp', 'udp'] as const) {
				const kept: Sections = {};
				for (const [section, entries] of Object.entries(config[proto] ?? {})) {
					if (filter !== 'all' && section !== filter) continue;
					const hits = Object.entries(entries ?? {}).filter(matches);
					if (hits.length > 0) kept[section] = Object.fromEntries(hits);
				}
				if (Object.keys(kept).length > 0) result[proto] = kept;
			}
		}
		if (config.tls && (filter === 'all' || filter === 'tls') && matches(config.tls)) {
			result.tls = config.tls;
		}
		return result;
	}

	function resetFilters() {
		search = '';
		filter = 'all';
	}

	async function copy() {
		if (!formatted) return;
		if ((await clipboard.copy(formatted)) === 'failure') {
			toast.error("Couldn't copy to the clipboard", {
				description: 'Browsers only allow it on HTTPS or localhost.'
			});
		}
	}

	function download() {
		if (!formatted) return;
		const url = URL.createObjectURL(new Blob([formatted], { type: format.mime }));
		const a = document.createElement('a');
		a.href = url;
		a.download = `${env.current}.${format.value}`;
		document.body.appendChild(a);
		a.click();
		document.body.removeChild(a);
		URL.revokeObjectURL(url);
	}
</script>

{#if !hasConfig}
	<Empty.Root class="border">
		<Empty.Header>
			<Empty.Media variant="icon">
				<FileBracesIcon />
			</Empty.Media>
			<Empty.Title>No routes yet</Empty.Title>
			<Empty.Description>Nothing in this environment for Traefik to load.</Empty.Description>
		</Empty.Header>
	</Empty.Root>
{:else}
	<div class="flex items-center gap-2">
		<InputGroup.Root>
			<InputGroup.Input
				bind:value={search}
				placeholder="Search by name, host or URL"
				aria-label="Search configuration"
			/>
			<InputGroup.Addon>
				<SearchIcon />
			</InputGroup.Addon>
			{#if search}
				<InputGroup.Addon align="inline-end">
					<InputGroup.Button
						size="icon-xs"
						aria-label="Clear search"
						title="Clear search"
						onclick={() => (search = '')}
					>
						<XIcon />
					</InputGroup.Button>
				</InputGroup.Addon>
			{/if}
		</InputGroup.Root>

		<DropdownMenu.Root>
			<DropdownMenu.Trigger>
				{#snippet child({ props })}
					<Button {...props} variant="outline" title="Filter by type">
						<FunnelIcon data-icon="inline-start" />
						{activeFilter.label}
					</Button>
				{/snippet}
			</DropdownMenu.Trigger>
			<DropdownMenu.Content align="end">
				<DropdownMenu.Group>
					<DropdownMenu.RadioGroup bind:value={filter}>
						{#each FILTERS as f (f.value)}
							<DropdownMenu.RadioItem value={f.value}>
								<f.icon data-icon="inline-start" />
								{f.label}
							</DropdownMenu.RadioItem>
						{/each}
					</DropdownMenu.RadioGroup>
				</DropdownMenu.Group>
			</DropdownMenu.Content>
		</DropdownMenu.Root>
	</div>

	<div class="flex flex-col overflow-hidden rounded-xl border bg-card">
		<div class="flex items-center justify-between gap-2 border-b py-1.5 pr-2 pl-1.5">
			<Tabs.Root value={format.value} onValueChange={(v) => (lang.current = v)}>
				<Tabs.List>
					{#each FORMATS as f (f.value)}
						<Tabs.Trigger value={f.value} class="px-2.5 text-xs">{f.label}</Tabs.Trigger>
					{/each}
				</Tabs.List>
			</Tabs.Root>

			<div class="flex items-center gap-1">
				<Button
					variant="ghost"
					size="icon-sm"
					aria-label="Copy"
					title="Copy"
					disabled={!hasMatches || !formatted}
					onclick={copy}
				>
					{#if clipboard.copied}
						<CheckIcon />
					{:else}
						<CopyIcon />
					{/if}
				</Button>
				<Button
					variant="ghost"
					size="icon-sm"
					aria-label="Download {env.current}.{format.value}"
					title="Download {env.current}.{format.value}"
					disabled={!hasMatches || !formatted}
					onclick={download}
				>
					<DownloadIcon />
				</Button>
			</div>
		</div>

		{#if !hasMatches}
			<Empty.Root>
				<Empty.Header>
					<Empty.Media variant="icon">
						<SearchXIcon />
					</Empty.Media>
					<Empty.Title>No matches</Empty.Title>
					<Empty.Description>
						{#if search.trim()}
							Nothing {filter === 'all' ? '' : `in ${activeFilter.noun} `}matches "{search.trim()}".
						{:else}
							There are no {activeFilter.noun} in this environment.
						{/if}
					</Empty.Description>
				</Empty.Header>
				<Empty.Content>
					<Button variant="outline" size="sm" onclick={resetFilters}>Clear filters</Button>
				</Empty.Content>
			</Empty.Root>
		{:else if formatted === null}
			<Empty.Root>
				<Empty.Header>
					<Empty.Media variant="icon">
						<FileBracesIcon />
					</Empty.Media>
					<Empty.Title>Can't show this as {format.label}</Empty.Title>
					<Empty.Description>Some values don't map to {format.label}.</Empty.Description>
				</Empty.Header>
				<Empty.Content>
					<Button variant="outline" size="sm" onclick={() => (lang.current = 'yaml')}>
						Show YAML
					</Button>
				</Empty.Content>
			</Empty.Root>
		{:else}
			<ScrollArea.Root orientation="horizontal">
				<div class="code text-sm leading-relaxed">
					{#if html}
						{@html html}
					{:else}
						<pre>{formatted}</pre>
					{/if}
				</div>
			</ScrollArea.Root>
		{/if}
	</div>
{/if}

<style>
	.code :global(pre) {
		margin: 0;
		padding: 0.75rem 1rem;
		background: transparent !important;
	}
	:global(.dark) .code :global(.shiki),
	:global(.dark) .code :global(.shiki span) {
		color: var(--shiki-dark) !important;
		font-style: var(--shiki-dark-font-style) !important;
		font-weight: var(--shiki-dark-font-weight) !important;
		text-decoration: var(--shiki-dark-text-decoration) !important;
	}
</style>
