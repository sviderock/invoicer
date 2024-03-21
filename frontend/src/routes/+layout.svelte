<script lang="ts">
	import { Toaster } from '$lib/components/ui/sonner';
	import { QueryClientProvider } from '@tanstack/svelte-query';
	import { SvelteQueryDevtools } from '@tanstack/svelte-query-devtools';
	import '../app.css';
	import type { LayoutData } from './$types';

	export let data: LayoutData;
</script>

<QueryClientProvider client={data.queryClient}>
	<div class="flex h-full w-full bg-slate-800 p-2 text-white">
		<slot />
	</div>
	<Toaster />
	<SvelteQueryDevtools />
</QueryClientProvider>
