<script lang="ts">
	import Config from '$lib/components/Config.svelte';
	import * as Select from '$lib/components/ui/select';
	import * as Empty from '$lib/components/ui/empty';
	import { Cloud } from '@lucide/svelte';
	import { loggedIn } from '$lib/store.svelte';
	import { api } from '$lib/api';

	let envs = $state<string[]>([]);
	let env = $state('');

	$effect(() => {
		if (!loggedIn.current) return;

		let timeoutId: number | undefined;
		async function pollEnvs() {
			try {
				const data = await api.envs();
				envs = Array.isArray(data) ? data : [];

				if (envs.length > 0) {
					if (!env) {
						env = envs.includes('default') ? 'default' : envs[0];
					}
					return; // Stop polling once an environment is available
				}
			} catch (_) {}
			timeoutId = window.setTimeout(pollEnvs, 5000);
		}

		pollEnvs();

		return () => {
			if (timeoutId) clearTimeout(timeoutId);
		};
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

		{#if envs.length > 0}
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
		{/if}
	</div>

	{#if env}
		<Config {env} />
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
