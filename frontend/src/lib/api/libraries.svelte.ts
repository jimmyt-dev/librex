import { apiFetch } from './client';

export type Library = {
  id: string;
  title: string;
  url: string;
  icon?: string;
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
        bookCount: number;
        fileNamingPattern: string | null;
      }[] = await apiFetch('/api/libraries');
      
      this.items = dbItems.map((l) => ({
        id: l.id,
        title: l.name,
        url: '#',
        icon: l.icon ?? undefined,
        books: l.bookCount,
        fileNamingPattern: l.fileNamingPattern
      }));
    } catch (e) {
      console.error('Failed to fetch libraries', e);
    }
  };

  create = async (name: string, folder: string, icon?: string, fileNamingPattern?: string) => {
    await apiFetch('/api/libraries', {
      method: 'POST',
      body: JSON.stringify({
        name,
        folder,
        icon: icon ?? null,
        fileNamingPattern: fileNamingPattern || null
      })
    });
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
