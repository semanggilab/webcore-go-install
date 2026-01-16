# WebCore Go Template Installer

A command-line installer that helps you set up a new WebCore Go project from a template repository.

## Features

- **Template Download**: Automatically downloads the WebCore Go template from GitHub
- **Interactive Configuration**: Interactive CLI for easy project setup
- **Module Name Configuration**: Customize your Go module name
- **Library Selection**: Choose which libraries to include in your project
- **Project Modes**: Support for both mono-repo and simple project structures
- **Feature Selection**: Select which features to include (config, database, handler)
- **Automated Setup**: Automatically renames folders, updates imports, and cleans up

## Prerequisites

- Go 1.25.0 or higher
- Git
- Internet connection (for downloading template and libraries)

## Installation

### Build from Source

```bash
cd install
go build -o install
```

The `install` binary will be created in the `install/` directory.

## Usage

Run the installer from the project root directory:

```bash
cd /path/to/your/project
./install/install
```

Or if you're in the install directory:

```bash
./install
```

## Installation Steps

The installer will guide you through the following steps:

### 1. Download Template
The installer will download the WebCore Go template from:
```
https://github.com/semanggilab/webcore-go-template.git
```

### 2. Configure Go Module Name
Enter your Go module name (default: `github.com/semanggilab/project1`)

Example:
```
Enter Go module name: github.com/yourusername/yourproject
```

### 3. Select Libraries
Choose which libraries to include in your project:

Available libraries:
- **database:postgres** - PostgreSQL database support
- **database:mysql** - MySQL database support
- **database:sqlite** - SQLite database support
- **database:mongodb** - MongoDB database support
- **redis** - Redis cache support
- **kafka:producer** - Kafka producer support
- **kafka:consumer** - Kafka consumer support
- **pubsub** - Pub/Sub support
- **authstorage:yaml** - YAML-based authentication storage
- **authentication:apikey** - API key authentication
- **authentication:basic** - Basic authentication

### 4. Choose Project Mode

#### Mono-Repo Mode (Default)
For projects with multiple modules:
- Enter module folder name (e.g., `mymodule`)
- Enter Go module name for this module (e.g., `github.com/yourusername/mymodule`)
- The installer will:
  - Rename `modules/dummy` to `modules/{folder-name}`
  - Update all package names and import paths
  - Update the module's `go.mod` file

#### Simple Mode
For single-module projects:
- The installer will:
  - Move contents from `modules/dummy` to `webcore/app`
  - Update package names to `app`
  - Update import paths

### 5. Select Features
Choose which features to include:

- **specific config** - Include the `config` folder
- **database repository** - Include `service` and `repository` folders
- **http request handler** - Include `handler` folder

Unselected features will have their corresponding folders removed.

### 6. Automatic Configuration
The installer will automatically:
- Update `webcore/go.mod` with your module name
- Update `webcore/deps/libraries.go` with selected libraries
- Run `go get` for each selected library
- Apply project mode configuration
- Update `webcore/deps/packages.go` with the correct module import
- Clean up the `modules/dummy` folder

## Project Structure After Installation

### Mono-Repo Mode
```
/{project-dir}/
├── webcore/
│   └── deps/
│   │   ├── libraries.go
│   │   └── packages.go
│   ├── go.mod
│   ├── main.go
│   └── ...
├── modules/
│   └── {folder-name}/
│       ├── go.mod
│       ├── module.go
│       ├── config/          (if selected)
│       ├── handler/         (if selected)
│       ├── service/         (if selected)
│       ├── repository/      (if selected)
│       └── model/
└── install/
    ├── install.go
    ├── go.mod
    └── README.md
```

### Simple Mode
```
/{project-dir}/
├── webcore/
│   ├── go.mod
│   ├── main.go
│   └── deps/
│   │   ├── libraries.go
│   │   └── packages.go
│   ├── app/               (moved from modules/dummy)
│   │   ├── module.go
│   │   ├── config/        (if selected)
│   │   ├── handler/       (if selected)
│   │   ├── service/       (if selected)
│   │   ├── repository/    (if selected)
│   │   └── model/
│   └── ...
└── install/
    ├── install.go
    ├── go.mod
    └── README.md
```

## Running Your Project

After installation, you can run your project:

```bash
cd {project-dir}
make sync 
make run
```

## Troubleshooting

### Git Clone Fails
Make sure you have Git installed and have internet access. The installer uses `git clone --depth 1` to download the template.

### Go Get Fails
Make sure you have Go installed and configured properly. The installer runs `go get` in the `webcore` directory.

### Folder Already Exists
If the `webcore` directory already exists, the installer will skip the download step and use the existing directory.

### Module Name Format
The installer validates module names but will continue even if the format is non-standard. A standard Go module name should contain at least one slash and no spaces.

## Development

To rebuild the installer after making changes:

```bash
cd install
go build -o install
```

## License

This installer is part of the WebCore Go Template project.

## Support

For issues or questions, please refer to the main project repository.
