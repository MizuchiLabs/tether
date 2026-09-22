<script lang="ts">
	import Config from '$lib/components/Config.svelte';
	import * as Empty from '$lib/components/ui/empty';
	import { env } from '$lib/store.svelte';
	import { Cloud } from '@lucide/svelte';
</script>

<div class="mx-auto mt-4 flex w-full max-w-5xl flex-1 flex-col gap-6">
	{#if env.current}
		<Config />
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
