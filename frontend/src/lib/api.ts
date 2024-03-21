import { env } from '$env/dynamic/public';

export const api = {
	uploadFile: async (formData: FormData) => {
		const res = await fetch(`${env.PUBLIC_API_URL}/file-upload`, {
			method: 'POST',
			body: formData
		});

		return res.json();
	}
};
