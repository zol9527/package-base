package interfaces

// Logger 日志接口
//
// 描述: 统一的日志接口
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	With(fields ...Field) Logger
	Named(name string) Logger
	Sync() error
}

// Field 日志字段
//
// 描述: 日志记录中的字段信息
type Field interface {
	Key() string
	Value() interface{}
}

// 预定义的字段构造函数
func String(key, val string) Field {
	return &stringField{key: key, val: val}
}

func Int(key string, val int) Field {
	return &intField{key: key, val: val}
}

func Int64(key string, val int64) Field {
	return &int64Field{key: key, val: val}
}

func Uint(key string, val uint) Field {
	return &uintField{key: key, val: val}
}

func Bool(key string, val bool) Field {
	return &boolField{key: key, val: val}
}

func Any(key string, val interface{}) Field {
	return &anyField{key: key, val: val}
}

func ErrorField(err error) Field {
	return &errorField{key: "error", val: err}
}

// 字段实现
type stringField struct {
	key string
	val string
}

func (f *stringField) Key() string        { return f.key }
func (f *stringField) Value() interface{} { return f.val }

type intField struct {
	key string
	val int
}

func (f *intField) Key() string        { return f.key }
func (f *intField) Value() interface{} { return f.val }

type int64Field struct {
	key string
	val int64
}

func (f *int64Field) Key() string        { return f.key }
func (f *int64Field) Value() interface{} { return f.val }

type uintField struct {
	key string
	val uint
}

func (f *uintField) Key() string        { return f.key }
func (f *uintField) Value() interface{} { return f.val }

type boolField struct {
	key string
	val bool
}

func (f *boolField) Key() string        { return f.key }
func (f *boolField) Value() interface{} { return f.val }

type anyField struct {
	key string
	val interface{}
}

func (f *anyField) Key() string        { return f.key }
func (f *anyField) Value() interface{} { return f.val }

type errorField struct {
	key string
	val error
}

func (f *errorField) Key() string        { return f.key }
func (f *errorField) Value() interface{} { return f.val }
