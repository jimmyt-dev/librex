import { apiFetch } from './client';

export type BookAuthor = {
  id: string;
  name: string;
};

export type BookGenre = {
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
  seriesTotal: number | null;
  rating: number | null;
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
  genres: BookGenre[];
  tags: BookTag[];
  progress?: ReadingProgress;
};

class BooksState {
  private byLibrary = $state<Record<string, Book[]>>({});
  all = $state<Book[]>([]);

  async fetchAll(): Promise<void> {
    try {
      this.all = await apiFetch('/api/books/all');
    } catch (e) {
      console.error('Failed to fetch all books', e);
    }
  }

  get(libraryId: string): Book[] {
    return this.byLibrary[libraryId] ?? [];
  }

  has(libraryId: string): boolean {
    return libraryId in this.byLibrary;
  }

  async fetchForLibrary(libraryId: string): Promise<void> {
    try {
      const books: Book[] = await apiFetch(`/api/libraries/${libraryId}/books`);
      this.byLibrary = { ...this.byLibrary, [libraryId]: books };
    } catch (e) {
      console.error(`Failed to fetch books for library ${libraryId}`, e);
    }
  }

  upsert(book: Book) {
    // Update byLibrary
    const list = this.byLibrary[book.libraryId];
    if (list) {
      const idx = list.findIndex((b) => b.id === book.id);
      if (idx !== -1) {
        this.byLibrary[book.libraryId] = list.map((b) => (b.id === book.id ? book : b));
      } else {
        this.byLibrary[book.libraryId] = [...list, book];
      }
    }
    // Update all
    const idx = this.all.findIndex((b) => b.id === book.id);
    if (idx !== -1) {
      this.all = this.all.map((b) => (b.id === book.id ? book : b));
    } else {
      this.all = [...this.all, book];
    }
  }

  find(bookId: string): Book | undefined {
    const book = this.all.find((b) => b.id === bookId);
    if (book) return book;
    for (const list of Object.values(this.byLibrary)) {
      const b = list.find((bk) => bk.id === bookId);
      if (b) return b;
    }
    return undefined;
  }

  async patchMetadata(bookId: string, metadata: Partial<BookMetadata>): Promise<void> {
    const book = this.find(bookId);
    if (!book) return;

    const originalBook = JSON.parse(JSON.stringify(book));

    // Optimistic update
    const updatedBook = {
      ...book,
      metadata: { ...book.metadata, ...metadata }
    };
    this.upsert(updatedBook);

    try {
      const serverUpdated = await apiFetch(`/api/books/${bookId}`, {
        method: 'PUT',
        body: JSON.stringify({
          metadata: updatedBook.metadata,
          authors: updatedBook.authors.map((a) => a.name),
          genres: updatedBook.genres.map((g) => g.name),
          tags: updatedBook.tags.map((t) => t.name)
        })
      });
      this.upsert(serverUpdated);
    } catch (err) {
      this.upsert(originalBook);
      throw err;
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
    await apiFetch(url, { method: 'DELETE' });
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
