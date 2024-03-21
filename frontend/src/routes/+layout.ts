import { browser } from '$app/environment';
import { MutationCache, QueryCache, QueryClient } from '@tanstack/svelte-query';
import type { LayoutLoad } from './$types';
import { createConnectTransport } from '@connectrpc/connect-web';
import { toast } from 'svelte-sonner';

export const load: LayoutLoad = async () => {
	const queryClient = new QueryClient({
		defaultOptions: {
			queries: {
				enabled: browser,
				staleTime: 60 * 1000
			}
		},
		queryCache: new QueryCache({
			onError: (err) => toast.error(err.message)
		}),
		mutationCache: new MutationCache({
			onError: (err) => toast.error(err.message)
		})
	});

	const grpcTransport = createConnectTransport({
		baseUrl: 'http://localhost:9002',
		useBinaryFormat: true
	});

	return { queryClient, grpcTransport };
};
