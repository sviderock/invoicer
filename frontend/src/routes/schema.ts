import { type PlainMessage } from '@bufbuild/protobuf';
import type { File, UpdateFileNameRequest } from 'proto/file_pb';
import * as v from 'valibot';

// Define outside the load function so the adapter can be cached
export type FileSchema = typeof FileSchema;
export const FileSchema = v.object({
	id: v.string(),
	name: v.string(),
	path: v.string(),
	thumbnail: v.string(),
	size: v.number(),
	ext: v.string()
}) satisfies v.BaseSchema<PlainMessage<File>>;

export type UpdateFileSchema = typeof UpdateFileSchema;
export const UpdateFileSchema = v.object({
	id: v.string([v.minLength(1)]),
	name: v.string([v.minLength(1)])
}) satisfies v.BaseSchema<PlainMessage<UpdateFileNameRequest>>;
