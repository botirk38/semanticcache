# backends -- Agent Instructions

## What this package does
Re-exports backend constructors from `inmemory/` and `remote/` for convenience. No logic of its own.

## Rules
- When adding a new backend subpackage, add a re-export function here.
- Do not put implementation code in this package.

## Subpackages
- `inmemory/` -- LRU, LFU, FIFO
- `remote/` -- Redis
