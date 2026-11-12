const CACHE_NAME = 'mc-images-v1';
const IMAGE_HOSTS = ['cards.scryfall.io', 'c1.scryfall.com', 'svgs.scryfall.io'];

self.addEventListener('install', () => self.skipWaiting());

self.addEventListener('activate', e => {
  e.waitUntil(
    caches.keys().then(keys =>
      Promise.all(keys.filter(k => k !== CACHE_NAME).map(k => caches.delete(k)))
    ).then(() => self.clients.claim())
  );
});

self.addEventListener('fetch', e => {
  let host;
  try { host = new URL(e.request.url).hostname; } catch { return; }
  if (!IMAGE_HOSTS.includes(host)) return;

  e.respondWith(
    caches.open(CACHE_NAME).then(async cache => {
      const cached = await cache.match(e.request);
      if (cached) return cached;
      try {
        const res = await fetch(e.request);
        if (res.ok) cache.put(e.request, res.clone());
        return res;
      } catch {
        return new Response(null, { status: 503, statusText: 'Offline' });
      }
    })
  );
});
