import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async (event) => {
  const sidebarOpen = event.cookies.get('sidebar:state') !== 'false';
  return { user: event.locals.user ?? null, sidebarOpen };
};
