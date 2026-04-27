<script lang="ts">
	import { authState } from '$lib/auth.svelte';
	import { Lock, RefreshCw, Shield } from '@lucide/svelte';
	import { Button } from '$lib/ui/button';

	let { onLogin } = $props<{ onLogin: () => void }>();
	let inputToken = $state(authState.token);
	let isLoading = $state(false);
	let error = $state('');

	async function handleSubmit(e: Event) {
		e.preventDefault();
		isLoading = true;
		error = '';
		try {
			const res = await fetch('/envs', {
				headers: inputToken ? { Authorization: `Bearer ${inputToken}` } : {}
			});
			if (res.ok) {
				authState.setToken(inputToken);
				authState.isAuthed = true;
				onLogin();
			} else {
				error = 'Invalid token or unauthorized';
			}
		} catch (err: any) {
			error = err.message || 'Connection error';
		} finally {
			isLoading = false;
		}
	}
</script>

<div class="mx-auto mt-20 w-full max-w-md">
	<div class="rounded-xl border p-8 shadow-xl bg-card">
		<div class="mb-6 flex flex-col items-center text-center">
			<div
				class="mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-indigo-100 dark:bg-indigo-800/30"
			>
				<Shield class="h-6 w-6 text-indigo-600 dark:text-indigo-300" />
			</div>
			<h2 class="text-2xl font-bold">Authentication</h2>
			<p class="mt-2 text-sm">Enter your access token to view configurations.</p>
		</div>

		<form onsubmit={handleSubmit} class="space-y-4">
			<div>
				<div class="relative">
					<Lock class="absolute left-3 top-1/2 h-5 w-5 -translate-y-1/2" />
					<input
						type="password"
						bind:value={inputToken}
						placeholder="Bearer Token"
						class="w-full rounded-lg border bg-transparent py-2.5 pl-10 pr-4 text-sm outline-none transition-all"
						disabled={isLoading}
					/>
				</div>
			</div>
			{#if error}
				<p class="text-sm text-red-500 font-medium">{error}</p>
			{/if}
			<Button type="submit" disabled={isLoading} class="w-full">
				{#if isLoading}
					<RefreshCw class="animate-spin" />
					Verifying...
				{:else}
					Access Configuration
				{/if}
			</Button>
		</form>
	</div>
</div>
