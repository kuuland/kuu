package kuu

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	"os"
)

// DefaultCron (set option 5 cron to convet 6 cron)
var DefaultCron = cron.New(cron.WithSeconds())

// Job
type Job struct {
	Spec string
	Cmd  func()
}

// AddJob
func AddJob(spec string, name string, cmd func() error) (cron.EntryID, error) {
	return DefaultCron.AddFunc(spec, func() {
		if v := os.Getenv("JOB_PARALLEL"); v == "" {
			// 基于缓存进行拦截，避免多实例重复执行
			k := fmt.Sprintf("job_%s_running", name)
			if GetCacheString(k) != "" {
				return
			}
			SetCacheString(k, "1")
			defer DelCache(k)
		}

		INFO("----------- Job '%s' start -----------", name)
		if err := cmd(); err != nil {
			ERROR(errors.Wrap(err, fmt.Sprintf("Job '%s'", name)))
		}
		INFO("----------- Job '%s' finish -----------", name)
	})
}
