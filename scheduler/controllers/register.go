package controllers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/tobg/scheduler/helpers"
	"github.com/tobg/scheduler/helpers/validations"
	"github.com/tobg/scheduler/usecases"
)

// RegisterController represents a register controller
type RegisterController struct {
	ru usecases.RegisterInterface
}

// NewRegisterController returns a register controller
func NewRegisterController(ru usecases.RegisterInterface) *RegisterController {
	return &RegisterController{
		ru: ru,
	}
}

// Register save a job to execute at specific time
func (rc *RegisterController) Register(w http.ResponseWriter, r *http.Request) {
	err := validations.IsMethodAllowed(r.Method, http.MethodPost)
	if err != nil {
		helpers.SendResponseMessage(w, http.StatusMethodNotAllowed, fmt.Sprintf("invalid method: %v, POST method allowed only", r.Method))
		return
	}

	job, err := rc.ru.ParseBody(r)
	if err != nil {
		helpers.SendResponseMessage(w, http.StatusBadRequest, fmt.Errorf("could not parse body: %w", err).Error())
		return
	}

	err = rc.ru.ValidateJob(job)
	if err != nil {
		helpers.SendResponseMessage(w, http.StatusBadRequest, fmt.Errorf("could not validate body: %w", err).Error())
		return
	}

	tul, err := rc.ru.VerifyDate(job.Schedule)
	if err != nil {
		helpers.SendResponseMessage(w, http.StatusBadRequest, fmt.Errorf("date is invalid: %w", err).Error())
		return
	}

	err = rc.ru.SetCronFrequency(&job)
	if err != nil {
		helpers.SendResponseMessage(w, http.StatusInternalServerError, fmt.Errorf("could not create cron frequency: %w", err).Error())
		return
	}

	jobID, err := rc.ru.RegisterJob(job, tul, false)
	if err != nil {
		helpers.SendResponseMessage(w, http.StatusInternalServerError, fmt.Errorf("could not register job: %w", err).Error())
		return
	}

	rc.ru.CleanPayload(&job, jobID)

	log.Printf("\n job '%v' received - id %d \n time until starts: %v \n frequency: %v \n", job.Label, jobID, time.Until(job.Schedule), job.Frequency)

	helpers.SendResponseData(w, http.StatusOK, job)
}

func (rc *RegisterController) GetJobs(w http.ResponseWriter, r *http.Request) {
	err := validations.IsMethodAllowed(r.Method, http.MethodGet)
	if err != nil {
		helpers.SendResponseMessage(w, http.StatusMethodNotAllowed, fmt.Sprintf("invalid method: %v, POST method allowed only", r.Method))
		return
	}

	jobs, err := rc.ru.GetJobs()
	if err != nil {
		helpers.SendResponseMessage(w, http.StatusInternalServerError, fmt.Errorf("could not retrieve jobs: %w", err).Error())
		return
	}

	helpers.SendResponseData(w, http.StatusOK, jobs)
}

func (rc *RegisterController) ReloadJobs() error {
	err := rc.ru.ReloadJobs()
	if err != nil {
		return err
	}

	return nil
}
