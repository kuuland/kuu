package kuu

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	"os"
	"sync"
)

// DefaultCron (set option 5 cron to convet 6 cron)
var DefaultCron = cron.New(cron.WithSeconds())

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

// AddJob
func AddJob(spec string, name string, cmd func(c *JobContext)) (cron.EntryID, error) {
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

		c := &JobContext{
			name: name,
			l:    new(sync.RWMutex),
		}
		cmd(c)
		if len(c.errs) > 0 {
			for i, err := range c.errs {
				c.errs[i] = errors.Wrap(err, fmt.Sprintf("Job '%s' execute error", name))
			}
			ERROR(c.errs)
		}
		INFO("----------- Job '%s' finish -----------", name)
	})
}
