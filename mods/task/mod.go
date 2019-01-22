package task

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kuuland/kuu"
	"github.com/robfig/cron"
)

// 任务运行模式
const (
	SerialMode   = iota // 串行
	ParallelMode        // 并行
)

var (
	// Cron 调度器实例
	Cron = cron.New()
	// Tasks 任务列表
	Tasks = map[string]*Task{}
)

func init() {
	kuu.On("OnNew", func(args ...interface{}) {
		k := args[0].(*kuu.Kuu)
		if url := k.Config["taskURL"]; url != nil {
			// 支持通过URL加载远程任务列表
			loadTasksFromURL(url.(string))
			Cron.Start()
		}
	})
}

// Task 任务
type Task struct {
	Name    string `json:"name" displayName:"任务名称"`      // 任务名称，为空时与Spec一致
	Spec    string `json:"spec" displayName:"任务cron表达式"` // 规则详见：https://godoc.org/github.com/robfig/cron
	Cmd     func() `json:"-" displayName:"任务触发函数"`       // 任务触发时，调用此函数，与URL不能同时使用
	URL     string `json:"url" displayName:"远程任务URL"`    // 任务被触发时，调用此URL
	Running bool   `json:"running" displayName:"是否正在运行"` // 任务正在运行时，该值为true
	Mode    int    `json:"mode" displayName:"任务运行模式"`    // 当任务频率过高时，需选择正确的运行模式：串行表示下一次的任务执行必须在上一次执行结束后才会触发，并行无此限制。
}

func loadTasksFromURL(taskURL string) {
	body := fetch(taskURL)
	var data []Task
	json.Unmarshal(body, &data)
	for _, task := range data {
		Add(&task)
	}
}

func fetch(path string) []byte {
	split := strings.Split(path, " ")
	if len(split) != 2 {
		return nil
	}
	var (
		method = split[0]
		url    = split[1]
		client = &http.Client{}
		body   []byte
	)
	if req, err := http.NewRequest(method, url, nil); err == nil {
		if resp, err := client.Do(req); err == nil {
			defer resp.Body.Close()
			body, _ = ioutil.ReadAll(resp.Body)
		}
	}
	return body
}

// Add 添加任务
func Add(ts ...*Task) error {
	if ts == nil || len(ts) == 0 {
		return nil
	}
	for _, t := range ts {
		if t.Cmd == nil && t.URL == "" {
			continue
		} else if t.URL != "" {
			if len(strings.Split(t.URL, " ")) != 2 {
				continue
			}
			t.Cmd = func() {
				fetch(t.URL)
			}
		}
		cmd := t.Cmd
		t.Cmd = func() {
			if t.Running == true && t.Mode == SerialMode {
				return
			}
			t.Running = true
			cmd()
			t.Running = false
		}
		err := Cron.AddFunc(t.Spec, t.Cmd)
		if err != nil {
			return err
		}
		if t.Name != "" {
			Tasks[t.Name] = t
		}
	}
	return nil
}

// AddFunc 快捷调用
func AddFunc(spec string, cmd func()) error {
	return Add(&Task{
		Name: spec,
		Spec: spec,
		Cmd:  cmd,
	})
}

// TasksHandler 任务列表路由
func TasksHandler(c *gin.Context) {
	c.JSON(http.StatusOK, kuu.StdOK(Tasks))
}

// All 模块声明
func All() *kuu.Mod {
	return &kuu.Mod{
		Models: []interface{}{
			&Task{},
		},
		Routes: kuu.Routes{
			kuu.RouteInfo{
				Method:  "GET",
				Path:    "/tasks",
				Handler: TasksHandler,
			},
		},
	}
}
