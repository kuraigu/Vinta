# Vinta

[![Go Version](https://img.shields.io/badge/Go-1.26%2B-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Svelte Version](https://img.shields.io/badge/Svelte-v5-FF3E00?style=flat&logo=svelte)](https://svelte.dev/)
[![TypeScript](https://img.shields.io/badge/TypeScript-v5-3178C6?style=flat&logo=typescript)](https://www.typescriptlang.org/)
[![Wails Framework](https://img.shields.io/badge/Wails-v3-E64A19?style=flat&logo=wails)](https://wails.io/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

Vinta is an open-source, beautifully designed Dynamic Island widget built specifically for Windows PC. Floating effortlessly at the top of your screen, Vinta turns desktop notifications, live weather updates, system statistics, and everyday widgets into an intuitive and fluid interactive experience.

---

## Origin & Vision

The name **Vinta** comes from the iconic outrigger boats (_vinta_) of Mindanao, Philippines, celebrated for their colorful sails that glide gracefully across open waters.

Inspired by this heritage, Vinta brings that same fluid grace to your desktop interface. It blends clean glassmorphic visuals with responsive interactions, offering a modern desktop companion that stays out of your way until you need it.

---

## Features

- **Floating Island Experience:** Seamlessly floats at the top of your screen with smooth backdrop blurs and transparent glass borders.
- **Interactive Component Gallery:** Swipe or drag through your favorite widgets including weather forecasts, live clocks, system meters, and calendar views.
- **Live Weather & Forecasts:** Real-time local temperatures, atmospheric conditions, daily high and low ranges, and multi-day forecasts.
- **Intelligent Notification Hub:** Captures incoming Windows notifications in a clean pop-up banner with customizable audio chimes and quiet hours.
- **Customizable Appearance & Themes:** Personalize your island with dark and light modes, customizable opacity, ambient glow intensity, and compact width options.
- **Flexible Screen Positioning:** Drag the island horizontally to align with your setup, or double-click anytime to return it to center.

> [!NOTE]
> **PC Gaming Note:** Vinta may not work well with **Exclusive Fullscreen** games, but works smoothly with **Fullscreen** and **Borderless** window modes. If Vinta hides or minimizes when launching a game, switch your game's display settings to **Fullscreen** or **Borderless**.

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

# 3. Daily development with Taskfile (recommended)
task dev      # Run live-reloading development server with DevTest panel
task run      # Compile and launch release binary
task lint     # Check formatting, Go linting, and TypeScript checks
task build    # Build production executable to build/bin/Vinta.exe
```

> [!TIP]
> **Taskfile Standard:** All daily development tasks should be run using **[go-task / Taskfile](https://taskfile.dev)** (`task dev`, `task run`, `task lint`, `task doctor`). The `.\scripts.bat` wrapper is provided primarily for **first-time setup (`.\scripts.bat setup`)** and as a **fallback** if `task` is not yet installed or not running as expected. Refer to [CONTRIBUTING.md](CONTRIBUTING.md) for full documentation.

---

## License

Vinta is open-source software licensed under the [MIT License](LICENSE).
