import type { ActionReturn } from 'svelte/action';

export function draggable<T extends string>(node: HTMLElement, data: T): ActionReturn<T> {
	let state = data;

	node.draggable = true;
	node.style.cursor = 'grab';

	function onDragStart(e: DragEvent) {
		if (!e.dataTransfer) return;
		e.dataTransfer.setData('text/plain', state);
	}

	node.addEventListener('dragstart', onDragStart);

	return {
		update(data) {
			state = data;
		},
		destroy() {
			node.removeEventListener('dragstart', onDragStart);
		}
	};
}

export type DropzoneAction = typeof dropzone;
export function dropzone<
	El extends HTMLElement,
	T extends { onDrop: (data: string, e: DragEvent) => void }
>(node: El, options: T): ActionReturn<T> {
	const mandatoryState = {
		dropEffect: 'copy' as DataTransfer['dropEffect'],
		dragOverClass: 'bg-amber-500'
	};
	let state = { ...mandatoryState, ...options };

	function onDrageEnter(e: DragEvent) {
		if (!(e.target instanceof HTMLElement)) return;
		e.target.classList.add(state.dragOverClass);
	}

	function onDragLeave(e: DragEvent) {
		if (!(e.target instanceof HTMLElement)) return;
		e.target.classList.remove(state.dragOverClass);
	}

	function onDragOver(e: DragEvent) {
		e.preventDefault();
		if (!e.dataTransfer) return;
		e.dataTransfer.dropEffect = state.dropEffect;
	}

	function onDrop(e: DragEvent) {
		e.preventDefault();
		if (!e.dataTransfer || !(e.target instanceof HTMLElement)) return;
		const data = e.dataTransfer.getData('text/plain');
		e.target.classList.remove(state.dragOverClass);
		state.onDrop(data, e);
	}

	node.addEventListener('dragenter', onDrageEnter);
	node.addEventListener('dragleave', onDragLeave);
	node.addEventListener('dragover', onDragOver);
	node.addEventListener('drop', onDrop);

	return {
		update(options) {
			state = { ...mandatoryState, ...options };
		},
		destroy() {
			node.removeEventListener('dragenter', onDrageEnter);
			node.removeEventListener('dragleave', onDragLeave);
			node.removeEventListener('dragover', onDragOver);
			node.removeEventListener('drop', onDrop);
		}
	};
}
