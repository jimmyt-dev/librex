import type { Handle } from '@sveltejs/kit';
import { redirect } from '@sveltejs/kit';
import { building } from '$app/environment';
import { env } from '$env/dynamic/private';
import { auth } from '$lib/server/auth';
import { svelteKitHandler } from 'better-auth/svelte-kit';

const PUBLIC_PATHS = ['/login', '/register', '/api/auth'];
const API_URL = env.API_URL ?? `http://127.0.0.1:${env.API_PORT ?? '6001'}`;

const handleBetterAuth: Handle = async ({ event, resolve }) => {
  const { pathname } = event.url;

  // In production (adapter-node) there is no Vite proxy, so the Node server
  // receives all browser fetch calls. Proxy non-auth /api/* to the Go backend.
  if (pathname.startsWith('/api/') && !pathname.startsWith('/api/auth/')) {
    const url = `${API_URL}${pathname}${event.url.search}`;
    const headers = new Headers(event.request.headers);
    headers.delete('host');
    const res = await fetch(url, {
      method: event.request.method,
      headers,
      body: ['GET', 'HEAD'].includes(event.request.method) ? undefined : event.request.body,
      // @ts-expect-error duplex required for streaming request bodies in Node 18+
      duplex: 'half'
    });
    return new Response(res.body, { status: res.status, statusText: res.statusText, headers: res.headers });
  }

  let session = null;
  try {
    session = await auth.api.getSession({ headers: event.request.headers });
  } catch {
    // treat as unauthenticated
  }

  if (session) {
    event.locals.session = session.session;
    event.locals.user = session.user;
  }

  const isPublic = PUBLIC_PATHS.some((p) => pathname.startsWith(p));

  if (!event.locals.user && !isPublic) {
    return redirect(302, '/login');
  }
  if (event.locals.user && isPublic && !pathname.startsWith('/api/')) {
    return redirect(302, '/');
  }

  return svelteKitHandler({ event, resolve, auth, building });
};

export const handle: Handle = handleBetterAuth;
