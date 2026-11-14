<script lang="ts">
	import Github from '#lib/assets/github.svelte';
</script>

<!-- GitHub on the left, every other link on the right. In the page flow, a fixed one covers content on a phone. -->
<footer
	class="flex flex-wrap items-center justify-between gap-x-4 gap-y-2 pb-3 text-xs text-muted-foreground"
>
	<a
		href="https://github.com/mizuchilabs/tether"
		target="_blank"
		rel="noopener noreferrer"
		class="flex items-center gap-1.5 rounded-sm font-mono transition-colors outline-none hover:text-foreground focus-visible:ring-[3px] focus-visible:ring-ring/50"
	>
		<Github class="size-4" />
		<span>mizuchilabs/<span class="font-medium text-foreground">tether</span></span>
	</a>
</footer>
