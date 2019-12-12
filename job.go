package kuu

import "github.com/robfig/cron/v3"

// DefaultCron (set option 5 cron to convet 6 cron)
var DefaultCron = cron.New(cron.WithSeconds())

// Job
type Job struct {
	Spec string
	Cmd  func()
}

// AddJob
func AddJob(spec string, cmd func()) (cron.EntryID, error) {
	return DefaultCron.AddFunc(spec, cmd)
}
