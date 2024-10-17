package validations

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tobg/scheduler/models"
)

func TestIsMethodAllowed(t *testing.T) {
	tests := map[string]struct {
		rMethod      string
		expectMethod string
		wantErr      assert.ErrorAssertionFunc
	}{
		"nominal": {
			rMethod:      http.MethodGet,
			expectMethod: http.MethodGet,
			wantErr:      assert.NoError,
		},
		"invalid method": {
			rMethod:      http.MethodPost,
			expectMethod: http.MethodGet,
			wantErr:      assert.Error,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := IsMethodAllowed(tt.rMethod, tt.expectMethod)
			tt.wantErr(t, err)
		})
	}
}

func TestIsValidOccurrences(t *testing.T) {
	tests := map[string]struct {
		occurences int
		wantErr    assert.ErrorAssertionFunc
	}{
		"nominal single run": {
			occurences: 1,
			wantErr:    assert.NoError,
		},
		"nominal multiple run": {
			occurences: 10,
			wantErr:    assert.NoError,
		},
		"nominal infinite run": {
			occurences: -1,
			wantErr:    assert.NoError,
		},
		"invalid method, negative integer": {
			occurences: -2,
			wantErr:    assert.Error,
		},
		"invalid method, no occurences": {
			occurences: 0,
			wantErr:    assert.Error,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := IsValidOccurrences(tt.occurences)
			tt.wantErr(t, err)
		})
	}
}

func TestIsValidFrequency(t *testing.T) {
	tests := map[string]struct {
		frequency string
		wantErr   assert.ErrorAssertionFunc
	}{
		"nominal week frequency": {
			frequency: "W",
			wantErr:   assert.NoError,
		},
		"nominal hour frequency": {
			frequency: "H",
			wantErr:   assert.NoError,
		},
		"unknown frequency, return error": {
			frequency: "Z",
			wantErr:   assert.Error,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := IsValidFrequency(tt.frequency)
			tt.wantErr(t, err)
		})
	}
}

func TestIsValidAction(t *testing.T) {
	tests := map[string]struct {
		task    models.Task
		wantErr assert.ErrorAssertionFunc
	}{
		"nominal": {
			task: models.Task{
				Action: "deploy",
				Args:   []string{"civic-assistant", "/home/apps/civic-assistant"},
			},
			wantErr: assert.NoError,
		},
		"nominal, preprod": {
			task: models.Task{
				Action: "deploy",
				Args:   []string{"civic-assistant", "/home/apps/civic-assistant/preprod"},
			},
			wantErr: assert.NoError,
		},
		"unknown task, return error": {
			task: models.Task{
				Action: "unknown",
				Args:   []string{"aze", "aze"},
			},
			wantErr: assert.Error,
		},
		"invalid args, return error": {
			task: models.Task{
				Action: "deploy",
				Args:   []string{"aze", "aze"},
			},
			wantErr: assert.Error,
		},
		"no args, return error": {
			task: models.Task{
				Action: "deploy",
				Args:   nil,
			},
			wantErr: assert.Error,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := IsValidAction(tt.task)
			tt.wantErr(t, err)
		})
	}
}
