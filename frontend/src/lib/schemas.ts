import { type PlainMessage } from '@bufbuild/protobuf';
import type { Template, UpdateTemplateRequest } from 'proto/template_pb';
import * as v from 'valibot';

// Define outside the load function so the adapter can be cached
export type TemplateSchema = typeof TemplateSchema;
export const TemplateSchema = v.object({
	id: v.number(),
	name: v.string(),
	path: v.string(),
	thumbnail: v.string(),
	size: v.number(),
	ext: v.string(),
	createdAt: v.any(),
	updatedAt: v.any()
}) satisfies v.BaseSchema<PlainMessage<Template>>;

export type UpdateTemplateSchema = typeof UpdateTemplateSchema;
export const UpdateTemplateSchema = v.object({
	name: v.string([v.minLength(1)])
}) satisfies v.BaseSchema<PlainMessage<UpdateTemplateRequest>>;
