import { createError, defineEventHandler, getQuery, getRouterParam } from "h3";
import { eq, or } from "drizzle-orm";
import { db } from "@/db";
import { timer_rooms } from "@/db/schema";
import { createTimerToken } from "@/lib/timer/jwt";

export default defineEventHandler(async (event) => {
  const roomId = getRouterParam(event, "roomId");
  const query = getQuery(event);
  const userId = query.userId;

  if (!roomId) {
    throw createError({
      statusCode: 400,
      message: "Room ID is required",
    });
  }

  // Find room by ID or slug
  const rooms = await db
    .select()
    .from(timer_rooms)
    .where(or(eq(timer_rooms.id, roomId), eq(timer_rooms.slug, roomId)))
    .limit(1);

  if (rooms.length === 0) {
    throw createError({
      statusCode: 404,
      message: "Room not found",
    });
  }

  const room = rooms[0];
  const isAdmin = userId ? room.creatorId === userId : false;

  // Generate token if userId provided
  let token: string | null = null;
  if (userId) {
    token = await createTimerToken(userId, room.id, isAdmin);
  }

  return {
    id: room.id,
    slug: room.slug,
    isAdmin,
    token,
    settings: {
      focusDuration: room.focusDuration,
      breakDuration: room.breakDuration,
      longBreakDuration: room.longBreakDuration,
    },
  };
});
