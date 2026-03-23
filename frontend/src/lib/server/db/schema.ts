import { pgTable, text, uuid } from 'drizzle-orm/pg-core';
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
  libraryId: text('library_id').notNull(),
  title: text('title').notNull(),
  author: text('author'),
  filePath: text('file_path').notNull()
});

export const stagedBooks = pgTable('staged_books', {
  id: uuid('id').defaultRandom().primaryKey(),
  title: text('title').notNull(),
  author: text('author'),
  fileName: text('file_name').notNull(),
  ext: text('ext').notNull(),
  originalPath: text('original_path').notNull(),
  userId: text('user_id')
    .notNull()
    .references(() => user.id, { onDelete: 'cascade' })
});

export * from './auth.schema';
