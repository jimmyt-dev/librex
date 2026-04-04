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
    // Force connection close so undici does not reuse stale pooled connections.
    headers.set('connection', 'close');
    const hasBody = !['GET', 'HEAD'].includes(event.request.method);
    // For HEAD requests, proxy as GET to avoid undici HEAD response body handling issues,
    // then return null body with the response headers.
    const proxyMethod = event.request.method === 'HEAD' ? 'GET' : event.request.method;
    const fetchInit: RequestInit = { method: proxyMethod, headers };
    if (hasBody) {
      fetchInit.body = event.request.body;
      // @ts-expect-error duplex required for streaming request bodies in Node 18+
      fetchInit.duplex = 'half';
    }
    try {
      const res = await fetch(url, fetchInit);
      if (event.request.method === 'HEAD') {
        // Drain the body so undici doesn't leave the socket in a dirty state.
        await res.body?.cancel();
        return new Response(null, { status: res.status, statusText: res.statusText, headers: res.headers });
      }
      return new Response(res.body, { status: res.status, statusText: res.statusText, headers: res.headers });
    } catch (err: unknown) {
      const cause = (err as { cause?: { code?: string } })?.cause;
      console.error(`[proxy] ${event.request.method} ${url} failed:`, cause?.code ?? String(err));
      return new Response('Bad Gateway', { status: 502 });
    }
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

  return svelteKitHandler({ event, resolve, auth, building });
};

export const handle: Handle = handleBetterAuth;
