package kuu

// Hook 钩子接口
type Hook interface {
	Fire(*Kuu, ...interface{}) error
}

// hooks 钩子集合
var hooks = map[string][]Hook{}

// AddHook 添加钩子
func AddHook(when string, hook Hook) {
	hooks[when] = append(hooks[when], hook)
}

// FireHooks 触发钩子
func FireHooks(when string, k *Kuu, args ...interface{}) error {
	for _, hook := range hooks[when] {
		if err := hook.Fire(k, args...); err != nil {
			return err
		}
	}
	return nil
}
