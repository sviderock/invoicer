import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig, searchForWorkspaceRoot } from 'vite';
import { viteCommonjs } from '@originjs/vite-plugin-commonjs';
import { dirname } from 'path';
import { fileURLToPath } from 'url';
import Icons from 'unplugin-icons/vite';

const cwd = dirname(fileURLToPath(import.meta.url));

export default defineConfig({
	plugins: [
		sveltekit(),
		viteCommonjs(),
		Icons({
			compiler: 'svelte',
			autoInstall: true
		})
	],
	build: {
		terserOptions: {
			compress: {
				drop_console: false
			}
		}
	},
	server: {
		fs: {
			allow: [searchForWorkspaceRoot(cwd)]
		}
	}
});
