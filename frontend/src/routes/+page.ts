import { api } from '$lib/api';
import type { PageLoad } from './$types';

export const prerender = true;
export const ssr = true;

export const load: PageLoad = async ({ parent, fetch }) => {
	const { queryClient } = await parent();
	await queryClient.prefetchQuery({
		queryKey: ['get-templates', fetch],
		queryFn: () => api(fetch).getTemplates()
	});
	return { fetch };
};
