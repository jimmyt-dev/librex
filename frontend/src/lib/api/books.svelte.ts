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
  private booksById = $state<Record<string, Book>>({});
  private fetchedLibraries = $state(new Set<string>());

  // Derived views for efficient access
  all = $derived(Object.values(this.booksById));

  hasLibrary(libraryId: string): boolean {
    return this.fetchedLibraries.has(libraryId);
  }

  get(libraryId: string): Book[] {
    return this.all.filter((b) => b.libraryId === libraryId);
  }

  async fetchAll(): Promise<void> {
    try {
      const books: Book[] = await apiFetch('/api/books/all');
      const newMap: Record<string, Book> = {};
      for (const b of books) {
        newMap[b.id] = b;
      }
      this.booksById = newMap;
    } catch (e) {
      console.error('Failed to fetch all books', e);
    }
  }

  async fetchForLibrary(libraryId: string): Promise<void> {
    try {
      const books: Book[] = await apiFetch(`/api/libraries/${libraryId}/books`);
      // Update the map while preserving books from other libraries
      const nextMap = { ...this.booksById };
      for (const b of books) {
        nextMap[b.id] = b;
      }
      this.booksById = nextMap;
      this.fetchedLibraries = new Set([...this.fetchedLibraries, libraryId]);
    } catch (e) {
      console.error(`Failed to fetch books for library ${libraryId}`, e);
    }
  }

  upsert(book: Book) {
    this.booksById[book.id] = book;
  }

  find(bookId: string): Book | undefined {
    return this.booksById[bookId];
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

  async updateProgress(
    bookId: string,
    data: {
      status?: string;
      progress?: number;
      personalRating?: number;
      dateStarted?: string | null;
      dateFinished?: string | null;
    }
  ): Promise<ReadingProgress> {
    const result: ReadingProgress = await apiFetch(`/api/books/${bookId}/progress`, {
      method: 'PUT',
      body: JSON.stringify(data)
    });
    const book = this.find(bookId);
    if (book) {
      this.upsert({ ...book, progress: result });
    }
    return result;
  }

  async fetchOne(bookId: string): Promise<Book> {
    const book: Book = await apiFetch(`/api/books/${bookId}`);
    this.upsert(book);
    return book;
  }

  async delete(bookId: string, deleteFile = false): Promise<void> {
    const url = `/api/books/${bookId}${deleteFile ? '?deleteFile=true' : ''}`;
    await apiFetch(url, { method: 'DELETE' });
    delete this.booksById[bookId];
  }

  invalidate(libraryId: string) {
    const nextMap = { ...this.booksById };
    for (const id of Object.keys(nextMap)) {
      if (nextMap[id].libraryId === libraryId) {
        delete nextMap[id];
      }
    }
    this.booksById = nextMap;
  }
}

export const booksState = new BooksState();
