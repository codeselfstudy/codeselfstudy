/**
 * Collapse a multi-line string to one line: trim each line, drop blank lines,
 * and join with a single space. Handy for inlining snippets like the Google
 * Analytics config into a one-line <script>.
 *
 * This is a whitespace flattener, not a real minifier: it assumes the input has
 * no single-line `//` comments, which would swallow the following code once
 * everything is on one line.
 */
export function minify(text: string): string {
  return text
    .split("\n")
    .map((line) => line.trim())
    .filter((line) => line)
    .join(" ");
}
