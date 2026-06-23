# Frontend smoke test

Expected local development URL:

```text
http://localhost:8325/
```

For remote development, use the remote machine host or IP with port `8325`.

## Checklist

- Start MATE with `go run ./cmd/mate`.
- Open `/` in a browser.
- Confirm the MATE header is visible.
- Confirm the toolbox is visible.
- Confirm the graph workspace is visible.
- Confirm the status indicator is visible.
- Confirm the browser console has no errors on initial load.
- Confirm `/health` returns HTTP 200.
- Confirm `/static/css/main.css` returns HTTP 200.
- Confirm `/static/js/app.js` returns HTTP 200.
- Confirm a missing static file returns HTTP 404.
- Refresh `/` and confirm the page still loads.
