<script lang="ts">
	import { api } from '$lib/api';
	import FileFormEdit from '$lib/components/template-card/template-form-edit.svelte';
	import { Button } from '$lib/components/ui/button';
	import { FormControl, FormField } from '$lib/components/ui/form';
	import FormButton from '$lib/components/ui/form/form-button.svelte';
	import { Input } from '$lib/components/ui/input';
	import {
		Popover,
		PopoverArrow,
		PopoverContent,
		PopoverTrigger
	} from '$lib/components/ui/popover';
	import { bytesToSize, cn } from '$lib/utils';
	import { type PlainMessage } from '@bufbuild/protobuf';
	import { createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { nanoid } from 'nanoid';
	import type { Template } from 'proto/template_pb';
	import { toast } from 'svelte-sonner';
	import { defaults, superForm } from 'sveltekit-superforms';
	import { valibot } from 'sveltekit-superforms/adapters';
	import BytesizeEdit from '~icons/bytesize/edit';
	import LineMdUploadingLoop from '~icons/line-md/uploading-loop';
	import MingcuteDelete2Line from '~icons/mingcute/delete-2-line';
	import PhCheckBold from '~icons/ph/check-bold';
	import PhSealCheckBold from '~icons/ph/seal-check-bold';
	import { UpdateTemplateSchema } from '../../schemas';

	type Props = {
		template: PlainMessage<Template>;
		pending?: 'uploading' | 'uploaded';
	};

	let { template, pending }: Props = $props();
	let editingName = $state(false);
	const queryClient = useQueryClient();

	const updateTemplate = createMutation({
		mutationKey: ['update-file'],
		mutationFn: api().updateTemplate,
		onSuccess: (data) => {
			editingName = false;
			queryClient.invalidateQueries({ queryKey: ['get-templates'] });
			if (data) {
				form.reset({ data: { name: data.name } });
			}
		}
	});

	const deleteTemplate = createMutation({
		mutationKey: ['delete-file'],
		mutationFn: api().deleteTemplate,
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['get-templates'] });
			toast.info('File was deleted');
		}
	});

	const form = superForm(
		defaults(valibot(UpdateTemplateSchema), {
			defaults: { name: template.name.replace(template.ext, '') }
		}),
		{
			id: nanoid(),
			SPA: true,
			dataType: 'json',
			validators: valibot(UpdateTemplateSchema),
			onSubmit: ({ formData }) => {
				$updateTemplate.mutate({ id: template.id, name: formData.get('name')!.toString() });
			}
		}
	);
	const { form: formData, enhance } = form;
</script>

<div class="rounded-md bg-slate-900 p-4 text-sm text-slate-300">
	<div class="flex h-full flex-col justify-between gap-4">
		{#if pending}
			{@render uploadingPreview()}
		{:else}
			<FileFormEdit {template} />
		{/if}

		<div class="flex flex-col items-start gap-2">
			{#if pending}
				<h3 class="wrap break-all border">{template.name || ''}</h3>
			{:else}
				{@render fileName()}
			{/if}

			{@render fileDetails()}
		</div>
	</div>
</div>

{#snippet fileName()}
	<form method="POST" use:enhance class="flex w-full items-start justify-between gap-2">
		<FormButton
			variant="ghost"
			size="icon"
			type={editingName ? 'submit' : 'button'}
			on:click={(e) => {
				if (editingName) return;
				editingName = true;
				e.preventDefault();
			}}
		>
			{#if editingName}
				<PhCheckBold class="text-xs" />
			{:else}
				<BytesizeEdit class="text-xs" />
			{/if}
		</FormButton>
		<FormField {form} name="name" class="w-full">
			<FormControl let:attrs>
				{#if editingName}
					<div class="flex items-center gap-1">
						<Input
							{...attrs}
							bind:value={$formData.name}
							class="box-border h-auto py-0 leading-7"
						/>
						<span>{template.ext}</span>
					</div>
				{:else}
					<h3 class="wrap break-all border border-transparent leading-7">{template.name}</h3>
				{/if}
			</FormControl>
		</FormField>
	</form>
{/snippet}

{#snippet fileDetails()}
	<div class="flex w-full items-center justify-between gap-2">
		<div class="flex items-center gap-2">
			<span class="rounded-sm border bg-slate-200 px-2 text-slate-600">
				{bytesToSize(template.size)}
			</span>
		</div>
		<div class="flex items-center gap-2">
			<Popover>
				<PopoverTrigger asChild let:builder>
					<Button variant="destructive" size="icon" builders={[builder]}>
						<MingcuteDelete2Line class="text-xs" />
					</Button>
				</PopoverTrigger>
				<PopoverContent
					class="flex flex-col items-center justify-between gap-2"
					side="top"
					align="end"
				>
					<PopoverArrow />
					You sure?
					<Button
						variant="destructive"
						size="sm"
						class="px-2"
						on:click={() => $deleteTemplate.mutate({ id: template.id })}>Yes, delete</Button
					>
				</PopoverContent>
			</Popover>
		</div>
	</div>
{/snippet}

{#snippet uploadingPreview()}
	<div
		class={cn(
			'aspect-a4 bg-file-upload flex w-full flex-col items-center justify-center gap-4 rounded-md border-2 border-slate-300 bg-[length:400%_400%] text-xl text-blue-800 transition-colors',
			pending === 'uploaded' && 'animate-upload-finished text-emerald-800'
		)}
	>
		<div class="h-7 overflow-hidden">
			<div
				class={cn(
					'relative flex flex-col items-center transition-transform duration-500',
					pending === 'uploaded' && '-translate-y-1/2'
				)}
			>
				<span>Uploading</span>
				<span>Uploaded</span>
			</div>
		</div>

		<div class="h-11 overflow-hidden">
			<div
				class={cn(
					'relative flex -translate-y-1/2 flex-col items-center transition-transform duration-500',
					pending === 'uploaded' && 'translate-y-0'
				)}
			>
				<PhSealCheckBold class="text-4xl" />
				<LineMdUploadingLoop class="text-4xl" />
			</div>
		</div>
	</div>
{/snippet}
