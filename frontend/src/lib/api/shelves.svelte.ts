export type Shelf = {
  id: string;
  title: string;
  url: string;
  icon?: string;
  books: number;
};

class ShelvesState {
  items = $state<Shelf[]>([]);

  fetchAll = async () => {
    const token = localStorage.getItem('bearer_token') || '';
    const res = await fetch('/api/shelves', {
      headers: { Authorization: `Bearer ${token}` }
    });
    if (res.ok) {
      const dbItems: { id: string; name: string; icon: string | null }[] = await res.json();
      
      const unshelved: Shelf = {
        id: 'unshelved',
        title: 'Unshelved',
        url: '#',
        icon: 'frame',
        books: 0
      };

      this.items = [
        unshelved,
        ...dbItems.map((s) => ({
          id: s.id,
          title: s.name,
          url: '#',
          icon: s.icon ?? undefined,
          books: 0
        }))
      ];
    }
  };

  create = async (name: string, icon?: string) => {
    const token = localStorage.getItem('bearer_token') || '';
    const res = await fetch('/api/shelves', {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${token}`,
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
    if (id === 'unshelved') return; // Safety check
    const token = localStorage.getItem('bearer_token') || '';
    
    // Keep local UI instantly in sync
    this.items = this.items.filter((s) => s.id !== id);

    const res = await fetch(`/api/shelves/${id}`, {
      method: 'DELETE',
      headers: {
        Authorization: `Bearer ${token}`
      }
    });
    
    if (!res.ok) {
      const msg = await res.text();
      await this.fetchAll(); // rollback on error
      throw new Error(msg || 'Failed to delete shelf');
    }
  };
}

export const shelvesState = new ShelvesState();
