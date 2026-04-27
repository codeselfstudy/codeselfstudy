import { TermPane } from "./term-pane";
import type { ForumPost } from "@/content/mock-home-data";
import { FORUM_ACTIVITY } from "@/content/mock-home-data";

type ForumActivityPaneProps = {
  posts?: Array<ForumPost>;
};

export function ForumActivityPane({
  posts = FORUM_ACTIVITY,
}: ForumActivityPaneProps = {}) {
  return (
    <TermPane
      index={3}
      label="recent activity · forum"
      labelColor="magenta"
      keys="↻ live · enter open"
      variant="full"
    >
      <table
        style={{
          width: "100%",
          borderCollapse: "collapse",
          fontSize: 12,
          fontFamily: "inherit",
        }}
      >
        <thead>
          <tr>
            <Th>when</Th>
            <Th>by</Th>
            <Th>category</Th>
            <Th>topic</Th>
            <Th align="right">replies</Th>
          </tr>
        </thead>
        <tbody>
          {posts.map((post) => (
            <tr key={`${post.when}-${post.by}`}>
              <Td color="var(--term-fg-dim)">{post.when}</Td>
              <Td color="var(--term-cyan)">{post.by}</Td>
              <Td>
                <span style={{ color: post.categoryColor }}>
                  {post.category}
                </span>
              </Td>
              <Td color="var(--term-fg)" wrap>
                {post.topic}
              </Td>
              <Td color="var(--term-accent)" align="right">
                {post.replies}
              </Td>
            </tr>
          ))}
        </tbody>
      </table>
    </TermPane>
  );
}

function Th({
  children,
  align = "left",
}: {
  children: React.ReactNode;
  align?: "left" | "right";
}) {
  return (
    <th
      style={{
        textAlign: align,
        padding: "2px 8px 2px 0",
        whiteSpace: "nowrap",
        color: "var(--term-fg-dim)",
        borderBottom: "1px dashed var(--term-border)",
        fontWeight: "normal",
        textTransform: "uppercase",
        letterSpacing: "0.06em",
        fontSize: 11,
      }}
    >
      {children}
    </th>
  );
}

function Td({
  children,
  color,
  align = "left",
  wrap = false,
}: {
  children: React.ReactNode;
  color?: string;
  align?: "left" | "right";
  wrap?: boolean;
}) {
  return (
    <td
      style={{
        textAlign: align,
        padding: "2px 8px 2px 0",
        whiteSpace: wrap ? "normal" : "nowrap",
        width: wrap ? "100%" : undefined,
        color,
      }}
    >
      {children}
    </td>
  );
}
