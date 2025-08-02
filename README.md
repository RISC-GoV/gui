# RISC-GoV GUI

A graphical user interface for the **RISC-GoV** IDE — a RISC-V assembly development environment written in Go with bindings to Qt via \[therecipe/qt].

## Features

* Qt-based UI built using the [`therecipe/qt`](https://github.com/therecipe/qt) framework
* Project browsing, code editing, assembly, and simulation toolchain integrations
* Context-aware help menu, error reporting, and bug submission
* Fully cross-platform (Windows, macOS, Linux) support via Qt

## Prerequisites

Before building:

* Go 1.20+ installed
* Qt 5 / Qt 6 and the **therecipe/qt** bindings installed
* A working RISC-GoV toolchain for assembling and running RISC-V assembly code

## Building from Source

```bash
# Cloning the repo to your machine
git clone https://github.com/RISC-GoV/gui.git
cd gui

# Download the dependencies
go get .

# Build the GUI application
go build .
```

## Running

```bash
./risc-gov-ide
```

This launches the IDE window.

## Usage

* Create or open `.s` (RISC-V assembly) projects
* Assemble and simulate using your configured toolchain
* View results, errors, or runtime output within the GUI
* Access **Help → Report Bug** to log issues or feature requests

## Bug Reporting 

To report a bug, go to the **Help → Report Bug** menu. It will open your browser to the GitHub issue form:

```
https://github.com/RISC-GoV/gui/issues/new
```

Allowing users to quickly report bugs or request features.

## License

This project is licensed under the **MIT License** — see [LICENSE](LICENSE) for details.
