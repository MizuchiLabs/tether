<script lang="ts">
	import { RefreshCw } from '@lucide/svelte';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { loggedIn } from '$lib/store.svelte';
	import Logo from '$lib/assets/logo.svelte';
	import { api } from '$lib/api';

	let secret = $state('');
	let isLoading = $state(false);
	let error = $state<string | null>(null);

	async function handleLogin(e: SubmitEvent) {
		e.preventDefault();

		isLoading = true;
		error = null;
		try {
			await api.login(secret);
		} catch (err: any) {
			error = err.message || 'Failed to sign in. Please check your token.';
		} finally {
			isLoading = false;
		}
	}
</script>

{#if !loggedIn.current}
	<section class="flex min-h-screen px-4 py-16 md:py-32 dark:bg-transparent">
		<form
			onsubmit={handleLogin}
			class="m-auto h-fit w-full max-w-sm overflow-hidden rounded-[calc(var(--radius)+.125rem)] border bg-muted shadow-md shadow-zinc-950/5 dark:[--color-muted:var(--color-zinc-900)]"
		>
			<div class="-m-px rounded-[calc(var(--radius)+.125rem)] border bg-card p-8 pb-6">
				<div class="text-center">
					<a href="/" aria-label="go home" class="mx-auto block w-fit">
						<Logo class="size-7" />
					</a>
					<h1 class="mt-4 mb-1 text-xl font-semibold">Sign In to Tether</h1>
					<p class="text-sm">Enter your access token to view configurations</p>
				</div>

				<div class="mt-6 space-y-6">
					<div class="space-y-2">
						<Label for="pwd" class="text-title text-sm">Bearer Token</Label>
						<Input
							bind:value={secret}
							type="password"
							required
							name="pwd"
							placeholder="Enter your token"
							disabled={isLoading}
						/>
						{#if error}
							<p class="text-sm text-red-500">{error}</p>
						{/if}
					</div>

					<Button type="submit" disabled={isLoading} class="w-full">
						{#if isLoading}
							<RefreshCw class="animate-spin" />
							Verifying...
						{:else}
							Sign In
						{/if}
					</Button>
				</div>
			</div>
		</form>
	</section>
{/if}
