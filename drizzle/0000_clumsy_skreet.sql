CREATE TABLE `timer_room_state` (
	`room_id` text PRIMARY KEY NOT NULL,
	`phase` text DEFAULT 'focus' NOT NULL,
	`remaining_ms` integer NOT NULL,
	`is_running` integer DEFAULT false NOT NULL,
	`cycle_number` integer DEFAULT 1 NOT NULL,
	`updated_at` integer DEFAULT (unixepoch()),
	FOREIGN KEY (`room_id`) REFERENCES `timer_rooms`(`id`) ON UPDATE no action ON DELETE cascade
);
--> statement-breakpoint
CREATE TABLE `timer_rooms` (
	`id` text PRIMARY KEY NOT NULL,
	`creator_id` text NOT NULL,
	`slug` text,
	`focus_duration` integer DEFAULT 1500000 NOT NULL,
	`break_duration` integer DEFAULT 300000 NOT NULL,
	`long_break_duration` integer DEFAULT 900000 NOT NULL,
	`created_at` integer DEFAULT (unixepoch()),
	`last_active_at` integer DEFAULT (unixepoch())
);
--> statement-breakpoint
CREATE UNIQUE INDEX `timer_rooms_slug_unique` ON `timer_rooms` (`slug`);--> statement-breakpoint
CREATE TABLE `todos` (
	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	`title` text NOT NULL,
	`created_at` integer DEFAULT (unixepoch())
);
