export type Author = {
  id: string;
  name: string;
  bookCount: number;
};

function getToken() {
  return localStorage.getItem('bearer_token') || '';
}

class AuthorsState {
  items = $state<Author[]>([]);

  fetchAll = async () => {
    const res = await fetch('/api/authors', {
      headers: { Authorization: `Bearer ${getToken()}` }
    });
    if (res.ok) {
      this.items = await res.json();
    }
  };

  create = async (name: string): Promise<Author> => {
    const res = await fetch('/api/authors', {
      method: 'POST',
      headers: { Authorization: `Bearer ${getToken()}`, 'Content-Type': 'application/json' },
      body: JSON.stringify({ name })
    });
    if (!res.ok) throw new Error(await res.text());
    const author = await res.json();
    await this.fetchAll();
    return author;
  };

  update = async (id: string, name: string): Promise<void> => {
    const res = await fetch(`/api/authors/${id}`, {
      method: 'PUT',
      headers: { Authorization: `Bearer ${getToken()}`, 'Content-Type': 'application/json' },
      body: JSON.stringify({ name })
    });
    if (!res.ok) throw new Error(await res.text());
    await this.fetchAll();
  };

  delete = async (id: string): Promise<void> => {
    this.items = this.items.filter((a) => a.id !== id);
    const res = await fetch(`/api/authors/${id}`, {
      method: 'DELETE',
      headers: { Authorization: `Bearer ${getToken()}` }
    });
    if (!res.ok) {
      await this.fetchAll();
      throw new Error('Failed to delete author');
    }
  };
}

export const authorsState = new AuthorsState();
