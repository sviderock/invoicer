<script lang="ts">
	import { api } from '$lib/api';
	import { Button } from '$lib/components/ui/button';
	import { createMutation, createQuery } from '@tanstack/svelte-query';
	import { PDFDocument, StandardFonts, rgb, degrees } from 'pdf-lib';
	import { toast } from 'svelte-sonner';

	let files = $state<FileList | null>();

	const uploadedFiles = createQuery({ queryKey: ['uploadedFiles'], queryFn: api.getUploadedFiles });
	const createFile = createMutation({ mutationKey: ['upload'], mutationFn: api.uploadFile });

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

	async function onCreateFile() {
		if (!files?.length) return;
		const file = files[0];
		const formData = new FormData();
		formData.append('file', file, file.name);

		$createFile.mutate(formData, {
			onSuccess: (data) => {
				toast.success(`Uploaded file: ${data.file?.name}`);
			}
		});
	}

	$effect(() => {
		console.log($uploadedFiles.data);
	});
</script>

<div class="flex flex-col gap-2">
	<span class="text-xl">Uploaded images</span>
	<div class="grid gap-2"></div>
</div>

<div class="flex h-full w-auto items-center gap-4 rounded-md border-2 p-8">
	<label for="img">Upload a picture:</label>
	<input accept="application/pdf" bind:files id="img" name="img" type="file" />
	<Button on:click={onCreateFile} disabled={!files?.length}>Call FileUpload:UploadFile()</Button>
</div>
