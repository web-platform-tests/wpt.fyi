self.addEventListener(
  'install',
  e => e.waitUntil(
    caches.open('{{ .Version }}').then(cache => cache.addAll([
      {{- range .Files }}
      '/{{ . }}',
      {{- end }}
    ])).catch(e => {
      console.error(e);
    })
  )
);

self.addEventListener(
  'fetch',
  e => {
    return e.respondWith(
      caches.match(e.request)
        .then(r => {
          if (r) {
            return r;
          }
          return fetch(e.request)
            .then(r => {
              // NOTE(lukebjerring): wait for actually-used bower_components,
              // to avoid bloating the cache with unused files.
              if (e.request.url.toString().includes('/bower_components/')) {
                let clone = r.clone();
                caches.open('{{ .Version }}').then(cache => {
                  cache.put(e.request, clone);
                });
              }
              return r;
            });
        }
      )
    );
  }
);
