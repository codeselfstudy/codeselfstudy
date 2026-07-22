// Kill-switch service worker.
//
// codeselfstudy.com previously ran Gatsby (gatsby-plugin-offline) and
// create-react-app, both of which registered a root-scope service worker that
// precached the HTML and CSS bundle. After the migration to Astro those became
// "zombies": a returning visitor's browser keeps serving the old cached
// index.html, which links a CSS hash that no longer exists on the server, so
// the page loads unstyled. A 404 on the old /sw.js does NOT clear the
// registration, so the stale worker persists until site data is cleared.
//
// This script replaces the old /sw.js. Browsers re-fetch the top-level worker
// script on their next navigation (it bypasses the HTTP cache on update
// checks), install this version, and it unregisters itself, clears all caches,
// and reloads open tabs so they fetch fresh, worker-free assets. The current
// Astro site registers no service worker, so fresh visitors never fetch or run
// this file. Safe to delete once stale-worker traffic has drained (keep it for
// several months).
self.addEventListener("install", () => self.skipWaiting());

self.addEventListener("activate", (event) => {
  event.waitUntil(
    (async () => {
      try {
        const keys = await caches.keys();
        await Promise.all(keys.map((key) => caches.delete(key)));
      } catch {
        /* Cache Storage unavailable */
      }
      await self.registration.unregister();
      const clients = await self.clients.matchAll({
        type: "window",
        includeUncontrolled: true,
      });
      for (const client of clients) {
        client.navigate(client.url);
      }
    })()
  );
});
