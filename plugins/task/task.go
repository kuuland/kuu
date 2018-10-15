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

var c = cron.New()

// 任务运行模式
const (
	SerialMode   = iota // 线性
	ParallelMode        // 并行
)

// Task 任务
type Task struct {
	Name    string `json:"name"`
	Spec    string `json:"spec"`
	Func    func() `json:"-"`
	URL     string `json:"url"`
	RunInit int    `json:"runInit"`
	Running bool   `json:"running"`
	Mode    int    `json:"mode"`
}

// Tasks 任务实例
var Tasks = map[string]*Task{}

// loadTasksFromURL 从远程URL获取任务配置
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
func Add(ts ...*Task) {
	if ts == nil || len(ts) == 0 {
		return
	}
	for _, t := range ts {
		if t.Func == nil && t.URL == "" {
			continue
		} else if t.URL != "" {
			if len(strings.Split(t.URL, " ")) != 2 {
				continue
			}
			t.Func = func() {
				fetch(t.URL)
			}
		}
		if t.Func != nil {
			fn := t.Func
			t.Func = func() {
				if t.Running == true && t.Mode == SerialMode {
					return
				}
				t.Running = true
				fn()
				t.Running = false
			}
			c.AddFunc(t.Spec, t.Func)
			if t.Name != "" {
				Tasks[t.Name] = t
			}
		}
	}
}

// P 插件声明
var P = &kuu.Plugin{
	Name: "task",
	OnLoad: func(k *kuu.Kuu) {
		var taskURL string
		if k.Config["taskURL"] != nil {
			taskURL = k.Config["taskURL"].(string)
			loadTasksFromURL(taskURL)
			c.Start()
		}
	},
	Methods: kuu.Methods{
		"add": func(args ...interface{}) interface{} {
			if args != nil && len(args) > 0 {
				for _, item := range args {
					if item == nil {
						continue
					}
					t := item.(*Task)
					Add(t)
				}
			}
			return nil
		},
		"tasks": func(args ...interface{}) interface{} {
			return Tasks
		},
	},
	Routes: kuu.R{
		"list": &kuu.Route{
			Method: "GET",
			Path:   "/tasks",
			Handler: func(c *gin.Context) {
				c.JSON(http.StatusOK, Tasks)
			},
		},
	},
}
