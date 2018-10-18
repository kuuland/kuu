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
	SerialMode   = iota // 线性
	ParallelMode        // 并行
)

var (
	c     = cron.New()
	tasks = map[string]*Task{}
)

func init() {
	kuu.Emit("OnNew", func(args ...interface{}) {
		k := args[0].(*kuu.Kuu)
		if url := k.Config["taskURL"]; url != nil {
			loadTasksFromURL(url.(string))
			c.Start()
		}
	})
}

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
				tasks[t.Name] = t
			}
		}
	}
}

// TasksHandler 任务列表路由
func TasksHandler(c *gin.Context) {
	c.JSON(http.StatusOK, kuu.StdOK(tasks))
}

// All 插件声明
func All() *kuu.Plugin {
	return &kuu.Plugin{
		Routes: kuu.Routes{
			kuu.RouteInfo{
				Method:  "GET",
				Path:    "/tasks",
				Handler: TasksHandler,
			},
		},
	}
}
