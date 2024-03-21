<script lang="ts">
	import { Toaster } from '$lib/components/ui/sonner';
	import { QueryClientProvider } from '@tanstack/svelte-query';
	import { SvelteQueryDevtools } from '@tanstack/svelte-query-devtools';
	import '../app.css';
	import type { LayoutData } from './$types';

	export let data: LayoutData;
</script>

<QueryClientProvider client={data.queryClient}>
	<main class="bg-slate-800 w-full h-full text-white p-24">
		<slot />
	</main>
	<Toaster />
	<SvelteQueryDevtools />
</QueryClientProvider>
