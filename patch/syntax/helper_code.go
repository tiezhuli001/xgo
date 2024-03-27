//go:build ignore
// +build ignore

package syntax

type __xgo_local_func_stub struct {
	PkgPath      string
	Fn           interface{}
	PC           uintptr // filled later
	Interface    bool
	Generic      bool
	Closure      bool // is the given function a closure
	RecvTypeName string
	RecvPtr      bool
	Name         string
	IdentityName string // name without pkgPath

	RecvName string
	ArgNames []string
	ResNames []string

	// can be retrieved at runtime
	FirstArgCtx bool // first argument is context.Context or sub type?
	LastResErr  bool // last res is error or sub type?

	File string
	Line int
}

func __xgo_link_generated_register_func(fn interface{}) {
	// linked later by compiler
	panic("failed to link __xgo_link_generated_register_func")
}

func __xgo_local_register_func(pkgPath string, fn interface{}, closure bool, recvName string, argNames []string, resNames []string, file string, line int) {
	__xgo_link_generated_register_func(__xgo_local_func_stub{PkgPath: pkgPath, Fn: fn, Closure: closure, RecvName: recvName, ArgNames: argNames, ResNames: resNames})
}

func __xgo_local_register_interface(pkgPath string, interfaceName string, file string, line int) {
	__xgo_link_generated_register_func(__xgo_local_func_stub{PkgPath: pkgPath, Interface: true, File: file, Line: line})
}
