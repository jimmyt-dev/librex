import { apiFetch } from './client';
import { booksState, type Book } from './books.svelte';

export type Author = {
  id: string;
  name: string;
  bookCount: number;
};

class AuthorsState {
  items = $state<Author[]>([]);

  fetchOne = async (id: string): Promise<Author> => {
    return await apiFetch(`/api/authors/${id}`);
  };

  fetchBooksForAuthor = async (id: string): Promise<Book[]> => {
    const books: Book[] = await apiFetch(`/api/authors/${id}/books`);
    for (const b of books) {
      booksState.upsert(b);
    }
    return books;
  };

  fetchAll = async () => {
    try {
      this.items = await apiFetch('/api/authors');
    } catch (e) {
      console.error('Failed to fetch authors', e);
    }
  };

  create = async (name: string): Promise<Author> => {
    const author = await apiFetch('/api/authors', {
      method: 'POST',
      body: JSON.stringify({ name })
    });
    await this.fetchAll();
    return author;
  };

  update = async (id: string, name: string): Promise<void> => {
    await apiFetch(`/api/authors/${id}`, {
      method: 'PUT',
      body: JSON.stringify({ name })
    });
    await this.fetchAll();
  };

  delete = async (id: string): Promise<void> => {
    this.items = this.items.filter((a) => a.id !== id);
    try {
      await apiFetch(`/api/authors/${id}`, { method: 'DELETE' });
    } catch (e) {
      await this.fetchAll();
      throw e;
    }
  };
}

export const authorsState = new AuthorsState();
