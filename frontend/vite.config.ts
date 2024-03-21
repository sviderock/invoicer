import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig, searchForWorkspaceRoot } from 'vite';
import { viteCommonjs } from '@originjs/vite-plugin-commonjs';
import { dirname } from 'path';
import { fileURLToPath } from 'url';

const cwd = dirname(fileURLToPath(import.meta.url));

export default defineConfig({
	plugins: [sveltekit(), viteCommonjs()],
	build: {
		terserOptions: {
			compress: {
				drop_console: false
			}
		}
	},
	resolve: {
		preserveSymlinks: true
	},
	server: {
		fs: {
			allow: [searchForWorkspaceRoot(cwd)]
		}
	}
});
