<script lang="ts">
	import { api } from '$lib/api';
	import { Button } from '$lib/components/ui/button';
	import type { PlainMessage } from '@bufbuild/protobuf';
	import { createMutation, createQuery } from '@tanstack/svelte-query';
	import { GetTemplatesResponse } from 'proto/template_pb';
	import { toast } from 'svelte-sonner';
	import { slide } from 'svelte/transition';
	import LineMdUploadLoop from '~icons/line-md/upload-loop';
	import { TemplateCard } from '../lib/components/template-card';

	type TemplateItem = PlainMessage<GetTemplatesResponse['templates'][number]>;
	let { data } = $props();
	let files = $state<FileList | null>();
	let newTemplateUploaded = $state(false);

	const templatesList = createQuery({
		queryKey: ['get-templates', data.fetch],
		queryFn: () => api(data.fetch).getTemplates()
	});

	const createFile = createMutation({
		mutationKey: ['upload'],
		mutationFn: api().uploadFile,
		onMutate: async () => {
			await data.queryClient.cancelQueries({ queryKey: ['get-templates'] });
			const previousTemplates = data.queryClient.getQueryData<TemplateItem[]>(['get-templates']);
			const file = files?.[0];

			if (previousTemplates && file) {
				data.queryClient.setQueryData<TemplateItem[]>(
					['get-templates'],
					[
						...previousTemplates,
						{
							id: 0,
							name: file.name,
							size: file.size,
							path: '',
							thumbnail: '',
							ext: '',
							createdAt: 0n,
							updatedAt: 0n
						} satisfies TemplateItem
					]
				);
			}

			return { previousTemplates };
		},
		onSuccess: (data) => {
			toast.success(`Template created: ${data?.name}`);
			// files = null;
			newTemplateUploaded = true;
			setTimeout(() => {
				newTemplateUploaded = false;
			}, 2000);
		},
		onError: (_err, _vars, context) => {
			if (context?.previousTemplates) {
				data.queryClient.setQueryData<TemplateItem[]>(['get-templates'], context.previousTemplates);
			}
		},
		onSettled: () => {
			data.queryClient.invalidateQueries({ queryKey: ['get-templates'] });
		}
	});

	function getFormData() {
		if (!files?.length) return;
		const file = files[0];
		const formData = new FormData();
		formData.append('file', file, file.name);
		return formData;
	}

	async function onCreateFile() {
		const formData = getFormData();
		if (!formData) return;
		$createFile.mutate(formData);
	}

	const getUploadingStatus = $derived((i: number) => {
		if (!$templatesList.data) return;
		if (i !== $templatesList.data?.length - 1) return;

		if ($templatesList.isPending && !newTemplateUploaded) return 'uploading';
		if (newTemplateUploaded) return 'uploaded';
		return undefined;
	});
</script>

<div class="flex w-full flex-col items-start justify-start gap-12 rounded-sm border-2 p-16">
	<div class="flex flex-col gap-4">
		<label for="file-upload">Upload PDF file:</label>
		<div class="flex items-center justify-between gap-4">
			<input
				accept="application/pdf"
				id="file-upload"
				type="file"
				class="border-input cursor-pointer rounded-sm border file:mr-4 file:h-9 file:border-0 file:bg-slate-900 file:bg-transparent file:px-4 file:text-sm file:font-medium"
				bind:files
			/>
			<Button
				variant="outline"
				on:click={onCreateFile}
				disabled={!files?.length || $createFile.isPending}
				class="h-auto w-auto gap-2"
			>
				Upload
				{#if files?.length}
					<div transition:slide={{ duration: 300, axis: 'x' }}>
						<LineMdUploadLoop class="text-md" />
					</div>
				{/if}
			</Button>
		</div>
	</div>

	<div class="flex w-full flex-col gap-4">
		<span class="text-xl">Uploaded images</span>
		<div class="grid grid-cols-1 gap-6 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5">
			{#if $templatesList.data}
				{#each $templatesList.data as template, i}
					<TemplateCard {template} pending={getUploadingStatus(i)} />
				{/each}
			{/if}
		</div>
	</div>
</div>
