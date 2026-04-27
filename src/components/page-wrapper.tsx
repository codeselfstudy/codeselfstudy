import { useLocation } from "@tanstack/react-router";
import type { ReactNode } from "react";
import { TermPane } from "@/components/terminal/term-pane";

type PageWrapperProps = {
  children: ReactNode;
  /** Pane label override. Defaults to `~/<first route segment>`. */
  paneTitle?: string;
  paneIndex?: number;
};

function deriveTitle(pathname: string): string {
  const seg = pathname.replace(/^\/+/, "").split("/")[0] ?? "";
  return seg ? `~/${seg}` : "~/home";
}

export function PageWrapper({
  children,
  paneTitle,
  paneIndex = 0,
}: PageWrapperProps) {
  const { pathname } = useLocation();
  return (
    <TermPane
      index={paneIndex}
      label={paneTitle ?? deriveTitle(pathname)}
      variant="full"
    >
      {children}
    </TermPane>
  );
}
