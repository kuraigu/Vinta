# Vinta

[![Go Version](https://img.shields.io/badge/Go-1.24%2B-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Svelte Version](https://img.shields.io/badge/Svelte-v5-FF3E00?style=flat&logo=svelte)](https://svelte.dev/)
[![TypeScript](https://img.shields.io/badge/TypeScript-v5-3178C6?style=flat&logo=typescript)](https://www.typescriptlang.org/)
[![Wails Framework](https://img.shields.io/badge/Wails-v2-E64A19?style=flat&logo=wails)](https://wails.io/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

Vinta is an open-source, highly customizable Dynamic Island desktop widget for Windows PC.

---

## Origin & Vision

The name **Vinta** originates from the traditional outrigger boats (_vinta_) of Mindanao, Philippines, famous for their vibrant, colorful sails that glide fluidly over water.

Inspired by this maritime heritage, Vinta is designed to be a sleek, vibrant, and floating desktop Dynamic Island for PC—combining minimalist glassmorphic aesthetics, fluid 60fps micro-animations, and native system integration into a customizable desktop experience.

---

## Features

- **Frameless & Transparent:** Seamlessly floats on your desktop with native window clipping and custom backdrop blurs.
- **Component Gallery Carousel:** Interactive slideshow supporting smooth touch and mouse drag gestures.
- **Clock & Timezone Dial:** Circular glass progress ring, steady colon, and automatic timezone detection with GMT offsets.
- **Weather Overview:** Real-time temperature displays, low/high range indicators, metrics bar, and 4-day forecast cards.
- **Calendar & Bento Grid:** Modern date grid views and layout customization.
- **System & Notification Hub:** Integrated notifications, system stats monitoring, and app shortcuts.

---

## Quick Navigation & Documentation

- [Contributing Guide](CONTRIBUTING.md) — How to contribute code, test packages, or report bugs.
- [Design Principles](docs/DESIGN.md) — Theme CSS variable compliance, UI standards, and aesthetics.
- [Contributors Hall of Fame](docs/CONTRIBUTORS.md) — Recognition for our core team, code authors, and QA testers.
- [Security Policy](SECURITY.md) — Vulnerability reporting and security guidelines.
- [Code of Conduct](CODE_OF_CONDUCT.md) — Community standards and behavior expectations.
- [Support & Help](SUPPORT.md) — Community channels, GitHub Discussions, and issue guidelines.

---

## Development Setup

Prerequisites: **Go 1.24+**, **Node.js 20+**, and **Wails v2 CLI**.

```powershell
# 1. Clone repository
git clone https://github.com/kuraigu/Vinta.git
cd Vinta

# 2. Run live-reloading development server
.\scripts.bat dev

# 3. Check linting and formatting
.\scripts.bat lint

# 4. Build release executable
.\scripts.bat build
```

The compiled executable will be placed in `build/bin/Vinta.exe`. See [CONTRIBUTING.md](CONTRIBUTING.md) for full developer details.

---

## License

Vinta is open-source software licensed under the [MIT License](LICENSE).
