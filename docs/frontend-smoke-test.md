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
- If an owner exists, open `Create account`, create a self-service account, and confirm it signs in.
- Sign in with an existing account.
- Confirm the browser redirects to `/`.
- Confirm the MATE header is visible.
- Confirm the current account menu is visible in the header.
- Open the account menu and confirm `Sign out` is available.
- Confirm the toolbox is visible.
- Confirm the graph workspace is visible.
- Confirm the status indicator is visible.
- Confirm the browser console has no errors on initial load.
- Refresh `/` and confirm the session still loads the page.
- Click account menu `Sign out` and confirm the browser returns to `/login`.
- Open `/` without a valid session and confirm it redirects to `/login`.
- Confirm `/health` returns HTTP 200.
- Confirm `/login` returns HTTP 200.
- Confirm `/static/css/main.css` returns HTTP 200.
- Confirm `/static/js/app.js` returns HTTP 200.
- Confirm a missing static file returns HTTP 404.

## 0.11 Graph UI Checklist

- Confirm the left toolbox has grouped `Navigation`, `Selection`, `Graph`, `Data`, and `Create` areas.
- Confirm action tools render as three-column square tiles without trailing explanatory text.
- Collapse and expand a toolbox section and confirm only that section changes.
- Refresh the browser and confirm section collapse state persists.
- Confirm `Fit all nodes`, `Fit selection`, `Zoom in`, `Zoom out`, and `Reset zoom` respond without console errors.
- Create a male person from the blue person icon and confirm the modal has gender `m`.
- Create a female person from the pink person icon and confirm the modal has gender `f`.
- Create an other/unknown person from the gray person icon and confirm the modal has gender `o`.
- Confirm male person nodes render blue, female person nodes render pink, and other/unknown person nodes render neutral gray.
- Reload a network with several nodes and confirm the graph is framed automatically on first load.
- Pan or zoom manually, reload the same network, and confirm the viewport is not forced back.
- Switch to another network and confirm that network is framed automatically.
- Create at least one organization subtype from the organization icon grid.
- Create a project from the Project icon.
- Create a person whose name matches an existing person and confirm the duplicate suggestion opens as an HTML dialog.
- Confirm duplicate suggestions show confidence, reasons, and useful person details before choosing.
- Choose `Use selected` and confirm the existing person is linked to the selected network.
- Repeat and choose `Create new person` to confirm a new global person can still be created.
- Link a person to an organization and confirm only person-to-organization relationship options are shown.
- Create a `works_at` relationship and set `role`, `start`, `end`, and `current`.
- Link a person to a project and confirm `works_on` is available.
- Link an organization to a project and confirm `sponsors` is available.
- Create a custom person-to-person relationship type and confirm it appears again for the same source/target type pair.
- Reload the graph and confirm the custom relationship type is still available in that network.
- Click a relationship and confirm the relationship property dialog opens.
- Edit relationship role, dates, current flag, notes, and custom label where applicable.
- Enter invalid relationship dates and confirm an HTML validation dialog appears.
- Enter valid relationship dates using `YYYY`, `YYYY-MM`, and `YYYY-MM-DD`.
- Save the relationship and confirm the updated label/context survives reload.
- Use the relationship dialog Delete button and confirm delete confirmation opens as an HTML dialog, not a browser dialog.
- Cancel relationship deletion and confirm the relationship remains.
- Confirm relationship deletion through the dialog and confirm the relationship is removed.
- Use Ctrl/Cmd-click to multi-select nodes.
- Drag on empty canvas to box-select several nodes.
- Use `Select connected` and confirm neighboring nodes are selected.
- Use `Select same type` and confirm nodes of the same broad type are selected.
- Use `Hide selected` and confirm selected nodes and their links disappear.
- Use `Show hidden` and confirm hidden nodes return.
- Use `Auto layout` and confirm visible nodes are repositioned.
- Lock a selected node and confirm it cannot be moved by drag.
- Run `Auto layout` and confirm locked nodes do not move.
- Unlock the node and confirm it can be dragged again.
- Select a person or organization with no attributes and confirm a neutral empty state is shown.
- Add an attribute and confirm the empty state is replaced by the attribute.
- Enter invalid attribute dates and confirm an HTML validation dialog appears.
- Enter valid attribute dates using `YYYY`, `YYYY-MM`, and `YYYY-MM-DD`.
- Delete an organization or project and confirm it disappears after reload.
- Delete a node from the edit dialog and confirm the confirmation is an HTML dialog, not a browser dialog.
- Press Delete or Backspace on a selected node and confirm the confirmation is an HTML dialog, not a browser dialog.
- Confirm the browser console has no errors during these operations.

## 0.11 Permission Checklist

- Sign in as a second user if available.
- Confirm the second user cannot open or edit another user's network graph.
- Confirm the second user cannot create custom relationship types in another user's network.
- Confirm network search results do not expose graph nodes, relationship notes, or positions for networks the user does not own.

## 0.13 Account UI Checklist

- Log in as owner.
- Open the account menu.
- Open Preferences and confirm node color controls are visible.
- Change a supported node color and confirm existing graph nodes update.
- Reset colors and confirm defaults return.
- Confirm Account Administration is visible.
- Open Account Administration.
- Confirm accounts load without password hashes or password fields.
- Create a `viewer` account.
- Create an `editor` account.
- Change an account role.
- Disable a non-owner test account using the two-step Disable/Confirm flow.
- Confirm the UI blocks disabling the current account.
- Confirm backend rejects disabling or demoting the last usable owner.
- Log in as admin if available.
- Confirm admin can manage `editor` and `viewer` accounts.
- Confirm admin cannot manage `owner` or `admin` accounts.
- Log in as editor or viewer if available.
- Confirm Account Administration is hidden.
- Create a self-service account from `/login` and confirm Account Administration is hidden for that account.

## 0.13 Network UI Checklist

- Confirm the selected network area has direct create/rename buttons and a network actions menu.
- Create a network from the direct create button.
- Create a network from the network actions menu.
- Rename the current network from the direct rename button.
- Edit network metadata from the network actions menu.
- Confirm network create, rename, and metadata errors render in the HTML dialog rather than browser alerts.
- Delete a disposable network from the network actions menu.
- Confirm delete requires the second `Confirm delete` click.
- Confirm deleted networks disappear from the owned network list.
- Confirm deleting a Social Network network does not delete global person identities.

## 0.13 Context Menu Checklist

- Right-click a person node and confirm a node context menu opens.
- From the node context menu, test Edit, Create relationship from node, Fit selection, Hide, and Delete.
- Confirm disabled node actions are visibly disabled.
- Right-click a relationship and confirm a relationship context menu opens.
- From the relationship context menu, test Edit relationship, Select connected nodes, Show connected nodes, and Delete relationship.
- Right-click empty canvas and confirm a canvas context menu opens.
- From the canvas context menu, test Create person here, Fit all nodes, Reset zoom, Show hidden, and Clear selection.
- Confirm context menus close on outside click, Escape, and action selection.

## 0.13 Toolbox Layout Checklist

- Confirm toolbox sections expand and collapse vertically.
- Confirm collapsed sections hide their tools without changing toolbox width.
- Confirm action tools and create tools use a consistent three-column tile layout.
- Confirm tool labels are available through browser tooltip or accessible label.
- Confirm toolbox section state persists after browser reload.
- Confirm Data section Edit and Delete are disabled when nothing is selected.
- Select a node and confirm Data section Edit opens the edit dialog.
- Select a node and confirm Data section Delete opens the HTML delete confirmation.
- Confirm Data section Duplicate is visibly disabled.

## 0.13 Selection Checklist

- Drag a box on empty canvas and confirm nodes inside the box become selected.
- Ctrl/Cmd-drag a box and confirm the matched nodes are added to the existing selection.
- Confirm drag-box selection does not start when dragging an existing node.
- Select one person and use `Select same type`; confirm all visible people are selected.
- Select one organization subtype and use `Select same type`; confirm visible organizations are selected.
- Select one project and use `Select same type`; confirm visible projects are selected.

## 0.13 Node Lock Checklist

- Select one or more nodes and use `Lock selected`.
- Confirm locked nodes show a lock marker.
- Confirm locked nodes can still be selected and edited.
- Confirm locked nodes do not move when dragged.
- Confirm auto layout moves unlocked visible nodes but leaves locked nodes in place.
- Use node context menu `Unlock` and confirm dragging works again.

## 0.13 Attribute State Checklist

- Select a person without attributes and confirm `No attributes yet.` appears as a neutral state.
- Select an organization without attributes and confirm the same neutral state.
- Add an attribute and confirm the attribute list replaces the neutral state.
- Simulate or observe an attribute API failure and confirm the error state is visually distinct from the empty state.
- Confirm attribute save errors use an HTML dialog, not a browser alert.

## 0.15 Domain Toolbox Checklist

- Create or open a Social Network network and confirm the domain selector shows `Social Network`.
- Confirm the Social Network Create toolbox shows person, organization, project, and family-placeholder tools.
- Create a person in a Social Network network and confirm duplicate detection still appears for matching names.
- Create a Flowchart network and confirm the domain selector shows `Flowchart`.
- Confirm the Flowchart Create toolbox hides Social Network create tools.
- Confirm Flowchart Start, Stop, Process, Decision, Input, Output, Merge, and Delay tools can create nodes.
- Reload the Flowchart network and confirm created Flowchart nodes remain.
- Create a Flowchart relationship using `next`, `yes`, `no`, `loop`, or `error`.
- Confirm Flowchart relationship styles differ for `yes`, `no`, `loop`, and `error`.
- Create two or more relationships between the same two nodes and confirm the links do not overlap.
- Confirm direct API calls cannot create Flowchart nodes or Flowchart relationships in Social Network networks.
- Confirm direct API calls cannot add persons or social relationships in Flowchart networks.
- Confirm Flowchart search only searches the loaded graph, not global person or organization identity data.
- Search networks and confirm result metadata includes the network domain.
- Switch domains on an owned network and reload; confirm the selected domain is restored.
- Confirm a viewer can see the selected domain but cannot change it.

## 0.13 Preference Checklist

- Open account menu Preferences.
- Change male, female, other, organization, and project colors.
- Confirm existing graph nodes update after saving.
- Refresh the browser and confirm saved color preferences remain.
- Reset colors and confirm defaults return.

## 0.13 Date Input Checklist

- Confirm relationship start and end fields accept `YYYY`, `YYYY-MM`, and `YYYY-MM-DD`.
- Confirm relationship start and end fields reject malformed dates.
- Confirm attribute start and end fields accept `YYYY`, `YYYY-MM`, and `YYYY-MM-DD`.
- Confirm attribute start and end fields reject malformed dates.
