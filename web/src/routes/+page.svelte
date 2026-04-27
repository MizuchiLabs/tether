<script lang="ts">
	import ConfigViewer from '$lib/components/ConfigViewer.svelte';
	import * as Select from '$lib/components/ui/select';
	import * as Empty from '$lib/components/ui/empty';
	import { Cloud } from '@lucide/svelte';
	import { loggedIn } from '$lib/store.svelte';

	let envs = $state<string[]>([]);
	let env = $state('');

	async function fetchEnvs() {
		try {
			const res = await fetch('/api/envs');
			if (res.ok) {
				const data = await res.json();
				envs = Array.isArray(data) ? data : [];

				if (envs.length > 0 && !env) {
					env = envs.includes('default') ? 'default' : envs[0];
				}
			} else {
				loggedIn.current = false;
			}
		} catch (_) {}
	}

	$effect(() => {
		if (loggedIn.current) fetchEnvs();
	});
</script>

<div class="mx-auto w-full max-w-5xl flex-1 flex flex-col gap-6">
	<div class="flex flex-col items-start justify-between gap-4 sm:flex-row sm:items-center">
		<div>
			<h2 class="text-2xl font-bold tracking-tight">Configuration Explorer</h2>
			<p class="text-sm text-muted-foreground">
				Viewing Traefik routing rules and service definitions.
			</p>
		</div>

		<div class="flex items-center gap-4">
			<span class="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
				Environment
			</span>
			<Select.Root type="single" bind:value={env}>
				<Select.Trigger class="w-40">
					{env || 'Select...'}
				</Select.Trigger>
				<Select.Content>
					{#each envs as env}
						<Select.Item value={env}>{env}</Select.Item>
					{/each}
				</Select.Content>
			</Select.Root>
		</div>
	</div>

	{#if env}
		<ConfigViewer {env} />
	{:else}
		<Empty.Root class="border border-dashed max-h-80">
			<Empty.Header>
				<Empty.Media variant="icon">
					<Cloud />
				</Empty.Media>
				<Empty.Title>No environment selected</Empty.Title>
				<Empty.Description>
					Environments are created automatically once agents push data to the server. Waiting for
					incoming data...
				</Empty.Description>
			</Empty.Header>
		</Empty.Root>
	{/if}
</div>
