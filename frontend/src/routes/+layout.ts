import { browser } from '$app/environment';
import { MutationCache, QueryCache, QueryClient, type QueryKey } from '@tanstack/svelte-query';
import type { LayoutLoad } from './$types';
import { createConnectTransport } from '@connectrpc/connect-web';
import { toast } from 'svelte-sonner';
import { env } from '$env/dynamic/public';

function getPath(url: QueryKey): string {
	const validUrl = url.some((i) => typeof i === 'string' || typeof i === 'number');
	if (!validUrl) {
		throw new Error('Invalid QueryKey');
	}
	return (url as Array<string | number>).map((i) => `${i}`.toLowerCase()).join('/');
}

export const load: LayoutLoad = async () => {
	const queryClient = new QueryClient({
		defaultOptions: {
			queries: {
				enabled: browser,
				staleTime: 60 * 1000,
				queryFn: async ({ queryKey }) => {
					const path = getPath(queryKey);
					const data = await fetch(`${env.PUBLIC_API_URL}/${path}`, { method: 'GET' });
					const json = await data.json();
					return json;
				}
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
		baseUrl: env.PUBLIC_API_URL!,
		useBinaryFormat: true
	});

	return { queryClient, grpcTransport };
};
