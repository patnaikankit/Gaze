# Gaze

**Gaze** is a hot-reload tool for Go applications. It watches your source files for changes, automatically rebuilds your app, and restarts the running process. This streamlines your development workflow by providing instant feedback and minimizing manual restarts.

---

## Features

- **Recursive File Watching:** Monitors your project directory and subdirectories for changes.
- **Configurable Ignore Patterns:** Exclude files and folders (e.g., `vendor`, `node_modules`, temp files) from watching.
- **Automatic Build & Run:** Rebuilds your Go application and restarts it on file changes.
- **Cross-Platform Support:** Handles process and port cleanup for both Windows and Unix systems.
- **Configurable via File or CLI:** Use a JSON/YAML config file or command-line flags.
- **Colored Logging:** Clear, leveled logs for build, run, and watcher events.

---

## Getting Started

### 1. Installation

Clone the repo and build the Gaze binary:

```sh
git clone https://github.com/patnaikankit/Gaze.git
cd Gaze
go build -o gaze.exe ./cmd/gaze/main.go
```

Or use the prebuilt binary in `cmd/gaze/gaze.exe`.

### 2. Configuration

Create a `config.json` in `cmd/gaze/` or specify a config file via CLI.

Example `config.json`:

```json
{
  "watchDir": "E:/Projects/Go/test",
  "ignorePattern": [
    "vendor/*",
    "node_modules/*",
    "*.tmp"
  ],
  "buildCmd": "go build -o ./temp/app.exe ./main.go",
  "runCmd": "./temp/app.exe",
  "main": "E:/Projects/Go/test/main.go",
  "port": "",
  "logLevel": "debug"
}
```

### 3. Usage

Run Gaze from the command line:

```sh
./gaze.exe --config ./cmd/gaze/config.json
```

Or use CLI flags:

```sh
./gaze.exe --main ./path/to/main.go --watch ./src --port 8080
```

---

## How It Works

1. **Startup:** Loads configuration, sets up file watcher, builder, and runner.
2. **Initial Build & Run:** Builds your Go app and starts it.
3. **Watching:** Monitors for file changes, ignoring specified patterns.
4. **On Change:** Debounces rapid changes, rebuilds the app, stops the previous process, cleans up ports if needed, and restarts the app.
5. **Logging:** All actions are logged with color and level for clarity.
6. **Shutdown:** Handles OS signals for graceful exit.

---

## Project Structure

```
cmd/
  gaze/
    config.json      # Example config file
    gaze.exe         # Built binary
    main.go          # Entry point
    temp/            # Build output directory
internal/
  builder/           # Build logic
  config/            # Config parsing
  runner/            # Process management
  watcher/           # File watcher
pkg/
  logger/            # Colored logging
```

---

## Advanced

- **Platform Helpers:**  
  - Windows: Uses `taskkill` and `netstat` for process/port cleanup.
  - Unix: Uses `kill` and `lsof`.
- **Web Server Detection:**  
  Warns if your main file looks like a web server but no port is specified.

---

## Contributing

Pull requests and issues are welcome! Please open an issue for bugs or feature requests.

---

## License

MIT

---

## Author

[patnaikankit](https://github.com/patnaikankit)