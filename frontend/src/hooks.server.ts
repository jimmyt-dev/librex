import type { Handle } from '@sveltejs/kit';
import { redirect } from '@sveltejs/kit';
import { building } from '$app/environment';
import { auth } from '$lib/server/auth';
import { svelteKitHandler } from 'better-auth/svelte-kit';

const PUBLIC_PATHS = ['/login', '/register', '/api/auth'];

const handleBetterAuth: Handle = async ({ event, resolve }) => {
  const session = await auth.api.getSession({ headers: event.request.headers });

  if (session) {
    event.locals.session = session.session;
    event.locals.user = session.user;
  }

  const { pathname } = new URL(event.request.url);
  const isPublic = PUBLIC_PATHS.some((p) => pathname.startsWith(p));

  if (!event.locals.user && !isPublic) {
    return redirect(302, '/login');
  }
  if (event.locals.user && isPublic) {
    return redirect(302, '/');
  }

  return svelteKitHandler({ event, resolve, auth, building });
};

export const handle: Handle = handleBetterAuth;
