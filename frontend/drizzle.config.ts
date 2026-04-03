import { defineConfig } from 'drizzle-kit';
import * as dotenv from 'dotenv';
import { resolve } from 'path';
const isDev = process.env.NODE_ENV === 'development' || !process.env.NODE_ENV;
dotenv.config({ path: resolve(__dirname, `../.env${isDev ? '.dev' : ''}`) });

if (!process.env.DATABASE_URL) {
  throw new Error('DATABASE_URL is not set. Check if ../.env exists and contains the variable.');
}

export default defineConfig({
  schema: './src/lib/server/db/schema.ts',
  dialect: 'postgresql',
  dbCredentials: {
    url: process.env.DATABASE_URL,
    ssl: false
  },
  verbose: true,
  // strict prompts for confirmation in interactive terminals; disable so
  // automated deployments (Docker entrypoint) can push without a TTY.
  strict: false
});
