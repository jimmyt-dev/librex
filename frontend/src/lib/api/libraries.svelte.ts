import { apiFetch } from './client';
import { booksState } from './books.svelte';

export type Library = {
  id: string;
  title: string;
  url: string;
  icon?: string;
  folder?: string;
  books: number;
  fileNamingPattern?: string | null;
};

class LibrariesState {
  items = $state<Library[]>([]);

  fetchAll = async () => {
    try {
      const dbItems: {
        id: string;
        name: string;
        icon: string | null;
        folder: string | null;
        bookCount: number;
        fileNamingPattern: string | null;
      }[] = await apiFetch('/api/libraries');

      this.items = dbItems.map((l) => ({
        id: l.id,
        title: l.name,
        url: '#',
        icon: l.icon ?? undefined,
        folder: l.folder ?? undefined,
        books: l.bookCount,
        fileNamingPattern: l.fileNamingPattern
      }));
    } catch (e) {
      console.error('Failed to fetch libraries', e);
    }
  };

  create = async (
    name: string,
    folder: string,
    icon?: string,
    fileNamingPattern?: string
  ): Promise<string> => {
    const lib: { id: string } = await apiFetch('/api/libraries', {
      method: 'POST',
      body: JSON.stringify({
        name,
        folder,
        icon: icon ?? null,
        fileNamingPattern: fileNamingPattern || null
      })
    });
    await this.fetchAll();
    return lib.id;
  };

  scan = async (id: string) => {
    await apiFetch(`/api/libraries/${id}/scan`, { method: 'POST' });
    booksState.invalidate(id);
    await this.fetchAll();
  };

  delete = async (id: string) => {
    this.items = this.items.filter((l) => l.id !== id);
    try {
      await apiFetch(`/api/libraries/${id}`, { method: 'DELETE' });
    } catch (e) {
      await this.fetchAll();
      throw e;
    }
  };
}

export const librariesState = new LibrariesState();
