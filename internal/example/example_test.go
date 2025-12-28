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
	"strings"
	"testing"
)

func TestGenerateConfig(t *testing.T) {
	tests := []struct {
		name          string
		complexity    templateComplexity
		projType      projectType
		includeComments bool
		checkStrings  []string
	}{
		{
			name:          "Simple nodejs config",
			complexity:    simple,
			projType:      nodejs,
			includeComments: true,
			checkStrings:  []string{"project:", "stages:", "recipe:"},
		},
		{
			name:          "Standard go config",
			complexity:    standard,
			projType:      golang,
			includeComments: true,
			checkStrings:  []string{"tasks:", "hooks:", "build:"},
		},
		{
			name:          "Advanced python config",
			complexity:    advanced,
			projType:      python,
			includeComments: false,
			checkStrings:  []string{"vars:", "python_version:"},
		},
		{
			name:          "Full php config",
			complexity:    full,
			projType:      php,
			includeComments: true,
			checkStrings:  []string{"slack_webhook:", "notification_email:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := generateConfig(tt.complexity, tt.projType, tt.includeComments)

			// Check that content is not empty
			if content == "" {
				t.Error("Generated config is empty")
			}

			// Check for expected strings
			for _, check := range tt.checkStrings {
				if !strings.Contains(content, check) {
					t.Errorf("Generated config missing expected string: %s", check)
				}
			}

			// If comments are enabled, check for comment prefix
			if tt.includeComments {
				if !strings.Contains(content, "#") {
					t.Error("Comments requested but no comments found in output")
				}
			}
		})
	}
}

func TestGenerateBasicConfig(t *testing.T) {
	content := generateBasicConfig(standard, nodejs, true)

	expectedStrings := []string{
		"project: my-awesome-app",
		"default_stage: dev",
		"# Basic Project Configuration",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(content, expected) {
			t.Errorf("Basic config missing expected string: %s", expected)
		}
	}
}

func TestGenerateStagesConfig(t *testing.T) {
	// Test with simple complexity
	content := generateStagesConfig(simple, true)
	if !strings.Contains(content, "dev:") {
		t.Error("Simple stages config missing dev stage")
	}

	// Test with advanced complexity (should include staging)
	content = generateStagesConfig(advanced, true)
	if !strings.Contains(content, "staging:") {
		t.Error("Advanced stages config missing staging stage")
	}
	if !strings.Contains(content, "prod:") {
		t.Error("Advanced stages config missing prod stage")
	}
}

func TestGenerateTasksConfig(t *testing.T) {
	tests := []struct {
		projType   projectType
		buildCheck string
	}{
		{nodejs, "npm run build"},
		{golang, "go build"},
		{python, "python -m build"},
		{php, "composer install"},
	}

	for _, tt := range tests {
		t.Run(string(tt.projType), func(t *testing.T) {
			content := generateTasksConfig(standard, tt.projType, true)
			if !strings.Contains(content, tt.buildCheck) {
				t.Errorf("Tasks config for %s missing expected build command: %s", tt.projType, tt.buildCheck)
			}
		})
	}
}

func TestGenerateRecipeConfig(t *testing.T) {
	tests := []struct {
		complexity templateComplexity
		minTasks   int
	}{
		{simple, 3},
		{standard, 6},
		{advanced, 8},
		{full, 7}, // full has comments so fewer actual tasks but more detailed
	}

	for _, tt := range tests {
		t.Run(string(tt.complexity), func(t *testing.T) {
			content := generateRecipeConfig(tt.complexity, true)
			if content == "" {
				t.Error("Recipe config is empty")
			}

			// Count the number of tasks (lines starting with "  -")
			taskCount := strings.Count(content, "  -")
			if taskCount < tt.minTasks {
				t.Errorf("Expected at least %d tasks, got %d", tt.minTasks, taskCount)
			}
		})
	}
}

func TestGenerateHooksConfig(t *testing.T) {
	// Test with standard complexity (only has before_all)
	content := generateHooksConfig(standard, true)

	expectedHooks := []string{
		"before_all:",
	}

	for _, hook := range expectedHooks {
		if !strings.Contains(content, hook) {
			t.Errorf("Hooks config missing expected hook: %s", hook)
		}
	}

	// Test with advanced complexity (has both before_all and after_all)
	content = generateHooksConfig(advanced, true)

	advancedHooks := []string{
		"before_all:",
		"after_all:",
		"before_build:",
		"after_build:success:",
		"after_build:failed:",
		"on_failed:",
	}

	for _, hook := range advancedHooks {
		if !strings.Contains(content, hook) {
			t.Errorf("Advanced hooks config missing expected hook: %s", hook)
		}
	}
}

func TestGenerateVarsConfig(t *testing.T) {
	tests := []struct {
		projType  projectType
		varCheck  string
	}{
		{nodejs, "node_version:"},
		{golang, "go_version:"},
		{python, "python_version:"},
		{php, "php_version:"},
	}

	for _, tt := range tests {
		t.Run(string(tt.projType), func(t *testing.T) {
			content := generateVarsConfig(advanced, tt.projType, true)
			if !strings.Contains(content, tt.varCheck) {
				t.Errorf("Vars config for %s missing expected var: %s", tt.projType, tt.varCheck)
			}
		})
	}
}
