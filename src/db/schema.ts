import { integer, sqliteTable, text } from "drizzle-orm/sqlite-core";
import { sql } from "drizzle-orm";

export const todos = sqliteTable("todos", {
  id: integer({ mode: "number" }).primaryKey({
    autoIncrement: true,
  }),
  title: text().notNull(),
  createdAt: integer("created_at", { mode: "timestamp" }).default(
    sql`(unixepoch())`
  ),
});

// Timer room configuration
export const timer_rooms = sqliteTable("timer_rooms", {
  id: text("id").primaryKey(), // Short code (e.g., "abc123") or custom slug
  creatorId: text("creator_id").notNull(), // WorkOS user ID
  slug: text("slug").unique(), // Optional custom slug
  focusDuration: integer("focus_duration")
    .notNull()
    .default(25 * 60 * 1000), // 25 min in ms
  breakDuration: integer("break_duration")
    .notNull()
    .default(5 * 60 * 1000), // 5 min in ms
  longBreakDuration: integer("long_break_duration")
    .notNull()
    .default(15 * 60 * 1000), // 15 min in ms
  createdAt: integer("created_at", { mode: "timestamp" }).default(
    sql`(unixepoch())`
  ),
  lastActiveAt: integer("last_active_at", { mode: "timestamp" }).default(
    sql`(unixepoch())`
  ),
});

// Timer room state for recovery after server restart
export const timer_room_state = sqliteTable("timer_room_state", {
  roomId: text("room_id")
    .primaryKey()
    .references(() => timer_rooms.id, { onDelete: "cascade" }),
  phase: text("phase").notNull().default("focus"), // focus, break, longBreak, overtime
  remainingMs: integer("remaining_ms").notNull(),
  isRunning: integer("is_running", { mode: "boolean" })
    .notNull()
    .default(false),
  cycleNumber: integer("cycle_number").notNull().default(1), // 1-4 before long break
  updatedAt: integer("updated_at", { mode: "timestamp" }).default(
    sql`(unixepoch())`
  ),
});
