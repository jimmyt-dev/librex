import type { Book } from './books.svelte';

export type Shelf = {
  id: string;
  title: string;
  icon?: string;
  books: number;
};

function getToken() {
  return localStorage.getItem('bearer_token') || '';
}

class ShelvesState {
  items = $state<Shelf[]>([]);
  unshelvedCount = $state(0);
  private byShelf = $state<Record<string, Book[]>>({});

  fetchAll = async () => {
    const res = await fetch('/api/shelves', {
      headers: { Authorization: `Bearer ${getToken()}` }
    });
    if (res.ok) {
      const dbItems: { id: string; name: string; icon: string | null; books: number }[] =
        await res.json();

      this.items = dbItems.map((s) => ({
        id: s.id,
        title: s.name,
        icon: s.icon ?? undefined,
        books: s.books
      }));
    }

    await this.fetchUnshelvedCount();
  };

  fetchUnshelvedCount = async () => {
    const res = await fetch('/api/shelves/unshelved', {
      headers: { Authorization: `Bearer ${getToken()}` }
    });
    if (res.ok) {
      const books: Book[] = await res.json();
      this.unshelvedCount = books.length;
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
    const res = await fetch(url, {
      headers: { Authorization: `Bearer ${getToken()}` }
    });
    if (res.ok) {
      const books: Book[] = await res.json();
      this.byShelf = { ...this.byShelf, [shelfId]: books };
    }
  }

  invalidate(shelfId: string) {
    const { [shelfId]: _, ...rest } = this.byShelf;
    this.byShelf = rest;
  }

  create = async (name: string, icon?: string) => {
    const res = await fetch('/api/shelves', {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${getToken()}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ name, icon: icon ?? null })
    });
    if (!res.ok) {
      const msg = await res.text();
      throw new Error(msg || 'Failed to create shelf');
    }
    await this.fetchAll();
  };

  delete = async (id: string) => {
    if (id === 'unshelved') return;

    this.items = this.items.filter((s) => s.id !== id);

    const res = await fetch(`/api/shelves/${id}`, {
      method: 'DELETE',
      headers: { Authorization: `Bearer ${getToken()}` }
    });

    if (!res.ok) {
      const msg = await res.text();
      await this.fetchAll();
      throw new Error(msg || 'Failed to delete shelf');
    }
  };

  addBooks = async (shelfId: string, bookIds: string[]) => {
    const res = await fetch(`/api/shelves/${shelfId}/books`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${getToken()}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ bookIds })
    });
    if (!res.ok) throw new Error('Failed to add books to shelf');
    this.invalidate(shelfId);
    this.invalidate('unshelved');
    await this.fetchAll();
  };

  removeBooks = async (shelfId: string, bookIds: string[]) => {
    const res = await fetch(`/api/shelves/${shelfId}/books`, {
      method: 'DELETE',
      headers: {
        Authorization: `Bearer ${getToken()}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ bookIds })
    });
    if (!res.ok) throw new Error('Failed to remove books from shelf');
    this.invalidate(shelfId);
    this.invalidate('unshelved');
    await this.fetchAll();
  };
}

export const shelvesState = new ShelvesState();
