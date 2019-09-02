package kuu

import "github.com/jinzhu/gorm"

const (
	// BizCreateKind
	BizCreateKind = "create"
	// BizUpdateKind
	BizUpdateKind = "update"
	// BizDeleteKind
	BizDeleteKind = "delete"
	// BizQueryKind
	BizQueryKind = "query"
)

// DefaultCallback
var DefaultCallback = &Callback{}

// Callback
type Callback struct {
	creates    []*func(scope *Scope)
	updates    []*func(scope *Scope)
	deletes    []*func(scope *Scope)
	queries    []*func(scope *Scope)
	processors []*CallbackProcessor
}

// CallbackProcessor contains callback informations
type CallbackProcessor struct {
	name      string              // current callback's name
	before    string              // register current callback before a callback
	after     string              // register current callback after a callback
	replace   bool                // replace callbacks with same name
	remove    bool                // delete callbacks with same name
	kind      string              // callback type: create, update, delete, query
	processor *func(scope *Scope) // callback handler
	parent    *Callback
}

// Create
func (c *Callback) Create() *CallbackProcessor {
	return &CallbackProcessor{kind: "create", parent: c}
}

// Update
func (c *Callback) Update() *CallbackProcessor {
	return &CallbackProcessor{kind: "update", parent: c}
}

// Delete
func (c *Callback) Delete() *CallbackProcessor {
	return &CallbackProcessor{kind: "delete", parent: c}
}

// Query
func (c *Callback) Query() *CallbackProcessor {
	return &CallbackProcessor{kind: "query", parent: c}
}

// After insert a new callback after callback `callbackName`, refer `Callbacks.Create`
func (cp *CallbackProcessor) After(callbackName string) *CallbackProcessor {
	cp.after = callbackName
	return cp
}

// Before insert a new callback before callback `callbackName`, refer `Callbacks.Create`
func (cp *CallbackProcessor) Before(callbackName string) *CallbackProcessor {
	cp.before = callbackName
	return cp
}

// Register a new callback, refer `Callbacks.Create`
func (cp *CallbackProcessor) Register(callbackName string, callback func(scope *Scope)) {
	cp.name = callbackName
	cp.processor = &callback
	cp.parent.processors = append(cp.parent.processors, cp)
	cp.parent.reorder()
}

// Remove a registered callback
//     db.Callback().Create().Remove("gorm:update_time_stamp_when_create")
func (cp *CallbackProcessor) Remove(callbackName string) {
	INFO("removing callback `%v` from %v\n", callbackName, fileWithLineNum())
	cp.name = callbackName
	cp.remove = true
	cp.parent.processors = append(cp.parent.processors, cp)
	cp.parent.reorder()
}

// Replace a registered callback with new callback
//     db.Callback().Create().Replace("gorm:update_time_stamp_when_create", func(*Scope) {
//		   scope.SetColumn("Created", now)
//		   scope.SetColumn("Updated", now)
//     })
func (cp *CallbackProcessor) Replace(callbackName string, callback func(scope *Scope)) {
	INFO("replacing callback `%v` from %v", callbackName, fileWithLineNum())
	cp.name = callbackName
	cp.processor = &callback
	cp.replace = true
	cp.parent.processors = append(cp.parent.processors, cp)
	cp.parent.reorder()
}

// Get registered callback
//    db.Callback().Create().Get("gorm:create")
func (cp *CallbackProcessor) Get(callbackName string) (callback func(scope *Scope)) {
	for _, p := range cp.parent.processors {
		if p.name == callbackName && p.kind == cp.kind && !cp.remove {
			return *p.processor
		}
	}
	return nil
}

// getRIndex get right index from string slice
func getRIndex(strs []string, str string) int {
	for i := len(strs) - 1; i >= 0; i-- {
		if strs[i] == str {
			return i
		}
	}
	return -1
}

// sortProcessors sort callback processors based on its before, after, remove, replace
func sortProcessors(cps []*CallbackProcessor) []*func(scope *Scope) {
	var (
		allNames, sortedNames []string
		sortCallbackProcessor func(c *CallbackProcessor)
	)

	for _, cp := range cps {
		// show warning message the callback name already exists
		if index := getRIndex(allNames, cp.name); index > -1 && !cp.replace && !cp.remove {
			WARN("duplicated callback `%v` from %v", cp.name, fileWithLineNum())
		}
		allNames = append(allNames, cp.name)
	}

	sortCallbackProcessor = func(c *CallbackProcessor) {
		if getRIndex(sortedNames, c.name) == -1 { // if not sorted
			if c.before != "" { // if defined before callback
				if index := getRIndex(sortedNames, c.before); index != -1 {
					// if before callback already sorted, append current callback just after it
					sortedNames = append(sortedNames[:index], append([]string{c.name}, sortedNames[index:]...)...)
				} else if index := getRIndex(allNames, c.before); index != -1 {
					// if before callback exists but haven't sorted, append current callback to last
					sortedNames = append(sortedNames, c.name)
					sortCallbackProcessor(cps[index])
				}
			}

			if c.after != "" { // if defined after callback
				if index := getRIndex(sortedNames, c.after); index != -1 {
					// if after callback already sorted, append current callback just before it
					sortedNames = append(sortedNames[:index+1], append([]string{c.name}, sortedNames[index+1:]...)...)
				} else if index := getRIndex(allNames, c.after); index != -1 {
					// if after callback exists but haven't sorted
					cp := cps[index]
					// set after callback's before callback to current callback
					if cp.before == "" {
						cp.before = c.name
					}
					sortCallbackProcessor(cp)
				}
			}

			// if current callback haven't been sorted, append it to last
			if getRIndex(sortedNames, c.name) == -1 {
				sortedNames = append(sortedNames, c.name)
			}
		}
	}

	for _, cp := range cps {
		sortCallbackProcessor(cp)
	}

	var sortedFuncs []*func(scope *Scope)
	for _, name := range sortedNames {
		if index := getRIndex(allNames, name); !cps[index].remove {
			sortedFuncs = append(sortedFuncs, cps[index].processor)
		}
	}

	return sortedFuncs
}

// reorder all registered processors, and reset CRUD callbacks
func (c *Callback) reorder() {
	var creates, updates, deletes, queries []*CallbackProcessor

	for _, processor := range c.processors {
		if processor.name != "" {
			switch processor.kind {
			case BizCreateKind:
				creates = append(creates, processor)
			case BizUpdateKind:
				updates = append(updates, processor)
			case BizDeleteKind:
				deletes = append(deletes, processor)
			case BizQueryKind:
				queries = append(queries, processor)
			}
		}
	}

	c.creates = sortProcessors(creates)
	c.updates = sortProcessors(updates)
	c.deletes = sortProcessors(deletes)
	c.queries = sortProcessors(queries)
}

func createOrUpdateItem(scope *Scope, item interface{}) {
	tx := scope.DB
	if tx.NewRecord(item) {
		if err := tx.Create(item).Error; err != nil {
			_ = scope.Err(err)
			return
		}
	} else {
		itemScope := tx.NewScope(item)
		if field, ok := itemScope.FieldByName("DeletedAt"); ok && !field.IsBlank {
			if err := tx.Delete(item).Error; err != nil {
				_ = scope.Err(err)
				return
			}
		} else {
			if err := tx.Model(item).Updates(item).Error; err != nil {
				_ = scope.Err(err)
				return
			}
		}
	}
}

func checkCreateOrUpdateField(scope *Scope, field *gorm.Field) {
	if field.Relationship != nil && !field.IsBlank {
		switch field.Relationship.Kind {
		case "has_many", "many_to_many":
			for i := 0; i < field.Field.Len(); i++ {
				createOrUpdateItem(scope, field.Field.Index(i).Addr().Interface())
			}
		case "has_one", "belongs_to":
			createOrUpdateItem(scope, field.Field.Addr().Interface())
		}
	}
}
