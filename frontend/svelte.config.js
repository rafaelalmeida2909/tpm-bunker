import autoprefixer from 'autoprefixer';
import preprocess from 'svelte-preprocess';
import tailwindcss from 'tailwindcss';

/** @type {import('@sveltejs/vite-plugin-svelte').Options} */
const config = {
  preprocess: preprocess({
    postcss: {
      plugins: [
        tailwindcss(),
        autoprefixer()
      ]
    }
  })
};

export default config;