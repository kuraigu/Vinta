# Vinta

[![Go Version](https://img.shields.io/badge/Go-1.26%2B-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Svelte Version](https://img.shields.io/badge/Svelte-v5-FF3E00?style=flat&logo=svelte)](https://svelte.dev/)
[![TypeScript](https://img.shields.io/badge/TypeScript-v5-3178C6?style=flat&logo=typescript)](https://www.typescriptlang.org/)
[![Wails Framework](https://img.shields.io/badge/Wails-v3-E64A19?style=flat&logo=wails)](https://wails.io/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

Vinta is an open-source, beautifully designed Adaptive Island widget built specifically for Windows PC. Floating effortlessly at the top of your screen, Vinta turns desktop notifications, live weather updates, system statistics, and everyday widgets into an intuitive and fluid interactive experience.

---

## Download

Download the latest build for Windows 10/11 (64-bit):

| File                                                                       | Type      | Notes                                                              |
| :------------------------------------------------------------------------- | :-------- | :----------------------------------------------------------------- |
| **[Vinta-Setup.exe](https://github.com/kuraigu/Vinta/releases/latest)**    | Installer | Standard setup with Start Menu and desktop shortcuts (Recommended) |
| **[Vinta-Portable.zip](https://github.com/kuraigu/Vinta/releases/latest)** | Portable  | Just unzip and run `Vinta.exe` from anywhere                       |

Browse previous releases, release notes, and SHA-256 integrity checksums on the official **[GitHub Releases](https://github.com/kuraigu/Vinta/releases)** page.

---

## Origin & Vision

The name **Vinta** comes from the iconic outrigger boats (_vinta_) of Mindanao, Philippines, celebrated for their colorful sails that glide gracefully across open waters.

Inspired by this heritage, Vinta brings that same fluid grace to your desktop interface. It blends clean glassmorphic visuals with responsive interactions, offering a modern desktop companion that stays out of your way until you need it.

---

## Features

Vinta sits at the top edge of the screen and stays out of the way until needed. It slides horizontally with smooth inertia, switches compact modes via scroll wheel or swipe, and can hide completely along the top bezel during focused work.

Inside the island:

- **Media Player**: Automatically surfaces when supported media starts playing, with playback controls, source switching, and an audio mixer drawer.
- **Dual-Tab Launcher & File Shelf**: Launch installed desktop apps or stage files from different folders to drag out, copy, hash (SHA-256), or bundle into a ZIP.
- **Productivity & Study**: Quick notes with checklists, spaced-repetition Flashcards, Daily Habits with break reminders, a scientific calculator, and a timer/focus/alarm suite.
- **Calendar & Holidays**: Agenda view with search, source filters, public holidays, iCal feed subscriptions, meeting links, and optional Google Calendar sync.
- **Developer Hub & Cloud Sync**: Optional GitHub integration (contribution heatmap, streak, pull requests, issues, notifications), DPAPI-encrypted AI provider key tracking, and optional Google/Firebase Cloud Sync.
- **System & Notifications**: Live CPU/RAM/GPU Performance Monitor, incoming Windows toast notifications with app filters and privacy mode, live weather, and customizable themes and locales.

> [!NOTE]
> **PC Gaming Note:** Vinta works best with **Fullscreen** and **Borderless** window modes rather than **Exclusive Fullscreen**. If Vinta disappears when you launch a game, switching the game's display settings to **Fullscreen** or **Borderless** usually fixes it (though a few games block floating windows completely). On multi-monitor setups, if Vinta has to hide because a game doesn't support overlays, it will automatically move to your other display.

---

## Repository Structure

- [backend/](backend/README.md): Go backend services, Wails v3 integration, Win32 window management, and native C++/WinRT bridge.
- [frontend/](frontend/README.md): Svelte 5 application, runes state management, UI components, styling, and themes.
- [build/](build/README.md): Windows packaging assets, icon templates, and application manifest resources.
- [docs/](docs/README.md): Design system guidelines, contributor records, and project documentation.
- [scripts/](scripts/README.md): Build automation scripts, environment setup, and development task helpers.

---

## Documentation & Community

- [Contributing Guide](CONTRIBUTING.md): Instructions for contributing code, testing features, or reporting issues.
- [Translation Guide](docs/TRANSLATIONS.md): Instructions for translating Vinta into other languages and testing locally.
- [Design Principles](docs/DESIGN.md): Guidelines for UI component structure, styling rules, and design tokens.
- [Contributors Hall of Fame](docs/CONTRIBUTORS.md): Recognition for developers, designers, and community members.
- [Security Policy](SECURITY.md): Vulnerability reporting procedures and safety guidelines.
- [Code of Conduct](CODE_OF_CONDUCT.md): Community standards and collaboration guidelines.
- [Support & Help](SUPPORT.md): Community discussion links, issue reporting, and FAQ.

---

## Development Setup

Prerequisites: **Go 1.26+**, **Node.js 20+**, **Wails v3 CLI (wails3)**, **CMake**, and **MSVC C++ Build Tools**.

```powershell
# 1. Clone repository
git clone https://github.com/kuraigu/Vinta.git
cd Vinta

# 2. First-time setup (installs Taskfile, dependencies, and configures native build)
.\scripts.bat setup

# Optional: copy .env.example to .env to configure local Google/GitHub OAuth or Firebase keys

# 3. Daily development with Taskfile (recommended)
task dev      # Run live-reloading development server with DevTest panel
task run      # Compile and launch release binary
task lint     # Check formatting, Go linting, and TypeScript checks
task build    # Build production executable to build/bin/Vinta.exe

task help     # Check all commands
```

> [!TIP]
> **Taskfile Standard:** All daily development tasks should be run using **[go-task / Taskfile](https://taskfile.dev)** (`task dev`, `task run`, `task lint`, `task doctor`). The `.\scripts.bat` wrapper is provided primarily for **first-time setup (`.\scripts.bat setup`)** and as a **fallback** if `task` is not yet installed or not running as expected. Refer to [CONTRIBUTING.md](CONTRIBUTING.md) for full documentation.

---

## License

Vinta is open-source software licensed under the [MIT License](LICENSE).
