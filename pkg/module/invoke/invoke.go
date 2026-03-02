package invoke

import (
	"fmt"
	"reflect"
)

var errorType = reflect.TypeOf((*error)(nil)).Elem()

type Invoker struct {
	functions map[string]*FunctionInfo
}

func NewInvoker() *Invoker {
	return &Invoker{
		functions: make(map[string]*FunctionInfo),
	}
}

// Register validates and stores a function for later invocation.
// Panics if f is not a function, has more than 2 return values,
// has a non-error second return value, or fn is already registered.
func (iv *Invoker) Register(fn string, f interface{}) {
	v := reflect.ValueOf(f)
	t := v.Type()

	if t.Kind() != reflect.Func {
		panic(fmt.Sprintf("invoke: Register(%q): expected function, got %s", fn, t.Kind()))
	}

	numOut := t.NumOut()
	if numOut > 2 {
		panic(fmt.Sprintf("invoke: Register(%q): function has %d return values, max 2", fn, numOut))
	}
	if numOut == 2 && !t.Out(1).Implements(errorType) {
		panic(fmt.Sprintf("invoke: Register(%q): second return value must implement error, got %v", fn, t.Out(1)))
	}

	if _, exists := iv.functions[fn]; exists {
		panic(fmt.Sprintf("invoke: Register(%q): duplicate registration", fn))
	}

	fi := &FunctionInfo{
		FuncValue: v,
		FuncType:  t,
		InType:    make([]reflect.Type, t.NumIn()),
	}
	for i := 0; i < t.NumIn(); i++ {
		fi.InType[i] = t.In(i)
	}
	iv.functions[fn] = fi
}

func (iv *Invoker) Invoke(fn string, args ...interface{}) (interface{}, error) {
	fi, ok := iv.functions[fn]
	if !ok {
		return nil, fmt.Errorf("invoke: function %q not found", fn)
	}

	if len(args) != len(fi.InType) {
		return nil, fmt.Errorf("invoke: function %q expects %d args, got %d", fn, len(fi.InType), len(args))
	}

	var in []reflect.Value
	if len(args) > 0 {
		in = make([]reflect.Value, len(args))
		for i, arg := range args {
			if arg == nil {
				k := fi.InType[i].Kind()
				if k == reflect.Ptr || k == reflect.Interface || k == reflect.Map ||
					k == reflect.Slice || k == reflect.Chan || k == reflect.Func {
					in[i] = reflect.Zero(fi.InType[i])
				} else {
					return nil, fmt.Errorf("invoke: function %q param %d: nil not assignable to %v", fn, i, fi.InType[i])
				}
				continue
			}
			av := reflect.ValueOf(arg)
			if !av.Type().AssignableTo(fi.InType[i]) {
				return nil, fmt.Errorf("invoke: function %q param %d: type %v not assignable to %v", fn, i, av.Type(), fi.InType[i])
			}
			in[i] = av
		}
	}

	out := fi.FuncValue.Call(in)

	switch len(out) {
	case 0:
		return nil, nil
	case 1:
		r := out[0].Interface()
		if e, ok := r.(error); ok {
			return nil, e
		}
		return r, nil
	default:
		var rerr error
		if e, ok := out[1].Interface().(error); ok {
			rerr = e
		}
		return out[0].Interface(), rerr
	}
}

type FunctionInfo struct {
	FuncValue reflect.Value
	FuncType  reflect.Type
	InType    []reflect.Type
}