import type { Book } from '$lib/api/books.svelte';
import { SvelteMap } from 'svelte/reactivity';

export type ItemState = 'include' | 'exclude';
export type FilterMode = 'or' | 'and';

function cycleItem(selections: Map<string, ItemState>, name: string): Map<string, ItemState> {
  const m = new SvelteMap(selections);
  const cur = m.get(name);
  if (!cur) m.set(name, 'include');
  else if (cur === 'include') m.set(name, 'exclude');
  else m.delete(name);
  return m;
}

function applyCategory(
  bookValues: string[],
  selections: Map<string, ItemState>,
  mode: FilterMode
): boolean {
  if (selections.size === 0) return true;

  const included: string[] = [];
  const excluded: string[] = [];
  for (const [k, v] of selections) {
    if (v === 'include') included.push(k);
    else excluded.push(k);
  }

  // Exclusions always apply regardless of mode
  if (excluded.some((v) => bookValues.includes(v))) return false;

  if (included.length === 0) return true;
  if (mode === 'or') return included.some((v) => bookValues.includes(v));
  return included.every((v) => bookValues.includes(v));
}

class FilterState {
  open = $state(false);

  authorSelections = $state<Map<string, ItemState>>(new Map());
  authorMode = $state<FilterMode>('or');

  genreSelections = $state<Map<string, ItemState>>(new Map());
  genreMode = $state<FilterMode>('or');

  tagSelections = $state<Map<string, ItemState>>(new Map());
  tagMode = $state<FilterMode>('or');

  // Status and language are single-value per book, no AND/OR needed
  statusSelections = $state<Map<string, ItemState>>(new Map());
  languageSelections = $state<Map<string, ItemState>>(new Map());

  minRating = $state<number | null>(null);

  get activeCount(): number {
    return (
      (this.authorSelections.size > 0 ? 1 : 0) +
      (this.genreSelections.size > 0 ? 1 : 0) +
      (this.tagSelections.size > 0 ? 1 : 0) +
      (this.statusSelections.size > 0 ? 1 : 0) +
      (this.languageSelections.size > 0 ? 1 : 0) +
      (this.minRating !== null ? 1 : 0)
    );
  }

  apply(books: Book[]): Book[] {
    if (this.activeCount === 0) return books;
    return books.filter((b) => {
      if (
        !applyCategory(
          b.authors.map((a) => a.name),
          this.authorSelections,
          this.authorMode
        )
      )
        return false;
      if (
        !applyCategory(
          b.genres.map((g) => g.name),
          this.genreSelections,
          this.genreMode
        )
      )
        return false;
      if (
        !applyCategory(
          b.tags.map((t) => t.name),
          this.tagSelections,
          this.tagMode
        )
      )
        return false;
      if (!applyCategory([b.progress?.status ?? 'unread'], this.statusSelections, 'or'))
        return false;
      if (
        !applyCategory(
          b.metadata.language ? [b.metadata.language] : [],
          this.languageSelections,
          'or'
        )
      )
        return false;
      if (this.minRating !== null && (b.metadata.rating ?? 0) < this.minRating) return false;
      return true;
    });
  }

  toggle(v?: boolean) {
    this.open = v ?? !this.open;
  }

  clear() {
    this.authorSelections = new Map();
    this.genreSelections = new Map();
    this.tagSelections = new Map();
    this.statusSelections = new Map();
    this.languageSelections = new Map();
    this.minRating = null;
  }

  toggleAuthor(name: string) {
    this.authorSelections = cycleItem(this.authorSelections, name);
  }
  toggleGenre(name: string) {
    this.genreSelections = cycleItem(this.genreSelections, name);
  }
  toggleTag(name: string) {
    this.tagSelections = cycleItem(this.tagSelections, name);
  }
  toggleStatus(status: string) {
    this.statusSelections = cycleItem(this.statusSelections, status);
  }
  toggleLanguage(lang: string) {
    this.languageSelections = cycleItem(this.languageSelections, lang);
  }

  setRating(r: number) {
    this.minRating = this.minRating === r ? null : r;
  }
}

export const filterState = new FilterState();
