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
		content += "# ====================\n"
		content += "# Each stage represents a deployment environment (dev, staging, prod, etc.)\n"
		content += "#\n"
		content += "# Stage options:\n"
		content += "#   server       - SSH server address (user@server or just server)\n"
		content += "#   remote_dir   - Base directory for deployment on remote server\n"
		content += "#   keep_releases - Number of releases to keep (old ones are auto-cleaned)\n"
		content += "#   private_key  - Path to SSH private key (optional, uses default if not set)\n"
		content += "#   port         - SSH port (optional, default: 22)\n"
		content += "#   user         - SSH user (optional, default: current user)\n"
		content += "#   vars         - Environment-specific variables\n\n"
	}

	content += "stages:\n"

	// dev environment
	if includeComments {
		content += "  # Development environment\n"
	}
	content += "  dev:\n"
	content += "    server: dev-server.example.com\n"
	content += "    remote_dir: /var/www/app-dev\n"
	content += "    keep_releases: 3\n"
	if complexity == advanced || complexity == full {
		if includeComments {
			content += "    # SSH connection settings\n"
		}
		content += "    private_key: ~/.ssh/id_rsa\n"
		content += "    port: 22\n"
		content += "    user: deploy\n"
	}
	if includeComments {
		content += "    # Environment variables accessible in tasks and hooks\n"
	}
	content += "    vars:\n"
	content += "      APP_ENV: development\n"
	content += "      DEBUG: \"true\"\n"
	content += "      LOG_LEVEL: debug\n\n"

	// staging environment (only for advanced and full)
	if complexity == advanced || complexity == full {
		if includeComments {
			content += "  # Staging environment (pre-production testing)\n"
		}
		content += "  staging:\n"
		content += "    server: staging-server.example.com\n"
		content += "    remote_dir: /var/www/app-staging\n"
		content += "    keep_releases: 5\n"
		content += "    private_key: ~/.ssh/id_rsa\n"
		content += "    vars:\n"
		content += "      APP_ENV: staging\n"
		content += "      DEBUG: \"false\"\n"
		content += "      LOG_LEVEL: info\n\n"
	}

	// prod environment
	if includeComments {
		content += "  # Production environment\n"
	}
	content += "  prod:\n"
	content += "    server: prod-server.example.com\n"
	content += "    remote_dir: /var/www/app-prod\n"
	content += "    keep_releases: 5\n"
	if complexity == advanced || complexity == full {
		content += "    private_key: ~/.ssh/id_rsa\n"
		content += "    port: 22\n"
		content += "    user: deploy\n"
	}
	if includeComments {
		content += "    # Production variables\n"
	}
	content += "    vars:\n"
	content += "      APP_ENV: production\n"
	content += "      DEBUG: \"false\"\n"
	content += "      LOG_LEVEL: warn\n\n"

	return content
}

// generateOptionsConfig generates global options configuration section
func generateOptionsConfig(complexity templateComplexity, includeComments bool) string {
	content := ""

	if includeComments {
		content += "# ====================\n"
		content += "# Global Options Configuration\n"
		content += "# ====================\n"
		content += "# These options apply to all deployment stages\n\n"
	}

	content += "options:\n"

	if includeComments {
		content += "  # Release directory naming format\n"
		content += "  # Supports strftime format specifiers:\n"
		content += "  #   %Y - 4-digit year, %m - month (01-12), %d - day (01-31)\n"
		content += "  #   %H - hour (00-23), %M - minute (00-59), %S - second (00-59)\n"
	}
	content += "  release_dir_format: \"releases/%Y%m%d%H%M%S\"\n\n"

	if complexity != simple {
		if includeComments {
			content += "  # Shared directories - persist across deployments\n"
			content += "  # On first deployment: moved from release to shared/\n"
			content += "  # On subsequent deployments: symlinked from shared/ to release/\n"
			content += "  # Common use cases: logs, user uploads, cache, session storage\n"
		}
		content += "  shared_dirs:\n"
		content += "    - logs\n"
		content += "    - uploads\n"
		content += "    - storage\n\n"

		if includeComments {
			content += "  # Shared files - persist across deployments\n"
			content += "  # Useful for configuration files that shouldn't be overwritten\n"
			content += "  # Common use cases: .env files, config files\n"
		}
		content += "  shared_files:\n"
		content += "    - .env\n"
		content += "    - config/database.yml\n\n"
	}

	return content
}

// generateTasksConfig generates custom tasks configuration section
func generateTasksConfig(complexity templateComplexity, projType projectType, includeComments bool) string {
	content := ""

	if includeComments {
		content += "# ====================\n"
		content += "# Custom Tasks Configuration\n"
		content += "# ====================\n"
		content += "# Tasks can be one of three types:\n"
		content += "# 1. local:  Run commands on your local machine before deployment\n"
		content += "# 2. remote: Run commands on the remote server after deployment\n"
		content += "# 3. upload: Upload files/directories from local to remote\n"
		content += "#\n"
		content += "# Available variables:\n"
		content += "#   {{release_path}}  - Path to the current release directory\n"
		content += "#   {{remote_dir}}    - Base remote directory path\n"
		content += "#   {{stage}}         - Current deployment stage\n"
		content += "#   {{project}}       - Project name\n\n"
	}

	content += "tasks:\n"

	// Generate build task based on project type
	if complexity != simple {
		if includeComments {
			content += "  # Local build task - runs on your machine\n"
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
			content += "  # Remote task example - runs on server after deployment\n"
			content += "  # Using pipe | for multi-line commands\n"
		}
		content += "  install-deps:\n"
		content += "    remote: |\n"
		content += "      cd {{release_path}}\n"
		content += "      npm install --production\n\n"

		if includeComments {
			content += "  # Upload task - sync files to remote server\n"
			content += "  # source: local path, dest: remote path, options: rsync options\n"
		}
		content += "  upload-assets:\n"
		content += "    upload:\n"
		content += "      source: ./dist\n"
		content += "      dest: \"{{release_path}}\"\n"
		content += "      options: \"--delete --exclude=node_modules\"\n\n"

		if includeComments {
			content += "  # Multiple task types in one task\n"
		}
		content += "  migrate:\n"
		content += "    local: echo 'Running database migration...'\n"
		content += "    remote: |\n"
		content += "      cd {{release_path}} && php artisan migrate --force\n\n"

		if includeComments {
			content += "  # Another remote task example\n"
		}
		content += "  restart-server:\n"
		content += "    remote: sudo systemctl restart nginx\n\n"
	}

	return content
}

// generateRecipeConfig generates deployment recipe configuration section
func generateRecipeConfig(complexity templateComplexity, includeComments bool) string {
	content := ""

	if includeComments {
		content += "# ====================\n"
		content += "# Deployment Recipe Configuration\n"
		content += "# ====================\n"
		content += "# Recipe defines the deployment workflow steps.\n"
		content += "# You can use both built-in tasks and custom tasks defined above.\n"
		content += "#\n"
		content += "# Available built-in tasks:\n"
		content += "#   update_code      - Upload code to release directory via rsync\n"
		content += "#   publish_release  - Prepare release (create shared dirs/files links)\n"
		content += "#   symlink_release  - Update 'current' symlink to point to new release\n"
		content += "#   cleanup          - Remove old releases (keep only specified number)\n"
		content += "#   rollback         - Rollback to previous release\n"
		content += "#\n"
		content += "# Custom tasks are defined in the 'tasks' section above.\n\n"
	}

	content += "recipe:\n"

	if complexity == simple {
		content += "  - update_code        # Upload code to server\n"
		content += "  - symlink_release     # Point 'current' to new release\n"
		content += "  - cleanup             # Remove old releases\n\n"
	} else if complexity == standard {
		content += "  - build               # Build locally (custom task)\n"
		content += "  - update_code         # Upload to release directory\n"
		content += "  - publish_release     # Setup shared directories/files\n"
		content += "  - symlink_release     # Update current symlink\n"
		content += "  - cleanup             # Remove old releases\n\n"
	} else if complexity == advanced {
		content += "  - install-deps        # Install dependencies on server (custom)\n"
		content += "  - update_code         # Upload code to release directory\n"
		content += "  - publish_release     # Prepare release (shared dirs/files)\n"
		content += "  - migrate             # Run database migrations (custom)\n"
		content += "  - symlink_release     # Update 'current' symlink\n"
		content += "  - restart-server      # Restart web server (custom)\n"
		content += "  - cleanup             # Remove old releases\n\n"
	} else if complexity == full {
		if includeComments {
			content += "  # Step 1: Prepare and build\n"
		}
		content += "  - build               # Build locally\n\n"

		if includeComments {
			content += "  # Step 2: Upload code\n"
		}
		content += "  - update_code         # Upload to release directory\n\n"

		if includeComments {
			content += "  # Step 3: Setup release\n"
		}
		content += "  - publish_release     # Create shared dirs/files symlinks\n\n"

		if includeComments {
			content += "  # Step 4: Install dependencies\n"
		}
		content += "  - install-deps        # Install server dependencies\n\n"

		if includeComments {
			content += "  # Step 5: Run migrations\n"
		}
		content += "  - migrate             # Database migrations\n\n"

		if includeComments {
			content += "  # Step 6: Go live\n"
		}
		content += "  - symlink_release     # Point current to new release\n\n"

		if includeComments {
			content += "  # Step 7: Restart services\n"
		}
		content += "  - restart-server      # Restart web server\n\n"

		if includeComments {
			content += "  # Step 8: Cleanup\n"
		}
		content += "  - cleanup             # Remove old releases\n\n"
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
		content += "# Hooks allow you to execute custom commands at specific points.\n"
		content += "#\n"
		content += "# Available hook types:\n"
		content += "#   before_all         - Before deployment starts\n"
		content += "#   after_all          - After deployment completes successfully\n"
		content += "#   before_<task>      - Before a specific task runs\n"
		content += "#   after_<task>       - After a specific task completes (success or failure)\n"
		content += "#   after_<task>:success - After a specific task succeeds\n"
		content += "#   after_<task>:failed  - After a specific task fails\n"
		content += "#   on_failed          - When any task fails during deployment\n"
		content += "#\n"
		content += "# Available variables in hooks:\n"
		content += "#   {{stage}}      - Current deployment stage\n"
		content += "#   {{release_path}} - Path to current release\n"
		content += "#   {{error}}      - Error message (only in on_failed)\n\n"
	}

	content += "hooks:\n"

	if includeComments {
		content += "  # Global hooks - run at start/end of deployment\n"
	}
	content += "  before_all:\n"
	content += "    - echo \"Starting deployment to {{stage}} environment...\"\n"
	content += "    - echo \"Timestamp: $(date)\"\n\n"

	if complexity == advanced || complexity == full {
		if includeComments {
			content += "  # After all tasks complete successfully\n"
		}
		content += "  after_all:\n"
		content += "    - echo \"Deployment completed successfully!\"\n"
		content += "    - echo \"Release path: {{release_path}}\"\n\n"

		if includeComments {
			content += "  # Task-specific hooks\n"
			content += "  # These run before/after specific tasks in your recipe\n"
		}
		content += "  before_build:\n"
		content += "    - echo \"Preparing to build...\"\n"
		content += "    - echo \"Free disk space: $(df -h .)\"\n\n"

		if includeComments {
			content += "  # Success hook - only runs if task succeeds\n"
		}
		content += "  after_build:success:\n"
		content += "    - echo \"Build succeeded!\"\n"
		content += "    - echo \"Notifying team...\"\n"
		content += "    - echo \"Build size: $(du -sh dist)\"\n\n"

		if includeComments {
			content += "  # Failure hook - only runs if task fails\n"
		}
		content += "  after_build:failed:\n"
		content += "    - echo \"Build failed!\"\n"
		content += "    - echo \"Sending alert to team...\"\n\n"

		if includeComments {
			content += "  # More task-specific examples\n"
		}
		content += "  before_migrate:\n"
		content += "    - echo \"Creating database backup before migration...\"\n"
		content += "    - remote: mysqldump db_name > backup.sql\n\n"

		content += "  after_migrate:success:\n"
		content += "    - echo \"Migration completed successfully\"\n\n"

		if includeComments {
			content += "  # Global failure hook - runs when ANY task fails\n"
		}
		content += "  on_failed:\n"
		content += "    - echo \"Deployment failed! Error: {{error}}\"\n"
		content += "    - echo \"Rolling back changes...\"\n"
		content += "    - echo \"Sending alert to administrators\"\n\n"
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
