package trap_with_overlay

import (
	"context"
	"testing"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/functab"
	"github.com/xhd2015/xgo/runtime/trap"
)

func TestTrapWithOverlay(t *testing.T) {
	funcInfo := functab.GetFuncByPkg("github.com/xhd2015/xgo/runtime/test/trap_with_overlay", "A")
	if funcInfo == nil {
		t.Fatalf("cannot get function A")
	}
	var haveCalled bool
	trap.AddFuncInfoInterceptor(funcInfo, &trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (data interface{}, err error) {
			haveCalled = true
			return
		},
	})
	// do the call
	// padding
	if !haveCalled {
		t.Fatalf("expect have called, actually not")
	}
}
