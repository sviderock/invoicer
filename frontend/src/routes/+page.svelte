<script lang="ts">
	import { createPromiseClient, type Transport } from '@connectrpc/connect';
	import { PingService } from 'proto/ping/v1/ping_connect';
	import { ExampleService } from 'proto/examples/v1/examples_connect';
	import { getContext, onMount } from 'svelte';
	import { Example, type Examples } from 'proto/examples/v1/examples_pb';
	import { FileUploadService } from 'proto/file_upload/v1/file_upload_connect';

	let examples = $state<Example[]>([]);
	let files = $state<FileList | null>();

	const transport: Transport = getContext('transport');
	const pingClient = createPromiseClient(PingService, transport);
	const examapleClient = createPromiseClient(ExampleService, transport);
	const fileUploadClient = createPromiseClient(FileUploadService, transport);
</script>

<main class="bg-slate-800 w-full h-full text-white p-24">
	<div class="border-2 rounded-md h-full p-8 flex-col flex gap-4">
		<button
			class="border-2 p-4 bg-red-300"
			on:click={async () => {
				const response = await pingClient.ping({ number: 123n, text: '123' });
				console.log('PING', response);
			}}
		>
			Call Ping()
		</button>
		<button
			class="border-2 p-4 bg-blue-300"
			on:click={async () => {
				// examples = [...examples, new Example()];
				const response = await examapleClient.index({ number: 123n, text: '123' });
				console.log('Example:Index', response);
			}}
		>
			Call Example:Index()
		</button>
		<button
			class="border-2 p-4 bg-green-300"
			on:click={async () => {
				if (!files?.length) return;

				const formData = new FormData();
				formData.append('file', files[0]);

				const data = await fetch('http://localhost:9002/file-upload', {
					method: 'post',
					body: formData
					// headers: {
					// 	'Content-Type': 'multipart/form-data'
					// }
				});
				console.log(data);
			}}
		>
			Call FileUpload:UploadFile()
		</button>

		<label for="img">Upload a picture:</label>
		<input accept="image/*" bind:files id="img" name="img" type="file" />
	</div>
</main>
