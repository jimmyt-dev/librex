import type { Handle } from '@sveltejs/kit';
import { redirect } from '@sveltejs/kit';
import { building } from '$app/environment';
import { env } from '$env/dynamic/private';
import { auth } from '$lib/server/auth';
import { svelteKitHandler } from 'better-auth/svelte-kit';

const PUBLIC_PATHS = ['/login', '/register', '/api/auth', '/opds'];
const API_URL = env.API_URL ?? `http://127.0.0.1:${env.API_PORT ?? '6001'}`;

const handleBetterAuth: Handle = async ({ event, resolve }) => {
  const { pathname } = event.url;

  // In production (adapter-node) there is no Vite proxy, so the Node server
  // receives all browser fetch calls. Proxy non-auth /api/* and /opds/* to the Go backend.
  if (
    (pathname.startsWith('/api/') && !pathname.startsWith('/api/auth/')) ||
    pathname.startsWith('/opds')
  ) {
    const url = `${API_URL}${pathname}${event.url.search}`;
    const headers = new Headers(event.request.headers);

    // Set forwarding headers so the backend knows the real origin
    headers.set('X-Forwarded-Host', event.url.host);
    headers.set('X-Forwarded-Proto', event.url.protocol.replace(':', ''));

    headers.delete('host');
    headers.delete('content-length');
    headers.delete('transfer-encoding');
    const hasBody = !['GET', 'HEAD'].includes(event.request.method);
    const fetchOptions: RequestInit & { duplex?: string } = {
      method: event.request.method,
      headers,
      body: hasBody ? event.request.body : undefined
    };
    if (hasBody) {
      fetchOptions.duplex = 'half';
    }
    const res = await fetch(url, fetchOptions);
    return new Response(res.body, {
      status: res.status,
      statusText: res.statusText,
      headers: res.headers
    });
  }

  const session = await auth.api.getSession({ headers: event.request.headers });

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

  // Use a custom resolve to disable preloading of JS/CSS chunks in the Link header.
  // This prevents the header from becoming too large and triggering 502 errors 
  // in Nginx/OpenResty on pages with many components (like the book list).
  return svelteKitHandler({
    event,
    resolve: (e) => resolve(e, { preload: () => false }),
    auth,
    building
  });
};

export const handle: Handle = handleBetterAuth;
