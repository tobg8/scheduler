package repositories

import (
	"database/sql"
	_ "embed"
	"fmt"
	"log"
	"strings"

	"github.com/tobg/scheduler/models"
)

//go:embed queries/insert_job.sql
var insertJob string

//go:embed queries/insert_tasks.sql
var insertTasks string

//go:embed queries/get_job_by_id.sql
var getJob string

//go:embed queries/delete_job_by_id.sql
var deleteJob string

//go:embed queries/update_job_occurences.sql
var decrementOccurrences string

//go:embed queries/get_jobs.sql
var getJobs string

type RegisterRepository struct {
	db *sql.DB
}

type RegisterInterface interface {
	RegisterJob(j models.Job) (int, error)
	RetrieveJob(id int) (models.Job, error)
	DeleteJob(id int) error
	DecrementJobOccurrences(id int) (int, error)
	RetrieveJobs() ([]models.Job, error)
}

func NewRegisterRepository(db *sql.DB) *RegisterRepository {
	return &RegisterRepository{
		db: db,
	}
}

func (rr *RegisterRepository) RegisterJob(j models.Job) (int, error) {
	tx, err := rr.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("could not begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	result, err := tx.Exec(insertJob, j.Schedule.Local(), j.UserSchedule, j.Occurrences, j.Frequency, j.Label, j.CronTime)
	if err != nil {
		return 0, fmt.Errorf("could not insert job: %w", err)
	}

	jobID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("could not get last inserted job ID: %w", err)
	}

	for _, v := range j.Workflow {
		args := strings.Join(v.Args, ",")

		_, err := tx.Exec(insertTasks, jobID, v.Action, args)
		if err != nil {
			return 0, fmt.Errorf("could not insert tasks: %w", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return 0, fmt.Errorf("could not commit transaction: %w", err)
	}

	return int(jobID), nil
}

func (rr *RegisterRepository) RetrieveJob(id int) (models.Job, error) {
	var j models.Job
	rows, err := rr.db.Query(getJob, id)
	if err != nil {
		return models.Job{}, fmt.Errorf("could not retrieve job: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var task models.Task
		var action sql.NullString
		var args sql.NullString

		err := rows.Scan(
			&j.ID,
			&j.Schedule,
			&j.UserSchedule,
			&j.Occurrences,
			&j.Frequency,
			&j.Label,
			&j.CreatedAt,

			&action,
			&args,
		)
		if err != nil {
			return models.Job{}, fmt.Errorf("could not scan job row: %w", err)
		}

		if action.Valid {
			task.Action = action.String
			task.Args = strings.Split(args.String, ",")
			task.JobID = j.ID
			j.Workflow = append(j.Workflow, task)
		}
	}

	if err := rows.Err(); err != nil {
		return models.Job{}, fmt.Errorf("error during rows iteration: %w", err)
	}

	if len(j.Workflow) == 0 {
		return models.Job{}, fmt.Errorf("no job found with id: %d", id)
	}

	return j, nil
}

func (rr *RegisterRepository) DeleteJob(id int) error {
	_, err := rr.db.Exec(deleteJob, id)
	if err != nil {
		return fmt.Errorf("could not delete job: %w", err)
	}

	log.Print("deleted")

	return nil
}

func (rr *RegisterRepository) DecrementJobOccurrences(id int) (int, error) {
	var occurrencesLeft int

	row := rr.db.QueryRow(decrementOccurrences, id)
	err := row.Scan(&occurrencesLeft)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("no job found with id: %d", id)
		}
		return 0, fmt.Errorf("could not update occurrences: %w", err)

	}

	return occurrencesLeft, nil
}

func (rr *RegisterRepository) RetrieveJobs() ([]models.Job, error) {
	jobMap := make(map[int]*models.Job)
	var jobs []models.Job

	rows, err := rr.db.Query(getJobs)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve jobs: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var j models.Job
		var t models.Task
		var action sql.NullString
		var args sql.NullString

		err := rows.Scan(
			&j.ID,
			&j.Schedule,
			&j.UserSchedule,
			&j.Occurrences,
			&j.Frequency,
			&j.Label,
			&j.CreatedAt,
			&action,
			&args,
		)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve jobs: %w", err)
		}

		if existingJob, exists := jobMap[j.ID]; exists {
			if action.Valid {
				t.Action = action.String
				t.Args = strings.Split(args.String, ",")
				t.JobID = j.ID
				existingJob.Workflow = append(existingJob.Workflow, t)
			}
		} else {
			if action.Valid {
				t.Action = action.String
				t.Args = strings.Split(args.String, ",")
				t.JobID = j.ID
				j.Workflow = append(j.Workflow, t)
			}
			jobMap[j.ID] = &j
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error retrieving jobs: %w", err)
	}

	for _, job := range jobMap {
		jobs = append(jobs, *job)
	}

	if len(jobs) == 0 {
		log.Print("no jobs")
		return nil, nil
	}

	return jobs, nil
}
