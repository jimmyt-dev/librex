import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { playwright } from '@vitest/browser-playwright';
import devtoolsJson from 'vite-plugin-devtools-json';
import { defineConfig } from 'vitest/config';
import dotenv from 'dotenv';
import { fileURLToPath } from 'url';
import { resolve, dirname } from 'path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

export default defineConfig(({ mode }) => {
  const envPath = resolve(__dirname, `../.env${mode === 'development' ? '.dev' : ''}`);
  dotenv.config({ path: envPath });

  const apiPort = process.env.API_PORT || '5321';
  const target = process.env.API_URL || `http://localhost:${apiPort}`;

  return {
    plugins: [tailwindcss(), sveltekit(), devtoolsJson()],
    server: {
      proxy: {
        '/api': {
          target,

          bypass: (req) => {
            // Allow SvelteKit to handle better-auth routes instead of the Go proxy
            if (req.url && (req.url === '/api/auth' || req.url.startsWith('/api/auth/'))) {
              return req.url;
            }
          }
        }
      }
    },
    test: {
      expect: { requireAssertions: true },
      projects: [
        {
          extends: './vite.config.ts',
          test: {
            name: 'client',
            browser: {
              enabled: true,
              provider: playwright(),
              instances: [{ browser: 'chromium', headless: true }]
            },
            include: ['src/**/*.svelte.{test,spec}.{js,ts}'],
            exclude: ['src/lib/server/**']
          }
        },

        {
          extends: './vite.config.ts',
          test: {
            name: 'server',
            environment: 'node',
            include: ['src/**/*.{test,spec}.{js,ts}'],
            exclude: ['src/**/*.svelte.{test,spec}.{js,ts}']
          }
        }
      ]
    }
  };
});
