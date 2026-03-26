import { apiFetch } from './client';
import { booksState, type Book } from './books.svelte';

export type Shelf = {
  id: string;
  title: string;
  icon?: string;
  books: number;
};

class ShelvesState {
  items = $state<Shelf[]>([]);
  private booksByShelf = $state<Record<string, string[]>>({}); // shelfId -> bookIds[]

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
    } catch (e) {
      console.error('Failed to fetch shelves', e);
    }
  };

  unshelvedCount = $derived(
    booksState.all.filter((b) => {
      // Find if this book is in ANY of the shelf mappings
      const isShelved = Object.values(this.booksByShelf).some((ids) => ids.includes(b.id));
      return !isShelved;
    }).length
  );

  get(shelfId: string): Book[] {
    const ids = this.booksByShelf[shelfId] ?? [];
    return ids.map((id) => booksState.find(id)).filter((b): b is Book => !!b);
  }

  async fetchBooksForShelf(shelfId: string): Promise<void> {
    const url =
      shelfId === 'unshelved' ? '/api/shelves/unshelved' : `/api/shelves/${shelfId}/books`;
    try {
      const books: Book[] = await apiFetch(url);
      // Update global books state with full book objects
      for (const b of books) {
        booksState.upsert(b);
      }
      // Store only the IDs in our local mapping
      this.booksByShelf = { ...this.booksByShelf, [shelfId]: books.map((b) => b.id) };
    } catch (e) {
      console.error(`Failed to fetch books for shelf ${shelfId}`, e);
    }
  }

  invalidate(shelfId: string) {
    const next = { ...this.booksByShelf };
    delete next[shelfId];
    this.booksByShelf = next;
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
    await this.fetchAll();
  };

  removeBooks = async (shelfId: string, bookIds: string[]) => {
    await apiFetch(`/api/shelves/${shelfId}/books`, {
      method: 'DELETE',
      body: JSON.stringify({ bookIds })
    });
    this.invalidate(shelfId);
    await this.fetchAll();
  };
}

export const shelvesState = new ShelvesState();
