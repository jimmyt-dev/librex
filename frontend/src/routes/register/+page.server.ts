import { fail, redirect } from '@sveltejs/kit';
import type { Actions } from './$types';
import { auth } from '$lib/server/auth';
import { APIError } from 'better-auth/api';

export const actions: Actions = {
  default: async (event) => {
    const formData = await event.request.formData();
    const name = formData.get('name')?.toString() ?? '';
    const email = formData.get('email')?.toString() ?? '';
    const password = formData.get('password')?.toString() ?? '';

    try {
      await auth.api.signUpEmail({ body: { name, email, password } });
    } catch (error) {
      if (error instanceof APIError) {
        return fail(400, { message: error.message || 'Registration failed' });
      }
      return fail(500, { message: 'An unexpected error occurred' });
    }

    return redirect(302, '/');
  }
};
