import { svelte } from '@sveltejs/vite-plugin-svelte';
import { defineConfig } from 'vite';

export default defineConfig({
  plugins: [svelte()],
  build: {
    rollupOptions: {
      external: [
        'svelte-icons/fa/FaShield.svelte',
        'svelte-icons/fa/FaShieldAlt.svelte',
        'svelte-icons/fa/FaDownload.svelte',
        'svelte-icons/fa/FaLock.svelte',
        'svelte-icons/fa/FaSync.svelte',
        'svelte-icons/fa/FaUpload.svelte'
      ]
    }
  }
});