package route

const (
	String  = "string"
	Integer = "integer"
	Number  = "number"
	Boolean = "boolean"
	Object  = "object"
	Array   = "array"
)

type Param struct {
	Name        string
	Type        string
	Ref         string
	Required    bool
	Description string
	Default     interface{}
	Format      string `description:"https://json-schema.org/understanding-json-schema/reference/string.html#built-in-formats"`
	Properties  map[string]Param
	Items       []Param
}

type RequestParams struct {
	Query    []Param
	Body     []Param
	FormData []Param
	Headers  []Param
	Path     []Param
}

type ResponseParams struct {
	STDResponse bool
	Success     []Param
	Failure     map[int]string
}

func NewParam(name string, dataType string, description string, required bool, defaultValue ...interface{}) Param {
	p := Param{
		Name:        name,
		Type:        dataType,
		Required:    required,
		Description: description,
	}
	if len(defaultValue) > 0 {
		p.Default = defaultValue[0]
	}
	return p
}

func NewStringParam(name string, description string, required bool, defaultValue ...interface{}) Param {
	return NewParam(name, String, description, required, defaultValue)
}

func NewIntParam(name string, description string, required bool, defaultValue ...interface{}) Param {
	return NewParam(name, Integer, description, required, defaultValue)
}

func NewNumberParam(name string, description string, required bool, defaultValue ...interface{}) Param {
	return NewParam(name, Number, description, required, defaultValue)
}

func NewBooleanParam(name string, description string, required bool, defaultValue ...interface{}) Param {
	return NewParam(name, Boolean, description, required, defaultValue)
}

func NewObjectParam(name string, description string, required bool, properties map[string]Param) Param {
	return Param{
		Name:        name,
		Type:        Object,
		Required:    required,
		Description: description,
		Properties:  properties,
	}
}

func NewRefObjectParam(name string, description string, required bool, ref string) Param {
	return Param{
		Name:        name,
		Type:        Object,
		Required:    required,
		Description: description,
		Ref:         ref,
	}
}

func NewExtendedRefObjectParam(name string, description string, required bool, ref string, properties map[string]Param) Param {
	return Param{
		Name:        name,
		Type:        Object,
		Required:    required,
		Description: description,
		Ref:         ref,
		Properties:  properties,
	}
}

func NewArrayParam(name string, description string, required bool, items []Param) Param {
	return Param{
		Name:        name,
		Type:        Array,
		Required:    required,
		Description: description,
		Items:       items,
	}
}

func NewRefArrayParam(name string, description string, required bool, ref string) Param {
	return Param{
		Name:        name,
		Type:        Array,
		Required:    required,
		Description: description,
		Ref:         ref,
	}
}

func NewExtendedRefArrayParam(name string, description string, required bool, ref string, items []Param) Param {
	return Param{
		Name:        name,
		Type:        Array,
		Required:    required,
		Description: description,
		Ref:         ref,
		Items:       items,
	}
}
