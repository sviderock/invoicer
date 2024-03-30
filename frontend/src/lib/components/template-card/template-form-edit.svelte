<script lang="ts">
	import { api } from '$lib/api';
	import { Button } from '$lib/components/ui/button';
	import { Label } from '$lib/components/ui/label';
	import { Sheet, SheetContent, SheetTrigger } from '$lib/components/ui/sheet';
	import type { PlainMessage } from '@bufbuild/protobuf';
	import { createQuery } from '@tanstack/svelte-query';
	import DOMPurify from 'dompurify';
	import 'pdfjs-dist/build/pdf.worker.min.mjs';
	import 'pdfjs-dist/web/pdf_viewer.css';
	import type { Template } from 'proto/template_pb';

	type Props = {
		template: PlainMessage<Template>;
	};

	let { template }: Props = $props();
	let sheetOpen = $state(false);

	const testHtml = createQuery({
		queryKey: ['test-html', template.path],
		queryFn: () => api().getHtml(template.path),
		select: (text) => {
			DOMPurify.addHook('uponSanitizeElement', (node, data) => {
				if (node.id === 'sidebar' || node.id === 'outline') {
					return node.parentNode?.removeChild(node);
				}

				// if (node.classList.contains('loading-indicator')) {
				// 	return node.parentNode?.removeChild(node);
				// }

				if (data.tagName === 'style') {
					const content = node.textContent?.split('*/');
					if (content?.length === 2) {
						node.textContent = content[1];
					}
					return;
				}

				if (data.tagName === '#text' && node.parentElement) {
					node.parentElement.setAttribute('contenteditable', 'true');
					node.parentElement.classList.add(
						'transition-all',
						'border',
						'border-transparent',
						'hover:border-red-500'
					);
				}
			});
			const sanitized = DOMPurify.sanitize(text, { WHOLE_DOCUMENT: true, FORCE_BODY: true });
			return { sanitized };
		}
	});

	$effect(() => {});
</script>

<Sheet closeOnOutsideClick={false} bind:open={sheetOpen}>
	<SheetTrigger asChild let:builder>
		<Button
			variant="ghost"
			builders={[builder]}
			class="relative h-full w-full cursor-pointer p-0 after:absolute after:left-0 after:top-0 after:flex after:h-full after:w-full after:items-center after:justify-center after:rounded-md after:bg-transparent after:text-xl after:text-slate-800 after:opacity-0 after:transition after:content-['Open_Preview'] hover:after:bg-slate-800/50 hover:after:opacity-100"
		>
			<img src={template.thumbnail} alt={template.name} class="aspect-a4 rounded-md" />
		</Button>
	</SheetTrigger>

	<SheetContent class="flex sm:max-w-[unset]">
		<div class="flex w-full gap-4">
			<div class="min-w-[300px] rounded-sm border p-8">
				<div>
					<Label for="scale-slider">Scale</Label>
					<!-- <Slider
						id="scale-slider"
						value={[scale]}
						onValueChange={(value) => {
							scale = value[0];
						}}
						min={1}
						max={1.5}
						step={0.1}
					/> -->
				</div>
			</div>
			<div
				class="flex-3 relative h-full w-full flex-grow-0 overflow-auto rounded-sm *:hover:border *:hover:border-slate-800"
			>
				{@html $testHtml.data?.sanitized}
			</div>
		</div>
	</SheetContent>
</Sheet>
