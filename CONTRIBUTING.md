# Contributing to Vinta

First off, thank you for considering contributing to Vinta! It's people like you that make Vinta such a great tool.

## Development Setup

1. Ensure you have **Go 1.26+** and **Node.js 20+** installed.
2. Install the Wails CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
3. Clone the repository and navigate to the root directory.
4. Run `.\scripts.bat dev` to start the live-reloading development server.

## Code Standards & Linting

Vinta enforces strict code quality standards to ensure cross-platform stability. 
Before submitting a Pull Request, you **must** ensure your code passes the linters.

- Run `.\scripts.bat lint` to automatically check both Go (backend) and Svelte/TS (frontend) code.
- If you have formatting errors in the frontend, run `cd frontend && npm run format` to auto-fix them with Prettier.

> **Note:** Our GitHub Actions pipeline will automatically run these checks on your Pull Request. PRs with linting errors will not be merged.

## Building for Production

To compile a final release binary for your OS, simply run:
`.\scripts.bat build`

The resulting executable will be placed in the `build/bin/` directory.

## Submitting Changes

1. Fork the repository and create your branch from `main`.
2. Make your changes, following the code standards above.
3. Test your changes locally.
4. Issue a Pull Request with a clear description of the problem and your solution.
