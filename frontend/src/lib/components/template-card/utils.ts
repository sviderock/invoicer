import DOMPurify from 'dompurify';
import * as v from 'valibot';

export type IncrementField = v.Output<typeof IncrementField>;
export const IncrementField = v.object({
	type: v.literal('increment'),
	startFrom: v.number([v.minValue(0)])
});

export type FieldSchema = v.Output<typeof FieldSchema>;
export const FieldSchema = v.object({
	id: v.string(),
	name: v.string([v.minLength(1)]),
	data: v.variant('type', [
		IncrementField,
		v.object({
			type: v.literal('')
		})
	])
});

export const EDITABLE_CLASS = 'editable';
export const EDIT_ACTIVE_CLASS = 'edit-active';

export function initiallySanitizeHtml(text: string) {
	let styles = '';
	let fields: FieldSchema[] = [];
	DOMPurify.addHook('uponSanitizeElement', (node, data) => {
		if (data.tagName === 'style') {
			if (node.textContent) {
				const content = node.textContent.split('*/');
				styles += content?.length === 2 ? content[1] : node.textContent;
			}
			return node.parentNode?.removeChild(node);
		}

		if (node.id === 'sidebar' || node.id === 'outline') {
			return node.parentNode?.removeChild(node);
		}

		if (node.classList?.contains('loading-indicator')) {
			return node.parentNode?.removeChild(node);
		}

		if (node.classList?.contains('_') && node.textContent === '') {
			return node.parentNode?.removeChild(node);
		}

		if (data.tagName !== '#text' && node.textContent && node.textContent.trimEnd() === '') {
			return node.parentNode?.removeChild(node);
		}

		if (node instanceof HTMLElement && node.dataset.fieldId) {
			console.log(node.dataset);
			fields.push({
				id: node.dataset.fieldId,
				name: node.dataset.fieldName || '',
				data: {
					type: (node.dataset.fieldDataType || '') as any,
					startFrom: node.dataset.fieldDataStartFrom as any
				} satisfies FieldSchema['data']
			});
		}
	});

	const sanitizedHtml = DOMPurify.sanitize(text, {
		WHOLE_DOCUMENT: true,
		FORCE_BODY: true,
		FORBID_TAGS: ['title']
	});

	DOMPurify.removeAllHooks();

	DOMPurify.addHook('uponSanitizeElement', (node) => {
		if (
			node.nodeType === Node.TEXT_NODE &&
			node.parentElement &&
			node.parentElement.childNodes.length > 1 &&
			node.parentElement.id !== 'page-container' &&
			node.parentElement.tagName !== 'BODY'
		) {
			const span = document.createElement('span');
			span.textContent = node.textContent;
			return node.replaceWith(span);
		}

		if (
			node.nodeType === Node.TEXT_NODE &&
			node.parentElement &&
			(node.parentElement.tagName === 'SPAN' ||
				(node.parentElement.tagName === 'DIV' && node.parentElement.childNodes.length === 1))
		) {
			node.parentElement.classList.add(EDITABLE_CLASS);
			node.parentElement.tabIndex = 0;
		}
	});

	// DOMPurify.addHook('afterSanitizeElements', (node) => {
	// 	if (node.nodeType === Node.TEXT_NODE) {
	// 		console.log(node, node.parentElement);
	// 	}
	// });

	const temp = DOMPurify.sanitize(sanitizedHtml);

	DOMPurify.removeAllHooks();

	// For those weird ${''} pieces: https://github.com/sveltejs/svelte/issues/5292#issuecomment-787743573
	const stylesWithTag = `<${''}style>${styles}</${''}style>`;
	return { sanitizedHtml: temp, styles: stylesWithTag, fields };
}
