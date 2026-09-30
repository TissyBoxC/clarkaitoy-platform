# UI Text Contracts

Owns the versioned contracts for remote user-interface text, including the
package manifest, text entry, language, platform scope, signature metadata,
and validation error schema.

## Boundaries

- `device_platform.internal.modules.ui_text` publishes and serves packages.
- The parent application and admin console consume the same contract.
- Firmware consumes a constrained package through its `ui_text` module.
- Security-sensitive text such as guardian consent keeps its own version and
  approval record. It must not be treated as ordinary UI copy.

## Rules

- Keep contract identifiers `lower_snake_case` and JSON fields `snake_case`.
- Add `schema_version` to every breaking contract change.
- Never place provider credentials, signing private keys, or personalized child
  data in a text package.
- Every text value declares its allowed variables and maximum length.
