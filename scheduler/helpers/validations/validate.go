package validations

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/tobg/scheduler/models"
)

// Tasks represent all single task available for scheduler
var Tasks = map[string]models.TaskHandler{
	"deploy": {
		Execute: nil,
		Verify:  validateDeployArgs,
	},
}

// IsMethodAllowed returns wether or not the method is allowed
func IsMethodAllowed(rMethod, method string) error {
	if rMethod == method {
		return nil
	}
	return fmt.Errorf("invalid method %s, accepted method is %s", rMethod, method)
}

func IsValidOccurrences(i int) error {
	if i >= 1 || i == -1 {
		return nil
	}
	return fmt.Errorf("invalid occurence: %v", i)
}

func IsValidFrequency(f string) error {
	validFrequencies := map[string]bool{
		"W": true,
		"M": true,
		"D": true,
		"H": true,
		"m": true,
		"Y": true,
	}

	if !validFrequencies[f] {
		return fmt.Errorf("invalid frequency: %v", f)
	}

	return nil
}

func IsValidAction(t models.Task) error {
	action, exists := Tasks[t.Action]
	if !exists {
		return fmt.Errorf("task %v does not exist", t.Action)
	}

	err := action.Verify(t.Args)
	if err != nil {
		return fmt.Errorf("could not verify task %v : %w", t.Action, err)
	}

	return nil
}

func validateDeployArgs(args []string) error {
	if len(args) != 2 {
		return errors.New("invalid number of arguments: expected service name and deployment path")
	}

	serviceName := args[0]
	deploymentPath := args[1]

	// Validate the deployment path format
	expectedPath := filepath.Join("/home/apps", serviceName)

	// Check if the deployment path is correct
	if !strings.HasPrefix(deploymentPath, expectedPath) {
		return fmt.Errorf("invalid deployment path: expected '%s', got '%s'", expectedPath, deploymentPath)
	}
	return nil
}
