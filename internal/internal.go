package internal

import (
	"errors"
	"reflect"
	"runtime"
	"strings"
)

// 全局唯一
var GlobalMethod = &Method{methods: map[string]reflect.Value{}}

type Method struct {
	methods map[string]reflect.Value
}

func (m *Method) register(impl interface{}) error {
	pl := reflect.ValueOf(impl)
	if pl.Kind() != reflect.Func {
		return errors.New("impl should be function")
	}
	// 获取函数名
	methodName := runtime.FuncForPC(pl.Pointer()).Name()
	if len(strings.Split(methodName, ".")) < 1 {
		return errors.New("invalid function name")
	}
	lastFuncName := strings.Split(methodName, ".")[1]
	m.methods[lastFuncName] = pl
	return nil
}

func (m *Method) call(methodName string, callParams ...interface{}) ([]interface{}, error) {
	fn, ok := m.methods[methodName]
	if !ok {
		return nil, errors.New("impl method not found! Please Register first")
	}
	in := make([]reflect.Value, len(callParams))
	for i := 0; i < len(callParams); i++ {
		in[i] = reflect.ValueOf(callParams[i])
	}
	res := fn.Call(in)
	out := make([]interface{}, len(res))
	for i := 0; i < len(res); i++ {
		out[i] = res[i].Interface()
	}
	return out, nil
}

func Call(methodName string, callParams ...interface{}) ([]interface{}, error) {
	return GlobalMethod.call(methodName, callParams...)
}

func Register(impl interface{}) error {
	return GlobalMethod.register(impl)
}

type RpcRequest struct {
	MethodName string
	Params     []interface{}
}

type RpcResponses struct {
	Returns []interface{}
	Err     error
}
