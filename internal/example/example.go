/*
 * GoDeployer - Flexible and extensible deployment tool
 * Copyright (C) 2025-2025 amoydavid
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published
 * by the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package example

import (
	"fmt"
	"os"
	"time"

	"github.com/AlecAivazis/survey/v2"
)

// templateComplexity represents the configuration template complexity level
type templateComplexity string

const (
	simple   templateComplexity = "simple"
	standard templateComplexity = "standard"
	advanced templateComplexity = "advanced"
	full     templateComplexity = "full"
)

// projectType represents the type of project being deployed
type projectType string

const (
	nodejs  projectType = "nodejs"
	golang  projectType = "go"
	python  projectType = "python"
	php     projectType = "php"
	static  projectType = "static"
	generic projectType = "generic"
)

// Run executes the example configuration file generator
func Run(args []string) error {
	// Validate arguments
	if len(args) == 0 {
		return fmt.Errorf("please specify output filename, e.g.: deployer example deploy.example.yaml")
	}

	filename := args[0]

	// Check if file exists
	if _, err := os.Stat(filename); err == nil {
		// File exists, ask what to do
		var action string
		prompt := &survey.Select{
			Message: fmt.Sprintf("File '%s' already exists. Choose an action:", filename),
			Options: []string{"Overwrite", "Rename", "Cancel"},
			Default: "Cancel",
		}
		if err := survey.AskOne(prompt, &action); err != nil {
			return fmt.Errorf("failed to ask: %w", err)
		}

		switch action {
		case "Cancel":
			fmt.Println("Operation cancelled")
			return nil
		case "Rename":
			var newName string
			prompt := &survey.Input{
				Message: "Enter new filename:",
				Default: filename + ".new",
			}
			if err := survey.AskOne(prompt, &newName); err != nil {
				return fmt.Errorf("failed to get input: %w", err)
			}
			filename = newName
		case "Overwrite":
			// Continue with original filename
		}
	}

	// Interactive menu for configuration complexity
	var complexityStr string
	prompt1 := &survey.Select{
		Message: "Choose configuration template complexity:",
		Options: []string{
			"simple   - Simple project configuration (for static sites, simple apps)",
			"standard - Standard project configuration (for most web applications)",
			"advanced - Advanced project configuration (multi-environment, complex hooks)",
			"full     - Complete example configuration (showcases all features)",
		},
		Default: "standard - Standard project configuration (for most web applications)",
	}
	if err := survey.AskOne(prompt1, &complexityStr); err != nil {
		return fmt.Errorf("failed to select complexity: %w", err)
	}
	complexity := templateComplexity(complexityStr[0:7]) // Extract first 7 characters as type

	// Interactive menu for project type
	var projTypeStr string
	prompt2 := &survey.Select{
		Message: "Choose project type:",
		Options: []string{
			"nodejs  - Node.js application",
			"go      - Go application",
			"python  - Python application",
			"php     - PHP application",
			"static  - Static website",
			"generic - Generic project",
		},
		Default: "nodejs  - Node.js application",
	}
	if err := survey.AskOne(prompt2, &projTypeStr); err != nil {
		return fmt.Errorf("failed to select project type: %w", err)
	}
	projType := projectType(projTypeStr[0:7]) // Extract first 7 characters as type

	// Ask whether to include detailed comments
	var includeComments bool
	prompt3 := &survey.Confirm{
		Message: "Include detailed comments?",
		Default: true,
	}
	if err := survey.AskOne(prompt3, &includeComments); err != nil {
		return fmt.Errorf("failed to confirm: %w", err)
	}

	// Generate configuration content
	configContent := generateConfig(complexity, projType, includeComments)

	// Write to file
	if err := os.WriteFile(filename, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}

	fmt.Printf("\n✓ Example configuration generated: %s\n", filename)
	fmt.Printf("  Complexity: %s\n", complexity)
	fmt.Printf("  Project type: %s\n", projType)
	if includeComments {
		fmt.Printf("  Detailed comments: Yes\n")
	} else {
		fmt.Printf("  Detailed comments: No\n")
	}

	return nil
}

// generateConfig generates configuration content based on options
func generateConfig(complexity templateComplexity, projType projectType, includeComments bool) string {
	commentPrefix := ""
	if includeComments {
		commentPrefix = "# "
	}

	var content string

	// File header
	content += fmt.Sprintf("%sGoDeployer Configuration File\n", commentPrefix)
	content += fmt.Sprintf("%sGenerated: %s\n", commentPrefix, time.Now().Format("2006-01-02"))
	content += "\n"

	// Basic configuration
	content += generateBasicConfig(complexity, projType, includeComments)

	// Stages configuration
	content += generateStagesConfig(complexity, includeComments)

	// Options configuration
	content += generateOptionsConfig(complexity, includeComments)

	// Tasks configuration
	content += generateTasksConfig(complexity, projType, includeComments)

	// Recipe configuration
	content += generateRecipeConfig(complexity, includeComments)

	// Hooks configuration
	if complexity == standard || complexity == advanced || complexity == full {
		content += generateHooksConfig(complexity, includeComments)
	}

	// Vars configuration
	if complexity == advanced || complexity == full {
		content += generateVarsConfig(complexity, projType, includeComments)
	}

	return content
}

// generateBasicConfig generates basic project configuration section
func generateBasicConfig(complexity templateComplexity, projType projectType, includeComments bool) string {
	content := ""

	if includeComments {
		content += "# ====================\n"
		content += "# Basic Project Configuration\n"
		content += "# ====================\n\n"
		content += "# Project name\n"
	}

	content += "project: my-awesome-app\n\n"

	if includeComments {
		content += "# Default deployment stage (environment)\n"
		content += "# Available values: dev, staging, prod\n"
	}
	content += "default_stage: dev\n\n"

	return content
}

// generateStagesConfig generates deployment stages configuration section
func generateStagesConfig(complexity templateComplexity, includeComments bool) string {
	content := ""

	if includeComments {
		content += "# ====================\n"
		content += "# Deployment Stages (Environments) Configuration\n"
		content += "# ====================\n\n"
	}

	content += "stages:\n"

	// dev environment
	if includeComments {
		content += "  # Development environment configuration\n"
	}
	content += "  dev:\n"
	content += "    server: dev-server.example.com\n"
	content += "    remote_dir: /var/www/app-dev\n"
	content += "    keep_releases: 3\n"
	if complexity == advanced || complexity == full {
		content += "    private_key: ~/.ssh/id_rsa\n"
	}
	if includeComments {
		content += "    # Development-specific variables\n"
	}
	content += "    vars:\n"
	content += "      APP_ENV: development\n"
	content += "      DEBUG: \"true\"\n\n"

	// staging environment (only for advanced and full)
	if complexity == advanced || complexity == full {
		if includeComments {
			content += "  # Staging environment configuration\n"
		}
		content += "  staging:\n"
		content += "    server: staging-server.example.com\n"
		content += "    remote_dir: /var/www/app-staging\n"
		content += "    keep_releases: 5\n"
		content += "    private_key: ~/.ssh/id_rsa\n"
		content += "    vars:\n"
		content += "      APP_ENV: staging\n"
		content += "      DEBUG: \"false\"\n\n"
	}

	// prod environment
	if includeComments {
		content += "  # Production environment configuration\n"
	}
	content += "  prod:\n"
	content += "    server: prod-server.example.com\n"
	content += "    remote_dir: /var/www/app-prod\n"
	content += "    keep_releases: 5\n"
	if complexity == advanced || complexity == full {
		content += "    private_key: ~/.ssh/id_rsa\n"
	}
	if includeComments {
		content += "    # Production-specific variables\n"
	}
	content += "    vars:\n"
	content += "      APP_ENV: production\n"
	content += "      DEBUG: \"false\"\n\n"

	return content
}

// generateOptionsConfig generates global options configuration section
func generateOptionsConfig(complexity templateComplexity, includeComments bool) string {
	content := ""

	if includeComments {
		content += "# ====================\n"
		content += "# Global Options Configuration\n"
		content += "# ====================\n\n"
	}

	content += "options:\n"

	if includeComments {
		content += "  # Release directory naming format (supports strftime format)\n"
	}
	content += "  release_dir_format: \"releases/%Y%m%d%H%M%S\"\n\n"

	if complexity != simple {
		if includeComments {
			content += "  # Shared directories (shared across all releases)\n"
			content += "  # These directories will be moved from release to shared directory on first deployment\n"
			content += "  # Subsequent deployments will create symlinks to shared directory\n"
		}
		content += "  shared_dirs:\n"
		content += "    - logs\n"
		content += "    - uploads\n\n"

		if includeComments {
			content += "  # Shared files (shared across all releases)\n"
		}
		content += "  shared_files:\n"
		content += "    - .env\n\n"
	}

	return content
}

// generateTasksConfig generates custom tasks configuration section
func generateTasksConfig(complexity templateComplexity, projType projectType, includeComments bool) string {
	content := ""

	if includeComments {
		content += "# ====================\n"
		content += "# Custom Tasks Configuration\n"
		content += "# ====================\n\n"
	}

	content += "tasks:\n"

	// Generate build task based on project type
	if complexity != simple {
		if includeComments {
			content += "  # Local build task\n"
		}
		switch projType {
		case nodejs:
			content += "  build:\n"
			content += "    local: npm run build\n\n"
		case golang:
			content += "  build:\n"
			content += "    local: go build -o bin/app ./cmd/app\n\n"
		case python:
			content += "  build:\n"
			content += "    local: python -m build\n\n"
		case php:
			content += "  build:\n"
			content += "    local: composer install --no-dev\n\n"
		case static:
			content += "  build:\n"
			content += "    local: echo 'No build needed for static site'\n\n"
		default:
			content += "  build:\n"
			content += "    local: echo 'Building...'\n\n"
		}
	}

	if complexity == advanced || complexity == full {
		if includeComments {
			content += "  # Backup task\n"
		}
		content += "  backup:\n"
		content += "    remote: |\n"
		content += "      if [ -d {{remote_dir}}/current ]; then\n"
		content += "        cp -r {{remote_dir}}/current {{remote_dir}}/backups/backup_$(date +%Y%m%d_%H%M%S)\n"
		content += "      fi\n\n"

		if includeComments {
			content += "  # Upload task\n"
		}
		content += "  upload-app:\n"
		content += "    upload:\n"
		content += "      source: ./dist\n"
		content += "      dest: \"{{release_path}}\"\n"
		content += "      options: \"--delete\"\n\n"

		if includeComments {
			content += "  # Database migration task\n"
		}
		content += "  migrate:\n"
		content += "    remote: |\n"
		content += "      cd {{release_path}} && php artisan migrate --force\n\n"
	}

	return content
}

// generateRecipeConfig generates deployment recipe configuration section
func generateRecipeConfig(complexity templateComplexity, includeComments bool) string {
	content := ""

	if includeComments {
		content += "# ====================\n"
		content += "# Deployment Recipe Configuration\n"
		content += "# ====================\n\n"
	}

	content += "recipe:\n"

	if complexity == simple {
		content += "  - update_code\n"
		content += "  - symlink_release\n"
		content += "  - cleanup\n\n"
	} else if complexity == standard {
		content += "  - build\n"
		content += "  - update_code\n"
		content += "  - publish_release\n"
		content += "  - symlink_release\n"
		content += "  - restart_app\n"
		content += "  - cleanup\n\n"
	} else if complexity == advanced {
		content += "  - backup\n"
		content += "  - build\n"
		content += "  - update_code\n"
		content += "  - publish_release\n"
		content += "  - migrate\n"
		content += "  - symlink_release\n"
		content += "  - restart_app\n"
		content += "  - cleanup\n\n"
	} else if complexity == full {
		if includeComments {
			content += "  # Code update\n"
		}
		content += "  - update_code\n\n"

		if includeComments {
			content += "  # Local build\n"
		}
		content += "  - build\n\n"

		if includeComments {
			content += "  # Release preparation (shared directories/files)\n"
		}
		content += "  - publish_release\n\n"

		if includeComments {
			content += "  # Create symlink\n"
		}
		content += "  - symlink_release\n\n"

		if includeComments {
			content += "  # Database migration\n"
		}
		content += "  - migrate\n\n"

		if includeComments {
			content += "  # Restart application\n"
		}
		content += "  - restart_app\n\n"

		if includeComments {
			content += "  # Cleanup old releases\n"
		}
		content += "  - cleanup\n\n"
	}

	return content
}

// generateHooksConfig generates deployment hooks configuration section
func generateHooksConfig(complexity templateComplexity, includeComments bool) string {
	content := ""

	if includeComments {
		content += "# ====================\n"
		content += "# Deployment Hooks Configuration\n"
		content += "# ====================\n"
		content += "# Hooks allow you to execute custom commands before/after specific events\n\n"
	}

	content += "hooks:\n"

	if includeComments {
		content += "  # Global before hook (before all tasks)\n"
	}
	content += "  before_all:\n"
	content += "    - echo \"Starting deployment to {{stage}} environment...\"\n\n"

	if complexity == advanced || complexity == full {
		if includeComments {
			content += "  # Global after hook (after all tasks)\n"
		}
		content += "  after_all:\n"
		content += "    - echo \"Deployment completed successfully!\"\n\n"

		if includeComments {
			content += "  # Before specific task hook\n"
		}
		content += "  before_build:\n"
		content += "    - echo \"Preparing to build...\"\n\n"

		if includeComments {
			content += "  # Success hook for specific task\n"
		}
		content += "  after_build:success:\n"
		content += "    - echo \"Build succeeded!\"\n"
		content += "    - echo \"Notifying team...\"\n\n"

		if includeComments {
			content += "  # Failure hook for specific task\n"
		}
		content += "  after_build:failed:\n"
		content += "    - echo \"Build failed!\"\n"
		content += "    - echo \"Sending alert...\"\n\n"

		if includeComments {
			content += "  # Global failure hook (when any task fails)\n"
		}
		content += "  on_failed:\n"
		content += "    - echo \"Deployment failed! Error: {{error}}\"\n"
		content += "    - echo \"Rolling back...\"\n\n"
	}

	return content
}

// generateVarsConfig generates global variables configuration section
func generateVarsConfig(complexity templateComplexity, projType projectType, includeComments bool) string {
	content := ""

	if includeComments {
		content += "# ====================\n"
		content += "# Global Variables Configuration\n"
		content += "# ====================\n"
		content += "# These variables can be used in hooks and tasks\n\n"
	}

	content += "vars:\n"
	content += "  app_name: my-awesome-app\n"

	switch projType {
	case nodejs:
		content += "  node_version: \"18\"\n"
	case golang:
		content += "  go_version: \"1.21\"\n"
	case python:
		content += "  python_version: \"3.11\"\n"
	case php:
		content += "  php_version: \"8.2\"\n"
	}

	content += "  deploy_user: deploy\n"

	if complexity == full {
		content += "  slack_webhook: \"https://hooks.slack.com/services/YOUR/WEBHOOK/URL\"\n"
		content += "  notification_email: \"team@example.com\"\n"
	}

	content += "\n"

	return content
}

// TestGenerateConfigForTesting is an exported helper function for testing/demo purposes
func TestGenerateConfigForTesting(complexity templateComplexity, projType projectType, includeComments bool) string {
	return generateConfig(complexity, projType, includeComments)
}
