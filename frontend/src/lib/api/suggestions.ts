import { apiFetch } from './client';

export async function fetchAuthorSuggestions(q: string): Promise<string[]> {
  try {
    const data: { name: string }[] = await apiFetch(`/api/authors?q=${encodeURIComponent(q)}`);
    return data.map((a) => a.name);
  } catch {
    return [];
  }
}

export async function fetchGenreSuggestions(q: string): Promise<string[]> {
  try {
    const data: { name: string }[] = await apiFetch(`/api/genres?q=${encodeURIComponent(q)}`);
    return data.map((g) => g.name);
  } catch {
    return [];
  }
}

export async function fetchTagSuggestions(q: string): Promise<string[]> {
  try {
    const data: { name: string }[] = await apiFetch(`/api/tags?q=${encodeURIComponent(q)}`);
    return data.map((t) => t.name);
  } catch {
    return [];
  }
}

export async function fetchSeriesSuggestions(q: string): Promise<string[]> {
  try {
    return await apiFetch(`/api/series?q=${encodeURIComponent(q)}`);
  } catch {
    return [];
  }
}
