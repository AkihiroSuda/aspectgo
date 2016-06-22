package weave

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"go/types"
	"log"
	"os"
	"strings"

	rewrite "github.com/tsuna/gorewrite"

	"golang.org/x/tools/go/loader"

	"golang.org/x/exp/aspectgo/aspect"
	"golang.org/x/exp/aspectgo/compiler/consts"
	"golang.org/x/exp/aspectgo/compiler/gopath"
	"golang.org/x/exp/aspectgo/compiler/util"
)

func rewriteProgram(wovenGOPATH string, rw *rewriter) ([]string, error) {
	if err := rw.init(); err != nil {
		return nil, err
	}
	oldGOPATH := os.Getenv("GOPATH")
	if oldGOPATH == "" {
		return nil, fmt.Errorf("GOPATH not set")
	}
	var rewrittenFnames []string
	for _, pkgInfo := range rw.Program.InitialPackages() {
		rw.currentPkg = pkgInfo.Pkg
		for _, file := range pkgInfo.Files {
			rw.currentFile = file
			posn := rw.Program.Fset.Position(file.Pos())
			if strings.HasSuffix(posn.Filename, "_aspect.go") {
				continue
			}
			outf, err := gopath.FileForNewGOPATH(posn.Filename,
				oldGOPATH, wovenGOPATH)
			if err != nil {
				return nil, err
			}
			defer outf.Close()
			log.Printf("Rewriting %s --> %s",
				posn.Filename, outf.Name())
			rewritten := rewrite.Rewrite(rw, file)
			outw := bufio.NewWriter(outf)
			outw.Write([]byte(consts.AutogenFileHeader))
			err = format.Node(outw, rw.Program.Fset, rewritten)
			if err != nil {
				return nil, err
			}
			for _, add := range rw.AddendumForASTFile() {
				outw.Write([]byte("\n"))
				format.Node(outw, rw.Program.Fset, add)
				outw.Write([]byte("\n"))
			}
			outw.Flush()
			rewrittenFnames = append(rewrittenFnames, outf.Name())
		}
	}
	return rewrittenFnames, nil
}

var gRewriterLastP = 0

// rewriter implements rewrite.Rewriter.
// usage:
//  Step 1: instatiate rewriter and call rewriter.init().
//  Step 2: set rewriter.currentPkg, for each pkg in the program
//  Step 3: set rewriter.currentFile, for each file in the pkg
//  Step 4: call rewrite.Rewrite(rewriter, rewriter.currentFile) for rewriting the file
//  Step 5: call rewriter.AddendumForAstFile() for getting the addendum for the file
type rewriter struct {
	Program          *loader.Program
	Matched          map[*ast.Ident]types.Object
	Aspects          map[aspect.Pointcut]*types.Named
	PointcutsByIdent map[*ast.Ident]aspect.Pointcut
	// fileAddendum is set by rewriter.Rewrite().
	// rewriteProgram() uses rewriter.AddendumForASTFile()
	// as a getter.
	fileAddendum []ast.Node
	proxyExprs   map[*ast.Ident]ast.Expr
	// currentPkg is set by the loop in rewriteProgram().
	// It is used for rewriter.typeString().
	currentPkg *types.Package
	// currentFile is set by the loop in rewriteProgram().
	// It is used for rewriter.typeString().
	currentFile *ast.File
}

func (r *rewriter) init() error {
	if r.Program == nil || r.Matched == nil ||
		r.Aspects == nil || r.PointcutsByIdent == nil {
		log.Fatal("impl error (nil args)")
	}

	// NOTE: r.fileAddendum is initialized in Rewrite():*ast.File
	r.proxyExprs = make(map[*ast.Ident]ast.Expr)
	return nil
}

func voidIntfArrayExpr() *ast.ArrayType {
	return &ast.ArrayType{
		Elt: &ast.InterfaceType{
			Methods: &ast.FieldList{},
		}}
}

// _proxy_decl generates _ag_proxy_func decl like this:
// `func _ag_proxy_0(s string)`
func (r *rewriter) _proxy_decl(node ast.Node, matched types.Object, proxyName string) *ast.FuncDecl {
	sig := matched.Type().(*types.Signature)
	funcDecl := &ast.FuncDecl{}
	funcDecl.Name = ast.NewIdent(proxyName)
	funcDecl.Type = &ast.FuncType{}
	params, results := &ast.FieldList{}, &ast.FieldList{}
	params.List, results.List = make([]*ast.Field, 0), make([]*ast.Field, 0)
	if sig.Recv() != nil {
		param := &ast.Field{
			Names: []*ast.Ident{ast.NewIdent("_ag_recv")},
			Type: &ast.ParenExpr{
				X: ast.NewIdent(r.typeString(sig.Recv().Type())),
			}}
		params.List = append(params.List, param)
	}
	for i := 0; i < sig.Params().Len(); i++ {
		sigParam := sig.Params().At(i)
		param := &ast.Field{}
		param.Names = []*ast.Ident{ast.NewIdent(sigParam.Name())}
		paramTypeStr := r.typeString(sigParam.Type())
		param.Type = ast.NewIdent(paramTypeStr)
		params.List = append(params.List, param)
	}
	for i := 0; i < sig.Results().Len(); i++ {
		sigResult := sig.Results().At(i)
		result := &ast.Field{}
		result.Names = []*ast.Ident{ast.NewIdent(sigResult.Name())}
		result.Type = ast.NewIdent(r.typeString(sigResult.Type()))
		results.List = append(results.List, result)
	}
	funcDecl.Type.Params, funcDecl.Type.Results = params, results
	return funcDecl
}

// _proxy_body_XArgs generates like this:
// `XArgs: []interface{}{"world"}`
func (r *rewriter) _proxy_body_XArgs(matched types.Object) []ast.Expr {
	sig := matched.Type().(*types.Signature)
	var xArgsExprs []ast.Expr
	for i := 0; i < sig.Params().Len(); i++ {
		sigParam := sig.Params().At(i)
		xArgsExprs = append(xArgsExprs, ast.NewIdent(sigParam.Name()))
	}

	return xArgsExprs
}

// _proxy_body_XFunc generates like this:
// `XFunc: func(_ag_args []interface{}) []interface {} {
//                _ag_arg0 := _ag_args[0].(string)
//                sayHello(_ag_arg0)
//                _ag_res := []interface{}{}
//                return _ag_res
//          }`
func (r *rewriter) _proxy_body_XFunc(node ast.Node, matched types.Object) *ast.FuncLit {
	sig := matched.Type().(*types.Signature)
	var xFuncBodyStmts []ast.Stmt
	var xFuncBodyArgExprs []ast.Expr
	for i := 0; i < sig.Params().Len(); i++ {
		sigParam := sig.Params().At(i)
		lhsName := fmt.Sprintf("_ag_arg%d", i)
		rhsType := ast.NewIdent(r.typeString(sigParam.Type()))
		if i == sig.Params().Len()-1 && sig.Variadic() {
			xFuncBodyArgExprs = append(xFuncBodyArgExprs,
				ast.NewIdent(lhsName+"..."))
		} else {
			xFuncBodyArgExprs = append(xFuncBodyArgExprs,
				ast.NewIdent(lhsName))
		}
		assignStmt := &ast.AssignStmt{
			Lhs: []ast.Expr{
				ast.NewIdent(lhsName),
			},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.TypeAssertExpr{
					X: &ast.IndexExpr{
						X: ast.NewIdent("_ag_args"),
						Index: &ast.BasicLit{
							Kind:  token.INT,
							Value: fmt.Sprintf("%d", i),
						}},
					Type: rhsType,
				}}}
		xFuncBodyStmts = append(xFuncBodyStmts, assignStmt)
	}
	var xFuncBodyCallFuncExp ast.Expr
	switch n := node.(type) {
	case *ast.Ident:
		xFuncBodyCallFuncExp = ast.NewIdent(n.Name)
	case *ast.SelectorExpr:
		var x ast.Expr
		if sig.Recv() != nil {
			x = ast.NewIdent("_ag_recv")
		} else {
			// FIXME FIXME FIXME: copy n.X
			x = n.X
		}
		xFuncBodyCallFuncExp = &ast.SelectorExpr{
			X:   x,
			Sel: ast.NewIdent(n.Sel.Name)}
	default:
		log.Fatalf("impl error: %s is unexpected type: %s", util.ASTDebugString(n))
	}
	var xFuncBodyCallLhs []ast.Expr
	var xFuncBodyCallLhs2 []ast.Expr
	for i := 0; i < sig.Results().Len(); i++ {
		s := fmt.Sprintf("_ag_res%d", i)
		xFuncBodyCallLhs = append(xFuncBodyCallLhs,
			ast.NewIdent(s))
		xFuncBodyCallLhs2 = append(xFuncBodyCallLhs2,
			ast.NewIdent(s))
	}
	var xFuncBodyCallStmt ast.Stmt
	xFuncBodyCallExpr := &ast.CallExpr{
		Fun:  xFuncBodyCallFuncExp,
		Args: xFuncBodyArgExprs}
	if len(xFuncBodyCallLhs) > 0 {
		xFuncBodyCallStmt = &ast.AssignStmt{
			Lhs: xFuncBodyCallLhs,
			Tok: token.DEFINE,
			Rhs: []ast.Expr{xFuncBodyCallExpr}}
	} else {
		xFuncBodyCallStmt = &ast.ExprStmt{
			X: xFuncBodyCallExpr}
	}
	xFuncBodyStmts = append(xFuncBodyStmts, xFuncBodyCallStmt)
	xFuncBodyResultAssignStmt := &ast.AssignStmt{
		Lhs: []ast.Expr{ast.NewIdent("_ag_res")},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			&ast.CompositeLit{
				Type: voidIntfArrayExpr(), Elts: xFuncBodyCallLhs2}}}
	xFuncBodyStmts = append(xFuncBodyStmts, xFuncBodyResultAssignStmt)
	xFuncBodyReturnStmt := &ast.ReturnStmt{
		Results: []ast.Expr{ast.NewIdent("_ag_res")}}
	xFuncBodyStmts = append(xFuncBodyStmts, xFuncBodyReturnStmt)

	xFuncLit := &ast.FuncLit{
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{
					&ast.Field{
						Names: []*ast.Ident{ast.NewIdent("_ag_args")},
						Type:  voidIntfArrayExpr()}}},
			Results: &ast.FieldList{
				List: []*ast.Field{
					&ast.Field{
						Type: voidIntfArrayExpr()}}}},
		Body: &ast.BlockStmt{List: xFuncBodyStmts}}

	return xFuncLit
}

func (r *rewriter) _proxy_body_XReceiver(node ast.Node, matched types.Object) ast.Expr {
	sig := matched.Type().(*types.Signature)
	recv := sig.Recv()
	if recv != nil {
		return ast.NewIdent("_ag_recv")
	}
	return ast.NewIdent("nil")
}

func (r *rewriter) _proxy_body_callExpr(node ast.Node, matched types.Object, asp *types.Named) *ast.CallExpr {
	callExpr := &ast.CallExpr{}
	adviceExpr := &ast.SelectorExpr{
		X: &ast.ParenExpr{
			X: &ast.UnaryExpr{
				Op: token.AND,
				X: &ast.CompositeLit{
					Type: &ast.SelectorExpr{
						X:   ast.NewIdent("agaspect"),
						Sel: ast.NewIdent(asp.Obj().Name()),
					}}}},
		Sel: &ast.Ident{
			Name: "Advice",
		}}

	ctxExpr := &ast.UnaryExpr{
		Op: token.AND,
		X: &ast.CompositeLit{
			Type: &ast.SelectorExpr{
				X:   ast.NewIdent("aspectrt"),
				Sel: ast.NewIdent("ContextImpl"),
			},
			Elts: []ast.Expr{
				&ast.KeyValueExpr{
					Key: ast.NewIdent("XArgs"),
					Value: &ast.CompositeLit{
						Type: voidIntfArrayExpr(),
						Elts: r._proxy_body_XArgs(matched),
					}},
				&ast.KeyValueExpr{
					Key:   ast.NewIdent("XFunc"),
					Value: r._proxy_body_XFunc(node, matched),
				},
				&ast.KeyValueExpr{
					Key:   ast.NewIdent("XReceiver"),
					Value: r._proxy_body_XReceiver(node, matched),
				}}}}

	callExpr.Fun = adviceExpr
	callExpr.Args = []ast.Expr{ctxExpr}
	return callExpr
}

// _proxy_body generates _ag_proxy_func body like this:
//
// _ag_res := (&dummyAspect{}).Advice(
// 	&ContextImpl{
// 		XArgs: []interface{}{"world"},
// 		XFunc: func(_ag_args []interface{}) []interface{} {
// 			_ag_arg0 := _ag_args[0].(string)
// 			sayHello(_ag_arg0)
// 			_ag_res := []interface{}{}
// 			return _ag_res
// 		}})
// _ = _ag_res
// return
func (r *rewriter) _proxy_body(node ast.Node, matched types.Object, asp *types.Named) *ast.BlockStmt {
	var stmts []ast.Stmt
	stmts = append(stmts,
		&ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent("_ag_res")},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{r._proxy_body_callExpr(node, matched, asp)}})

	sig := matched.Type().(*types.Signature)
	var resAssignStmts []ast.Stmt
	var resExprs []ast.Expr
	for i := 0; i < sig.Results().Len(); i++ {
		sigResult := sig.Results().At(i)
		stmt := &ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent(fmt.Sprintf("_ag_res%d", i)),
				ast.NewIdent("_")},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.TypeAssertExpr{
					X: &ast.IndexExpr{
						X: ast.NewIdent("_ag_res"),
						Index: &ast.BasicLit{
							Kind:  token.INT,
							Value: fmt.Sprintf("%d", i),
						}},
					Type: ast.NewIdent(r.typeString(sigResult.Type()))}}}
		resAssignStmts = append(resAssignStmts, stmt)
		resExprs = append(resExprs, ast.NewIdent(fmt.Sprintf("_ag_res%d", i)))
	}

	stmts = append(stmts,
		&ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent("_")},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{ast.NewIdent("_ag_res")}})
	stmts = append(stmts, resAssignStmts...)
	stmts = append(stmts, &ast.ReturnStmt{Results: resExprs})
	res := &ast.BlockStmt{List: stmts}
	return res
}

func (r *rewriter) _proxy(node ast.Node, matched types.Object, proxyName string, asp *types.Named) *ast.FuncDecl {
	funcDecl := r._proxy_decl(node, matched, proxyName)
	funcDecl.Body = r._proxy_body(node, matched, asp)
	return funcDecl
}

func (r *rewriter) _pgen_decl(matched types.Object, pdecl *ast.FuncDecl, pgenName string) *ast.FuncDecl {
	sig := matched.Type().(*types.Signature)
	receiver := sig.Recv()
	funcDecl := &ast.FuncDecl{}
	funcDecl.Name = ast.NewIdent(pgenName)
	funcDecl.Type = &ast.FuncType{}
	params, results := &ast.FieldList{}, &ast.FieldList{}
	params.List, results.List = make([]*ast.Field, 0), make([]*ast.Field, 0)

	if receiver != nil {
		pdeclRecv := pdecl.Type.Params.List[0]
		name := pdeclRecv.Names[0].Name
		typ := r.typeString(receiver.Type())
		param := &ast.Field{
			Names: []*ast.Ident{ast.NewIdent(name)},
			Type: &ast.ParenExpr{
				X: ast.NewIdent(typ)}}
		params.List = append(params.List, param)
	}

	pdParamsL, pdResultsL := make([]*ast.Field, 0), make([]*ast.Field, 0)
	pdParamScanBegin := 0
	if receiver != nil {
		pdParamScanBegin = 1
	}
	for i := pdParamScanBegin; i < len(pdecl.Type.Params.List); i++ {
		typIdent := pdecl.Type.Params.List[i].Type.(*ast.Ident)
		typ := typIdent.Name
		if sig.Variadic() && i == len(pdecl.Type.Params.List)-1 {
			typ = strings.Replace(typ, "[]", "...", 1)
		}
		pdParam := &ast.Field{
			Type: ast.NewIdent(typ),
		}
		pdParamsL = append(pdParamsL, pdParam)
	}
	for i := 0; i < len(pdecl.Type.Results.List); i++ {
		pdResult := &ast.Field{
			Type: pdecl.Type.Results.List[i].Type,
		}
		pdResultsL = append(pdResultsL, pdResult)
	}
	result := &ast.Field{
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: pdParamsL,
			},
			Results: &ast.FieldList{
				List: pdResultsL,
			}}}
	results.List = append(results.List, result)
	funcDecl.Type.Params, funcDecl.Type.Results = params, results
	return funcDecl
}

func (r *rewriter) _pgen_body(matched types.Object, pdecl *ast.FuncDecl) *ast.BlockStmt {
	sig := matched.Type().(*types.Signature)
	receiver := sig.Recv()

	funcLit := &ast.FuncLit{}
	funcLit.Type = &ast.FuncType{}
	params, results := &ast.FieldList{}, &ast.FieldList{}
	pdParamsL, pdResultsL := make([]*ast.Field, 0), make([]*ast.Field, 0)
	pdParamScanBegin := 0
	if receiver != nil {
		pdParamScanBegin = 1
	}
	for i := pdParamScanBegin; i < len(pdecl.Type.Params.List); i++ {
		typIdent := pdecl.Type.Params.List[i].Type.(*ast.Ident)
		typ := typIdent.Name
		if sig.Variadic() && i == len(pdecl.Type.Params.List)-1 {
			typ = strings.Replace(typ, "[]", "...", 1)
		}
		pdParam := &ast.Field{
			Names: pdecl.Type.Params.List[i].Names,
			Type:  ast.NewIdent(typ),
		}
		pdParamsL = append(pdParamsL, pdParam)
	}
	for i := 0; i < len(pdecl.Type.Results.List); i++ {
		pdResult := &ast.Field{
			Names: pdecl.Type.Results.List[i].Names,
			Type:  pdecl.Type.Results.List[i].Type,
		}
		pdResultsL = append(pdResultsL, pdResult)
	}
	params.List, results.List = pdParamsL, pdResultsL
	funcLit.Type.Params, funcLit.Type.Results = params, results

	var funcLitArgs []ast.Expr
	for i := 0; i < len(pdecl.Type.Params.List); i++ {
		for j := 0; j < len(pdecl.Type.Params.List[i].Names); j++ {
			funcLitArgs = append(funcLitArgs,
				ast.NewIdent(pdecl.Type.Params.List[i].Names[j].Name))
		}
	}

	funcLitBodyExpr := &ast.CallExpr{
		Fun:  ast.NewIdent(pdecl.Name.Name),
		Args: funcLitArgs,
	}
	var funcLitBodyStmt ast.Stmt
	if len(pdecl.Type.Results.List) == 0 {
		funcLitBodyStmt = &ast.ExprStmt{
			X: funcLitBodyExpr,
		}
	} else {
		funcLitBodyStmt = &ast.ReturnStmt{
			Results: []ast.Expr{funcLitBodyExpr},
		}
	}
	funcLit.Body = &ast.BlockStmt{List: []ast.Stmt{funcLitBodyStmt}}

	return &ast.BlockStmt{
		List: []ast.Stmt{
			&ast.ReturnStmt{
				Results: []ast.Expr{funcLit},
			},
		}}
}

// _pgen generates the pgen function.
//
// pgen is like this:
//
// var f func(int)
// f := (_ag_pgen_ag_proxy_0(i)) // orig: f := i.Foo
// f(42)
//
// func _ag_pgen_ag_proxy_0(i I) func(int) {
// 	return func(x int){_ag_proxy_0(i, x)}
// }
// ​
// func _ag_proxy_0(i I, x int) {
//   ..
// }
func (r *rewriter) _pgen(matched types.Object, pdecl *ast.FuncDecl, pgenName string) *ast.FuncDecl {
	funcDecl := r._pgen_decl(matched, pdecl, pgenName)
	funcDecl.Body = r._pgen_body(matched, pdecl)
	return funcDecl
}

func (r *rewriter) _proxy_fix_up(node ast.Node, matched types.Object, pgenName string) ast.Expr {
	sig := matched.Type().(*types.Signature)
	var args []ast.Expr
	recv := sig.Recv()
	if recv != nil {
		_, recvIsPointer := recv.Type().Underlying().(*types.Pointer)

		xs, ok := node.(*ast.SelectorExpr)
		if !ok {
			log.Fatalf("impl error: node=%s, recv=%s", util.ASTDebugString(node), recv)
		}
		typesInfo := r.Program.AllPackages[r.currentPkg]
		xTypeInfo := typesInfo.Types[xs.X.(ast.Expr)]
		_, xIsPointer := xTypeInfo.Type.Underlying().(*types.Pointer)

		// FIXME FIXME FIXME: copy xs.X
		x := xs.X
		var arg ast.Expr
		if recvIsPointer && !xIsPointer {
			arg = &ast.UnaryExpr{
				Op: token.AND,
				X:  x}
		} else {
			arg = x
		}
		args = append(args, arg)
	}
	callExpr := &ast.CallExpr{
		Fun:  ast.NewIdent(pgenName),
		Args: args,
	}
	parenExpr := &ast.ParenExpr{
		X: callExpr,
	}
	return parenExpr
}

// proxy generates addendum and returns new ast.Expr for the node.
// generated addendum can be obtained via AddendumForASTFile.
//
// How it works:
//   Step 1: calls _proxy for generating _ag_proxy_N addendum
//   Step 2: calls _pgen for generating _ag_pgen_ag_proxy_N addendum
//   Step 3: calls _proxy_fix_up for generating the new node
func (r *rewriter) proxy(node ast.Node, pointcut aspect.Pointcut) ast.Expr {
	var id *ast.Ident
	switch n := node.(type) {
	case *ast.Ident:
		id = n
	case *ast.SelectorExpr:
		id = n.Sel
	default:
		log.Fatalf("impl error: %s is unexpected type: %s", util.ASTDebugString(n))
	}
	// alreadyGen, ok := r.proxyExprs[id]
	// if ok {
	// 	return alreadyGen
	// }

	matched, ok := r.Matched[id]
	if !ok {
		log.Fatalf("impl error: obj not found for id %s", id)
	}
	asp, ok := r.Aspects[pointcut]
	if !ok {
		log.Fatalf("impl error: asp %s not found for pointcut %s", asp, pointcut)
	}

	proxyName := fmt.Sprintf("_ag_proxy_%d", gRewriterLastP)
	pgenName := fmt.Sprintf("_ag_pgen%s", proxyName)
	gRewriterLastP++

	proxyAst := r._proxy(node, matched, proxyName, asp)
	r.fileAddendum = append(r.fileAddendum, proxyAst)

	pgenAst := r._pgen(matched, proxyAst, pgenName)
	r.fileAddendum = append(r.fileAddendum, pgenAst)

	expr := r._proxy_fix_up(node, matched, pgenName)
	r.proxyExprs[id] = expr
	return expr
}

func (r *rewriter) Rewrite(node ast.Node) (ast.Node, rewrite.Rewriter) {
	switch n := node.(type) {
	case *ast.File:
		r.fileAddendum = make([]ast.Node, 0)
		newImports := []*ast.ImportSpec{
			&ast.ImportSpec{
				Name: ast.NewIdent("aspectrt"),
				Path: &ast.BasicLit{
					Kind:  token.STRING,
					Value: "\"" + consts.AspectGoPackagePath + "/aspect/rt\"",
				}},
			&ast.ImportSpec{
				Path: &ast.BasicLit{
					Kind:  token.STRING,
					Value: "\"agaspect\"",
				}},
		}
		newFile := &ast.File{}
		newFile.Name = ast.NewIdent(n.Name.Name)
		newFile.Decls = append([]ast.Decl{
			&ast.GenDecl{
				Tok:   token.IMPORT,
				Specs: []ast.Spec{newImports[0]}},
			&ast.GenDecl{
				Tok:   token.IMPORT,
				Specs: []ast.Spec{newImports[1]}},
		}, n.Decls...)
		newFile.Scope = n.Scope
		newFile.Imports = append(newImports, n.Imports...)
		newFile.Unresolved = n.Unresolved
		return newFile, r
	case *ast.Ident:
		pointcut, ok := r.PointcutsByIdent[n]
		if !ok {
			goto nop
		}
		newExpr := r.proxy(n, pointcut)
		return newExpr, nil
	case *ast.SelectorExpr:
		pointcut, ok := r.PointcutsByIdent[n.Sel]
		if !ok {
			goto nop
		}
		newExpr := r.proxy(n, pointcut)
		return newExpr, nil
	}
nop:
	return node, r
}

func (r *rewriter) AddendumForASTFile() []ast.Node {
	return r.fileAddendum
}

func (r *rewriter) typeString(typ types.Type) string {
	s, err := util.LocalTypeString(typ,
		r.currentPkg,
		r.currentFile.Imports)
	if err != nil {
		log.Fatal(err)
	}
	return s
}
