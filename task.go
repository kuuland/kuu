package kuu

import "github.com/robfig/cron/v3"

// DefaultCron
var DefaultCron = cron.New()

// Task
type Task struct {
	Spec string
	Cmd  func()
}

// AddTask
func AddTask(spec string, cmd func()) (cron.EntryID, error) {
	return DefaultCron.AddFunc(spec, cmd)
}
