<script lang="ts">
	import { api } from '$lib/api';
	import { Button } from '$lib/components/ui/button';
	import type { PlainMessage } from '@bufbuild/protobuf';
	import { createMutation, createQuery } from '@tanstack/svelte-query';
	import { PDFDocument, StandardFonts, degrees, rgb } from 'pdf-lib';
	import { GetFilesResponse } from 'proto/file_pb';
	import { toast } from 'svelte-sonner';
	import { slide } from 'svelte/transition';
	import LineMdUploadLoop from '~icons/line-md/upload-loop';
	import FileForm from './file-form.svelte';

	type FileItem = PlainMessage<GetFilesResponse['files'][number]>;
	let { data } = $props();
	let files = $state<FileList | null>();
	let newFileUploaded = $state(false);

	const uploadedFiles = createQuery({
		queryKey: ['get-files'],
		queryFn: () => api().getUploadedFiles()
	});

	const createFile = createMutation({
		mutationKey: ['upload'],
		mutationFn: api().uploadFile,
		onMutate: async () => {
			await data.queryClient.cancelQueries({ queryKey: ['get-files'] });
			const previousFiles = data.queryClient.getQueryData<FileItem[]>(['get-files']);
			const file = files?.[0];

			if (previousFiles && file) {
				data.queryClient.setQueryData<FileItem[]>(
					['get-files'],
					[
						...previousFiles,
						{
							id: 'new-file',
							name: file.name,
							size: file.size,
							path: '',
							thumbnail: '',
							ext: ''
						} satisfies FileItem
					]
				);
			}

			return { previousFiles };
		},
		onSuccess: (data) => {
			toast.success(`Uploaded file: ${data?.name}`);
			files = null;
			newFileUploaded = true;
			setTimeout(() => {
				newFileUploaded = false;
			}, 5000);
		},
		onError: (_err, _vars, context) => {
			if (context?.previousFiles) {
				data.queryClient.setQueryData<FileItem[]>(['get-files'], context.previousFiles);
			}
		},
		onSettled: () => {
			data.queryClient.invalidateQueries({ queryKey: ['get-files'] });
		}
	});

	async function processPDF(file: File) {
		const fileArrayBuffer = await file.arrayBuffer();
		const pdf = await PDFDocument.load(fileArrayBuffer);
		const helveticaFont = await pdf.embedFont(StandardFonts.Helvetica);
		const pages = pdf.getPages();
		const firstPage = pages[0];
		const { height } = firstPage.getSize();
		firstPage.drawText('This text was added with JavaScript!', {
			x: 5,
			y: height / 2 + 300,
			size: 50,
			font: helveticaFont,
			color: rgb(0.95, 0.1, 0.1),
			rotate: degrees(-45)
		});

		const pdfBytes = await pdf.save();
		const blob = new Blob([pdfBytes]);
		return blob;
	}

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
		if (!$uploadedFiles.data) return;
		if (i !== $uploadedFiles.data?.length - 1) return;

		if ($uploadedFiles.isPending && !newFileUploaded) return 'uploading';
		if (newFileUploaded) return 'uploaded';
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
			{#if $uploadedFiles.data}
				{#each $uploadedFiles.data as file, i}
					<FileForm {file} pending={getUploadingStatus(i)} />
				{/each}
			{/if}
		</div>
	</div>
</div>
