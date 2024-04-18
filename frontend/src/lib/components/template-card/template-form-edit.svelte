<script lang="ts">
	import { api } from '$lib/api';
	import {
		EDITABLE_CLASS,
		EDIT_ACTIVE_CLASS,
		FieldSchema,
		IncrementField,
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
	import { Textarea, type FormTextareaEvent } from '$lib/components/ui/textarea';
	import type { PlainMessage } from '@bufbuild/protobuf';
	import { createMutation, createQuery, useQueryClient } from '@tanstack/svelte-query';
	import DOMPurify from 'dompurify';
	import { nanoid } from 'nanoid';
	import 'pdfjs-dist/build/pdf.worker.min.mjs';
	import 'pdfjs-dist/web/pdf_viewer.css';
	import type { Template } from 'proto/template_pb';
	import { toast } from 'svelte-sonner';
	import SuperDebug, { defaults, superForm } from 'sveltekit-superforms';
	import { valibot } from 'sveltekit-superforms/adapters';
	import * as v from 'valibot';
	import LetsIconsBack from '~icons/lets-icons/back';
	import MingcuteDelete2Line from '~icons/mingcute/delete-2-line';

	type Props = {
		template: PlainMessage<Template>;
	};

	const Schema = v.object({
		editingField: v.object({
			originalText: v.string(),
			innerText: v.string()
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
		defaults(valibotSchema, {
			defaults: { fields: [], editingField: { originalText: '', innerText: '' } }
		}),
		{
			id: nanoid(),
			SPA: true,
			dataType: 'json',
			validators: valibotSchema,
			resetForm: false,
			onSubmit: () => {
				const htmlContent = [$htmlTemplate.data!.styles, containerRef!.innerHTML.trim()].join('');
				const body = DOMPurify.sanitize(htmlContent, { FORCE_BODY: true });
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
		if (!editingRef) return;
		editingRef.innerText = $formData.editingField.innerText;
	});

	$effect(() => {
		if (!containerRef) return;
		const all = containerRef.querySelectorAll('*[data-field-id]');
		all.forEach((el) => {
			if (!(el instanceof HTMLElement)) return;
			fieldRefs[el.dataset.fieldId!] = el;
		});

		function onFocus(e: FocusEvent) {
			console.log('focus');
			if (!(e.target instanceof HTMLElement)) return;

			if (editingRef) {
				editingRef.classList.remove(EDIT_ACTIVE_CLASS);
				editingRef.contentEditable = 'false';
			}

			editingRef = e.target;
			editingRef.classList.add(EDIT_ACTIVE_CLASS);
			editingRef.contentEditable = 'true';
			$formData.editingField.originalText = editingRef.innerText;
			$formData.editingField.innerText = editingRef.innerText;
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
				$formData.editingField.innerText = e.target.innerHTML;
			}
		}

		editingRef?.addEventListener('input', onInput);
		return () => editingRef?.removeEventListener('input', onInput);
	});

	$effect(() => {
		$formData.fields.forEach((field) => {
			const ref = fieldRefs[field.id]!;
			ref.dataset.fieldId = field.id;
			ref.dataset.fieldName = field.name;
			ref.dataset.fieldDataType = field.data.type;
			ref.dataset.fieldDataStartFrom = `${(field.data as IncrementField).startFrom}`;
		});
	});

	function removeEditable(el: HTMLElement) {
		el.removeAttribute('contenteditable');
		el.removeAttribute('tabindex');
	}

	function deleteField(i: number) {
		const node = fieldRefs[$formData.fields[i].id]!;
		node.parentNode?.removeChild(node);
		$formData.fields = $formData.fields.filter((_, idx) => idx !== i);
	}

	function createTextEl(text: string) {
		const textSpan = document.createElement('span');
		textSpan.innerText = text;
		textSpan.tabIndex = 0;
		textSpan.classList.add(EDITABLE_CLASS);
		return textSpan;
	}

	function createFieldEl() {
		const fieldSpan = document.createElement('span');
		fieldSpan.classList.add(EDITABLE_CLASS, 'field');
		fieldSpan.tabIndex = 0;
		return fieldSpan;
	}

	function processIncrementField(value: string, ref: HTMLElement) {
		const incrementIdx = value.indexOf('{i}');
		if (incrementIdx === -1) return;

		let spans: HTMLSpanElement[] = [];
		const textToReplace = ref.innerText.slice(incrementIdx, incrementIdx + 3);

		const leftText = ref.innerText.slice(0, incrementIdx);
		if (leftText.length) spans.push(createTextEl(leftText));

		const fieldSpan = createFieldEl();
		spans.push(fieldSpan);

		const rightText = ref.innerText.slice(incrementIdx + 3);
		if (rightText.length) spans.push(createTextEl(rightText));

		const parent = ref.parentNode!;
		parent.childNodes.forEach((child) => {
			if (child !== ref) return;

			spans.forEach((span) => (child as Element).insertAdjacentElement('beforebegin', span));
			child.remove();
		});

		const newField = {
			id: nanoid(),
			name: `Incremented field`,
			data: { type: 'increment', startFrom: 0 }
		} satisfies FieldSchema;
		fieldRefs[newField.id] = fieldSpan;
		$formData.fields = [...$formData.fields, newField];
		fieldSpan.focus();
	}

	function editField(e: FormTextareaEvent<InputEvent>) {
		if (!editingRef) return;
		$formData.editingField.innerText = e.currentTarget.value;
		processIncrementField(e.currentTarget.value, editingRef);
	}
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
				{@html $htmlTemplate.data.styles}
			{/if}

			<div class="flex h-min flex-[0.5] flex-col items-start gap-4 rounded-sm border p-8">
				<Button
					size="icon"
					variant="ghost"
					class="w-auto gap-2 pl-1 pr-3"
					on:click={() => {
						sheetOpen = false;
					}}
				>
					<LetsIconsBack />
					Back
				</Button>
				<form method="post" use:enhance class="flex w-full flex-col gap-8">
					<FormField {form} name="editingField.innerText" class="w-full">
						<FormControl let:attrs>
							<Label class="text-lg" for="">Text Value</Label>
							<Textarea {...attrs} value={$formData.editingField.innerText} on:input={editField} />
						</FormControl>
					</FormField>

					<FormButton>Submit</FormButton>

					<FormFieldset {form} name="fields" class="w-full">
						<FormLegend>Fields</FormLegend>
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

					<SuperDebug data={$formData} />
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

	:global(.field) {
		height: 24px;
		width: 24px;
		background-color: lightblue;
		border: 1px solid red;
		display: inline-block;
	}

	:global(.super-debug--code) {
		white-space: pre-wrap;
	}
</style>
