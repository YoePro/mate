# Frontend smoke test

Expected local development URL:

```text
http://localhost:8325/
```

For remote development, use the remote machine host or IP with port `8325`.

## Checklist

- Start Neo4j.
- Start MATE with `go run ./cmd/mate` or `./mate`.
- Open `/login` in a browser.
- If no owner exists, create the first owner account.
- Sign in with an existing account.
- Confirm the browser redirects to `/`.
- Confirm the MATE header is visible.
- Confirm the current account label is visible in the header.
- Confirm the toolbox is visible.
- Confirm the graph workspace is visible.
- Confirm the status indicator is visible.
- Confirm the browser console has no errors on initial load.
- Refresh `/` and confirm the session still loads the page.
- Click `Sign out` and confirm the browser returns to `/login`.
- Open `/` without a valid session and confirm it redirects to `/login`.
- Confirm `/health` returns HTTP 200.
- Confirm `/login` returns HTTP 200.
- Confirm `/static/css/main.css` returns HTTP 200.
- Confirm `/static/js/app.js` returns HTTP 200.
- Confirm a missing static file returns HTTP 404.
