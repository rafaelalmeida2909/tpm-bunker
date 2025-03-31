# TPM Bunker

This is the official Wails Svelte template for the **TPM Bunker** project. It combines the power of Go (backend) and Svelte (frontend) to deliver a modern desktop application.

## Table of Contents

- [About](#about)
- [Requirements](#requirements)
- [Live Development](#live-development)
- [Building](#building)
- [Project Structure](#project-structure)
- [Installation](#installation)
- [Usage](#usage)
- [Contributing](#contributing)
- [License](#license)

## About

The **TPM Bunker** project is built using the [Wails](https://wails.io/) framework, which allows for creating lightweight desktop applications with Go and modern frontend frameworks like Svelte. This template provides a starting point for building cross-platform applications with a fast development workflow.

## Requirements

### System Requirements

| Component | Version Required |
|-----------|-----------------|
| Python | 3.12.0 |
| Go | 1.24.0 (linux/amd64) |
| Node.js | 22.14.0 |
| npm | 10.9.2 |
| TPM 2.0 | Hardware Module |

## Live Development

To run the project in live development mode:

1. Open a terminal in the project directory
2. Run the development server:

    ```bash
    wails dev
    ```

3. Access the development interface:
   - Local development server: [http://localhost:34115](http://localhost:34115)
   - Use browser devtools to interact with Go code

## Building

Create a production-ready build:

```bash
wails build
```

## Project Structure

```
tpm-bunker/
├── frontend/           # Svelte frontend
│   ├── src/           # Source files
│   ├── public/        # Static assets
│   └── package.json   # Dependencies
├── internal/          # Internal Go packages
├── main.go           # Application entry point
└── wails.json        # Wails configuration
```

## Installation

1. **Clone the repository**

    ```bash
    git clone https://github.com/rafaelalmeida2909/tpm-bunker.git
    cd tpm-bunker
    ```

2. **Install Wails CLI**

    ```bash
    go install github.com/wailsapp/wails/v2/cmd/wails@latest
    ```

3. **Install frontend dependencies**

    ```bash
    cd frontend
    npm install
    ```

## Usage

### Development Mode

```bash
wails dev
```

### Production Build

```bash
wails build
./build/bin/tpm-bunker
```

## Contributing

1. Fork the repository
2. Create your feature branch:
    ```bash
    git checkout -b feature/amazing-feature
    ```
3. Commit your changes:
    ```bash
    git commit -m 'Add amazing feature'
    ```
4. Push to the branch:
    ```bash
    git push origin feature/amazing-feature
    ```
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

**Note:** Ensure your TPM 2.0 hardware module is properly configured in your system's BIOS/UEFI settings before running the application.
