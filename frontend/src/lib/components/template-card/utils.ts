import DOMPurify from 'dompurify';

export const EDITABLE_CLASS = 'editable';
export const EDIT_ACTIVE_CLASS = 'edit-active';

export function initiallySanitizeHtml(text: string) {
	let styles = '';
	let classes = new Set<string>();
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

		if (
			data.tagName === '#text' &&
			node.parentElement &&
			node.parentElement.id !== 'page-container'
		) {
			node.parentElement.classList.add(EDITABLE_CLASS);
			node.parentElement.tabIndex = 0;
		}

		if (node.classList?.length) {
			node.classList.forEach((i) => classes.add(i));
		}
	});

	const sanitizedHtml = DOMPurify.sanitize(text, {
		WHOLE_DOCUMENT: true,
		FORCE_BODY: true,
		FORBID_TAGS: ['title']
	});

	DOMPurify.removeAllHooks();

	return { sanitizedHtml, styles, classes };
}
