import { apiFetch } from './client';
import type { Book } from './books.svelte';

export type Shelf = {
  id: string;
  title: string;
  icon?: string;
  books: number;
};

class ShelvesState {
  items = $state<Shelf[]>([]);
  unshelvedCount = $state(0);
  private byShelf = $state<Record<string, Book[]>>({});

  fetchAll = async () => {
    try {
      const dbItems: { id: string; name: string; icon: string | null; books: number }[] =
        await apiFetch('/api/shelves');

      this.items = dbItems.map((s) => ({
        id: s.id,
        title: s.name,
        icon: s.icon ?? undefined,
        books: s.books
      }));

      await this.fetchUnshelvedCount();
    } catch (e) {
      console.error('Failed to fetch shelves', e);
    }
  };

  fetchUnshelvedCount = async () => {
    try {
      const books: Book[] = await apiFetch('/api/shelves/unshelved');
      this.unshelvedCount = books.length;
    } catch (e) {
      console.error('Failed to fetch unshelved count', e);
    }
  };

  get(shelfId: string): Book[] {
    return this.byShelf[shelfId] ?? [];
  }

  has(shelfId: string): boolean {
    return shelfId in this.byShelf;
  }

  async fetchBooksForShelf(shelfId: string): Promise<void> {
    const url =
      shelfId === 'unshelved' ? '/api/shelves/unshelved' : `/api/shelves/${shelfId}/books`;
    try {
      const books: Book[] = await apiFetch(url);
      this.byShelf = { ...this.byShelf, [shelfId]: books };
    } catch (e) {
      console.error(`Failed to fetch books for shelf ${shelfId}`, e);
    }
  }

  invalidate(shelfId: string) {
    delete this.byShelf[shelfId];
  }

  removeBook(bookId: string) {
    const updated: Record<string, Book[]> = {};
    for (const [key, books] of Object.entries(this.byShelf)) {
      updated[key] = books.filter((b) => b.id !== bookId);
    }
    this.byShelf = updated;
  }

  create = async (name: string, icon?: string) => {
    await apiFetch('/api/shelves', {
      method: 'POST',
      body: JSON.stringify({ name, icon: icon ?? null })
    });
    await this.fetchAll();
  };

  delete = async (id: string) => {
    if (id === 'unshelved') return;

    this.items = this.items.filter((s) => s.id !== id);
    try {
      await apiFetch(`/api/shelves/${id}`, { method: 'DELETE' });
    } catch (e) {
      await this.fetchAll();
      throw e;
    }
  };

  addBooks = async (shelfId: string, bookIds: string[]) => {
    await apiFetch(`/api/shelves/${shelfId}/books`, {
      method: 'POST',
      body: JSON.stringify({ bookIds })
    });
    this.invalidate(shelfId);
    this.invalidate('unshelved');
    await this.fetchAll();
  };

  removeBooks = async (shelfId: string, bookIds: string[]) => {
    await apiFetch(`/api/shelves/${shelfId}/books`, {
      method: 'DELETE',
      body: JSON.stringify({ bookIds })
    });
    this.invalidate(shelfId);
    this.invalidate('unshelved');
    await this.fetchAll();
  };
}

export const shelvesState = new ShelvesState();
