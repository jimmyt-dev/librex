import { sqliteTable, text } from 'drizzle-orm/sqlite-core';
import { user } from './auth.schema';

export const library = sqliteTable('libraries', {
  id: text('id')
    .primaryKey()
    .$defaultFn(() => crypto.randomUUID()),
  name: text('name').notNull(),
  icon: text('icon'),
  folder: text('folder').unique(),
  userId: text('user_id')
    .notNull()
    .references(() => user.id, { onDelete: 'cascade' })
});

export * from './auth.schema';
