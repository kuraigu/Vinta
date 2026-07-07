# Contributing to Vinta

First off, thank you for considering contributing to Vinta! It's people like you that make Vinta such a great tool.

## Project Structure

Vinta is built using the [Wails](https://wails.io/) framework, which combines a Go backend with a modern web frontend.

- `main.go`: The application entry point and Wails configuration.
- `backend/`: Contains all Go backend logic, including native Windows API calls (`app.go`), system tray management, and OS-level window clipping.
- `frontend/`: The Svelte/TypeScript frontend containing all UI components, animations, and CSS layouts. This is where the actual "island" is rendered.
- `scripts.bat`: A unified batch script for running development commands (`dev`, `build`, `lint`).

## Development Setup

1. Ensure you have **Go 1.24+** and **Node.js 20+** installed.
2. Install the Wails CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
3. Clone the repository and navigate to the root directory.
4. Run `.\scripts.bat dev` to start the live-reloading development server.

## Code Standards & Linting

Vinta enforces strict code quality standards to ensure cross-platform stability. 
Before submitting a Pull Request, you **must** ensure your code passes the linters.

## Building and Running

You **must** always run the lint script before attempting to run or build the project. The CI will fail your PR if there are linting errors.

1. First, check your code for formatting and lint errors: `.\scripts.bat lint`
2. If you have formatting errors in the frontend, run `cd frontend && npm run format` to auto-fix them with Prettier.
3. Once linting passes, run the development server: `.\scripts.bat run`
4. To compile a final release binary for your OS, run: `.\scripts.bat build`

The resulting executable will be placed in the `build/bin/` directory.

## Submitting Changes

1. Fork the repository and create your branch from `main`.
2. Make your changes, following the code standards above.
3. Test your changes locally.
4. Issue a Pull Request with a clear description of the problem and your solution.
