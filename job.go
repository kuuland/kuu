package kuu

import "github.com/robfig/cron/v3"

// DefaultCron
var DefaultCron = cron.New()

// Job
type Job struct {
	Spec string
	Cmd  func()
}

// AddJob
func AddJob(spec string, cmd func()) (cron.EntryID, error) {
	return DefaultCron.AddFunc(spec, cmd)
}
