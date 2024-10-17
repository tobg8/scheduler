package helpers

import (
	"fmt"

	"github.com/tobg/scheduler/models"
)

func GetCronFrequency(j *models.Job) error {
	var cronExpr string

	switch j.Frequency {
	case "m": // minute
		cronExpr = "@every 1m"
	case "H": // hourly
		cronExpr = "@hourly"
	case "D": // daily
		cronExpr = "@daily"
	case "W": // weekly
		cronExpr = "@weekly"
	case "M": // monthly
		cronExpr = "@monthly"
	case "Y": // yearly
		cronExpr = "@yearly"
	default:
		return fmt.Errorf("invalid frequency: %s", j.Frequency)
	}

	j.CronTime = cronExpr
	return nil
}
