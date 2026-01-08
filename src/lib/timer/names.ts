// Word lists for generating anonymous user names
const adjectives = [
  "happy",
  "swift",
  "bright",
  "calm",
  "bold",
  "cool",
  "warm",
  "soft",
  "wild",
  "quiet",
  "stormy",
  "sunny",
  "misty",
  "dusty",
  "frosty",
  "gentle",
  "brave",
  "clever",
  "eager",
  "fancy",
  "jolly",
  "lucky",
  "merry",
  "noble",
  "proud",
];

const nouns = [
  "tiger",
  "eagle",
  "river",
  "mountain",
  "forest",
  "ocean",
  "meadow",
  "canyon",
  "valley",
  "sunset",
  "salad",
  "robot",
  "cat",
  "panda",
  "phoenix",
  "dragon",
  "falcon",
  "dolphin",
  "koala",
  "otter",
  "badger",
  "raven",
  "wolf",
  "fox",
  "owl",
];

/**
 * Generates a deterministic guest name from a session ID.
 * Format: adjective_noun_number (e.g., "stormy_salad_4329")
 */
export function generateGuestName(sessionId: string): string {
  const hash = hashString(sessionId);
  const adj = adjectives[Math.abs(hash) % adjectives.length];
  const noun = nouns[Math.abs(hashString(sessionId + "noun")) % nouns.length];
  const num = Math.abs(hashString(sessionId + "num")) % 10000;

  return `${adj}_${noun}_${num.toString().padStart(4, "0")}`;
}

/**
 * Generates a short random room code.
 * Format: 6 alphanumeric characters (e.g., "abc123")
 */
export function generateRoomCode(): string {
  const chars = "abcdefghijklmnopqrstuvwxyz0123456789";
  let code = "";
  for (let i = 0; i < 6; i++) {
    code += chars[Math.floor(Math.random() * chars.length)];
  }
  return code;
}

/**
 * Validates a room slug.
 * Must be 3-50 characters, alphanumeric with hyphens only.
 */
export function isValidSlug(slug: string): boolean {
  if (slug.length < 3 || slug.length > 50) {
    return false;
  }
  return /^[a-zA-Z0-9-]+$/.test(slug);
}

// Simple string hash function
function hashString(s: string): number {
  let h = 0;
  for (let i = 0; i < s.length; i++) {
    h = (31 * h + s.charCodeAt(i)) | 0;
  }
  return h;
}
