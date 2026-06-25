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

## 0.11 Graph UI Checklist

- Confirm the left toolbox has grouped `Navigation`, `Selection`, `Graph`, and `Create` areas.
- Confirm `Fit all nodes`, `Fit selection`, `Zoom in`, `Zoom out`, and `Reset zoom` respond without console errors.
- Create a male person from the blue person icon and confirm the modal has gender `m`.
- Create a female person from the pink person icon and confirm the modal has gender `f`.
- Create an other/unknown person from the gray person icon and confirm the modal has gender `o`.
- Confirm male person nodes render blue, female person nodes render pink, and other/unknown person nodes render neutral gray.
- Create at least one organization subtype from the organization icon grid.
- Create a project from the Project icon.
- Link a person to an organization and confirm only person-to-organization relationship options are shown.
- Create a `works_at` relationship and set `role`, `start`, `end`, and `current`.
- Link a person to a project and confirm `works_on` is available.
- Link an organization to a project and confirm `sponsors` is available.
- Create a custom person-to-person relationship type and confirm it appears again for the same source/target type pair.
- Reload the graph and confirm the custom relationship type is still available in that network.
- Use Ctrl/Cmd-click to multi-select nodes.
- Use `Select connected` and confirm neighboring nodes are selected.
- Use `Hide selected` and confirm selected nodes and their links disappear.
- Use `Show hidden` and confirm hidden nodes return.
- Use `Auto layout` and confirm visible nodes are repositioned.
- Archive an organization or project and confirm it disappears after reload.
- Confirm the browser console has no errors during these operations.

## 0.11 Permission Checklist

- Sign in as a second user if available.
- Confirm the second user cannot open or edit another user's network graph.
- Confirm the second user cannot create custom relationship types in another user's network.
- Confirm network search results do not expose graph nodes, relationship notes, or positions for networks the user does not own.
