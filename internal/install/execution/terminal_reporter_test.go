//go:build unit
// +build unit

package execution

import (
	// "fmt"
	// "os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/newrelic/newrelic-cli/internal/install/types"
)

func TestTerminalStatusReporter_interface(t *testing.T) {
	var r StatusSubscriber = NewTerminalStatusReporter()
	require.NotNil(t, r)
}

func Test_ShouldGenerateEntityLink(t *testing.T) {
	r := NewTerminalStatusReporter()
	g := NewMockPlatformLinkGenerator()
	status := &InstallStatus{
		PlatformLinkGenerator: g,
	}
	recipeStatus := &RecipeStatus{
		Status: RecipeStatusTypes.INSTALLED,
	}
	status.Statuses = append(status.Statuses, recipeStatus)

	err := r.InstallComplete(status)
	require.NoError(t, err)
	require.Equal(t, 1, g.GenerateEntityLinkCallCount)
	require.Equal(t, 0, g.GenerateExplorerLinkCallCount)
}

func Test_ShouldNotGenerateEntityLink(t *testing.T) {
	r := NewTerminalStatusReporter()
	g := NewMockPlatformLinkGenerator()

	status := &InstallStatus{
		PlatformLinkGenerator: g,
	}
	recipeStatus := &RecipeStatus{
		Status: RecipeStatusTypes.FAILED,
	}
	status.Statuses = append(status.Statuses, recipeStatus)

	err := r.InstallComplete(status)
	require.NoError(t, err)
	require.Equal(t, 0, g.GenerateEntityLinkCallCount)
	require.Equal(t, 0, g.GenerateExplorerLinkCallCount)
}

func Test_ShouldNotGenerateEntityLinkWhenNoRecipes(t *testing.T) {
	r := NewTerminalStatusReporter()
	g := NewMockPlatformLinkGenerator()

	status := &InstallStatus{
		PlatformLinkGenerator: g,
	}

	err := r.InstallComplete(status)
	require.NoError(t, err)
	require.Equal(t, 0, g.GenerateEntityLinkCallCount)
	require.Equal(t, 0, g.GenerateExplorerLinkCallCount)
}

func Test_ShouldGenerateExplorerLink(t *testing.T) {
	r := NewTerminalStatusReporter()
	g := NewMockPlatformLinkGenerator()

	status := &InstallStatus{
		PlatformLinkGenerator: g,
	}
	recipeStatus := &RecipeStatus{
		Status: RecipeStatusTypes.INSTALLED,
	}
	status.Statuses = append(status.Statuses, recipeStatus)
	status.successLinkConfig = types.OpenInstallationSuccessLinkConfig{
		Type:   "explorer",
		Filter: "\"`tags.language` = 'java'\"",
	}

	err := r.InstallComplete(status)
	require.NoError(t, err)
	require.Equal(t, 0, g.GenerateEntityLinkCallCount)
	require.Equal(t, 1, g.GenerateExplorerLinkCallCount)
}

func Test_ShouldNotGenerateExplorerLink(t *testing.T) {
	r := NewTerminalStatusReporter()
	g := NewMockPlatformLinkGenerator()

	status := &InstallStatus{
		PlatformLinkGenerator: g,
	}
	recipeStatus := &RecipeStatus{
		Status: RecipeStatusTypes.FAILED,
	}
	status.Statuses = append(status.Statuses, recipeStatus)
	status.successLinkConfig = types.OpenInstallationSuccessLinkConfig{
		Type:   "explorer",
		Filter: "\"`tags.language` = 'java'\"",
	}

	err := r.InstallComplete(status)
	require.NoError(t, err)
	require.Equal(t, 0, g.GenerateEntityLinkCallCount)
	require.Equal(t, 0, g.GenerateExplorerLinkCallCount)
}

func Test_ShouldNotGenerateExplorerLinkWhenNoRecipes(t *testing.T) {
	r := NewTerminalStatusReporter()
	g := NewMockPlatformLinkGenerator()

	status := &InstallStatus{
		PlatformLinkGenerator: g,
	}
	status.successLinkConfig = types.OpenInstallationSuccessLinkConfig{
		Type:   "explorer",
		Filter: "\"`tags.language` = 'java'\"",
	}

	err := r.InstallComplete(status)
	require.NoError(t, err)
	require.Equal(t, 0, g.GenerateEntityLinkCallCount)
	require.Equal(t, 0, g.GenerateExplorerLinkCallCount)
}

func Test_ShouldNotPrintDetectedRecipeInSummary(t *testing.T) {
	r := NewTerminalStatusReporter()

	status := &InstallStatus{}
	recipeInstalled := &RecipeStatus{
		DisplayName: "test-recipe-installed",
		Status:      RecipeStatusTypes.INSTALLED,
	}
	recipeDetected := &RecipeStatus{
		DisplayName: "test-recipe-detected",
		Status:      RecipeStatusTypes.DETECTED,
	}

	status.Statuses = []*RecipeStatus{
		recipeInstalled,
		recipeDetected,
	}

	// expected := []*RecipeStatus{
	// 	recipeInstalled,
	// }

	recipesToSummarize := r.getRecipesStatusesForInstallationSummary(status)

	require.Equal(t, 1, len(recipesToSummarize))
}
