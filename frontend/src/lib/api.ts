import { env } from '$env/dynamic/public';
import { FileUploadResponse } from 'proto/proto_pb';

export const api = {
	getUploadedFiles: async () => {
		const res = await fetch(`${env.PUBLIC_API_URL}/get-files`, { method: 'GET' });
		return res.json();
	},
	uploadFile: async (formData: FormData) => {
		const res = await fetch(`${env.PUBLIC_API_URL}/file-upload`, {
			method: 'POST',
			body: formData
		});
		const json = await res.json();
		return json as FileUploadResponse;
	}
};
