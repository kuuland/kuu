package kuu

import (
	"errors"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"strings"
	"time"
)

var (
	repeatEventProcesserMap = make(map[string]RepeatEventProcesser)
)

type RepeatEventProcesser func(*REContext, map[string]interface{}) error

func init() {
	Enum("RepeatEventStatus").
		Add("-1", "已失败").
		Add("0", "未完成").
		Add("1", "已完成")
}

type REContext struct {
	Event   RepeatEvent
	Current int
	Max     int
	Data    map[string]interface{}
}

func (context *REContext) Name() string {
	return context.Event.Name
}

type RepeatEvent struct {
	Model         `rest:"*" displayName:"重试事件"`
	Name          string     `name:"任务类型"`
	EventID       string     `name:"任务ID"`
	RetryInterval string     `name:"重试间隔"`
	NextTime      *time.Time `name:"下次重试时间"`
	RetryCount    int        `name:"重试次数"`
	Status        string     `name:"状态" enum:"RepeatEventStatus"`
	Message       string     `name:"错误消息" gorm:"type:text"`
	Data          string     `name:"上下文数据" gorm:"type:text"`
}

// ReigsterRepeatEventProcesser
func RegisterRepeatEventProcesser(name string, processer RepeatEventProcesser) {
	if _, has := repeatEventProcesserMap[name]; has {
		WARN("ReigsterRepeatEventProcesser: [%s] is duplicate.", name)
	} else {
		repeatEventProcesserMap[name] = processer
	}
}

// ReigsterRepeatEvent
func RegisterRepeatEvent(name string, interval string, data map[string]interface{}) error {
	intervals := strings.Split(interval, "/")
	if len(intervals) == 0 {
		return errors.New("interval can not be empty")
	}
	var duration time.Duration
	for _, s := range intervals {
		d, err := time.ParseDuration(s)
		if err != nil {
			return fmt.Errorf("interval format error: %w", err)
		}
		if duration == 0 {
			duration = d
		}
	}
	t := time.Now().Add(duration)
	event := RepeatEvent{
		Name:          name,
		EventID:       uuid.NewV4().String(),
		RetryInterval: interval,
		NextTime:      &t,
		RetryCount:    0,
		Status:        "0",
		Message:       "",
		Data:          JSONStringify(data, true),
	}
	return DB().Create(&event).Error
}

var TriggerRepeatEvent = RouteInfo{
	Name:   "触发执行可重复事件",
	Method: "GET",
	Path:   "/TriggerRepeatEvent",
	HandlerFunc: func(c *Context) *STDReply {
		c.IgnoreAuth()
		name := c.Query("name")
		q := DB().Model(&RepeatEvent{}).Where("status = 0 and next_time <= ?", time.Now())
		if name != "" {
			q = q.Where("name = ?", name)
		}
		var repeatEvents []RepeatEvent
		q.Find(&repeatEvents)
		for _, event := range repeatEvents {
			context := &REContext{Event: event}
			JSONParse(context.Event.Data, &context.Data)
			intervals := strings.Split(event.RetryInterval, "/")
			context.Max = len(intervals)
			context.Current = event.RetryCount + 1
			if processer, has := repeatEventProcesserMap[event.Name]; has {
				go processRepeatEvent(processer, context)
			}
		}
		return c.STDOK()
	},
}

func processRepeatEvent(processer RepeatEventProcesser, context *REContext) {
	err := processer(context, context.Data)
	event := context.Event
	update := map[string]interface{}{}
	if err != nil {
		update["message"] = event.Message + repeatEventWarperMsg(err.Error())
		intervals := strings.Split(event.RetryInterval, "/")
		if len(intervals) > event.RetryCount {
			update["retry_count"] = event.RetryCount + 1
			interval := intervals[event.RetryCount]
			d, _ := time.ParseDuration(interval)
			t := time.Now().Add(d)
			event.NextTime = &t
			update["next_time"] = &t
		} else {
			update["status"] = "-1"
			update["message"] = event.Message + repeatEventWarperMsg("全部重试失败，事件标记为已失败.")
		}
	} else {
		update["status"] = "1"
		update["message"] = event.Message + repeatEventWarperMsg("执行成功，事件标记为已完成.")
	}
	if err := DB().Model(&RepeatEvent{}).Where("id = ?", event.ID).Update(update).Error; err != nil {
		ERROR(err)
	}
}

func repeatEventWarperMsg(msg string) string {
	return fmt.Sprintf("%s: %s\n", time.Now().Format("2006-01-02 15:04:05.000000"), msg)
}
