# Vinta Documentation

This directory contains design, style, and community documentation for Vinta.

## Documentation Index

- [DESIGN.md](DESIGN.md): Visual aesthetics, design tokens, color variables, typography, form control standards, and bento card architecture.
- [CONTRIBUTORS.md](CONTRIBUTORS.md): Contributors Hall of Fame recognizing developers, designers, and community members.

## Core Contributing Guides

For project setup, architecture overview, and git contribution workflows, refer to the root [CONTRIBUTING.md](../CONTRIBUTING.md).

For vulnerability disclosure and safety protocols, refer to [SECURITY.md](../SECURITY.md).

## Developer Checks

Run these checks locally via `task` or `scripts.bat` before opening a pull request (pass flags with `task <name> -- <flags>` or `scripts.bat <name> <flags>`, or view command help via `task <name> -- --help` or `task help -- <name>`):

- `task check:themes` (`scripts.bat check-themes`): Validates `--vinta-*` CSS token references and theme preset parity.
- `task check:locales` (`scripts.bat check-locales`): Validates translation key parity and placeholder tokens across all locale catalogs.
- `task check:settings` (`scripts.bat check-settings`): Verifies unique `vinta_*` keys, setting defaults, schema migrations, and audits direct `localStorage` usage. Use `task check:settings -- --defaults` to print every setting's default value.
- `task check:transitions` / `task check:motion` (`scripts.bat check-transitions`): Audits non-composited CSS transitions in production code (excluding `devtest`) and verifies `prefers-reduced-motion` support. Intentional container-bounds morphs use `/* vinta-ignore: transition */`.
- `task check:hardcoded-ui` (`scripts.bat check-hardcoded-ui`): Scans `frontend/src/components` and `frontend/src/pages` for hardcoded colors, native HTML controls, JS CSS imports, and hardcoded UI attribute strings. Pass `--changed --strict` or `--all` to inspect all findings.
