<script lang="ts">
	import type { Agent, Collision, SharedService } from '#lib/api.js';
	import StatusDot from '#lib/components/StatusDot.svelte';
	import * as Alert from '#lib/components/ui/alert/index.js';
	import { Badge } from '#lib/components/ui/badge/index.js';
	import * as Card from '#lib/components/ui/card/index.js';
	import * as Table from '#lib/components/ui/table/index.js';
	import { NetworkIcon, TriangleAlertIcon } from '@lucide/svelte';
	import { useInterval } from 'runed';

	let {
		agents,
		collisions,
		shared
	}: { agents: Agent[]; collisions: Collision[]; shared: SharedService[] } = $props();

	// Matches expireAfter in the backend.
	const EXPIRE_MS = 30_000;

	let now = $state(Date.now());
	useInterval(1000, { callback: () => (now = Date.now()) });

	const online = $derived(agents.filter((a) => a.connected).length);

	// Go sends year 1 for times that were never set.
	const isSet = (iso: string) => Date.parse(iso) > 0;

	function ago(iso: string) {
		if (!isSet(iso)) return 'never';
		const s = Math.max(0, Math.round((now - Date.parse(iso)) / 1000));
		if (s < 60) return `${s}s ago`;
		if (s < 3600) return `${Math.floor(s / 60)}m ago`;
		if (s < 86400) return `${Math.floor(s / 3600)}h ago`;
		return `${Math.floor(s / 86400)}d ago`;
	}

	function expiresIn(since: string) {
		return Math.max(0, Math.ceil((Date.parse(since) + EXPIRE_MS - now) / 1000));
	}

	function plural(n: number, one: string, many: string) {
		return `${n} ${n === 1 ? one : many}`;
	}

	function localTime(iso: string) {
		return isSet(iso) ? new Date(iso).toLocaleString() : undefined;
	}
</script>

<Card.Root>
	<Card.Header>
		<Card.Title>Agents</Card.Title>
		<Card.Description>
			{agents.length > 0 ? `${online} of ${agents.length} online` : 'No agents connected yet'}
		</Card.Description>
	</Card.Header>

	{#if agents.length > 0}
		<Card.Content class="flex flex-col gap-4">
			<Table.Root>
				<Table.Header>
					<Table.Row>
						<Table.Head>Agent</Table.Head>
						<Table.Head>Status</Table.Head>
						<Table.Head class="hidden text-right sm:table-cell">Routers</Table.Head>
						<Table.Head class="hidden text-right sm:table-cell">Services</Table.Head>
						<Table.Head class="hidden text-right sm:table-cell">Last update</Table.Head>
					</Table.Row>
				</Table.Header>
				<Table.Body>
					{#each agents as agent (agent.name)}
						<Table.Row>
							<Table.Cell>
								<div class="flex flex-col">
									<span class="font-medium">{agent.name}</span>
									<span class="font-mono text-xs text-muted-foreground">{agent.addr}</span>
								</div>
							</Table.Cell>
							<Table.Cell>
								{#if agent.connected}
									<span
										class="flex items-center gap-2"
										title="Connected since {localTime(agent.since)}"
									>
										<StatusDot />
										Online
									</span>
								{:else}
									<span
										class="flex items-center gap-2 text-destructive"
										title="Routes stay for 30s unless another agent serves the same app"
									>
										<StatusDot tone="error" />
										Offline, removed in {expiresIn(agent.since)}s
									</span>
								{/if}
							</Table.Cell>
							<Table.Cell class="hidden text-right tabular-nums sm:table-cell"
								>{agent.routers}</Table.Cell
							>
							<Table.Cell class="hidden text-right tabular-nums sm:table-cell">
								{agent.services}
							</Table.Cell>
							<Table.Cell
								class="hidden text-right text-muted-foreground sm:table-cell"
								title={localTime(agent.updated)}
							>
								{ago(agent.updated)}
							</Table.Cell>
						</Table.Row>
					{/each}
				</Table.Body>
			</Table.Root>

			{#if collisions.length > 0}
				<Alert.Root>
					<TriangleAlertIcon />
					<Alert.Title>
						{plural(collisions.length, 'entry', 'entries')} skipped because the name is already taken
					</Alert.Title>
					<Alert.Description>
						<ul class="flex flex-col gap-1">
							{#each collisions as c (`${c.kind}/${c.name}@${c.source}`)}
								<li>
									<code class="font-mono">{c.kind}/{c.name}</code> from
									<span class="font-medium">{c.source}</span>, kept the one from
									<span class="font-medium">{c.owner}</span>
								</li>
							{/each}
						</ul>
					</Alert.Description>
				</Alert.Root>
			{/if}

			{#if shared.length > 0}
				<Alert.Root>
					<NetworkIcon />
					<Alert.Title>
						{plural(shared.length, 'service', 'services')} load balanced across agents
					</Alert.Title>
					<Alert.Description>
						<ul class="flex flex-col gap-1">
							{#each shared as svc (svc.name)}
								<li class="flex flex-wrap items-center gap-1">
									<code class="font-mono">{svc.name}</code>
									<span>on</span>
									{#each svc.agents as agent (agent)}
										<Badge variant="outline">{agent}</Badge>
									{/each}
									<span title={svc.servers.join('\n')}>
										({plural(svc.servers.length, 'server', 'servers')})
									</span>
								</li>
							{/each}
						</ul>
					</Alert.Description>
				</Alert.Root>
			{/if}
		</Card.Content>
	{/if}
</Card.Root>
