<script lang="ts">
	import { resolve } from '$app/paths';
	import Logo from '$lib/assets/logo.svelte';
	import { Button } from '$lib/components/ui/button';
	import { LogOut, Moon, Sun } from '@lucide/svelte';
	import { mode, toggleMode } from 'mode-watcher';
	import { api } from '$lib/api';
	import * as Select from '$lib/components/ui/select';
	import { env, loggedIn } from '$lib/store.svelte';

	let envs = $state.raw<string[]>([]);

	$effect(() => {
		if (!loggedIn.current) return;

		let timeoutId: number | undefined;
		async function pollEnvs() {
			if (!loggedIn.current) return;
			try {
				const data = await api.envs();
				envs = Array.isArray(data) ? data : [];

				if (envs.length > 0) {
					if (!env.current) {
						env.current = envs.includes('default') ? 'default' : envs[0];
					}
					return;
				}
			} catch {
				// ignore
			}
			timeoutId = window.setTimeout(pollEnvs, 5000);
		}

		pollEnvs();
		return () => {
			if (timeoutId) clearTimeout(timeoutId);
		};
	});
</script>

<header class="pointer-events-none sticky z-50 mx-auto mt-4 mb-6 w-full max-w-5xl">
	<div class="flex items-center justify-between gap-3">
		<a
			href={resolve('/')}
			class="pointer-events-auto flex h-10 min-w-0 items-center gap-2 rounded-full border bg-background/80 px-3.5 shadow-sm backdrop-blur-md outline-none focus-visible:ring-[3px] focus-visible:ring-ring/50"
		>
			<Logo class="size-5 shrink-0" />
			<span class="truncate text-sm font-semibold tracking-tight">Tether</span>
		</a>

		<div
			class="pointer-events-auto flex h-10 shrink-0 items-center gap-2 rounded-full border bg-background/80 px-2 shadow-sm backdrop-blur-md"
		>
			<Select.Root type="single" bind:value={env.current}>
				<Select.Trigger class="bg-transparent" title="Select environment">
					{env.current || 'Select...'}
				</Select.Trigger>
				<Select.Content>
					<Select.Group>
						{#each envs as item (item)}
							<Select.Item value={item}>{item}</Select.Item>
						{/each}
					</Select.Group>
				</Select.Content>
			</Select.Root>

			<Button
				variant="ghost"
				size="icon-sm"
				class="rounded-full"
				onclick={toggleMode}
				aria-label="Toggle theme"
			>
				{#if mode.current === 'light'}
					<Moon />
				{:else}
					<Sun />
				{/if}
			</Button>
			<Button
				variant="ghost"
				onclick={api.logout}
				size="icon-sm"
				class="hover:bg-destructive/10 hover:text-destructive"
				title="Log out"
			>
				<LogOut />
			</Button>
		</div>
	</div>
</header>
