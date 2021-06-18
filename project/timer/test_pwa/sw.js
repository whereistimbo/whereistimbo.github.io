self.addEventListener('install', (e) => {
  e.waitUntil(
    caches.open('test_pwa').then((cache) => cache.addAll([
	  '.',
      'index.html',
      'app.js',
    ])),
  );
});

self.addEventListener('fetch', (e) => {
  console.log(e.request.url);
  e.respondWith(
    caches.match(e.request).then((response) => response || fetch(e.request)),
  );
});