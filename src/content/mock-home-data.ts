/**
 * Hardcoded data for the terminal/tmux home redesign (mockup 68).
 * Values copied verbatim from mockups/68-tmux-home-cta-below.html.
 * Real data wiring is follow-up work; for now feature panes import from here.
 */

export type TermColor = "green" | "cyan" | "yellow" | "magenta";

export type Language = {
  name: string;
  percent: number;
  color: TermColor;
};

export const LANGUAGES: Array<Language> = [
  { name: "python", percent: 62, color: "green" },
  { name: "javascript", percent: 54, color: "cyan" },
  { name: "typescript", percent: 48, color: "cyan" },
  { name: "go", percent: 35, color: "yellow" },
  { name: "rust", percent: 31, color: "yellow" },
  { name: "haskell", percent: 22, color: "magenta" },
  { name: "clojure", percent: 19, color: "magenta" },
  { name: "elixir", percent: 17, color: "magenta" },
  { name: "racket", percent: 14, color: "green" },
];

export const LANGUAGES_OVERFLOW_LABEL = "+ 12 more";

export type StatRow = {
  label: string;
  value: string | number;
  color: "green" | "cyan" | "magenta";
  meterPercent: number;
};

export type NetActivity = {
  icon: "↓" | "↑" | "⇄";
  value: number;
  label: string;
  color: TermColor;
};

export const STATS = {
  community: [
    { label: "members", value: "5,000+", color: "green", meterPercent: 92 },
    { label: "years up", value: 11, color: "cyan", meterPercent: 82 },
    { label: "langs", value: "20+", color: "magenta", meterPercent: 70 },
  ] satisfies Array<StatRow>,
  netThisWeek: [
    { icon: "↓", value: 12, label: "meetups / month", color: "green" },
    { icon: "↑", value: 234, label: "forum posts", color: "cyan" },
    { icon: "⇄", value: 38, label: "intros & jobs", color: "yellow" },
  ] satisfies Array<NetActivity>,
  loadAvgAscii: "▁▂▃▅▆▇█▇▆▅▆▇█▇▆▅▃▂▃▅▇█▇▆▅▆▇█▇▆▅",
  loadAvgCaption: "12mo · steady & rising",
} as const;

export type ForumPost = {
  when: string;
  by: string;
  category: string;
  /** CSS color string — uses the --term-* variables from styles.css. */
  categoryColor: string;
  topic: string;
  replies: number;
};

export const FORUM_ACTIVITY: Array<ForumPost> = [
  {
    when: "2m",
    by: "alice",
    category: "#rust",
    categoryColor: "var(--term-yellow)",
    topic:
      "code review on a tiny json parser — got it down to 180 lines, can it be smaller?",
    replies: 8,
  },
  {
    when: "14m",
    by: "bob",
    category: "#events",
    categoryColor: "var(--term-cyan)",
    topic: "in-person meetup · Berkeley library · sat 1pm · 18 RSVPs",
    replies: 3,
  },
  {
    when: "42m",
    by: "carol",
    category: "#haskell",
    categoryColor: "var(--term-magenta)",
    topic:
      "finally got `traverse` to click — short writeup with the aha moment",
    replies: 12,
  },
  {
    when: "1h",
    by: "dan",
    category: "#beginners",
    categoryColor: "var(--term-accent)",
    topic: "study plan for going from JS → systems programming?",
    replies: 21,
  },
  {
    when: "3h",
    by: "eve",
    category: "#jobs",
    /* mockup uses #ffb86b, an orange not in the token set; keep literal. */
    categoryColor: "#ffb86b",
    topic: "backend role at a small SF startup — intros welcome",
    replies: 5,
  },
  {
    when: "5h",
    by: "frank",
    category: "#tools",
    categoryColor: "var(--term-cyan)",
    topic: "regex-trainer · new contributor · added lookahead exercises",
    replies: 2,
  },
  {
    when: "8h",
    by: "grace",
    category: "#elixir",
    categoryColor: "var(--term-magenta)",
    topic: "phoenix study group · weds 7pm · open to all levels",
    replies: 7,
  },
];

export const WELCOME_COPY = {
  title: "code self study",
  tagline:
    "a friendly programming community in the san francisco bay area · est. 2014 · all languages, all levels, in person + online",
  /**
   * The "5,000+" substring is intended to be visually highlighted by the
   * <WelcomePane> component (e.g. wrapped in a styled span).
   */
  lead: "Self-study computer programming alongside 5,000+ members in Berkeley and the SF Bay Area. Bring a project, bring a question, or just show up. We meet in person and online.",
  leadHighlight: "5,000+",
  whatWeDo: [
    "build a local community of motivated programmers",
    "open to every language & level",
    "self-study with mentorship from experienced members",
    "bootcamp supplement or alternative",
  ],
  cta: {
    label: "Join",
    href: "https://www.meetup.com/codeselfstudy/",
  },
} as const;

export type TmuxWindow = {
  id: number;
  label: string;
  href?: string;
};

export const TMUX_SESSION = "codeselfstudy" as const;

export const TMUX_WINDOWS: Array<TmuxWindow> = [
  { id: 0, label: "home", href: "/" },
  { id: 1, label: "forum", href: "/forum" },
  { id: 2, label: "events", href: "/events" },
  { id: 3, label: "learn", href: "/learn" },
  { id: 4, label: "jobs", href: "/jobs" },
];

export const TMUX_RIGHT_TEXT = "members 5,000+ · est. 2014" as const;
