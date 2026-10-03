<script lang="ts">
	import Agents from '#lib/components/Agents.svelte';
	import Config from '#lib/components/Config.svelte';
	import * as Empty from '#lib/components/ui/empty/index.js';
	import { Spinner } from '#lib/components/ui/spinner/index.js';
	import { live } from '#lib/live.svelte.js';
	import { env } from '#lib/store.svelte.js';
	import { CloudIcon } from '@lucide/svelte';
	import { useInterval } from 'runed';

	useInterval(5000, { immediateCallback: true, callback: () => live.refreshEnvs() });

	const ready = $derived(live.envs?.includes(env.current) ?? false);

	// Live connection to the server, not derivable state.
	$effect(() => {
		if (ready) return live.connect(env.current);
	});
</script>

<div class="flex flex-1 flex-col gap-6 pb-6">
	{#if live.envs?.length === 0}
		<Empty.Root class="border">
			<Empty.Header>
				<Empty.Media variant="icon">
					<CloudIcon />
				</Empty.Media>
				<Empty.Title>Waiting for agents</Empty.Title>
				<Empty.Description>
					Start <a href="https://github.com/MizuchiLabs/tetherd" target="_blank" rel="noreferrer"
						>tetherd</a
					> on your servers. They show up here as soon as they connect.
				</Empty.Description>
			</Empty.Header>
		</Empty.Root>
	{:else if !live.snapshot}
		<Empty.Root>
			<Empty.Header>
				<Empty.Media variant="icon">
					<Spinner />
				</Empty.Media>
				<Empty.Title>Connecting</Empty.Title>
			</Empty.Header>
		</Empty.Root>
	{:else}
		<Agents
			agents={live.snapshot.agents ?? []}
			collisions={live.snapshot.collisions ?? []}
			shared={live.snapshot.shared ?? []}
		/>
		<Config config={live.snapshot.config} />
	{/if}
</div>
