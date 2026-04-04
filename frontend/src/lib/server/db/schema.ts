import {
  pgTable,
  text,
  uuid,
  customType,
  primaryKey,
  unique,
  integer,
  real,
  timestamp,
  boolean
} from 'drizzle-orm/pg-core';

const bytea = customType<{ data: Buffer; notNull: false; default: false }>({
  dataType() {
    return 'bytea';
  }
});
import { user } from './auth.schema';

export const library = pgTable('libraries', {
  id: uuid('id').defaultRandom().primaryKey(),
  name: text('name').notNull(),
  icon: text('icon'),
  folder: text('folder').unique().notNull(),
  fileNamingPattern: text('file_naming_pattern'),
  userId: text('user_id')
    .notNull()
    .references(() => user.id, { onDelete: 'cascade' })
});

export const shelf = pgTable('shelves', {
  id: uuid('id').defaultRandom().primaryKey(),
  name: text('name').notNull(),
  icon: text('icon'),
  userId: text('user_id')
    .notNull()
    .references(() => user.id, { onDelete: 'cascade' })
});

export const authors = pgTable(
  'authors',
  {
    id: uuid('id').defaultRandom().primaryKey(),
    name: text('name').notNull(),
    userId: text('user_id')
      .notNull()
      .references(() => user.id, { onDelete: 'cascade' })
  },
  (t) => [unique().on(t.name, t.userId)]
);

export const genres = pgTable(
  'genres',
  {
    id: uuid('id').defaultRandom().primaryKey(),
    name: text('name').notNull(),
    userId: text('user_id')
      .notNull()
      .references(() => user.id, { onDelete: 'cascade' })
  },
  (t) => [unique().on(t.name, t.userId)]
);

export const tags = pgTable(
  'tags',
  {
    id: uuid('id').defaultRandom().primaryKey(),
    name: text('name').notNull(),
    userId: text('user_id')
      .notNull()
      .references(() => user.id, { onDelete: 'cascade' })
  },
  (t) => [unique().on(t.name, t.userId)]
);

// --- Books (slim) ---

export const books = pgTable('books', {
  id: uuid('id').defaultRandom().primaryKey(),
  libraryId: uuid('library_id')
    .notNull()
    .references(() => library.id, { onDelete: 'cascade' }),
  userId: text('user_id')
    .notNull()
    .references(() => user.id, { onDelete: 'cascade' }),
  filePath: text('file_path').notNull(),
  addedOn: timestamp('added_on').defaultNow().notNull()
});

// --- Book Metadata (1-to-1 with books) ---

export const bookMetadata = pgTable('book_metadata', {
  bookId: uuid('book_id')
    .primaryKey()
    .references(() => books.id, { onDelete: 'cascade' }),
  title: text('title').notNull(),
  subtitle: text('subtitle'),
  description: text('description'),
  publisher: text('publisher'),
  publishedDate: text('published_date'),
  isbn13: text('isbn_13'),
  isbn10: text('isbn_10'),
  language: text('language'),
  pageCount: integer('page_count'),
  seriesName: text('series_name'),
  seriesNumber: real('series_number'),
  seriesTotal: integer('series_total'),
  rating: integer('rating'),
  coverPath: text('cover_path'),
  coverMime: text('cover_mime')
});

// --- Join tables ---

export const bookAuthors = pgTable(
  'book_authors',
  {
    bookId: uuid('book_id')
      .notNull()
      .references(() => books.id, { onDelete: 'cascade' }),
    authorId: uuid('author_id')
      .notNull()
      .references(() => authors.id, { onDelete: 'cascade' })
  },
  (t) => [primaryKey({ columns: [t.bookId, t.authorId] })]
);

export const bookGenres = pgTable(
  'book_genres',
  {
    bookId: uuid('book_id')
      .notNull()
      .references(() => books.id, { onDelete: 'cascade' }),
    genreId: uuid('genre_id')
      .notNull()
      .references(() => genres.id, { onDelete: 'cascade' })
  },
  (t) => [primaryKey({ columns: [t.bookId, t.genreId] })]
);

export const bookTags = pgTable(
  'book_tags',
  {
    bookId: uuid('book_id')
      .notNull()
      .references(() => books.id, { onDelete: 'cascade' }),
    tagId: uuid('tag_id')
      .notNull()
      .references(() => tags.id, { onDelete: 'cascade' })
  },
  (t) => [primaryKey({ columns: [t.bookId, t.tagId] })]
);

export const bookShelves = pgTable(
  'book_shelves',
  {
    bookId: uuid('book_id')
      .notNull()
      .references(() => books.id, { onDelete: 'cascade' }),
    shelfId: uuid('shelf_id')
      .notNull()
      .references(() => shelf.id, { onDelete: 'cascade' })
  },
  (t) => [primaryKey({ columns: [t.bookId, t.shelfId] })]
);

// --- Reading ---

export const readingProgress = pgTable(
  'reading_progress',
  {
    id: uuid('id').defaultRandom().primaryKey(),
    userId: text('user_id')
      .notNull()
      .references(() => user.id, { onDelete: 'cascade' }),
    bookId: uuid('book_id')
      .notNull()
      .references(() => books.id, { onDelete: 'cascade' }),
    status: text('status').notNull().default('unread'),
    progress: real('progress').default(0),
    lastReadAt: timestamp('last_read_at'),
    dateStarted: timestamp('date_started'),
    dateFinished: timestamp('date_finished'),
    personalRating: real('personal_rating')
  },
  (t) => [unique().on(t.userId, t.bookId)]
);

export const readingSessions = pgTable('reading_sessions', {
  id: uuid('id').defaultRandom().primaryKey(),
  userId: text('user_id')
    .notNull()
    .references(() => user.id, { onDelete: 'cascade' }),
  bookId: uuid('book_id')
    .notNull()
    .references(() => books.id, { onDelete: 'cascade' }),
  startTime: timestamp('start_time').notNull(),
  endTime: timestamp('end_time'),
  durationSeconds: integer('duration_seconds'),
  startProgress: real('start_progress'),
  endProgress: real('end_progress')
});

// --- Staged Books (unchanged) ---

export const stagedBooks = pgTable('staged_books', {
  id: uuid('id').defaultRandom().primaryKey(),
  title: text('title').notNull(),
  subtitle: text('subtitle'),
  author: text('author'),
  subject: text('subject'),
  description: text('description'),
  publisher: text('publisher'),
  contributor: text('contributor'),
  date: text('date'),
  type: text('type'),
  format: text('format'),
  identifier: text('identifier'),
  source: text('source'),
  language: text('language'),
  relation: text('relation'),
  coverage: text('coverage'),
  seriesName: text('series_name'),
  seriesNumber: real('series_number'),
  seriesTotal: integer('series_total'),
  pageCount: integer('page_count'),
  rating: integer('rating'),
  tags: text('tags'),
  coverImage: bytea('cover_image'),
  coverMime: text('cover_mime'),
  fileName: text('file_name').notNull(),
  ext: text('ext').notNull(),
  originalPath: text('original_path').notNull(),
  userId: text('user_id')
    .notNull()
    .references(() => user.id, { onDelete: 'cascade' })
});

// --- User Settings ---

export const userSettings = pgTable('user_settings', {
  id: uuid('id').defaultRandom().primaryKey(),
  userId: text('user_id')
    .notNull()
    .unique()
    .references(() => user.id, { onDelete: 'cascade' }),
  fileNamingPattern: text('file_naming_pattern').notNull().default('{authors}/{title}{ext}'),
  writeMetadataToFile: boolean('write_metadata_to_file').notNull().default(false),
  bookdropPath: text('bookdrop_path'),
  maxUploadSizeMb: integer('max_upload_size_mb').notNull().default(100)
});

export const opdsCredentials = pgTable('opds_credentials', {
  id: uuid('id').defaultRandom().primaryKey(),
  userId: text('user_id')
    .notNull()
    .unique()
    .references(() => user.id, { onDelete: 'cascade' }),
  username: text('username').notNull().unique(),
  passwordHash: text('password_hash').notNull(),
  enabled: boolean('enabled').notNull().default(true)
});

export * from './auth.schema';
