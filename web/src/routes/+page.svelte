<script lang="ts">
	import { onMount } from 'svelte';
	import { authState } from '$lib/auth.svelte';
	import Login from '$lib/components/Login.svelte';
	import ConfigViewer from '$lib/components/ConfigViewer.svelte';
	import * as Select from '$lib/ui/select/index.js';

	let envs = $state<string[]>([]);
	let currentEnv = $state('');

	async function fetchEnvs() {
		try {
			const res = await fetch('/envs', {
				headers: authState.token ? { Authorization: `Bearer ${authState.token}` } : {}
			});
			if (res.ok) {
				envs = (await res.json()) || [];
				if (envs.length > 0 && !currentEnv) {
					currentEnv = envs.includes('default') ? 'default' : envs[0];
				}
			}
		} catch (e) {
			// Fail silently
		}
	}

	onMount(() => {
		if (authState.isAuthed) fetchEnvs();
	});

	$effect(() => {
		if (authState.isAuthed) fetchEnvs();
	});
</script>

{#if !authState.isAuthed}
	<Login onLogin={fetchEnvs} />
{:else}
	<div class="mx-auto w-full max-w-5xl flex-1 flex flex-col gap-6">
		<div class="flex flex-col items-start justify-between gap-4 sm:flex-row sm:items-center">
			<div>
				<h2 class="text-2xl font-bold tracking-tight">Configuration Explorer</h2>
				<p class="text-sm">Viewing Traefik routing rules and service definitions.</p>
			</div>

			<div class="flex items-center gap-4">
				<span class="text-xs font-semibold uppercase tracking-wider">Environment</span>
				<Select.Root type="single" bind:value={currentEnv}>
					<Select.Trigger class="w-40">
						{currentEnv || 'Select...'}
					</Select.Trigger>
					<Select.Content>
						{#each envs as env}
							<Select.Item value={env}>{env}</Select.Item>
						{/each}
					</Select.Content>
				</Select.Root>
			</div>
		</div>

		{#if currentEnv}
			<ConfigViewer env={currentEnv} />
		{:else}
			<div class="flex h-64 items-center justify-center rounded-xl border-2 border-dashed">
				Waiting for environment selection...
			</div>
		{/if}
	</div>
{/if}
