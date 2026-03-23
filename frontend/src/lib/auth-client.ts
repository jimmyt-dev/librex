import { createAuthClient } from 'better-auth/svelte';

export const authClient = createAuthClient({
  fetchOptions: {
    auth: {
      type: 'Bearer',
      token: () => localStorage.getItem('bearer_token') || '' // get the token from localStorage
    },
    onSuccess: (ctx) => {
      const authToken = ctx.response.headers.get('set-auth-token'); // get the token from the response headers
      // Store the token securely (e.g., in localStorage)
      if (authToken) {
        localStorage.setItem('bearer_token', authToken);
      }
    }
  }
});
