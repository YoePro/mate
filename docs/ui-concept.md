# MATE UI concept

MATE uses a node based workspace inspired by Node RED, adapted for relationship mapping.

## Main areas

- Header: brand, search, graph controls, and connection status.
- Toolbox: available node types for people, organizations, locations, and tags.
- Workspace: canvas-like relationship map with SVG links and draggable nodes.
- Inspector: selected node details and related connections.
- Modals: create and edit nodes or relationships.

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
