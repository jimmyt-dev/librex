export function getToken() {
  if (typeof localStorage === 'undefined') return '';
  return localStorage.getItem('bearer_token') || '';
}

export async function apiFetch(url: string, options: RequestInit = {}) {
  const token = getToken();
  const headers = new Headers(options.headers || {});

  if (token && !headers.has('Authorization')) {
    headers.set('Authorization', `Bearer ${token}`);
  }

  if (options.body && !headers.has('Content-Type') && !(options.body instanceof FormData)) {
    headers.set('Content-Type', 'application/json');
  }

  const res = await fetch(url, { ...options, headers });
  
  if (!res.ok) {
    const errorText = await res.text();
    throw new Error(errorText || `API request failed with status ${res.status}`);
  }

  // Handle 204 No Content
  if (res.status === 204) return null;

  return res.json();
}
