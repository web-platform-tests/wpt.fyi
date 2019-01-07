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
              const components = new RegExp('^/((bower_)?components|static)/');
              if (components.test(e.request.url.toString())) {
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
