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
import redirects from "@/data/redirects";
import WorkOSProvider from "@/integrations/workos/provider";
import TanStackQueryDevtools from "@/integrations/tanstack-query/devtools";
import appCss from "@/styles.css?url";
import { Footer } from "@/components/footer";
import { Navbar } from "@/components/navbar";

interface MyRouterContext {
  queryClient: QueryClient;
}

export const Route = createRootRouteWithContext<MyRouterContext>()({
  // Redirect system
  beforeLoad: ({ location }) => {
    const redir = redirects[location.pathname];
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
});

function RootDocument({ children }: { children: React.ReactNode }) {
  const location = useLocation();

  return (
    <html lang="en">
      <head>
        <HeadContent />
        <link
          rel="canonical"
          href={`https://codeselfstudy.com${location.pathname}`}
        />
      </head>
      <body>
        <WorkOSProvider>
          <Navbar />
          {children}
          <Footer />
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
