package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// DevContainerImage represents a predefined image configuration
type DevContainerImage struct {
	Name        string `json:"name"`
	Image       string `json:"image"`
	DefaultPort string `json:"defaultPort"`
	Settings    string `json:"settings"`
	PostCommand string `json:"postCommand"`
	Features    string `json:"features"`
}

// DevContainerFile represents the structure of a devcontainer.json file
type DevContainerFile struct {
	Name              string                            `json:"name"`
	Image             string                            `json:"image"`
	ForwardPorts      []int                             `json:"forwardPorts,omitempty"`
	PostCreateCommand string                            `json:"postCreateCommand,omitempty"`
	Settings          map[string]any                    `json:"settings,omitempty"`
	Features          map[string]map[string]interface{} `json:"features,omitempty"`
}

var (
	// Predefined images with their configurations
	predefinedImages = []DevContainerImage{
		{
			Name:        "Bun (Latest)",
			Image:       "oven/bun:latest",
			DefaultPort: "3000",
			Settings:    `{"terminal.integrated.shell.linux": "/bin/bash"}`,
			PostCommand: "bun install",
			Features:    `{"ghcr.io/devcontainers/features/git:1": {}, "ghcr.io/devcontainers/features/common-utils:2": {}}`,
		},
		{
			Name:        "Node.js (Latest)",
			Image:       "mcr.microsoft.com/devcontainers/javascript-node:latest",
			DefaultPort: "3000",
			Settings:    `{"terminal.integrated.shell.linux": "/bin/bash"}`,
			PostCommand: "npm install",
			Features:    `{"ghcr.io/devcontainers/features/git:1": {}, "ghcr.io/devcontainers/features/common-utils:2": {}}`,
		},
		{
			Name:        "PHP 8.2 (Composer + Symfony)",
			Image:       "mcr.microsoft.com/devcontainers/php:8.2",
			DefaultPort: "8000",
			Settings:    `{"php.validate.executablePath": "/usr/local/bin/php"}`,
			PostCommand: "composer install && symfony check:requirements",
			Features:    `{"ghcr.io/devcontainers/features/composer:2": {}, "ghcr.io/devcontainers/features/node:1": {"version": "20"}, "ghcr.io/devcontainers/features/git:1": {}}`,
		},
		{
			Name:        "Go (Alpine)",
			Image:       "golang:alpine",
			DefaultPort: "8080",
			Settings:    `{"go.goroot": "/usr/local/go", "go.gopath": "/go"}`,
			PostCommand: "go mod download",
			Features:    `{"ghcr.io/devcontainers/features/git:1": {}}`,
		},
		{
			Name:        "Rust (Latest)",
			Image:       "mcr.microsoft.com/devcontainers/rust:latest",
			DefaultPort: "8000",
			Settings:    `{"rust-analyzer.server.path": "/usr/local/bin/rust-analyzer"}`,
			PostCommand: "cargo build",
			Features:    `{"ghcr.io/devcontainers/features/git:1": {}, "ghcr.io/devcontainers/features/common-utils:2": {}}`,
		},
		{
			Name:        "Python 3.12",
			Image:       "mcr.microsoft.com/devcontainers/python:latest",
			DefaultPort: "8000",
			Settings:    `{"python.defaultInterpreterPath": "/usr/local/bin/python"}`,
			PostCommand: "pip install -r requirements.txt",
			Features:    `{"ghcr.io/devcontainers/features/git:1": {}, "ghcr.io/devcontainers/features/common-utils:2": {}}`,
		},
		{
			Name:        ".NET 8",
			Image:       "mcr.microsoft.com/devcontainers/dotnet:8.0",
			DefaultPort: "5000",
			Settings:    `{"dotnet.server.useOmnisharp": false, "dotnet.defaultSolution": "**/*.sln"}`,
			PostCommand: "dotnet restore",
			Features:    `{"ghcr.io/devcontainers/features/dotnet:2": {}, "ghcr.io/devcontainers/features/git:1": {}, "ghcr.io/devcontainers/features/common-utils:2": {}}`,
		},
		{
			Name:        "Custom Image",
			Image:       "",
			DefaultPort: "",
			Settings:    "",
			PostCommand: "",
			Features:    "",
		},
	}

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED")).
			MarginBottom(1)

	successStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#22C55E"))

	errorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#EF4444"))
)

func main() {
	fmt.Println(titleStyle.Render("🐳 DevContainer Generator"))
	fmt.Println("Generate a devcontainer.json file for your project")
	fmt.Println()

	// Get current directory name as default project name
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println(errorStyle.Render("Error getting current directory: " + err.Error()))
		return
	}
	defaultName := filepath.Base(currentDir)

	var (
		name        string = defaultName
		selectedImg string
		customImage string
		ports       string
		postCommand string
		settings    string
		features    string
		finalImage  string
		confirm     bool
	)

	// Create image options for selection
	imageOptions := make([]huh.Option[string], len(predefinedImages))
	for i, img := range predefinedImages {
		imageOptions[i] = huh.NewOption(img.Name, strconv.Itoa(i))
	}

	// Step 1: Project name and image selection
	step1Form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Project Name").
				Description("Enter the name for your development container").
				Value(&name).
				Placeholder(defaultName).
				Validate(func(str string) error {
					if strings.TrimSpace(str) == "" {
						return fmt.Errorf("project name cannot be empty")
					}
					return nil
				}),
		),

		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select Base Image").
				Description("Choose a predefined image or select 'Custom Image' to specify your own").
				Options(imageOptions...).
				Value(&selectedImg),
		),

		huh.NewGroup(
			huh.NewInput().
				Title("Custom Image URL").
				Description("Enter the custom image URL (without https://)").
				Value(&customImage).
				Validate(func(str string) error {
					if selectedImg == strconv.Itoa(len(predefinedImages)-1) && strings.TrimSpace(str) == "" {
						return fmt.Errorf("custom image URL cannot be empty when custom image is selected")
					}
					return nil
				}),
		).WithHideFunc(func() bool {
			idx, _ := strconv.Atoi(selectedImg)
			return idx != len(predefinedImages)-1 // Hide if not custom image
		}),
	)

	err = step1Form.Run()
	if err != nil {
		fmt.Println(errorStyle.Render("Error: " + err.Error()))
		return
	}

	// Process the selected image and auto-populate fields
	selectedIndex, _ := strconv.Atoi(selectedImg)
	selectedImage := predefinedImages[selectedIndex]

	if selectedImage.Name == "Custom Image" {
		finalImage = customImage
	} else {
		finalImage = selectedImage.Image
		// Auto-populate with defaults
		ports = selectedImage.DefaultPort
		postCommand = selectedImage.PostCommand
		settings = selectedImage.Settings
		features = selectedImage.Features
	}

	// Step 2: Configuration with pre-populated values
	fmt.Println()
	fmt.Println(titleStyle.Render("🔧 Configure Options"))
	if selectedImage.Name != "Custom Image" {
		fmt.Println("Default values have been filled in. Press Enter to accept or modify as needed.")
	}
	fmt.Println()

	step2Form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Ports").
				Description("Enter port numbers separated by commas (e.g., 3000,8080)").
				Value(&ports),
		),

		huh.NewGroup(
			huh.NewInput().
				Title("Post Create Command").
				Description("Command to run after the container is created").
				Value(&postCommand),
		),

		huh.NewGroup(
			huh.NewText().
				Title("Settings").
				Description("VS Code settings in JSON format (optional)").
				Value(&settings).
				Lines(5),
		),

		huh.NewGroup(
			huh.NewText().
				Title("Features").
				Description("DevContainer features in JSON format (optional)").
				Value(&features).
				Lines(5),
		),
	)

	err = step2Form.Run()
	if err != nil {
		fmt.Println(errorStyle.Render("Error: " + err.Error()))
		return
	}

	// Display summary and ask for confirmation
	fmt.Println()
	fmt.Println(titleStyle.Render("📋 Configuration Summary"))
	fmt.Printf("Project Name: %s\n", name)
	fmt.Printf("Image: %s\n", finalImage)
	fmt.Printf("Ports: %s\n", ports)
	fmt.Printf("Post Create Command: %s\n", postCommand)
	fmt.Printf("Settings: %s\n", settings)
	fmt.Printf("Features: %s\n", features)
	fmt.Println()

	confirmForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Create devcontainer.json?").
				Description("Do you want to create the devcontainer.json file with the above configuration?").
				Value(&confirm),
		),
	)

	err = confirmForm.Run()
	if err != nil {
		fmt.Println(errorStyle.Render("Error: " + err.Error()))
		return
	}

	if !confirm {
		fmt.Println("Configuration cancelled.")
		return
	}

	// Create DevContainer struct
	devContainer := DevContainerFile{
		Name:  name,
		Image: finalImage,
	}

	// Parse ports
	if ports != "" {
		for port := range strings.SplitSeq(ports, ",") {
			port = strings.TrimSpace(port)
			if port != "" {
				if portNum, err := strconv.Atoi(port); err == nil {
					devContainer.ForwardPorts = append(devContainer.ForwardPorts, portNum)
				}
			}
		}
	}

	// Set post create command
	if postCommand != "" {
		devContainer.PostCreateCommand = postCommand
	}

	// Parse settings
	if settings != "" {
		var settingsMap map[string]any
		if err := json.Unmarshal([]byte(settings), &settingsMap); err == nil {
			devContainer.Settings = settingsMap
		}
	}

	// Parse features
	if features != "" {
		var featuresMap map[string]map[string]any
		if err := json.Unmarshal([]byte(features), &featuresMap); err == nil {
			devContainer.Features = featuresMap
		}
	}

	// Create .devcontainer directory if it doesn't exist
	if err := os.MkdirAll(".devcontainer", 0755); err != nil {
		fmt.Println(errorStyle.Render("Error creating .devcontainer directory: " + err.Error()))
		return
	}

	// Convert devContainer struct to JSON with indentation
	jsonData, err := json.MarshalIndent(devContainer, "", "  ")
	if err != nil {
		fmt.Println(errorStyle.Render("Error marshaling JSON: " + err.Error()))
		return
	}

	// Write to file
	filePath := ".devcontainer/devcontainer.json"
	err = os.WriteFile(filePath, jsonData, 0644)
	if err != nil {
		fmt.Println(errorStyle.Render("Error writing file: " + err.Error()))
		return
	}

	fmt.Println()
	fmt.Println(successStyle.Render("✅ Success!"))
	fmt.Printf("DevContainer configuration has been saved to: %s\n", filePath)
	fmt.Println()
	fmt.Println("You can now open this project in your terminal or IDE.")
}
