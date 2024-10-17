package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tobg/scheduler/models"
)

func TestGetCronFrequency(t *testing.T) {

	tests := map[string]struct {
		j        models.Job
		wantErr  assert.ErrorAssertionFunc
		wantCron string
	}{
		"nominal, every minute": {
			j: models.Job{
				Frequency: "m",
			},
			wantErr:  assert.NoError,
			wantCron: "@every 1m",
		},
		"nominal, every hour": {
			j: models.Job{
				Frequency: "H",
			},
			wantErr:  assert.NoError,
			wantCron: "@hourly",
		},
		"nominal, every day": {
			j: models.Job{
				Frequency: "D",
			},
			wantErr:  assert.NoError,
			wantCron: "@daily",
		},
		"nominal, every week": {
			j: models.Job{
				Frequency: "W",
			},
			wantErr:  assert.NoError,
			wantCron: "@weekly",
		},
		"nominal, every month": {
			j: models.Job{
				Frequency: "M",
			},
			wantErr:  assert.NoError,
			wantCron: "@monthly",
		},
		"nominal, every year": {
			j: models.Job{
				Frequency: "Y",
			},
			wantErr:  assert.NoError,
			wantCron: "@yearly",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := GetCronFrequency(&tt.j)
			tt.wantErr(t, err)
			assert.Equal(t, tt.wantCron, tt.j.CronTime)
		})
	}
}
