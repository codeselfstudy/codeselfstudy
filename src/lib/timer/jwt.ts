import { SignJWT, jwtVerify } from "jose";
import { env } from "@/env";

// Get the secret from environment - must match the Go server's secret
const getSecret = () => {
  return new TextEncoder().encode(env.JWT_SECRET);
};

interface TimerTokenPayload {
  sub: string; // User ID
  roomId: string;
  isAdmin: boolean;
}

/**
 * Creates a JWT token for timer room authentication.
 * This token is verified by the Go WebSocket server.
 */
export async function createTimerToken(
  userId: string,
  roomId: string,
  isAdmin: boolean
): Promise<string> {
  const secret = getSecret();

  const token = await new SignJWT({
    roomId,
    isAdmin,
  })
    .setProtectedHeader({ alg: "HS256" })
    .setSubject(userId)
    .setIssuedAt()
    .setExpirationTime("1h") // Token expires in 1 hour
    .sign(secret);

  return token;
}

/**
 * Verifies a timer token and returns the payload.
 */
export async function verifyTimerToken(
  token: string
): Promise<TimerTokenPayload | null> {
  try {
    const secret = getSecret();
    const { payload } = await jwtVerify(token, secret);

    return {
      sub: payload.sub as string,
      roomId: payload.roomId as string,
      isAdmin: payload.isAdmin as boolean,
    };
  } catch {
    return null;
  }
}
