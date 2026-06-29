# MATE UI concept

MATE uses a node based workspace inspired by Node RED, adapted for relationship mapping.

## Main areas

- Header: brand, selected network, selected domain, network search, global search, graph controls, account menu, and connection status.
- Toolbox: available tools for the selected network domain.
- Workspace: canvas-like relationship map with SVG links and draggable nodes.
- Inspector: selected node details and related connections.
- Modals: create and edit nodes or relationships, administer accounts, and manage network actions as 0.13 grows.

## Account menu

The account menu replaces the earlier plain username and sign-out button.

- The trigger shows the signed-in account display name or account name.
- All users can sign out.
- Profile and system settings are placeholders in the first 0.13 slice.
- Preferences opens browser-local node color preferences.
- Owner and admin users can open Account Administration.
- Owner users can see the System settings placeholder.

## Account administration

Account Administration is a modal window.

- Owners can list accounts, create accounts, assign roles, change roles, and disable accounts.
- Admins can list accounts and manage `editor` and `viewer` accounts.
- Existing passwords are never displayed.
- Disable uses an explicit two-step in-window confirmation instead of a browser confirm dialog.

## Network actions

The selected network area supports both fast actions and a network actions menu.

- The plus button opens the Create Network dialog.
- The pencil button opens the Rename Network dialog.
- The network actions menu groups create, rename, edit metadata, delete, and future manage-networks behavior.
- Create, rename, and metadata edits use the same HTML dialog foundation.
- Delete is treated as destructive and requires a second in-dialog confirmation.
- Deleting a Social Network network removes it from the owner's network list but does not delete global person identities.

## Domains and work modes

0.15 introduces a lightweight network domain concept.

- `social` is the default domain for relationship maps with global person identity behavior.
- `flowchart` is the first experimental non-social domain.
- The selected domain is stored on network metadata.
- The header domain selector updates the current network domain.
- The Create toolbox filters available node tools by domain.
- Social Network shows person, organization, project, and family-placeholder tools.
- Flowchart shows Start, Stop, Process, Decision, Input, Output, Merge, and Delay tools.
- Flowchart creates network-scoped diagram nodes, not global identity nodes.
- Duplicate person matching only runs for Social Network person creation.
- Flowchart search is scoped to the currently loaded graph instead of global person and organization identity search.
- Multiple relationships between the same two nodes render as separated curved links.
- Flowchart relationships have first-pass visual styles: `yes` is green, `no` and `error` are red, and `loop` is dashed amber.

Work modes remain lightweight in 0.15. The current domain can influence toolbox ordering, available tools, labels, and future export or print behavior, but domains are not database-configurable yet.

## Context menus

MATE 0.13 introduces first-pass right-click context menus.

- Node context menus expose edit, create relationship from node, fit selection, hide, and delete.
- Relationship context menus expose edit, select/show connected nodes, and delete.
- Canvas context menus expose create person here, fit all nodes, reset zoom, show hidden, and clear selection.
- Future actions such as duplicate, lock, unlock, paste, and broader create menus may be visible as disabled placeholders until their behavior is implemented.

## Toolbox layout

The toolbox uses collapsible vertical sections and node-like tool tiles.

- Navigation, Selection, Graph, Data, and Create can be expanded or collapsed independently.
- Section collapse state is stored in browser local storage.
- Tool actions render as three-column square tiles, matching the create-node visual model.
- Tool names are exposed through tooltip and accessibility labels rather than explanatory text beside the tile.
- Data tools expose selected-node edit and delete actions. Duplicate is visible as a disabled placeholder until a real duplicate flow is implemented.

## Selection

Selection supports direct node click, Ctrl/Cmd multi-select, drag-box selection from empty canvas, connected-node selection, and same-type selection. Organization subtypes are treated as the same broad organization type for same-type selection.

## Node locking

Node locking is currently a frontend graph-control state. Locked nodes remain selectable and editable, but they do not move by accidental drag and auto layout skips them. Persistence is deferred until the network-owned UI state model is finalized.

## Graph framing

When a network graph is first loaded or the user switches networks, MATE automatically fits the visible graph into the workspace. Manual pan, zoom, fit, and reset actions mark the viewport as user-controlled, so reloading the same network does not unexpectedly override the user's current view.

## Attribute states

Inspector attribute lists distinguish loading, empty, and error states. Empty means the API request succeeded and there are no active attributes. Error means the attribute request failed and is styled separately so it is not confused with missing data.

## Date Inputs

Relationship and attribute dialogs use the same date input pattern. The accepted formats are `YYYY`, `YYYY-MM`, and `YYYY-MM-DD`; invalid values are blocked before saving.

## Color Preferences

The first color preference design is implemented as browser-local, account-keyed UI state. Users can override supported node colors from the Preferences dialog. Person gender colors remain separate configurable values, and resetting preferences restores the default palette. Backend persistence is deferred until account-owned versus network-owned preference storage is finalized.

## Node types

- Person
- Company
- Association
- School
- Location
- Tag

## Interaction model

- Click a toolbox item, then click the workspace to create a temporary node.
- Drag a toolbox item onto the workspace to create a temporary node.
- Click a node to select it and open the inspector.
- Drag a node to reposition it.
- Shift-click a source node, then click a target node to create a relationship.

## Current limitation

In version 0.5, graph changes are kept in temporary frontend state. Backend persistence is planned for later API and storage versions.


## Next integration step

The next implementation step is backend API integration with temporary storage. The frontend should keep the current interaction model while replacing temporary graph state with API-backed create, list, update, and relationship calls.
