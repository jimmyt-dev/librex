export type Library = {
  id: string;
  title: string;
  url: string;
  icon?: string;
  books: number;
};

class LibrariesState {
  items = $state<Library[]>([]);

  fetchAll = async () => {
    const token = localStorage.getItem('bearer_token') || '';
    const res = await fetch('/api/libraries', {
      headers: token ? { Authorization: `Bearer ${token}` } : {}
    });
    if (res.ok) {
      const dbItems: { id: string; name: string; icon: string | null; bookCount: number }[] =
        await res.json();
      this.items = dbItems.map((l) => ({
        id: l.id,
        title: l.name,
        url: '#',
        icon: l.icon ?? undefined,
        books: l.bookCount
      }));
    }
  };

  create = async (name: string, folder: string, icon?: string) => {
    const token = localStorage.getItem('bearer_token') || '';
    const res = await fetch('/api/libraries', {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ name, folder, icon: icon ?? null })
    });
    if (!res.ok) {
      const msg = await res.text();
      throw new Error(msg || 'Failed to create library');
    }
    await this.fetchAll();
  };

  delete = async (id: string) => {
    const token = localStorage.getItem('bearer_token') || '';

    // Keep local UI instantly in sync
    this.items = this.items.filter((l) => l.id !== id);

    const res = await fetch(`/api/libraries/${id}`, {
      method: 'DELETE',
      headers: {
        Authorization: `Bearer ${token}`
      }
    });

    if (!res.ok) {
      const msg = await res.text();
      await this.fetchAll(); // rollback on error
      throw new Error(msg || 'Failed to delete library');
    }
  };
}

export const librariesState = new LibrariesState();
