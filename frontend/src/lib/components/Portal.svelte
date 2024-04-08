<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import type { HTMLAttributes } from 'svelte/elements';

	interface $$Props extends HTMLAttributes<HTMLElement> {
		target?: HTMLElement | undefined | null;
	}

	let {
		// eslint-disable-next-line no-undef
		target = globalThis.document?.body,
		...restProps
	}: $$Props = $props();
	let ref = $state<HTMLElement>();
	onMount(() => {
		if (target) {
			target.appendChild(ref!);
		}
	});

	// https://github.com/sveltejs/svelte/issues/3088#issuecomment-1065827485
	onDestroy(() => {
		setTimeout(() => {
			if (ref?.parentNode) {
				ref.parentNode?.removeChild(ref);
			}
		});
	});
</script>

<div bind:this={ref} {...restProps}>
	<slot />
</div>
