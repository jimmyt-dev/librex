import { apiFetch } from './client';

class BookdropState {
  stagedCount = $state(0);

  async fetchCount() {
    try {
      const books: unknown[] = await apiFetch('/api/bookdrop/staged');
      this.stagedCount = books.length;
    } catch { /* silent */ }
  }
}

export const bookdropState = new BookdropState();
