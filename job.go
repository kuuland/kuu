package kuu

import (
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	"os"
	"sync"
)

// DefaultCron (set option 5 cron to convet 6 cron)
var DefaultCron = cron.New(cron.WithSeconds())

var (
	jobCmds   = make(map[cron.EntryID]bool)
	jobCmdsMu sync.RWMutex
)

// Job
type Job struct {
	Spec        string              `json:"spec" valid:"required"`
	Cmd         func(c *JobContext) `json:"-,omitempty" valid:"required"`
	Name        string              `json:"name" valid:"required"`
	RunAfterAdd bool                `json:"runAfterAdd"`
	EntryID     cron.EntryID        `json:"entryID,omitempty"`
}

// JobContext
type JobContext struct {
	name string
	errs []error
	l    *sync.RWMutex
}

func (c *JobContext) Error(err error) {
	c.l.Lock()
	defer c.l.Unlock()

	c.errs = append(c.errs, err)
}

// AddJobEntry
func AddJobEntry(j *Job) error {
	if os.Getenv("KUU_JOB") == "" {
		return nil
	}

	if _, err := govalidator.ValidateStruct(j); err != nil {
		return err
	}

	cmd := func() {
		jobCmdsMu.Lock()
		defer jobCmdsMu.Unlock()

		if jobCmds[j.EntryID] {
			return
		}
		jobCmds[j.EntryID] = true
		INFO("----------- Job '%s' start -----------", j.Name)

		c := &JobContext{
			name: j.Name,
			l:    new(sync.RWMutex),
		}
		j.Cmd(c)
		if len(c.errs) > 0 {
			for i, err := range c.errs {
				c.errs[i] = errors.Wrap(err, fmt.Sprintf("Job '%s' execute error", j.Name))
			}
			ERROR(c.errs)
		}
		INFO("----------- Job '%s' finish -----------", j.Name)
		jobCmds[j.EntryID] = false
	}
	if j.RunAfterAdd {
		cmd()
	}
	v, err := DefaultCron.AddFunc(j.Spec, cmd)
	if err == nil {
		j.EntryID = v
	}
	return err
}

// AddJob
func AddJob(spec string, name string, cmd func(c *JobContext)) (cron.EntryID, error) {
	job := Job{
		Spec: spec,
		Name: name,
		Cmd:  cmd,
	}
	err := AddJobEntry(&job)
	return job.EntryID, err
}
