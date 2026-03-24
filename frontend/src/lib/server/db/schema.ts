import { pgTable, text, uuid, customType } from 'drizzle-orm/pg-core';

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

export const books = pgTable('books', {
  id: uuid('id').defaultRandom().primaryKey(),
  libraryId: uuid('library_id')
    .notNull()
    .references(() => library.id, { onDelete: 'cascade' }),
  userId: text('user_id')
    .notNull()
    .references(() => user.id, { onDelete: 'cascade' }),
  title: text('title').notNull(),
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
  coverImage: bytea('cover_image'),
  coverMime: text('cover_mime'),
  filePath: text('file_path').notNull()
});

export const stagedBooks = pgTable('staged_books', {
  id: uuid('id').defaultRandom().primaryKey(),
  title: text('title').notNull(),
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
  coverImage: bytea('cover_image'),
  coverMime: text('cover_mime'),
  fileName: text('file_name').notNull(),
  ext: text('ext').notNull(),
  originalPath: text('original_path').notNull(),
  userId: text('user_id')
    .notNull()
    .references(() => user.id, { onDelete: 'cascade' })
});

export * from './auth.schema';
