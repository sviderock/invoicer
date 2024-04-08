<script lang="ts">
	import { browser } from '$app/environment';
	import { api } from '$lib/api';
	import {
		EDIT_ACTIVE_CLASS,
		EDITABLE_CLASS,
		initiallySanitizeHtml
	} from '$lib/components/template-card/utils';
	import { Button } from '$lib/components/ui/button';
	import { FormButton, FormControl, FormField } from '$lib/components/ui/form';
	import { Label } from '$lib/components/ui/label';
	import { Sheet, SheetContent, SheetTrigger } from '$lib/components/ui/sheet';
	import { Textarea } from '$lib/components/ui/textarea';
	import type { PlainMessage } from '@bufbuild/protobuf';
	import { createMutation, createQuery, useQueryClient } from '@tanstack/svelte-query';
	import DOMPurify from 'dompurify';
	import { nanoid } from 'nanoid';
	import 'pdfjs-dist/build/pdf.worker.min.mjs';
	import 'pdfjs-dist/web/pdf_viewer.css';
	import type { Template } from 'proto/template_pb';
	import { onMount } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { defaults, superForm } from 'sveltekit-superforms';
	import { valibot } from 'sveltekit-superforms/adapters';
	import * as v from 'valibot';

	type Props = {
		template: PlainMessage<Template>;
	};

	const Schema = v.object({
		editingField: v.object({
			innerHtml: v.string()
		}),
		fields: v.array(
			v.object({
				id: v.number(),
				x: v.number(),
				y: v.number(),
				name: v.string([v.minLength(1)])
			})
		)
	});
	const valibotSchema = valibot(Schema);

	const queryClient = useQueryClient();
	let { template }: Props = $props();
	let sheetOpen = $state(true);
	let editingRef = $state<HTMLElement | null>(null);
	let containerRef = $state<HTMLElement | null>(null);

	const updateTemplateHtml = createMutation({
		mutationKey: ['update-template-html'],
		mutationFn: api().updateTemplateHtml,
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['html-template', template.path] });
			toast.success('Template successfully updated');
		}
	});

	const form = superForm(
		defaults(valibotSchema, { defaults: { fields: [], editingField: { innerHtml: '' } } }),
		{
			id: nanoid(),
			SPA: true,
			dataType: 'json',
			validators: valibotSchema,
			resetForm: false,
			onSubmit: () => {
				const body = DOMPurify.sanitize(containerRef!.innerHTML, { FORCE_BODY: true });
				$updateTemplateHtml.mutate({ id: template.id, html: body });
			}
		}
	);
	const { form: formData, enhance } = form;

	const htmlTemplate = createQuery({
		queryKey: ['html-template', template.path],
		queryFn: () => api().getHtml(template.path),
		select: initiallySanitizeHtml
	});

	onMount(() => {
		function onFocus(e: FocusEvent) {
			if (!(e.target instanceof HTMLElement)) return;

			if (editingRef) {
				editingRef.classList.remove(EDIT_ACTIVE_CLASS);
				editingRef.contentEditable = 'false';
			}

			editingRef = e.target;
			editingRef.classList.add(EDIT_ACTIVE_CLASS);
			editingRef.contentEditable = 'true';
			editingRef.addEventListener('input', (e) => {
				e.stopPropagation();
				if (e.target instanceof HTMLElement) {
					$formData.editingField.innerHtml = e.target.innerHTML;
				}
			});
			$formData.editingField.innerHtml = editingRef.innerHTML;
		}

		containerRef?.addEventListener('focus', onFocus, true);
		return () => containerRef?.removeEventListener('focus', onFocus, true);
	});

	$effect(() => {
		function onInput(e: Event) {
			e.stopPropagation();
			if (e.target instanceof HTMLElement) {
				$formData.editingField.innerHtml = e.target.innerHTML;
			}
		}

		editingRef?.addEventListener('input', onInput);
		return () => editingRef?.removeEventListener('input', onInput);
	});
</script>

<Sheet closeOnOutsideClick={false} bind:open={sheetOpen}>
	<div class="relative">
		<img src={template.thumbnail} alt={template.name} class="aspect-a4 rounded-md" />
		<SheetTrigger asChild let:builder>
			<Button
				variant="ghost"
				builders={[builder]}
				class="absolute left-0 top-0 h-full w-full bg-transparent text-xl text-slate-800 opacity-0 transition hover:bg-slate-800/50 hover:opacity-100"
			>
				Edit Template
			</Button>
		</SheetTrigger>
	</div>

	<SheetContent class="flex w-full p-0 px-8 pt-8 sm:max-w-[unset]">
		<div class="flex w-full gap-4">
			{#if $htmlTemplate.data}
				{@html `<${''}style>${$htmlTemplate.data.styles}</${''}style>`}
			{/if}

			<div class="h-min flex-1 rounded-sm border p-8">
				<form method="post" use:enhance class="flex flex-col gap-4">
					<FormField {form} name="editingField.innerHtml" class="w-full">
						<FormControl let:attrs>
							<Label class="text-lg" for="">Text Value</Label>
							<Textarea
								{...attrs}
								value={$formData.editingField.innerHtml}
								on:input={(e) => {
									$formData.editingField.innerHtml = e.currentTarget.value;
									editingRef!.innerHTML = DOMPurify.sanitize(e.currentTarget.value)
								}}
							/>
						</FormControl>
					</FormField>

					<FormButton>Submit</FormButton>
				</form>
			</div>
			<div bind:this={containerRef} class="relative h-full w-full flex-1 overflow-auto">
				{#if $htmlTemplate.data}
					<!-- For those weird ${''} pieces: https://github.com/sveltejs/svelte/issues/5292#issuecomment-787743573 -->
					{@html `<${''}style>${$htmlTemplate.data.styles}</${''}style>`}
					{@html $htmlTemplate.data.sanitizedHtml}
				{/if}
			</div>
		</div>
	</SheetContent>
</Sheet>

<style type="text/css">
	#page-container {
		background: none;
	}
	:global(#page-container) {
		background: none !important;
		width: max-content;
		position: relative !important;
		display: flex;
		flex-direction: column;
		gap: 10px;
		outline: none;
		overflow: hidden !important;
	}

	:global(#page-container > div) {
		margin: 0;
		border-radius: 4px;
	}

	:global(.editable) {
		/* user-select: none; */
		z-index: 1;
	}

	:global(.editable:before) {
		content: '';
		position: absolute;
		top: 0;
		left: 0;
		height: calc(100% + 8px);
		width: calc(100% + 8px);
		transform: translate(-4px, -4px);
		transition: all 0.1s ease;
		border: 2px solid transparent;
		border-radius: 4px;
		z-index: -1;
	}

	:global(.editable:hover:not(:has(.editable:hover)):not(.edit-active):before) {
		border-color: #64748b;
		background-color: #e2e8f0;
	}

	:global(.edit-active:before) {
		transition: none !important;
		/* background-color: #e2e8f0; */
		border-color: #64748b;
	}
</style>
