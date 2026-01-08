import { createError, defineEventHandler, readBody } from "h3";
import { eq } from "drizzle-orm";
import { db } from "@/db";
import { timer_rooms } from "@/db/schema";
import { generateRoomCode, isValidSlug } from "@/lib/timer/names";
import { createTimerToken } from "@/lib/timer/jwt";

interface CreateRoomBody {
  userId: string;
  slug?: string;
  focusDuration?: number;
  breakDuration?: number;
  longBreakDuration?: number;
}

export default defineEventHandler(async (event) => {
  const body = await readBody<CreateRoomBody>(event);

  if (!body.userId) {
    throw createError({
      statusCode: 400,
      message: "userId is required",
    });
  }

  // Validate slug if provided
  if (body.slug) {
    if (!isValidSlug(body.slug)) {
      throw createError({
        statusCode: 400,
        message:
          "Invalid slug. Must be 3-50 characters, alphanumeric and hyphens only.",
      });
    }

    // Check if slug already exists
    const existing = await db
      .select()
      .from(timer_rooms)
      .where(eq(timer_rooms.slug, body.slug))
      .limit(1);

    if (existing.length > 0) {
      throw createError({
        statusCode: 409,
        message: "This slug is already taken",
      });
    }
  }

  // Generate room code
  let roomId = generateRoomCode();

  // Make sure it's unique
  let attempts = 0;
  while (attempts < 10) {
    const existing = await db
      .select()
      .from(timer_rooms)
      .where(eq(timer_rooms.id, roomId))
      .limit(1);

    if (existing.length === 0) break;
    roomId = generateRoomCode();
    attempts++;
  }

  if (attempts >= 10) {
    throw createError({
      statusCode: 500,
      message: "Failed to generate unique room code",
    });
  }

  // Create room
  const room = {
    id: roomId,
    creatorId: body.userId,
    slug: body.slug || null,
    focusDuration: body.focusDuration ?? 25 * 60 * 1000,
    breakDuration: body.breakDuration ?? 5 * 60 * 1000,
    longBreakDuration: body.longBreakDuration ?? 15 * 60 * 1000,
  };

  await db.insert(timer_rooms).values(room);

  // Generate admin token
  const token = await createTimerToken(body.userId, roomId, true);

  return {
    roomId,
    slug: room.slug,
    token,
  };
});
