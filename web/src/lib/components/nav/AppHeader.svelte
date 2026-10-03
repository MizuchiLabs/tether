<script lang="ts">
	import { resolve } from '$app/paths';
	import { api } from '#lib/api.js';
	import Logo from '#lib/assets/logo.svelte';
	import StatusDot from '#lib/components/StatusDot.svelte';
	import { Button } from '#lib/components/ui/button/index.js';
	import * as Select from '#lib/components/ui/select/index.js';
	import { live } from '#lib/live.svelte.js';
	import { env } from '#lib/store.svelte.js';
	import { LogOut, Moon, Sun } from '@lucide/svelte';
	import { mode, toggleMode } from 'mode-watcher';
</script>

<header class="mx-auto mt-4 mb-6 flex w-full max-w-5xl items-center justify-between gap-3">
	<a
		href={resolve('/')}
		class="flex h-10 min-w-0 items-center gap-2 rounded-full border px-3.5 shadow-sm outline-none focus-visible:ring-[3px] focus-visible:ring-ring/50"
	>
		<Logo class="size-5 shrink-0" />
		<span class="truncate text-sm font-semibold tracking-tight">Tether</span>
	</a>

	<div class="flex h-10 shrink-0 items-center gap-1 rounded-full border px-2 shadow-sm">
		{#if live.offline}
			<span
				role="status"
				class="flex items-center gap-1.5 px-2 text-xs text-destructive"
				title="Lost connection to Tether. Showing the last known state until it reconnects."
			>
				<StatusDot tone="error" pulse />
				Offline
			</span>
		{/if}

		{#if live.envs?.length}
			<Select.Root type="single" bind:value={env.current}>
				<Select.Trigger class="bg-transparent" aria-label="Environment" title="Environment">
					{env.current}
				</Select.Trigger>
				<Select.Content>
					<Select.Group>
						{#each live.envs as name (name)}
							<Select.Item value={name}>{name}</Select.Item>
						{/each}
					</Select.Group>
				</Select.Content>
			</Select.Root>
		{/if}

		<Button
			variant="ghost"
			size="icon-sm"
			class="rounded-full"
			onclick={toggleMode}
			aria-label="Toggle theme"
			title="Toggle theme"
		>
			{#if mode.current === 'light'}
				<Moon />
			{:else}
				<Sun />
			{/if}
		</Button>
		<Button
			variant="ghost"
			size="icon-sm"
			class="rounded-full hover:bg-destructive/10 hover:text-destructive"
			onclick={api.logout}
			aria-label="Log out"
			title="Log out"
		>
			<LogOut />
		</Button>
	</div>
</header>
