import type { PageLoad } from './$types';

export const load: PageLoad = async ({ parent }) => {
	// const { queryClient } = await parent();

	console.log('load page data');
};
