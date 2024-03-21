import { env } from '$env/dynamic/public';
import type { PlainMessage } from '@bufbuild/protobuf';
import {
	FileUploadResponse,
	GetFilesResponse,
	UpdateFileNameRequest,
	UpdateFileNameResponse
} from 'proto/file_pb';

export const api = (customFetch = fetch) => ({
	getUploadedFiles: async () => {
		const res = await customFetch(`${env.PUBLIC_API_URL}/get-files`, { method: 'GET' });
		const json = await res.json();
		return new GetFilesResponse(json).files;
	},

	uploadFile: async (formData: FormData) => {
		const res = await customFetch(`${env.PUBLIC_API_URL}/file-upload`, {
			method: 'POST',
			body: formData
		});
		const json = await res.json();
		return new FileUploadResponse(json).file;
	},

	updateFile: async (data: PlainMessage<UpdateFileNameRequest>) => {
		const res = await customFetch(`${env.PUBLIC_API_URL}/update-file`, {
			method: 'POST',
			body: JSON.stringify(data),
			headers: { 'Content-Type': 'application/json' }
		});
		const json = await res.json();
		return new UpdateFileNameResponse(json).file;
	}
});
