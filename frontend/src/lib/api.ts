import { env } from '$env/dynamic/public';
import type { PlainMessage } from '@bufbuild/protobuf';
import {
	FileUploadResponse,
	GetTemplatesResponse,
	Template,
	UpdateTemplateRequest,
	UpdateTemplateResponse
} from 'proto/template_pb';

export const api = (customFetch = fetch) => ({
	getTemplates: async () => {
		const res = await customFetch(`${env.PUBLIC_API_URL}/templates`, { method: 'GET' });
		const json = await res.json();
		return json;
		// return new GetTemplatesResponse(json).templates;
	},

	uploadFile: async (formData: FormData) => {
		const res = await customFetch(`${env.PUBLIC_API_URL}/templates`, {
			method: 'POST',
			body: formData
		});
		const json = await res.json();
		return new FileUploadResponse(json).template;
	},

	updateTemplate: async ({
		id,
		name
	}: Pick<Template, 'id'> & PlainMessage<UpdateTemplateRequest>) => {
		const res = await customFetch(`${env.PUBLIC_API_URL}/templates/${id}`, {
			method: 'PATCH',
			body: JSON.stringify({ name } satisfies PlainMessage<UpdateTemplateRequest>),
			headers: { 'Content-Type': 'application/json' }
		});
		const json = await res.json();
		return new UpdateTemplateResponse(json).template;
	},

	deleteTemplate: async ({ id }: Pick<Template, 'id'>) => {
		return customFetch(`${env.PUBLIC_API_URL}/templates/${id}`, {
			method: 'DELETE',
			headers: { 'Content-Type': 'text/html' }
		});
	},

	getHtml: async (url: string) => {
		const res = await customFetch(url);
		const text = await res.text();
		return text;
	}
});
