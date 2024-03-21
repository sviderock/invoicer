<script>
	import { createConnectTransport } from '@connectrpc/connect-web';
	import { setContext } from 'svelte';
	import '../app.css';

	const transport = createConnectTransport({
		baseUrl: 'http://localhost:9002',
		useBinaryFormat: true
	});
	setContext('transport', transport);
</script>

<slot />
