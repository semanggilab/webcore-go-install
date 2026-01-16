package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/yarlson/tap"
)

const (
	templateRepoURL   = "https://github.com/semanggilab/webcore-go-template.git"
	defaultModuleName = "github.com/semanggilab/project1"
	defaultProjectDir = "./webcore"
)

// LibraryOption represents a library option for selection
type LibraryOption struct {
	Name        string
	Description string
	PackagePath string
	LoaderName  string
	Enabled     bool
}

// Available libraries based on webcore/deps/libraries.go
var availableLibraries = []LibraryOption{
	{
		Name:        "database:postgres",
		Description: "PostgreSQL",
		PackagePath: "github.com/webcore-go/lib-postgres",
		Enabled:     true,
	},
	{
		Name:        "database:mysql",
		Description: "MySQL",
		PackagePath: "github.com/webcore-go/lib-mysql",
		Enabled:     false,
	},
	{
		Name:        "database:sqlite",
		Description: "SQLite",
		PackagePath: "github.com/webcore-go/lib-mysql",
		Enabled:     false,
	},
	{
		Name:        "database:mongodb",
		Description: "MongoDB",
		PackagePath: "github.com/webcore-go/lib-mongo",
		Enabled:     false,
	},
	{
		Name:        "redis",
		Description: "Redis",
		PackagePath: "github.com/webcore-go/lib-redis",
		Enabled:     false,
	},
	{
		Name:        "kafka:producer",
		Description: "Kafka Producer",
		PackagePath: "github.com/webcore-go/lib-kafka",
		LoaderName:  "KafkaProducerLoader",
		Enabled:     false,
	},
	{
		Name:        "kafka:consumer",
		Description: "Kafka Consumer",
		PackagePath: "github.com/webcore-go/lib-kafka",
		LoaderName:  "KafkaConsumerLoader",
		Enabled:     false,
	},
	{
		Name:        "pubsub",
		Description: "Google Pub/Sub",
		PackagePath: "github.com/webcore-go/lib-pubsub",
		LoaderName:  "PubSubLoader",
		Enabled:     false,
	},
	{
		Name:        "authstorage:yaml",
		Description: "Authentication Storage: YAML",
		PackagePath: "github.com/webcore-go/webcore/lib/authstore/yaml",
		Enabled:     true,
	},
	{
		Name:        "authentication:apikey",
		Description: "Authentication: API key",
		PackagePath: "github.com/webcore-go/webcore/lib/auth/apikey",
		LoaderName:  "ApiKeyLoader",
		Enabled:     true,
	},
	{
		Name:        "authentication:basic",
		Description: "Authentication: Basic",
		PackagePath: "github.com/webcore-go/webcore/lib/auth/basic",
		LoaderName:  "BasicAuthLoader",
		Enabled:     false,
	},
}

// Feature represents a feature option
type Feature struct {
	Name        string
	Description string
	Enabled     bool
}

var availableFeatures = []Feature{
	{
		Name:        "specific config",
		Description: "Additional Config",
		Enabled:     true,
	},
	{
		Name:        "database repository",
		Description: "Service and Repository",
		Enabled:     true,
	},
	{
		Name:        "http request handler",
		Description: "HTTP Request Handler",
		Enabled:     true,
	},
}

// Config holds the installation configuration
type Config struct {
	ProjectDir        string
	ModuleName        string
	SelectedLibraries []LibraryOption
	ProjectMode       string // "simple" or "mono-repo"
	FolderName        string // for mono-repo mode
	ModuleModName     string // for mono-repo mode
	SelectedFeatures  []Feature
	GitInit           bool // whether to initialize git
}

func main() {
	ctx := context.Background()

	tap.Intro("WebCore Go Template Installer")
	tap.Message("This installer will help you set up a new WebCore Go project")

	config := &Config{}

	// Step 1: Ask for project directory
	config.ProjectDir = askProjectDir(ctx)

	// Step 2: Download template
	if err := downloadTemplate(config.ProjectDir); err != nil {
		tap.Outro(fmt.Sprintf("❌ Failed to download template: %v\n", err))
		os.Exit(1)
	}

	// Step 2: Ask for module name
	config.ModuleName = askModuleName(ctx)

	// Step 3: Select libraries
	config.SelectedLibraries = selectLibraries(ctx)

	// Step 4: Select project mode
	config.ProjectMode = selectProjectMode(ctx)

	// Step 5: Handle project mode specific configuration
	if config.ProjectMode == "mono-repo" {
		config.FolderName = askFolderName(ctx)
		config.ModuleModName = askModuleModName(ctx, config.ModuleName, config.FolderName)
	}

	// Step 6: Select features
	config.SelectedFeatures = selectFeatures(ctx)

	// Step 7: Ask about git initialization
	config.GitInit = askGitInit(ctx)

	// Step 8: Apply configuration
	if err := applyConfiguration(ctx, config); err != nil {
		tap.Outro(fmt.Sprintf("❌ Failed to apply configuration: %v\n", err))
		os.Exit(1)
	}

	tap.Outro(fmt.Sprintf("✅ Installation completed successfully!\nYou can now run your project with: cd %s && make run", config.ProjectDir))
}

// downloadTemplate downloads the template from GitHub
func downloadTemplate(projectDir string) error {
	sp := tap.NewSpinner(tap.SpinnerOptions{Indicator: "dots"})
	sp.Start("Downloading template from GitHub...")

	// Check if project directory already exists
	if _, err := os.Stat(fmt.Sprintf("%s/webcore/go.mod", projectDir)); err == nil {
		sp.Stop(fmt.Sprintf("⚠️ Project already initialized in %s directory, skipping download", projectDir), 0)
		return nil
	}

	cmd := exec.Command("git", "clone", "--depth", "1", templateRepoURL, projectDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		sp.Stop("❌ Failed to download template", 1)
		return fmt.Errorf("git clone failed: %w", err)
	}

	// Remove .git directory from cloned repo
	gitDir := filepath.Join(projectDir, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		os.RemoveAll(gitDir)
	}

	sp.Stop("✅ Template downloaded successfully", 0)
	return nil
}

// askProjectDir asks for the project directory
func askProjectDir(ctx context.Context) string {
	projectDir := tap.Text(ctx, tap.TextOptions{
		Message:      "Enter project directory",
		Placeholder:  defaultProjectDir,
		InitialValue: defaultProjectDir,
	})

	// Clean up the directory path (remove trailing slashes)
	projectDir = strings.TrimSuffix(projectDir, "/")
	projectDir = strings.TrimSuffix(projectDir, "\\")

	tap.Message(fmt.Sprintf("✅ Project directory: %s\n", projectDir))
	return projectDir
}

// askModuleName asks for the Go module name
func askModuleName(ctx context.Context) string {
	moduleName := tap.Text(ctx, tap.TextOptions{
		Message:      "Enter Go module name",
		Placeholder:  defaultModuleName,
		InitialValue: defaultModuleName,
	})

	// Validate module name format
	if !isValidModuleName(moduleName) {
		tap.Message("⚠️ Module name format is not standard, but continuing anyway")
	}

	tap.Message(fmt.Sprintf("✅ Module name set to: %s\n", moduleName))
	return moduleName
}

// isValidModuleName checks if the module name follows Go module naming conventions
func isValidModuleName(name string) bool {
	// Basic validation: should contain at least one slash and no spaces
	return strings.Contains(name, "/") && !strings.Contains(name, " ")
}

// selectLibraries displays library selection options
func selectLibraries(ctx context.Context) []LibraryOption {
	// Create options for MultiSelect
	options := make([]tap.SelectOption[string], len(availableLibraries))
	defaultValues := make([]string, 0)

	for i, lib := range availableLibraries {
		options[i] = tap.SelectOption[string]{
			Value: lib.Name,
			Label: fmt.Sprintf("%s", lib.Description),
		}
		if lib.Enabled {
			defaultValues = append(defaultValues, lib.Name)
		}
	}

	// Use MultiSelect for multiple choices
	selectedNames := tap.MultiSelect(ctx, tap.MultiSelectOptions[string]{
		Message:       "Select libraries to include in your project",
		Options:       options,
		InitialValues: defaultValues,
	})

	// Map selected names back to LibraryOption
	selected := make([]LibraryOption, 0, len(selectedNames))
	selectedMap := make(map[string]bool)
	for _, name := range selectedNames {
		selectedMap[name] = true
	}

	selectedStrings := make([]string, 0)
	for _, lib := range availableLibraries {
		if selectedMap[lib.Name] {
			selected = append(selected, lib)
			selectedStrings = append(selectedStrings, lib.Name)
		}
	}

	tap.Message(fmt.Sprintf("✅ Selected: %v\n", selectedStrings))
	return selected
}

// selectProjectMode asks for the project mode
func selectProjectMode(ctx context.Context) string {
	mode := tap.Select(ctx, tap.SelectOptions[string]{
		Message: "Choose project type",
		Options: []tap.SelectOption[string]{
			{Label: "Mono-repo (multiple modules)", Value: "mono-repo"},
			{Label: "Simple (single module)", Value: "simple"},
		},
		InitialValue: &[]string{"mono-repo"}[0],
	})

	tap.Message(fmt.Sprintf("✅ Project type: %s\n", mode))
	return mode
}

// askFolderName asks for the folder name in mono-repo mode
func askFolderName(ctx context.Context) string {
	for {
		folderName := tap.Text(ctx, tap.TextOptions{
			Message:      "Enter module folder name",
			Placeholder:  "mymodule",
			InitialValue: "mymodule",
		})

		// Validate folder name
		if !isValidFolderName(folderName) {
			tap.Message("❌ Invalid folder name. Use only lowercase letters, numbers, and hyphens")
			continue
		}

		tap.Message(fmt.Sprintf("✅ Folder name: %s\n", folderName))
		return folderName
	}
}

// isValidFolderName validates folder name
func isValidFolderName(name string) bool {
	// Only allow lowercase letters, numbers, and hyphens
	matched, _ := regexp.MatchString(`^[a-z0-9-]+$`, name)
	return matched && len(name) > 0
}

// askModuleModName asks for the Go module name in mono-repo mode
func askModuleModName(ctx context.Context, projectModuleName string, folderName string) string {
	defaultModName := fmt.Sprintf("%s-mod-%s", projectModuleName, folderName)
	moduleName := tap.Text(ctx, tap.TextOptions{
		Message:      "Enter Go module name for this module",
		Placeholder:  defaultModName,
		InitialValue: defaultModName,
	})

	tap.Message(fmt.Sprintf("✅ Module name: %s\n", moduleName))
	return moduleName
}

// selectFeatures displays feature selection options
func selectFeatures(ctx context.Context) []Feature {
	// Create options for MultiSelect
	options := make([]tap.SelectOption[string], len(availableFeatures))
	defaultValues := make([]string, 0)

	for i, feature := range availableFeatures {
		options[i] = tap.SelectOption[string]{
			Value: feature.Name,
			Label: fmt.Sprintf("%s", feature.Description),
		}
		if feature.Enabled {
			defaultValues = append(defaultValues, feature.Name)
		}
	}

	// Use MultiSelect for multiple choices
	selectedNames := tap.MultiSelect(ctx, tap.MultiSelectOptions[string]{
		Message:       "Select features to include in your module",
		Options:       options,
		InitialValues: defaultValues,
	})

	// Map selected names back to Feature
	selected := make([]Feature, 0, len(selectedNames))
	selectedMap := make(map[string]bool)
	for _, name := range selectedNames {
		selectedMap[name] = true
	}

	selectedStrings := make([]string, 0)
	for _, feature := range availableFeatures {
		if selectedMap[feature.Name] {
			selected = append(selected, feature)
			selectedStrings = append(selectedStrings, feature.Name)
		}
	}

	tap.Message(fmt.Sprintf("✅ Selected: %v\n", selectedStrings))
	return selected
}

// askGitInit asks if user wants to initialize git
func askGitInit(ctx context.Context) bool {
	gitInit := tap.Confirm(ctx, tap.ConfirmOptions{
		Message:      "Initialize git repository?",
		InitialValue: false,
	})

	if gitInit {
		fmt.Println("✅ Git initialization enabled")
	} else {
		fmt.Println("⏭️ Git initialization disabled")
	}

	return gitInit
}

// applyConfiguration applies all the configuration changes
func applyConfiguration(ctx context.Context, config *Config) error {
	// Step 0. Replace module name in webcore/main.go
	mainGoPath := filepath.Join(config.ProjectDir, "webcore", "main.go")
	if err := replaceInFile(mainGoPath, "github.com/semanggilab/webcorego-template-app", config.ModuleName); err != nil {
		return fmt.Errorf("failed to update webcore/main.go: %w", err)
	}

	// Step 1: Update webcore/go.mod with main module name
	if err := updateWebcoreGoMod(config.ProjectDir, config.ModuleName); err != nil {
		return fmt.Errorf("failed to update webcore/go.mod: %w", err)
	}

	// Step 2: Update webcore/deps/libraries.go with selected libraries
	if err := updateLibrariesGo(config.ProjectDir, config.SelectedLibraries); err != nil {
		return fmt.Errorf("failed to update libraries.go: %w", err)
	}

	// Step 3: Install selected libraries
	if err := installLibraries(config.ProjectDir, config.SelectedLibraries); err != nil {
		return fmt.Errorf("failed to install libraries: %w", err)
	}

	// Step 4: Copy example config files
	if err := copyConfigFiles(config); err != nil {
		return fmt.Errorf("failed to copy config files: %w", err)
	}

	// Step 5: Handle project mode
	if config.ProjectMode == "mono-repo" {
		if err := applyMonoRepoMode(config); err != nil {
			return fmt.Errorf("failed to apply mono-repo mode: %w", err)
		}
	} else {
		if err := applySimpleMode(config); err != nil {
			return fmt.Errorf("failed to apply simple mode: %w", err)
		}
	}

	// Step 5: Update webcore/deps/packages.go
	if err := updatePackagesGo(config); err != nil {
		return fmt.Errorf("failed to update packages.go: %w", err)
	}

	// Step 6: Cleanup dummy folder if it still exists
	if err := cleanupDummyFolder(config.ProjectDir); err != nil {
		return fmt.Errorf("failed to cleanup dummy folder: %w", err)
	}

	// Step 7: Update go.work file
	if err := updateGoWork(config); err != nil {
		return fmt.Errorf("failed to update go.work: %w", err)
	}

	// Step 8: Initialize git if requested
	if config.GitInit {
		if err := initGit(ctx, config.ProjectDir); err != nil {
			return fmt.Errorf("failed to initialize git: %w", err)
		}
	}

	return nil
}

// updateWebcoreGoMod updates the module name in webcore/go.mod
func updateWebcoreGoMod(projectDir, moduleName string) error {
	tap.Message("📝 Updating webcore/go.mod...")

	goModPath := filepath.Join(projectDir, "webcore/go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "module ") {
			lines[i] = fmt.Sprintf("module %s", moduleName)
			break
		}
	}

	newContent := strings.Join(lines, "\n")
	if err := os.WriteFile(goModPath, []byte(newContent), 0644); err != nil {
		return err
	}

	tap.Message(fmt.Sprintf("✅ Updated module name to: %s\n", moduleName))
	return nil
}

// updateLibrariesGo creates a new webcore/deps/libraries.go file with selected libraries
func updateLibrariesGo(projectDir string, libraries []LibraryOption) error {
	tap.Message("📝 Updating webcore/deps/libraries.go...")

	libPath := filepath.Join(projectDir, "webcore", "deps", "libraries.go")

	var buf bytes.Buffer
	buf.WriteString("package deps\n\n")
	buf.WriteString("import (\n")

	// Add required imports
	buf.WriteString("\t\"github.com/webcore-go/webcore/app/core\"\n")

	// Add imports for selected libraries
	imports := make(map[string]string)
	for _, lib := range libraries {
		imports[lib.PackagePath] = getImportAlias(lib.PackagePath)
	}

	// Write imports
	for pkg, alias := range imports {
		buf.WriteString(fmt.Sprintf("\t%s \"%s\"\n", alias, pkg))
	}
	buf.WriteString(")\n\n")

	// Write APP_LIBRARIES map
	buf.WriteString("var APP_LIBRARIES = map[string]core.LibraryLoader{\n")
	for _, lib := range libraries {
		alias := getImportAlias(lib.PackagePath)
		loaderName := getLoaderName(lib)
		buf.WriteString(fmt.Sprintf("\t\"%s\": &%s.%s{},\n", lib.Name, alias, loaderName))
	}
	buf.WriteString("\n\t// Add your library here\n}\n")

	if err := os.WriteFile(libPath, buf.Bytes(), 0644); err != nil {
		return err
	}

	tap.Message("✅ Updated webcore/deps/libraries.go with selected libraries")
	return nil
}

// getImportAlias extracts the import alias from package path
func getImportAlias(pkgPath string) string {
	parts := strings.Split(pkgPath, "/")
	return strings.ReplaceAll(strings.ReplaceAll(parts[len(parts)-1], "lib-", ""), "-", "_")
}

// getLoaderName generates the loader struct name from library name
func getLoaderName(lib LibraryOption) string {
	if lib.LoaderName != "" {
		return lib.LoaderName
	}

	parts := strings.Split(lib.Name, ":")
	name := parts[len(parts)-1]
	return strings.Title(name) + "Loader"
}

// installLibraries runs go get for each selected library
func installLibraries(projectDir string, libraries []LibraryOption) error {
	sp := tap.NewSpinner(tap.SpinnerOptions{Indicator: "dots"})
	sp.Start("Installing selected libraries...")

	webcoreDir := filepath.Join(projectDir, "webcore")
	for _, lib := range libraries {
		sp.Start(fmt.Sprintf("Installing: %s", lib.PackagePath))

		cmd := exec.Command("go", "get", lib.PackagePath)
		cmd.Dir = webcoreDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			tap.Message(fmt.Sprintf("⚠️ Failed to install %s: %v\n", lib.PackagePath, err))
			continue
		}
	}

	sp.Stop("✅ Libraries installed", 0)
	return nil
}

// copyConfigFiles copies example config files to project directory
func copyConfigFiles(config *Config) error {
	sp := tap.NewSpinner(tap.SpinnerOptions{Indicator: "dots"})
	sp.Start("Copying example config files...")

	// Copy config.yaml.example to config.yaml
	configSrc := filepath.Join(config.ProjectDir, "config.yaml.example")
	configDst := filepath.Join(config.ProjectDir, "config.yaml")
	if err := copyFile(configSrc, configDst); err != nil {
		sp.Stop("❌ Failed to copy config.yaml", 1)
		return fmt.Errorf("failed to copy config.yaml: %w", err)
	}

	// Comment out sections based on selected libraries
	if err := commentConfigSections(configDst, config.SelectedLibraries); err != nil {
		sp.Stop("❌ Failed to update config.yaml", 1)
		return fmt.Errorf("failed to update config.yaml: %w", err)
	}

	// Copy access.yaml.example to access.yaml
	accessSrc := filepath.Join(config.ProjectDir, "access.yaml.example")
	accessDst := filepath.Join(config.ProjectDir, "access.yaml")
	if err := copyFile(accessSrc, accessDst); err != nil {
		sp.Stop("❌ Failed to copy access.yaml", 1)
		return fmt.Errorf("failed to copy access.yaml: %w", err)
	}

	sp.Stop("✅ Example config files copied", 0)
	return nil
}

// copyFile copies a file from source to destination
func copyFile(src, dst string) error {
	content, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, content, 0644)
}

// commentConfigSections comments out sections in config.yaml based on selected libraries
func commentConfigSections(configPath string, libraries []LibraryOption) error {
	content, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")

	// Check which libraries are selected
	hasDatabase := hasLibraryPrefix(libraries, "database:")
	hasRedis := hasLibraryName(libraries, "redis")
	hasPubsub := hasLibraryName(libraries, "pubsub")

	// Comment out sections if corresponding library is not selected
	lines = commentSection(lines, 71, 92, !hasDatabase) // database section
	lines = commentSection(lines, 94, 104, !hasRedis)   // redis section
	lines = commentSection(lines, 107, 114, !hasPubsub) // pubsub section

	newContent := strings.Join(lines, "\n")
	return os.WriteFile(configPath, []byte(newContent), 0644)
}

// hasLibraryPrefix checks if any library has the given prefix
func hasLibraryPrefix(libraries []LibraryOption, prefix string) bool {
	for _, lib := range libraries {
		if strings.HasPrefix(lib.Name, prefix) {
			return true
		}
	}
	return false
}

// hasLibraryName checks if any library has the exact name
func hasLibraryName(libraries []LibraryOption, name string) bool {
	for _, lib := range libraries {
		if lib.Name == name {
			return true
		}
	}
	return false
}

// commentSection comments out lines from start to end (1-indexed, inclusive) if shouldComment is true
func commentSection(lines []string, start, end int, shouldComment bool) []string {
	if !shouldComment {
		return lines
	}

	// Convert to 0-indexed
	startIdx := start - 1
	endIdx := end - 1

	// Validate indices
	if startIdx < 0 || endIdx >= len(lines) || startIdx > endIdx {
		return lines
	}

	// Comment out each line
	for i := startIdx; i <= endIdx; i++ {
		trimmed := strings.TrimSpace(lines[i])
		// Only comment if not already commented and not empty
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			lines[i] = "# " + lines[i]
		}
	}

	return lines
}

// applyMonoRepoMode applies mono-repo mode configuration
func applyMonoRepoMode(config *Config) error {
	sp := tap.NewSpinner(tap.SpinnerOptions{Indicator: "dots"})
	sp.Start("Applying mono-repo mode...")

	dummyPath := filepath.Join(config.ProjectDir, "modules", "dummy")
	newPath := filepath.Join(config.ProjectDir, "modules", config.FolderName)

	// Rename dummy folder to new folder name
	if err := os.Rename(dummyPath, newPath); err != nil {
		sp.Stop("❌ Failed to rename folder", 1)
		return fmt.Errorf("failed to rename folder: %w", err)
	}

	tap.Message(fmt.Sprintf("✅ Renamed modules/dummy to modules/%s\n", config.FolderName))

	// Replace module name in go.mod
	goModPath := filepath.Join(newPath, "go.mod")
	if err := replaceInFile(goModPath, "github.com/semanggilab/webcorego-template-mod", config.ModuleModName); err != nil {
		sp.Stop("❌ Failed to update module go.mod", 1)
		return fmt.Errorf("failed to update module go.mod: %w", err)
	}

	// Replace package name and import paths in all Go files
	if err := replacePackageNames(newPath, "dummy", config.FolderName, "github.com/semanggilab/webcorego-template-mod", config.ModuleModName); err != nil {
		sp.Stop("❌ Failed to update package names", 1)
		return fmt.Errorf("failed to update package names: %w", err)
	}

	// Handle feature-based folder inclusion/exclusion
	if err := handleFeatureFolders(newPath, config.SelectedFeatures); err != nil {
		sp.Stop("❌ Failed to handle feature folders", 1)
		return fmt.Errorf("failed to handle feature folders: %w", err)
	}

	sp.Stop("✅ Mono-repo mode applied successfully", 0)
	return nil
}

// applySimpleMode applies simple mode configuration
func applySimpleMode(config *Config) error {
	sp := tap.NewSpinner(tap.SpinnerOptions{Indicator: "dots"})
	sp.Start("Applying simple mode...")

	dummyPath := filepath.Join(config.ProjectDir, "modules", "dummy")
	appPath := filepath.Join(config.ProjectDir, "webcore", "app")

	// Create app directory if it doesn't exist
	if err := os.MkdirAll(appPath, 0755); err != nil {
		sp.Stop("❌ Failed to create app directory", 1)
		return fmt.Errorf("failed to create app directory: %w", err)
	}

	// Move contents from dummy to app (except go.mod and go.sum)
	entries, err := os.ReadDir(dummyPath)
	if err != nil {
		sp.Stop("❌ Failed to read dummy directory", 1)
		return fmt.Errorf("failed to read dummy directory: %w", err)
	}

	for _, entry := range entries {
		if entry.Name() == "go.mod" || entry.Name() == "go.sum" {
			continue
		}

		srcPath := filepath.Join(dummyPath, entry.Name())
		dstPath := filepath.Join(appPath, entry.Name())

		if err := os.Rename(srcPath, dstPath); err != nil {
			sp.Stop("❌ Failed to move files", 1)
			return fmt.Errorf("failed to move %s: %w", entry.Name(), err)
		}
	}

	// Replace package name with "app" in all Go files
	if err := replacePackageNames(appPath, "dummy", "app", "github.com/semanggilab/webcorego-template-mod", config.ModuleName+"/app"); err != nil {
		sp.Stop("❌ Failed to update package names", 1)
		return fmt.Errorf("failed to update package names: %w", err)
	}

	// Handle feature-based folder inclusion/exclusion
	if err := handleFeatureFolders(appPath, config.SelectedFeatures); err != nil {
		sp.Stop("❌ Failed to handle feature folders", 1)
		return fmt.Errorf("failed to handle feature folders: %w", err)
	}

	sp.Stop("✅ Simple mode applied successfully", 0)
	return nil
}

// replacePackageNames replaces package names and import paths in Go files
func replacePackageNames(dir, oldPkg, newPkg, oldModule, newModule string) error {
	tap.Message(fmt.Sprintf("📝 Updating package names from %s to %s...\n", oldPkg, newPkg))

	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		fileContent := string(content)

		// Replace package declaration
		fileContent = regexp.MustCompile(`(?m)^package `+regexp.QuoteMeta(oldPkg)+`$`).ReplaceAllString(fileContent, "package "+newPkg)

		if strings.HasSuffix(path, "module.go") {
			fileContent = regexp.MustCompile(`(?m)^\s+ModuleName\s*=\s*"`+regexp.QuoteMeta(oldPkg)+`"$`).ReplaceAllString(fileContent, "\tModuleName    = \""+newPkg+"\"")
		}

		// Replace import paths
		fileContent = strings.ReplaceAll(fileContent, oldModule, newModule)

		if err := os.WriteFile(path, []byte(fileContent), 0644); err != nil {
			return err
		}

		return nil
	})
}

// replaceInFile replaces all occurrences of oldString with newString in a file
func replaceInFile(filePath, oldString, newString string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	fileContent := strings.ReplaceAll(string(content), oldString, newString)

	return os.WriteFile(filePath, []byte(fileContent), 0644)
}

// handleFeatureFolders includes or excludes folders based on selected features
func handleFeatureFolders(modulePath string, features []Feature) error {
	tap.Message("📝 Handling feature-based folders...")

	// Check which features are selected
	hasConfig := false
	hasDatabase := false
	hasHandler := false

	for _, feature := range features {
		switch feature.Name {
		case "specific config":
			hasConfig = true
		case "database repository":
			hasDatabase = true
		case "http request handler":
			hasHandler = true
		}
	}

	// Remove folders for unselected features
	if !hasConfig {
		configPath := filepath.Join(modulePath, "config")
		if _, err := os.Stat(configPath); err == nil {
			if err := os.RemoveAll(configPath); err != nil {
				return fmt.Errorf("failed to remove config folder: %w", err)
			}
			tap.Message("⏭️ Removed config folder (specific config not selected)")
		}
	}

	if !hasDatabase {
		servicePath := filepath.Join(modulePath, "service")
		repositoryPath := filepath.Join(modulePath, "repository")

		if _, err := os.Stat(servicePath); err == nil {
			if err := os.RemoveAll(servicePath); err != nil {
				return fmt.Errorf("failed to remove service folder: %w", err)
			}
			tap.Message("⏭️ Removed service folder (database repository not selected)")
		}

		if _, err := os.Stat(repositoryPath); err == nil {
			if err := os.RemoveAll(repositoryPath); err != nil {
				return fmt.Errorf("failed to remove repository folder: %w", err)
			}
			tap.Message("⏭️ Removed repository folder (database repository not selected)")
		}
	}

	if !hasHandler {
		handlerPath := filepath.Join(modulePath, "handler")
		if _, err := os.Stat(handlerPath); err == nil {
			if err := os.RemoveAll(handlerPath); err != nil {
				return fmt.Errorf("failed to remove handler folder: %w", err)
			}
			tap.Message("⏭️ Removed handler folder (http request handler not selected)")
		}
	}

	tap.Message("✅ Feature-based folders updated")
	return nil
}

// updatePackagesGo updates webcore/deps/packages.go with the correct module import
func updatePackagesGo(config *Config) error {
	tap.Message("📝 Updating webcore/deps/packages.go...")

	packagesPath := filepath.Join(config.ProjectDir, "webcore", "deps", "packages.go")
	content, err := os.ReadFile(packagesPath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	var importLine string
	var moduleCall string

	if config.ProjectMode == "mono-repo" {
		importLine = fmt.Sprintf("\t%s \"%s\"", config.FolderName, config.ModuleModName)
		moduleCall = fmt.Sprintf("\t%s.NewModule(),", config.FolderName)
	} else {
		importLine = "\tapp \"github.com/semanggilab/project1/app\""
		if config.ModuleName != defaultModuleName {
			importLine = fmt.Sprintf("\tapp \"%s/app\"", config.ModuleName)
		}
		moduleCall = "\tapp.NewModule(),"
	}

	// Update import line
	for i, line := range lines {
		if strings.Contains(line, "github.com/semanggilab/webcorego-template-mod") {
			lines[i] = importLine
			break
		}
	}

	// Update module call
	for i, line := range lines {
		if strings.Contains(line, "dummy.NewModule()") {
			lines[i] = moduleCall
			break
		}
	}

	newContent := strings.Join(lines, "\n")
	if err := os.WriteFile(packagesPath, []byte(newContent), 0644); err != nil {
		return err
	}

	tap.Message("✅ Updated webcore/deps/packages.go")
	return nil
}

// cleanupDummyFolder removes the dummy folder after all operations
func cleanupDummyFolder(projectDir string) error {
	dummyPath := filepath.Join(projectDir, "modules", "dummy")
	if _, err := os.Stat(dummyPath); err == nil {
		if err := os.RemoveAll(dummyPath); err != nil {
			return fmt.Errorf("failed to remove dummy folder: %w", err)
		}
		tap.Message("✅ Removed modules/dummy folder")
	}
	return nil
}

// updateGoWork updates the go.work file in the project directory
func updateGoWork(config *Config) error {
	// Only update go.work for mono-repo mode
	if config.ProjectMode != "mono-repo" {
		return nil
	}

	tap.Message("📝 Updating go.work file...")

	goWorkPath := filepath.Join(config.ProjectDir, "go.work")
	content, err := os.ReadFile(goWorkPath)
	if err != nil {
		return fmt.Errorf("failed to read go.work: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	updated := false

	// Find and replace ./modules/dummy with ./modules/<folder name>
	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "./modules/dummy" {
			lines[i] = fmt.Sprintf("\t./modules/%s", config.FolderName)
			updated = true
			tap.Message(fmt.Sprintf("✅ Replaced ./modules/dummy with ./modules/%s\n", config.FolderName))
			break
		}
	}

	if !updated {
		tap.Message("⚠️ No ./modules/dummy line found in go.work, skipping update")
		return nil
	}

	// Write updated go.work file
	newContent := strings.Join(lines, "\n")
	if err := os.WriteFile(goWorkPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write go.work: %w", err)
	}

	// Run go work sync
	sp := tap.NewSpinner(tap.SpinnerOptions{Indicator: "dots"})
	sp.Start("Running go work sync...")

	cmd := exec.Command("go", "work", "sync")
	cmd.Dir = config.ProjectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		sp.Stop("⚠️ go work sync completed with warnings", 0)
		tap.Message(fmt.Sprintf("⚠️ Warning: %v\n", err))
		return nil
	}

	sp.Stop("✅ go.work updated and synced", 0)
	return nil
}

// initGit initializes git repository
func initGit(ctx context.Context, projectDir string) error {
	sp := tap.NewSpinner(tap.SpinnerOptions{Indicator: "dots"})
	sp.Start("Initializing git repository...")

	cmd := exec.Command("git", "init")
	cmd.Dir = projectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		sp.Stop("❌ Failed to initialize git", 1)
		return fmt.Errorf("git init failed: %w", err)
	}

	sp.Stop("✅ Git repository initialized", 0)
	return nil
}
