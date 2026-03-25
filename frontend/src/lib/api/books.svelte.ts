export type BookAuthor = {
  id: string;
  name: string;
};

export type BookCategory = {
  id: string;
  name: string;
};

export type BookTag = {
  id: string;
  name: string;
};

export type BookMetadata = {
  bookId: string;
  title: string;
  subtitle: string | null;
  description: string | null;
  publisher: string | null;
  publishedDate: string | null;
  isbn13: string | null;
  isbn10: string | null;
  language: string | null;
  pageCount: number | null;
  seriesName: string | null;
  seriesNumber: number | null;
  coverPath: string | null;
  coverMime: string | null;
};

export type ReadingProgress = {
  id: string;
  userId: string;
  bookId: string;
  status: string;
  progress: number;
  lastReadAt: string | null;
  dateStarted: string | null;
  dateFinished: string | null;
  personalRating: number | null;
};

export type Book = {
  id: string;
  libraryId: string;
  userId: string;
  filePath: string;
  addedOn: string;
  metadata: BookMetadata;
  authors: BookAuthor[];
  categories: BookCategory[];
  tags: BookTag[];
  progress?: ReadingProgress;
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
    // Update byLibrary
    const list = this.byLibrary[book.libraryId];
    if (list) {
      this.byLibrary = {
        ...this.byLibrary,
        [book.libraryId]: list.map((b) => (b.id === book.id ? book : b))
      };
    }
    // Update all
    const idx = this.all.findIndex((b) => b.id === book.id);
    if (idx !== -1) {
      this.all = this.all.map((b) => (b.id === book.id ? book : b));
    }
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
