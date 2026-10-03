<script lang="ts">
	import { api, ApiError } from '#lib/api.js';
	import Logo from '#lib/assets/logo.svelte';
	import { Button } from '#lib/components/ui/button/index.js';
	import * as Card from '#lib/components/ui/card/index.js';
	import * as Field from '#lib/components/ui/field/index.js';
	import * as InputGroup from '#lib/components/ui/input-group/index.js';
	import { Spinner } from '#lib/components/ui/spinner/index.js';
	import { EyeIcon, EyeOffIcon } from '@lucide/svelte';

	let secret = $state('');
	let showSecret = $state(false);
	let loading = $state(false);
	let error = $state<string | null>(null);

	async function login(e: SubmitEvent) {
		e.preventDefault();
		loading = true;
		error = null;
		try {
			await api.login(secret);
			secret = '';
		} catch (err) {
			if (err instanceof ApiError) {
				error = err.status === 401 ? 'Wrong token, try again.' : err.message;
			} else {
				error = "Can't reach Tether right now.";
			}
		} finally {
			loading = false;
		}
	}
</script>

<main class="flex min-h-screen items-center justify-center px-4 py-16">
	<Card.Root class="w-full max-w-sm">
		<Card.Header class="justify-items-center text-center">
			<Logo class="mb-2 size-8" />
			<Card.Title>Sign in to Tether</Card.Title>
			<Card.Description>Use the token from <code>TETHER_TOKEN</code>.</Card.Description>
		</Card.Header>
		<Card.Content>
			<form onsubmit={login}>
				<Field.FieldGroup>
					<Field.Field data-invalid={!!error || undefined}>
						<Field.FieldLabel for="token">Token</Field.FieldLabel>
						<InputGroup.Root>
							<InputGroup.Input
								id="token"
								type={showSecret ? 'text' : 'password'}
								autocomplete="current-password"
								autofocus
								required
								bind:value={secret}
								aria-invalid={!!error || undefined}
							/>
							<InputGroup.Addon align="inline-end">
								<InputGroup.Button
									size="icon-xs"
									aria-label={showSecret ? 'Hide token' : 'Show token'}
									title={showSecret ? 'Hide token' : 'Show token'}
									onclick={() => (showSecret = !showSecret)}
								>
									{#if showSecret}
										<EyeOffIcon />
									{:else}
										<EyeIcon />
									{/if}
								</InputGroup.Button>
							</InputGroup.Addon>
						</InputGroup.Root>
						{#if error}
							<Field.FieldError>{error}</Field.FieldError>
						{/if}
					</Field.Field>

					<Button type="submit" disabled={loading || !secret}>
						{#if loading}
							<Spinner data-icon="inline-start" />
							Signing in...
						{:else}
							Sign in
						{/if}
					</Button>
				</Field.FieldGroup>
			</form>
		</Card.Content>
	</Card.Root>
</main>
