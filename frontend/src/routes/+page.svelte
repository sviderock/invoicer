<script lang="ts">
	import { api } from '$lib/api';
	import { Button } from '$lib/components/ui/button';
	import { createMutation } from '@tanstack/svelte-query';

	let files = $state<FileList | null>();
	const createFile = createMutation({ mutationKey: ['upload'], mutationFn: api.uploadFile });

	function onCreateFile() {
		if (!files?.length) return;
		const formData = new FormData();
		formData.append('file', files[0]);
		$createFile.mutate(formData);
	}
</script>

<div class="flex h-full w-auto items-center gap-4 rounded-md border-2 p-8">
	<label for="img">Upload a picture:</label>
	<input accept="application/pdf" bind:files id="img" name="img" type="file" />
	<Button on:click={onCreateFile} disabled={!files?.length}>Call FileUpload:UploadFile()</Button>
</div>
