# UI Text Runtime

Owns the local-first runtime for user-visible text.

Resolution order is remote override, last known good package, then built-in
text. The runtime validates the package before activation and must never make
the application unable to render when a network request fails.

Platform, language, package signature, client compatibility, key coverage,
variable compatibility, and maximum length are checked before a remote package
becomes active.
