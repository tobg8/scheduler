package usecases

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/robfig/cron"
	"github.com/tobg/scheduler/helpers"
	"github.com/tobg/scheduler/helpers/validations"
	"github.com/tobg/scheduler/models"
	"github.com/tobg/scheduler/repositories"
)

// RegisterUsecase represents a register controller
type RegisterUsecase struct {
	rr repositories.RegisterInterface
}

type RegisterInterface interface {
	ParseBody(r *http.Request) (models.Job, error)
	ValidateJob(j models.Job) error
	RegisterJob(j models.Job, timeUntilStart time.Duration, isReload bool) (int, error)
	VerifyDate(t time.Time) (time.Duration, error)
	SetCronFrequency(j *models.Job) error
	CleanPayload(j *models.Job, id int)
	GetJobs() ([]models.Job, error)
	ReloadJobs() error
}

// NewRegisterUsecase returns a register usecase
func NewRegisterUsecase(rr repositories.RegisterInterface) *RegisterUsecase {
	return &RegisterUsecase{
		rr: rr,
	}
}

// ParseBody reads body from request and create a Job
func (ru *RegisterUsecase) ParseBody(r *http.Request) (models.Job, error) {
	var job models.Job

	if r.Body == nil {
		return models.Job{}, errors.New("empty request body")
	}

	err := json.NewDecoder(r.Body).Decode(&job)
	if err != nil {
		return models.Job{}, err
	}

	location, err := time.LoadLocation("Local")
	if err != nil {
		return models.Job{}, errors.New("could not add local env timezone")
	}

	// Parse the Schedule string to proper time format "DD-MM-YYYY HH:MM"
	parsedTime, err := time.ParseInLocation("02-01-2006 15:04", job.UserSchedule, location)
	if err != nil {
		return models.Job{}, errors.New("invalid schedule format; expected DD-MM-YYYY HH:MM")
	}

	// Add date in UTC time format to job
	job.Schedule = parsedTime.UTC()
	job.IsOneTime = job.Occurrences == 1

	return job, nil
}

// ValidateJob checks if the job is valid
func (ru *RegisterUsecase) ValidateJob(j models.Job) error {
	err := validations.IsValidFrequency(j.Frequency)
	if err != nil {
		return err
	}

	err = validations.IsValidOccurrences(j.Occurrences)
	if err != nil {
		return err
	}

	for _, v := range j.Workflow {
		err := validations.IsValidAction(v)
		if err != nil {
			return err
		}
	}

	if j.Label == "" {
		return fmt.Errorf("please provide label to the job")
	}

	return nil
}

func (ru *RegisterUsecase) VerifyDate(t time.Time) (time.Duration, error) {
	timeUntilStart := time.Until(t)
	if timeUntilStart < 0 {
		return 0, fmt.Errorf("provided date is in past")
	}
	return timeUntilStart, nil
}

func (ru *RegisterUsecase) SetCronFrequency(j *models.Job) error {
	if j.Occurrences > 0 || j.Occurrences == -1 {
		err := helpers.GetCronFrequency(j)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ru *RegisterUsecase) CleanPayload(j *models.Job, id int) {
	j.CronTime = ""
	j.ID = id
	j.CreatedAt = time.Now()

	for i := range j.Workflow {
		j.Workflow[i].JobID = id // Assuming Task struct has JobID field
	}
}

// RegisterJob registers a new job
func (ru *RegisterUsecase) RegisterJob(j models.Job, timeUntilStart time.Duration, isReload bool) (int, error) {
	var jobID int
	var err error

	// save job in db
	if !isReload {
		jobID, err = ru.rr.RegisterJob(j)
		if err != nil {
			return 0, err
		}
		j.ID = jobID
	} else {
		// To ensure we have a cronFrequency on reload job
		// even if the job has been cut before triggering once.
		err := ru.SetCronFrequency(&j)
		if err != nil {
			return 0, fmt.Errorf("could not reset cron frequency")
		}

		// it's a reload we already have it in db
		jobID = j.ID
	}

	if timeUntilStart > 0 {
		// register a go routine that'll trigger at job start time
		time.AfterFunc(timeUntilStart, func() {
			log.Printf("registering job: %v", j.ID)
			err := handleJob(&j, ru.rr)
			if err != nil {
				log.Printf("could not register job: %v", err)
			}
		})
	}
	return jobID, nil
}

func (ru *RegisterUsecase) GetJobs() ([]models.Job, error) {
	jobs, err := ru.rr.RetrieveJobs()
	if err != nil {
		return nil, err
	}

	return jobs, err
}

func (ru *RegisterUsecase) ReloadJobs() error {
	jobs, err := ru.GetJobs()
	if err != nil {
		return fmt.Errorf("could not retrieve jobs: %w", err)
	}

	now := time.Now()

	for _, job := range jobs {
		if job.Schedule.Before(now) {
			nextSchedule := calculateNextValidSchedule(&job)
			timeUntilStart := time.Until(nextSchedule)
			job.Schedule = nextSchedule.Local()

			err := ru.SetCronFrequency(&job)
			if err != nil {
				return fmt.Errorf("could not set cron time on reload jobs: %w", err)
			}

			log.Printf("\n job '%v' - %d rescheduled to run at %v (in %v)", job.Label, job.ID, nextSchedule, timeUntilStart)
			ru.RegisterJob(job, timeUntilStart, true)
		} else {
			timeUntilStart := time.Until(job.Schedule)
			log.Printf("job %d scheduled to run at %v (in %v)", job.ID, job.Schedule.Local(), timeUntilStart)
			ru.RegisterJob(job, timeUntilStart, true)
		}
	}

	return nil
}

func calculateNextValidSchedule(j *models.Job) time.Time {
	currentTime := time.Now()
	scheduledTime := j.Schedule

	var nextRun time.Time

	hour := scheduledTime.Hour()
	minute := scheduledTime.Minute()
	day := scheduledTime.Day()
	weekday := scheduledTime.Weekday()
	month := scheduledTime.Month()

	switch j.Frequency {
	case "m": // Every minute
		nextRun = currentTime.Truncate(time.Minute).Add(time.Minute)

	case "H": // Every hour
		nextRun = currentTime.Truncate(time.Hour).Add(time.Hour).Add(time.Duration(minute) * time.Minute)

	case "D": // Daily
		nextRun = time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), hour, minute, 0, 0, scheduledTime.Location())
		if currentTime.After(nextRun) {
			nextRun = nextRun.Add(24 * time.Hour)
		}

	case "W": // Weekly
		nextRun = time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), hour, minute, 0, 0, scheduledTime.Location())
		if currentTime.Weekday() > weekday || (currentTime.Weekday() == weekday && currentTime.After(nextRun)) {
			nextRun = nextRun.AddDate(0, 0, 7)
		}

	case "M": // Monthly
		nextRun = time.Date(currentTime.Year(), currentTime.Month(), day, hour, minute, 0, 0, scheduledTime.Location())
		if currentTime.After(nextRun) {
			nextRun = nextRun.AddDate(0, 1, 0)
		}

	case "Y": // Yearly
		nextRun = time.Date(currentTime.Year(), month, day, hour, minute, 0, 0, scheduledTime.Location())
		if currentTime.After(nextRun) {
			nextRun = nextRun.AddDate(1, 0, 0)
		}

	default:
		log.Printf("unsupported frequency: %s", j.Frequency)
		return time.Time{}
	}

	return nextRun.Local()
}

// JobHandler handles job execution and management
type JobHandler struct {
	j  *models.Job
	cs *cron.Cron
	rr repositories.RegisterInterface
}

// NewJobHandler creates a new JobHandler
func NewJobHandler(j *models.Job, cs *cron.Cron, rr repositories.RegisterInterface) *JobHandler {
	return &JobHandler{
		j:  j,
		cs: cs,
		rr: rr,
	}
}

// handleJob runs the job and manages its scheduling
func handleJob(j *models.Job, rr repositories.RegisterInterface) error {
	cronJob := cron.New()
	job := NewJobHandler(j, cronJob, rr)
	job.Run()

	if !j.IsOneTime {
		err := job.cs.AddJob(j.CronTime, job)
		if err != nil {
			return err
		}
		log.Printf("scheduled job --  %v every %v", j.ID, j.Frequency)
		job.cs.Start()
	}

	return nil
}

// Run executes the job and manages its occurrences
func (j *JobHandler) Run() {
	log.Printf("run job -- %v", j.j.ID)

	for _, v := range j.j.Workflow {
		log.Printf("run task -- %v on job %v", v.Action, v.JobID)
	}

	if j.j.Occurrences != -1 {
		occurrencesLeft, err := j.rr.DecrementJobOccurrences(j.j.ID)
		if err != nil {
			log.Printf("could not decrement Occurrences: %v with error: %v", j.j.ID, err)
			return
		}

		log.Printf("occurrences decrement -- job %v occurrences: %d", j.j.ID, occurrencesLeft)
		if occurrencesLeft == 0 {
			log.Printf("job deletion -- %v", j.j.ID)
			err := j.rr.DeleteJob(j.j.ID)
			if err != nil {
				log.Printf("could not delete job: %v with error: %v", j.j.ID, err)
				return
			}
			j.cs.Stop()
		}
	}
}
