<script lang="ts">
	import { cn } from '#lib/utils.js';

	let { tone = 'ok', pulse = false }: { tone?: 'ok' | 'error'; pulse?: boolean } = $props();
</script>

<span
	aria-hidden="true"
	class={cn(
		'size-2 shrink-0 rounded-full',
		tone === 'ok' && 'bg-primary',
		tone === 'error' && 'bg-destructive',
		pulse && 'animate-pulse'
	)}
></span>
