package patch

import (
	"fmt"
	"go/constant"
	"io"
	"os"
	"strings"

	"cmd/compile/internal/base"
	"cmd/compile/internal/ir"
	"cmd/compile/internal/syntax"
	"cmd/compile/internal/typecheck"
	"cmd/compile/internal/types"

	xgo_ctxt "cmd/compile/internal/xgo_rewrite_internal/patch/ctxt"
	xgo_record "cmd/compile/internal/xgo_rewrite_internal/patch/record"
	xgo_syntax "cmd/compile/internal/xgo_rewrite_internal/patch/syntax"
)

func debugIR() {
	var dumpIRFile string
	dumpIR := os.Getenv("XGO_DEBUG_DUMP_IR")
	if dumpIR != "" {
		dumpIRFile = os.Getenv("XGO_DEBUG_DUMP_IR_FILE")
	} else {
		// fallback
		dumpIR = os.Getenv("COMPILER_DEBUG_IR_DUMP_FUNCS")
	}
	if dumpIR == "" || dumpIR == "false" {
		return
	}

	var outFile io.Writer

	if dumpIRFile != "" {
		file, err := os.OpenFile(dumpIRFile, os.O_WRONLY|os.O_CREATE, 0755)
		if err != nil {
			panic(fmt.Errorf("dump ir: %w", err))
		}
		defer file.Close()
		outFile = file
	}

	pkgName := types.LocalPkg.Name

	if pkgName == "" {
		files := xgo_syntax.GetFiles()
		if len(files) > 0 {
			pkgName = files[0].PkgName.Value
		}
	}

	namePatterns := strings.Split(dumpIR, ",")
	forEachFunc(func(fn *ir.Func) bool {
		// fn.Sym().Name evaluates to plain func name, if with receiver, the receiver name
		// e.g.  A.B, (*A).C
		// examples:
		//   pkgPath.*, *.funcName, funcName
		if !xgo_ctxt.MatchAnyPattern(xgo_ctxt.GetPkgPath(), pkgName, fn.Sym().Name, namePatterns) {
			return true
		}
		if outFile == nil {
			ir.Dump("debug:", fn)
		} else {
			fmt.Fprintf(outFile, "%+v\n", fn)
		}
		return true
	})
}

func debugPrint(s string) *ir.CallExpr {
	return ir.NewCallExpr(base.AutogeneratedPos, ir.OCALL, typecheck.LookupRuntime("printstring"), []ir.Node{
		NewBasicLit(base.AutogeneratedPos, types.Types[types.TSTRING], constant.MakeString(s)),
	})
}

func regFuncsV1() {
	files := xgo_syntax.GetFiles()
	xgo_syntax.ClearFiles() // help GC

	type declName struct {
		name         string
		recvTypeName string
		recvPtr      bool
	}
	var declFuncNames []*declName
	for _, f := range files {
		for _, decl := range f.DeclList {
			fn, ok := decl.(*syntax.FuncDecl)
			if !ok {
				continue
			}
			if fn.Name.Value == "init" {
				continue
			}
			var recvTypeName string
			var recvPtr bool
			if fn.Recv != nil {
				if starExpr, ok := fn.Recv.Type.(*syntax.Operation); ok && starExpr.Op == syntax.Mul {
					recvTypeName = starExpr.X.(*syntax.Name).Value
					recvPtr = true
				} else {
					recvTypeName = fn.Recv.Type.(*syntax.Name).Value
				}
			}
			declFuncNames = append(declFuncNames, &declName{
				name:         fn.Name.Value,
				recvTypeName: recvTypeName,
				recvPtr:      recvPtr,
			})
		}
	}

	regFunc := typecheck.LookupRuntime("__xgo_register_func")
	regMethod := typecheck.LookupRuntime("__xgo_register_method")
	_ = regMethod

	var regNodes []ir.Node
	for _, declName := range declFuncNames {
		var valNode ir.Node
		fnSym, ok := types.LocalPkg.LookupOK(declName.name)
		if !ok {
			panic(fmt.Errorf("func name symbol not found: %s", declName.name))
		}
		if declName.recvTypeName != "" {
			typeSym, ok := types.LocalPkg.LookupOK(declName.recvTypeName)
			if !ok {
				panic(fmt.Errorf("type name symbol not found: %s", declName.recvTypeName))
			}
			var recvNode ir.Node
			if !declName.recvPtr {
				recvNode = typeSym.Def.(*ir.Name)
				// recvNode = ir.NewNameAt(base.AutogeneratedPos, typeSym, nil)
			} else {
				// types.TypeSymLookup are for things like "int","func(){...}"
				//
				// typeSym2 := types.TypeSymLookup(declName.recvTypeName)
				// if typeSym2 == nil {
				// 	panic("empty typeSym2")
				// }
				// types.TypeSym()
				recvNode = ir.TypeNode(typeSym.Def.(*ir.Name).Type())
			}
			valNode = ir.NewSelectorExpr(base.AutogeneratedPos, ir.OMETHEXPR, recvNode, fnSym)
			continue
		} else {
			valNode = fnSym.Def.(*ir.Name)
			// valNode = ir.NewNameAt(base.AutogeneratedPos, fnSym, fnSym.Def.Type())
			// continue
		}
		_ = valNode

		node := ir.NewCallExpr(base.AutogeneratedPos, ir.OCALL, regFunc, []ir.Node{
			// NewNilExpr(base.AutogeneratedPos, types.AnyType),
			ir.NewConvExpr(base.AutogeneratedPos, ir.OCONV, types.Types[types.TINTER] /*types.AnyType*/, valNode),
			// ir.NewBasicLit(base.AutogeneratedPos, types.Types[types.TSTRING], constant.MakeString("hello init\n")),
		})

		// ir.MethodExprFunc()
		regNodes = append(regNodes, node)
	}

	// this typecheck is required
	// to make subsequent steps work
	typecheck.Stmts(regNodes)

	// regFuncs.Body = []ir.Node{
	// 	ir.NewCallExpr(base.AutogeneratedPos, ir.OCALL, typecheck.LookupRuntime("printstring"), []ir.Node{
	// 		ir.NewBasicLit(base.AutogeneratedPos, types.Types[types.TSTRING], constant.MakeString("hello init\n")),
	// 	}),
	// }
	prependInit(base.AutogeneratedPos, typecheck.Target, regNodes)
}

func debugReplaceBody(fn *ir.Func) {
	// debug
	if false {
		str := NewStringLit(fn.Pos(), "debug")
		nd := fn.Body[0]
		ue := nd.(*ir.UnaryExpr)
		ce := ue.X.(*ir.ConvExpr)
		ce.X = str
		xgo_record.SetRewrittenBody(fn, fn.Body)
		return
	}
	if false {
		fn.Body = []ir.Node{
			debugPrint("replaced body x\n"),
		}
		typeCheckBody(fn)
		xgo_record.SetRewrittenBody(fn, fn.Body)
		return
	}
	debugBody := ifConstant(fn.Pos(), true, []ir.Node{
		debugPrint("replaced body 1\n"),
		debugPrint("replaced body 2\n"),
		ir.NewReturnStmt(base.AutogeneratedPos, nil),
		fn.Body[0],
		debugPrint("replaced body 3\n"),
		// ir.NewReturnStmt(fn.Pos(), nil),
	}, nil)
	// debugBody := debugPrint("replaced body\n")
	fn.Body = []ir.Node{debugBody}
	typeCheckBody(fn)
	xgo_record.SetRewrittenBody(fn, fn.Body)
}
