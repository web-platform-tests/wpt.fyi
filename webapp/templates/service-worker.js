// Clean up old caches.
self.addEventListener(
  'activate',
  e => {
    e.waitUntil(
      caches.keys()
        .then(cacheNames => Promise.all(
          cacheNames
            .filter(v => '{{ .Version }}' !== v)
            .map(v => caches.delete(v))
        ))
    );
  }
);

const cacheablePath = new RegExp('^/(components|static|node_modules)/');

// Locally cache eligible components/files.
self.addEventListener(
  'fetch',
  e => {
    return e.respondWith(
      caches.match(e.request)
        .then(r => {
          if (r) {
            return r;
          }

          // IMPORTANT: Clone the request to reuse in fetch.
          const request = e.request.clone();
          const url = new URL(e.request.url);
          const path = url.pathname;

          return fetch(request)
            .then(r => {
              // Do not cache failed or CORS requests.
              if (r.ok && r.type === 'basic') {
                if (cacheablePath.test(path) && path !== '/components/wpt-env-flags.js') {
                  // IMPORTANT: Clone the response to reuse in caches.
                  const responseToCache = r.clone();
                  caches.open('{{ .Version }}')
                    .then(cache => cache.put(e.request, responseToCache));
                }
              }
              return r;
            });
        }
      )
    );
  }
);

self.addEventListener('message', function (event) {
  if (event.data.action === 'skipWaiting') {
    self.skipWaiting();
  }
});

self.addEventListener('push', async (event) => {
  const data = await event.data.json();
  registration.showNotification(data.title, data.options);
});

self.addEventListener('notificationclick', function (event) {
  const clickedNotification = event.notification;
  clickedNotification.close();
  const url = new URL(clickedNotification.data.url || '/', location);

  if ('openWindow' in clients) {
    event.waitUntil(clients.openWindow(url));
  }
});