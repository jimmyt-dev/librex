import type { Book } from '$lib/api/books.svelte';

export type SortField =
  | 'seriesName'
  | 'seriesNumber'
  | 'title'
  | 'author'
  | 'rating'
  | 'publishedDate';
export type SortDir = 'asc' | 'desc';
export type SortLevel = { field: SortField; dir: SortDir };

export const SORT_FIELDS: { value: SortField; label: string }[] = [
  { value: 'seriesName', label: 'Series Name' },
  { value: 'seriesNumber', label: 'Series #' },
  { value: 'title', label: 'Title' },
  { value: 'author', label: 'Author' },
  { value: 'rating', label: 'Rating' },
  { value: 'publishedDate', label: 'Published Date' }
];

const DEFAULT_SORT: SortLevel[] = [
  { field: 'seriesName', dir: 'asc' },
  { field: 'seriesNumber', dir: 'asc' },
  { field: 'title', dir: 'asc' }
];

const STORAGE_KEY = 'view_settings';

function getValue(book: Book, field: SortField): string | number | null {
  switch (field) {
    case 'seriesName':
      return book.metadata.seriesName;
    case 'seriesNumber':
      return book.metadata.seriesNumber;
    case 'title':
      return book.metadata.title;
    case 'author':
      return book.authors[0]?.name ?? null;
    case 'rating':
      return book.metadata.rating;
    case 'publishedDate':
      return book.metadata.publishedDate;
  }
}

function compareField(a: Book, b: Book, level: SortLevel): number {
  const av = getValue(a, level.field);
  const bv = getValue(b, level.field);
  // Nulls always last, regardless of direction
  if (av === null && bv === null) return 0;
  if (av === null) return 1;
  if (bv === null) return -1;
  let cmp: number;
  if (typeof av === 'number' && typeof bv === 'number') {
    cmp = av - bv;
  } else {
    cmp = String(av).localeCompare(String(bv));
  }
  return level.dir === 'asc' ? cmp : -cmp;
}

class ViewSettings {
  mode = $state<'grid' | 'table'>('grid');
  sortLevels = $state<SortLevel[]>(DEFAULT_SORT);

  constructor() {
    if (typeof localStorage === 'undefined') return;
    try {
      const raw = localStorage.getItem(STORAGE_KEY);
      if (!raw) return;
      const parsed = JSON.parse(raw);
      if (parsed.mode === 'grid' || parsed.mode === 'table') this.mode = parsed.mode;
      if (Array.isArray(parsed.sortLevels) && parsed.sortLevels.length > 0) {
        this.sortLevels = parsed.sortLevels;
      }
    } catch {
      // ignore malformed storage
    }
  }

  private save() {
    if (typeof localStorage === 'undefined') return;
    localStorage.setItem(
      STORAGE_KEY,
      JSON.stringify({ mode: this.mode, sortLevels: this.sortLevels })
    );
  }

  setMode(m: 'grid' | 'table') {
    this.mode = m;
    this.save();
  }

  setLevels(levels: SortLevel[]) {
    this.sortLevels = levels;
    this.save();
  }

  addLevel() {
    const used = new Set(this.sortLevels.map((l) => l.field));
    const next = SORT_FIELDS.find((f) => !used.has(f.value));
    if (!next) return;
    this.sortLevels = [...this.sortLevels, { field: next.value, dir: 'asc' }];
    this.save();
  }

  removeLevel(i: number) {
    this.sortLevels = this.sortLevels.filter((_, idx) => idx !== i);
    this.save();
  }

  updateLevel(i: number, patch: Partial<SortLevel>) {
    this.sortLevels = this.sortLevels.map((l, idx) => (idx === i ? { ...l, ...patch } : l));
    this.save();
  }

  reorderLevels(from: number, to: number) {
    if (from === to) return;
    const levels = [...this.sortLevels];
    const [item] = levels.splice(from, 1);
    levels.splice(to, 0, item);
    this.sortLevels = levels;
    this.save();
  }

  resetSort() {
    this.sortLevels = DEFAULT_SORT;
    this.save();
  }

  sort(books: Book[]): Book[] {
    if (this.sortLevels.length === 0) return books;
    return [...books].sort((a, b) => {
      for (const level of this.sortLevels) {
        const cmp = compareField(a, b, level);
        if (cmp !== 0) return cmp;
      }
      return 0;
    });
  }
}

export const viewSettings = new ViewSettings();
