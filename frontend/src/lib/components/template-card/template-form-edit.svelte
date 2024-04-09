<script lang="ts">
	import { api } from '$lib/api';
	import {
		EDIT_ACTIVE_CLASS,
		FieldSchema,
		initiallySanitizeHtml
	} from '$lib/components/template-card/utils';
	import { Button } from '$lib/components/ui/button';
	import {
		FormButton,
		FormControl,
		FormField,
		FormFieldset,
		FormLabel,
		FormLegend
	} from '$lib/components/ui/form';
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
	import SuperDebug, { defaults, superForm } from 'sveltekit-superforms';
	import { valibot } from 'sveltekit-superforms/adapters';
	import MingcuteDelete2Line from '~icons/mingcute/delete-2-line';
	import * as v from 'valibot';

	type Props = {
		template: PlainMessage<Template>;
	};

	const Schema = v.object({
		editingField: v.object({
			innerHtml: v.string()
		}),
		fields: v.array(FieldSchema)
	});
	const valibotSchema = valibot(Schema);

	const queryClient = useQueryClient();
	let { template }: Props = $props();
	let sheetOpen = $state(true);
	let editingRef = $state<HTMLElement | null>(null);
	let containerRef = $state<HTMLElement | null>(null);
	let fieldRefs = $state<Record<FieldSchema['id'], HTMLElement | undefined>>({});

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

	$effect(() => {
		if (!$htmlTemplate.data) return;
		$formData.fields = $htmlTemplate.data.fields;
	});

	$effect(() => {
		console.log(123);
		if (!containerRef) return;
		const all = containerRef.querySelectorAll('*[data-field-id]');
		all.forEach((el) => {
			if (!(el instanceof HTMLElement)) return;
			fieldRefs[el.dataset.fieldId!] = el;
		});

		function onFocus(e: FocusEvent) {
			if (!(e.target instanceof HTMLElement)) return;

			if (editingRef) {
				editingRef.classList.remove(EDIT_ACTIVE_CLASS);
				editingRef.contentEditable = 'false';
			}

			editingRef = e.target;
			editingRef.classList.add(EDIT_ACTIVE_CLASS);
			editingRef.contentEditable = 'true';
			$formData.editingField.innerHtml = editingRef.innerHTML;
		}

		function onFocusOut(e: FocusEvent) {
			if (!(e.target instanceof HTMLElement)) return;
			if (editingRef) {
				editingRef.classList.remove(EDIT_ACTIVE_CLASS);
				editingRef.contentEditable = 'false';
			}
		}

		containerRef?.addEventListener('focus', onFocus, true);
		containerRef?.addEventListener('focusout', onFocusOut, true);

		return () => {
			containerRef?.removeEventListener('focus', onFocus, true);
			containerRef?.removeEventListener('focusout', onFocusOut, true);
		};
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

	function addField() {
		if (!editingRef) return;

		if (!editingRef.dataset.fieldId) {
			const newField = {
				id: nanoid(),
				name: `Field #${$formData.fields.length + 1}`
			} satisfies FieldSchema;
			$formData.fields = [...$formData.fields, newField];

			editingRef.dataset.fieldId = newField.id;
			editingRef.dataset.fieldName = newField.name;
		}
	}

	function deleteField(i: number) {
		const ref = fieldRefs[$formData.fields[i].id]!;
		delete ref.dataset.fieldId;
		delete ref.dataset.fieldName;
		$formData.fields = $formData.fields.filter((_, idx) => idx !== i);
	}

	function focusField(field: HTMLElement) {}
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
				<!-- For those weird ${''} pieces: https://github.com/sveltejs/svelte/issues/5292#issuecomment-787743573 -->
				{@html `<${''}style>${$htmlTemplate.data.styles}</${''}style>`}
			{/if}

			<div class="h-min flex-[0.5] rounded-sm border p-8">
				<SuperDebug data={$formData} />
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

					<FormFieldset {form} name="fields" class="w-full">
						<FormLegend>Fields</FormLegend>
						<Button on:click={addField} disabled={!editingRef}>Turn into field</Button>
						{#each $formData.fields as _, i}
							<FormField
								{form}
								name="fields[{i}]"
								class="flex items-center justify-between gap-2 border border-slate-500 p-2"
							>
								<FormControl let:attrs>
									<FormLabel>{$formData.fields[i].name}</FormLabel>
									<span
										role="presentation"
										on:click={() => fieldRefs[$formData.fields[i].id]!.focus()}
									>
										{$formData.fields[i].id}
									</span>
									<Button variant="destructive" size="icon" on:click={() => deleteField(i)}>
										<MingcuteDelete2Line class="text-xs" />
									</Button>
								</FormControl>
							</FormField>
						{/each}
					</FormFieldset>

					<FormButton>Submit</FormButton>
				</form>
			</div>

			{#if $htmlTemplate.data}
				<div bind:this={containerRef} class="relative h-full w-full flex-1 overflow-auto">
					{@html $htmlTemplate.data.sanitizedHtml}
				</div>
			{/if}
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
