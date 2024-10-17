package models

import "time"

// Type Job represents a job to run composed of multiple tasks
type Job struct {
	ID           int       `json:"id"`
	Schedule     time.Time `json:"schedule,omitempty"`      // "DD-MM-YYY HH:MM"
	UserSchedule string    `json:"user_schedule,omitempty"` // "DD-MM-YYY HH:MM" in string
	Occurrences  int       `json:"occurrences"`
	Label        string    `json:"label"`
	Frequency    string    `json:"frequency"`
	Workflow     []Task    `json:"workflow"`
	CreatedAt    time.Time `json:"created_at"`

	CronTime  string `json:"-"`
	IsOneTime bool   `json:"-"`
}

// Task represent a single unit of work in a workflow
type Task struct {
	JobID  int      `json:"id"`
	Action string   `json:"action"`
	Args   []string `json:"args"`
}

// TaskHandler is used to create the execute function and verify function
// of each task
type TaskHandler struct {
	Execute ActionFunc
	Verify  VerifyFunc
}

// types of functions used in TaskHandler
type ActionFunc func([]string) error
type VerifyFunc func([]string) error
