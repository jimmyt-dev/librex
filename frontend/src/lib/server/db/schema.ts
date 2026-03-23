import { sqliteTable, text } from 'drizzle-orm/sqlite-core';
import { user } from './auth.schema';

export const library = sqliteTable('libraries', {
  id: text('id')
    .primaryKey()
    .$defaultFn(() => crypto.randomUUID()),
  name: text('name').notNull(),
  icon: text('icon'),
  folder: text('folder').unique().notNull(),
  userId: text('user_id')
    .notNull()
    .references(() => user.id, { onDelete: 'cascade' })
});

export const shelf = sqliteTable('shelves', {
  id: text('id')
    .primaryKey()
    .$defaultFn(() => crypto.randomUUID()),
  name: text('name').notNull(),
  icon: text('icon'),
  userId: text('user_id')
    .notNull()
    .references(() => user.id, { onDelete: 'cascade' })
});

export const books = sqliteTable('books', {
  id: text('id')
    .primaryKey()
    .$defaultFn(() => crypto.randomUUID()),
  libraryId: text('library_id').notNull(),
  title: text('title').notNull(),
  author: text('author'),
  filePath: text('file_path').notNull()
});

export * from './auth.schema';
