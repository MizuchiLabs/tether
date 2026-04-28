<script lang="ts">
	import { Eye, EyeOff, RefreshCw } from '@lucide/svelte';
	import { Button } from '$lib/components/ui/button';
	import { Label } from '$lib/components/ui/label';
	import * as InputGroup from '$lib/components/ui/input-group';

	import { loggedIn } from '$lib/store.svelte';
	import Logo from '$lib/assets/logo.svelte';
	import { api } from '$lib/api';

	let secret = $state('');
	let showPassword = $state(false);
	let isLoading = $state(false);
	let error = $state<string | null>(null);

	async function handleLogin(e: SubmitEvent) {
		e.preventDefault();

		isLoading = true;
		error = null;
		try {
			await api.login(secret);
			secret = '';
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
					<h1 class="mt-4 mb-1 text-xl font-semibold">Authenticate with Tether</h1>
					<p class="text-sm">Enter your shared secret token to view agent configurations.</p>
				</div>

				<div class="mt-6 space-y-6">
					<div class="space-y-2">
						<Label for="pwd" class="text-title text-sm">Access Token</Label>
						<InputGroup.Root>
							<InputGroup.Input type={showPassword ? 'text' : 'password'} bind:value={secret} />
							<InputGroup.Addon align="inline-end">
								<InputGroup.Button
									aria-label="Show password"
									title="Show password"
									variant="ghost"
									size="icon-xs"
									class="h-7 w-7"
									onclick={() => (showPassword = !showPassword)}
								>
									{#if showPassword}
										<Eye size={16} />
									{:else}
										<EyeOff size={16} />
									{/if}
								</InputGroup.Button>
							</InputGroup.Addon>
						</InputGroup.Root>
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
