package functab

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/xhd2015/xgo/runtime/core"
)

const __XGO_SKIP_TRAP = true

// rewrite at compile time by compiler, the body will be replaced with
// a call to runtime.__xgo_for_each_func
func __xgo_link_for_each_func(f func(pkgName string, funcName string, pc uintptr, fn interface{}, recvName string, argNames []string, resNames []string)) {
	panic("failed to link __xgo_link_for_each_func")
}

var funcInfos []*core.FuncInfo
var funcInfoMapping map[string]*core.FuncInfo
var funcPCMapping map[uintptr]*core.FuncInfo // pc->FuncInfo

func GetFuncs() []*core.FuncInfo {
	ensureMapping()
	return funcInfos
}

func GetFunc(fullName string) *core.FuncInfo {
	ensureMapping()
	return funcInfoMapping[fullName]
}

func Info(fn interface{}) *core.FuncInfo {
	ensureMapping()
	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func {
		panic(fmt.Errorf("given type is not a func: %T", fn))
	}
	// deref to pc
	pc := v.Pointer()
	return funcPCMapping[pc]
}

func InfoPC(pc uintptr) *core.FuncInfo {
	ensureMapping()
	return funcPCMapping[pc]
}

func GetFuncByPkg(pkgPath string, name string) *core.FuncInfo {
	ensureMapping()
	fn := funcInfoMapping[pkgPath+"."+name]
	if fn != nil {
		return fn
	}
	dotIdx := strings.Index(name, ".")
	if dotIdx < 0 {
		return fn
	}
	typName := name[:dotIdx]
	funcName := name[dotIdx+1:]

	return funcInfoMapping[pkgPath+".(*"+typName+")."+funcName]
}

var mappingOnce sync.Once

func ensureMapping() {
	mappingOnce.Do(func() {
		funcInfoMapping = map[string]*core.FuncInfo{}
		funcPCMapping = make(map[uintptr]*core.FuncInfo)
		__xgo_link_for_each_func(func(pkgPath string, funcName string, pc uintptr, fn interface{}, recvName string, argNames, resNames []string) {
			// prefix := pkgPath + "."
			_, recvTypeName, recvPtr, name := core.ParseFuncName(funcName[len(pkgPath)+1:], false)
			info := &core.FuncInfo{
				FullName: funcName,
				Name:     name,
				Pkg:      pkgPath,
				RecvType: recvTypeName,
				RecvPtr:  recvPtr,

				//
				PC:       pc,
				Func:     fn,
				RecvName: recvName,
				ArgNames: argNames,
				ResNames: resNames,
			}
			funcInfos = append(funcInfos, info)
			funcInfoMapping[info.FullName] = info
			funcPCMapping[info.PC] = info
		})
	})
}