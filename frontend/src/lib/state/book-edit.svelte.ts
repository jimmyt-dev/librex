import type { Book } from '$lib/api/books.svelte';

class BookEditState {
  open = $state(false);
  book = $state<Book | null>(null);

  openFor(book: Book) {
    this.book = book;
    this.open = true;
  }

  close() {
    this.open = false;
  }
}

export const bookEditState = new BookEditState();
