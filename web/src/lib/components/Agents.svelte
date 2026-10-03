<script lang="ts">
	import type { Agent, Collision, SharedService } from '#lib/api.js';
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

	const connected = $derived(agents.filter((a) => a.connected).length);

	function ago(iso: string) {
		const s = Math.max(0, Math.round((now - Date.parse(iso)) / 1000));
		if (s < 60) return `${s}s ago`;
		if (s < 3600) return `${Math.floor(s / 60)}m ago`;
		if (s < 86400) return `${Math.floor(s / 3600)}h ago`;
		return `${Math.floor(s / 86400)}d ago`;
	}

	function expiresIn(since: string) {
		return Math.max(0, Math.ceil((Date.parse(since) + EXPIRE_MS - now) / 1000));
	}
</script>

<Card.Root>
	<Card.Header>
		<Card.Title>Agents</Card.Title>
		<Card.Description>
			{connected} of {agents.length} connected. Disconnected agents keep their routes for 30s, unless
			another agent serves the same app.
		</Card.Description>
	</Card.Header>
	<Card.Content class="flex flex-col gap-4">
		{#if agents.length === 0}
			<p class="text-sm text-muted-foreground">No agents in this environment yet.</p>
		{:else}
			<Table.Root>
				<Table.Header>
					<Table.Row>
						<Table.Head>Agent</Table.Head>
						<Table.Head>Status</Table.Head>
						<Table.Head class="text-right">Routers</Table.Head>
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
									<Badge variant="secondary" title="Connected since {agent.since}">
										Connected {ago(agent.since)}
									</Badge>
								{:else}
									<Badge variant="destructive" title="Disconnected at {agent.since}">
										Offline, removed in {expiresIn(agent.since)}s
									</Badge>
								{/if}
							</Table.Cell>
							<Table.Cell class="text-right tabular-nums">{agent.routers}</Table.Cell>
							<Table.Cell class="hidden text-right tabular-nums sm:table-cell"
								>{agent.services}</Table.Cell
							>
							<Table.Cell
								class="hidden text-right text-muted-foreground sm:table-cell"
								title={agent.updated}
							>
								{agent.updated ? ago(agent.updated) : '-'}
							</Table.Cell>
						</Table.Row>
					{/each}
				</Table.Body>
			</Table.Root>
		{/if}

		{#if shared.length > 0}
			<Alert.Root>
				<NetworkIcon />
				<Alert.Title>
					{shared.length} load balanced across agents
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
								<span title={svc.servers.join('\n')}>({svc.servers.length} servers)</span>
							</li>
						{/each}
					</ul>
				</Alert.Description>
			</Alert.Root>
		{/if}

		{#if collisions.length > 0}
			<Alert.Root>
				<TriangleAlertIcon />
				<Alert.Title>{collisions.length} skipped because the name is already taken</Alert.Title>
				<Alert.Description>
					<ul class="flex flex-col gap-1">
						{#each collisions as c (c.kind + c.name + c.source)}
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
	</Card.Content>
</Card.Root>
