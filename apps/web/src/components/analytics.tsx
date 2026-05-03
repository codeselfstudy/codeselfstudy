// TODO: double-check that this works. I don't know how TanStack handles page changes.
export function Analytics() {
  return (
    <>
      {/* Google Analytics */}
      <script
        async
        src="https://www.googletagmanager.com/gtag/js?id=G-99DC6WSRWL"
      />
      <script
        dangerouslySetInnerHTML={{
          __html: `
            window.dataLayer = window.dataLayer || [];
            function gtag(){dataLayer.push(arguments);}
            gtag('js', new Date());

            gtag('config', 'G-99DC6WSRWL');
          `,
        }}
      />
    </>
  );
}
