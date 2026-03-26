import type { Book } from '$lib/api/books.svelte';

class BookEditState {
  open = $state(false);
  book = $state<Book | null>(null);
  queue = $state<Book[]>([]);
  queueIndex = $state(0);

  get inQueue(): boolean {
    return this.queue.length > 0;
  }

  get isFirst(): boolean {
    return this.queueIndex === 0;
  }

  get isLast(): boolean {
    return this.queueIndex === this.queue.length - 1;
  }

  openFor(book: Book) {
    this.queue = [];
    this.queueIndex = 0;
    this.book = book;
    this.open = true;
  }

  openQueue(books: Book[]) {
    if (books.length === 0) return;
    this.queue = books;
    this.queueIndex = 0;
    this.book = books[0];
    this.open = true;
  }

  goTo(index: number) {
    if (index < 0 || index >= this.queue.length) return;
    this.queueIndex = index;
    this.book = this.queue[index];
  }

  next() {
    this.goTo(this.queueIndex + 1);
  }

  prev() {
    this.goTo(this.queueIndex - 1);
  }

  close() {
    this.open = false;
    this.queue = [];
    this.queueIndex = 0;
  }
}

export const bookEditState = new BookEditState();
