/**
 * This file is the root of the application. It generates the layout.
 */
import {
  HeadContent,
  Scripts,
  createRootRouteWithContext,
  redirect,
  useLocation,
} from "@tanstack/react-router";
import { TanStackRouterDevtoolsPanel } from "@tanstack/react-router-devtools";
import { TanStackDevtools } from "@tanstack/react-devtools";
import type { QueryClient } from "@tanstack/react-query";
import { findRedirect } from "@/lib/redirects";
import WorkOSProvider from "@/integrations/workos/provider";
import TanStackQueryDevtools from "@/integrations/tanstack-query/devtools";
import appCss from "@/styles.css?url";
import { Analytics } from "@/components/analytics";
import { ErrorPage } from "@/components/error-page";
import { NotFound } from "@/components/not-found";
import { TerminalShell } from "@/components/terminal/terminal-shell";
import { TermNav } from "@/components/terminal/term-nav";
import { ThemeToggle } from "@/components/terminal/theme-toggle";
import { Clock, TmuxStatusBar } from "@/components/terminal/tmux-status-bar";
import {
  TMUX_RIGHT_TEXT,
  TMUX_SESSION,
  TMUX_WINDOWS,
} from "@/content/mock-home-data";

interface MyRouterContext {
  queryClient: QueryClient;
}

export const Route = createRootRouteWithContext<MyRouterContext>()({
  // Redirect system
  beforeLoad: ({ location }) => {
    const redir = findRedirect(location.pathname);
    if (redir) {
      throw redirect({
        to: redir.to,
        statusCode: redir.status,
      });
    }
  },
  head: () => ({
    meta: [
      {
        charSet: "utf-8",
      },
      {
        name: "viewport",
        content: "width=device-width, initial-scale=1",
      },
      {
        title: "Code Self Study",
      },
    ],
    links: [
      {
        rel: "stylesheet",
        href: appCss,
      },
    ],
  }),

  shellComponent: RootDocument,
  notFoundComponent: NotFound,
  errorComponent: ({ error }) => <ErrorPage error={error} />,
});

function activeWindowIndex(pathname: string): number {
  const i = TMUX_WINDOWS.findIndex(
    (w) =>
      w.href === pathname ||
      (w.href !== "/" && pathname.startsWith(`${w.href}/`))
  );
  return i === -1 ? 0 : i;
}

function RootDocument({ children }: { children: React.ReactNode }) {
  const location = useLocation();

  return (
    <html lang="en">
      <head>
        <HeadContent />
        {import.meta.env.PROD && <Analytics />}
        <link
          rel="canonical"
          href={`https://codeselfstudy.com${location.pathname}`}
        />
      </head>
      <body>
        <WorkOSProvider>
          <TerminalShell
            nav={<TermNav themeToggle={<ThemeToggle />} />}
            statusBar={
              <TmuxStatusBar
                session={TMUX_SESSION}
                windows={TMUX_WINDOWS}
                activeIndex={activeWindowIndex(location.pathname)}
                right={
                  <>
                    {TMUX_RIGHT_TEXT} · <Clock />
                  </>
                }
              />
            }
          >
            {children}
          </TerminalShell>
          <TanStackDevtools
            config={{
              position: "bottom-right",
            }}
            plugins={[
              {
                name: "Tanstack Router",
                render: <TanStackRouterDevtoolsPanel />,
              },
              TanStackQueryDevtools,
            ]}
          />
        </WorkOSProvider>
        <Scripts />
      </body>
    </html>
  );
}
