<script lang="ts">
	import { api, type Snapshot } from '$lib/api';
	import Agents from '$lib/components/Agents.svelte';
	import Config from '$lib/components/Config.svelte';
	import * as Empty from '$lib/components/ui/empty';
	import { env } from '$lib/store.svelte';
	import { Cloud } from '@lucide/svelte';

	let snapshot = $state.raw<Snapshot | null>(null);

	// Live connection to the server, not derivable state.
	$effect(() => {
		if (!env.current) return;
		snapshot = null;
		const source = api.events(env.current, (s) => (snapshot = s));
		return () => source.close();
	});
</script>

<div class="mx-auto mt-4 flex w-full max-w-5xl flex-1 flex-col gap-6">
	{#if env.current}
		<Agents
			agents={snapshot?.agents ?? []}
			collisions={snapshot?.collisions ?? []}
			shared={snapshot?.shared ?? []}
		/>
		<Config config={snapshot?.config} />
	{:else}
		<Empty.Root class="border border-dashed">
			<Empty.Header>
				<Empty.Media variant="icon">
					<Cloud />
				</Empty.Media>
				<Empty.Title>Waiting for agents...</Empty.Title>
				<Empty.Description>
					Environments are discovered automatically when tetherd agents push their local
					configurations. Waiting for the first agent to connect...
				</Empty.Description>
			</Empty.Header>
		</Empty.Root>
	{/if}
</div>
