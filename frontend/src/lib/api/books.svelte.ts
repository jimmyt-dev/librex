export type Book = {
  id: string;
  libraryId: string;
  title: string;
  cover: string | null;
  author: string | null;
  subject: string | null;
  description: string | null;
  publisher: string | null;
  contributor: string | null;
  date: string | null;
  type: string | null;
  format: string | null;
  identifier: string | null;
  source: string | null;
  language: string | null;
  relation: string | null;
  coverage: string | null;
  filePath: string;
};

function getToken() {
  return localStorage.getItem('bearer_token') || '';
}

class BooksState {
  private byLibrary = $state<Record<string, Book[]>>({});
  all = $state<Book[]>([]);

  async fetchAll(): Promise<void> {
    const res = await fetch('/api/books/all', {
      headers: { Authorization: `Bearer ${getToken()}` }
    });
    if (res.ok) {
      this.all = await res.json();
    }
  }

  get(libraryId: string): Book[] {
    return this.byLibrary[libraryId] ?? [];
  }

  has(libraryId: string): boolean {
    return libraryId in this.byLibrary;
  }

  async fetchForLibrary(libraryId: string): Promise<void> {
    const res = await fetch(`/api/libraries/${libraryId}/books`, {
      headers: { Authorization: `Bearer ${getToken()}` }
    });
    if (res.ok) {
      const books: Book[] = await res.json();
      this.byLibrary = { ...this.byLibrary, [libraryId]: books };
    }
  }

  upsert(book: Book) {
    const list = this.byLibrary[book.libraryId];
    if (!list) return;
    this.byLibrary = {
      ...this.byLibrary,
      [book.libraryId]: list.map((b) => (b.id === book.id ? book : b))
    };
  }

  remove(libraryId: string, bookId: string) {
    const list = this.byLibrary[libraryId];
    if (!list) return;
    this.byLibrary = {
      ...this.byLibrary,
      [libraryId]: list.filter((b) => b.id !== bookId)
    };
  }

  async delete(bookId: string, deleteFile = false): Promise<void> {
    const url = `/api/books/${bookId}${deleteFile ? '?deleteFile=true' : ''}`;
    const res = await fetch(url, {
      method: 'DELETE',
      headers: { Authorization: `Bearer ${getToken()}` }
    });
    if (!res.ok) throw new Error('Failed to delete book');
    this.all = this.all.filter((b) => b.id !== bookId);
    for (const libraryId of Object.keys(this.byLibrary)) {
      this.remove(libraryId, bookId);
    }
  }

  invalidate(libraryId: string) {
    delete this.byLibrary[libraryId];
  }
}

export const booksState = new BooksState();
