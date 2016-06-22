// Copyright 2009 The Go Authors. All rights reserved.
// Copyright 2015 Benoit Sigoure. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package gorewrite

import (
	"fmt"
	. "go/ast"
)

// A Rewriter's Rewrite method is invoked for each node encountered by Walk.
// If the result visitor w is not nil, Walk visits each of the children
// of node with the visitor w, followed by a call of w.Rewrite(nil).
type Rewriter interface {
	Rewrite(node Node) (rewritten Node, w Rewriter)
}

// Helper functions for common node lists. They may be empty.

func rewriteIdentList(v Rewriter, list []*Ident) {
	for i, x := range list {
		list[i] = Rewrite(v, x).(*Ident)
	}
}

func rewriteExprList(v Rewriter, list []Expr) {
	for i, x := range list {
		list[i] = Rewrite(v, x).(Expr)
	}
}

func rewriteStmtList(v Rewriter, list []Stmt) {
	for i, x := range list {
		list[i] = Rewrite(v, x).(Stmt)
	}
}

func rewriteDeclList(v Rewriter, list []Decl) {
	for i, x := range list {
		list[i] = Rewrite(v, x).(Decl)
	}
}

// TODO(gri): Investigate if providing a closure to Rewrite leads to
//            simpler use (and may help eliminate Inspect in turn).

// Rewrite traverses an AST in depth-first order: It starts by calling
// v.Rewrite(node); node must not be nil. If the visitor w returned by
// v.Rewrite(node) is not nil, Rewrite is invoked recursively with visitor
// w for each of the non-nil children of node, followed by a call of
// w.Rewrite(nil).
//
func Rewrite(v Rewriter, node Node) (rewritten Node) {
	if rewritten, v = v.Rewrite(node); v == nil {
		return rewritten
	}

	// rewrite children
	// (the order of the cases matches the order
	// of the corresponding node types in ast.go)
	switch n := node.(type) {
	// Comments and fields
	case *Comment:
		// nothing to do

	case *CommentGroup:
		for i, c := range n.List {
			n.List[i] = Rewrite(v, c).(*Comment)
		}

	case *Field:
		if n.Doc != nil {
			n.Doc = Rewrite(v, n.Doc).(*CommentGroup)
		}
		rewriteIdentList(v, n.Names)
		n.Type = Rewrite(v, n.Type).(Expr)
		if n.Tag != nil {
			n.Tag = Rewrite(v, n.Tag).(*BasicLit)
		}
		if n.Comment != nil {
			n.Comment = Rewrite(v, n.Comment).(*CommentGroup)
		}

	case *FieldList:
		for i, f := range n.List {
			n.List[i] = Rewrite(v, f).(*Field)
		}

	// Expressions
	case *BadExpr, *Ident, *BasicLit:
		// nothing to do

	case *Ellipsis:
		if n.Elt != nil {
			n.Elt = Rewrite(v, n.Elt).(Expr)
		}

	case *FuncLit:
		n.Type = Rewrite(v, n.Type).(*FuncType)
		n.Body = Rewrite(v, n.Body).(*BlockStmt)

	case *CompositeLit:
		if n.Type != nil {
			n.Type = Rewrite(v, n.Type).(Expr)
		}
		rewriteExprList(v, n.Elts)

	case *ParenExpr:
		n.X = Rewrite(v, n.X).(Expr)

	case *SelectorExpr:
		n.X = Rewrite(v, n.X).(Expr)
		n.Sel = Rewrite(v, n.Sel).(*Ident)

	case *IndexExpr:
		n.X = Rewrite(v, n.X).(Expr)
		n.Index = Rewrite(v, n.Index).(Expr)

	case *SliceExpr:
		n.X = Rewrite(v, n.X).(Expr)
		if n.Low != nil {
			n.Low = Rewrite(v, n.Low).(Expr)
		}
		if n.High != nil {
			n.High = Rewrite(v, n.High).(Expr)
		}
		if n.Max != nil {
			n.Max = Rewrite(v, n.Max).(Expr)
		}

	case *TypeAssertExpr:
		n.X = Rewrite(v, n.X).(Expr)
		if n.Type != nil {
			n.Type = Rewrite(v, n.Type).(Expr)
		}

	case *CallExpr:
		n.Fun = Rewrite(v, n.Fun).(Expr)
		rewriteExprList(v, n.Args)

	case *StarExpr:
		n.X = Rewrite(v, n.X).(Expr)

	case *UnaryExpr:
		n.X = Rewrite(v, n.X).(Expr)

	case *BinaryExpr:
		n.X = Rewrite(v, n.X).(Expr)
		n.Y = Rewrite(v, n.Y).(Expr)

	case *KeyValueExpr:
		n.Key = Rewrite(v, n.Key).(Expr)
		n.Value = Rewrite(v, n.Value).(Expr)

	// Types
	case *ArrayType:
		if n.Len != nil {
			n.Len = Rewrite(v, n.Len).(Expr)
		}
		n.Elt = Rewrite(v, n.Elt).(Expr)

	case *StructType:
		n.Fields = Rewrite(v, n.Fields).(*FieldList)

	case *FuncType:
		if n.Params != nil {
			n.Params = Rewrite(v, n.Params).(*FieldList)
		}
		if n.Results != nil {
			n.Results = Rewrite(v, n.Results).(*FieldList)
		}

	case *InterfaceType:
		n.Methods = Rewrite(v, n.Methods).(*FieldList)

	case *MapType:
		n.Key = Rewrite(v, n.Key).(Expr)
		n.Value = Rewrite(v, n.Value).(Expr)

	case *ChanType:
		n.Value = Rewrite(v, n.Value).(Expr)

	// Statements
	case *BadStmt:
		// nothing to do

	case *DeclStmt:
		n.Decl = Rewrite(v, n.Decl).(Decl)

	case *EmptyStmt:
		// nothing to do

	case *LabeledStmt:
		n.Label = Rewrite(v, n.Label).(*Ident)
		n.Stmt = Rewrite(v, n.Stmt).(Stmt)

	case *ExprStmt:
		n.X = Rewrite(v, n.X).(Expr)

	case *SendStmt:
		n.Chan = Rewrite(v, n.Chan).(Expr)
		n.Value = Rewrite(v, n.Value).(Expr)

	case *IncDecStmt:
		n.X = Rewrite(v, n.X).(Expr)

	case *AssignStmt:
		rewriteExprList(v, n.Lhs)
		rewriteExprList(v, n.Rhs)

	case *GoStmt:
		n.Call = Rewrite(v, n.Call).(*CallExpr)

	case *DeferStmt:
		n.Call = Rewrite(v, n.Call).(*CallExpr)

	case *ReturnStmt:
		rewriteExprList(v, n.Results)

	case *BranchStmt:
		if n.Label != nil {
			n.Label = Rewrite(v, n.Label).(*Ident)
		}

	case *BlockStmt:
		rewriteStmtList(v, n.List)

	case *IfStmt:
		if n.Init != nil {
			n.Init = Rewrite(v, n.Init).(Stmt)
		}
		n.Cond = Rewrite(v, n.Cond).(Expr)
		n.Body = Rewrite(v, n.Body).(*BlockStmt)
		if n.Else != nil {
			n.Else = Rewrite(v, n.Else).(Stmt)
		}

	case *CaseClause:
		rewriteExprList(v, n.List)
		rewriteStmtList(v, n.Body)

	case *SwitchStmt:
		if n.Init != nil {
			n.Init = Rewrite(v, n.Init).(Stmt)
		}
		if n.Tag != nil {
			n.Tag = Rewrite(v, n.Tag).(Expr)
		}
		n.Body = Rewrite(v, n.Body).(*BlockStmt)

	case *TypeSwitchStmt:
		if n.Init != nil {
			n.Init = Rewrite(v, n.Init).(Stmt)
		}
		n.Assign = Rewrite(v, n.Assign).(Stmt)
		n.Body = Rewrite(v, n.Body).(*BlockStmt)

	case *CommClause:
		if n.Comm != nil {
			n.Comm = Rewrite(v, n.Comm).(Stmt)
		}
		rewriteStmtList(v, n.Body)

	case *SelectStmt:
		n.Body = Rewrite(v, n.Body).(*BlockStmt)

	case *ForStmt:
		if n.Init != nil {
			n.Init = Rewrite(v, n.Init).(Stmt)
		}
		if n.Cond != nil {
			n.Cond = Rewrite(v, n.Cond).(Expr)
		}
		if n.Post != nil {
			n.Post = Rewrite(v, n.Post).(Stmt)
		}
		n.Body = Rewrite(v, n.Body).(*BlockStmt)

	case *RangeStmt:
		if n.Key != nil {
			n.Key = Rewrite(v, n.Key).(Expr)
		}
		if n.Value != nil {
			n.Value = Rewrite(v, n.Value).(Expr)
		}
		n.X = Rewrite(v, n.X).(Expr)
		n.Body = Rewrite(v, n.Body).(*BlockStmt)

	// Declarations
	case *ImportSpec:
		if n.Doc != nil {
			n.Doc = Rewrite(v, n.Doc).(*CommentGroup)
		}
		if n.Name != nil {
			n.Name = Rewrite(v, n.Name).(*Ident)
		}
		n.Path = Rewrite(v, n.Path).(*BasicLit)
		if n.Comment != nil {
			n.Comment = Rewrite(v, n.Comment).(*CommentGroup)
		}

	case *ValueSpec:
		if n.Doc != nil {
			n.Doc = Rewrite(v, n.Doc).(*CommentGroup)
		}
		rewriteIdentList(v, n.Names)
		if n.Type != nil {
			n.Type = Rewrite(v, n.Type).(Expr)
		}
		rewriteExprList(v, n.Values)
		if n.Comment != nil {
			n.Comment = Rewrite(v, n.Comment).(*CommentGroup)
		}

	case *TypeSpec:
		if n.Doc != nil {
			n.Doc = Rewrite(v, n.Doc).(*CommentGroup)
		}
		n.Name = Rewrite(v, n.Name).(*Ident)
		n.Type = Rewrite(v, n.Type).(Expr)
		if n.Comment != nil {
			n.Comment = Rewrite(v, n.Comment).(*CommentGroup)
		}

	case *BadDecl:
		// nothing to do

	case *GenDecl:
		if n.Doc != nil {
			n.Doc = Rewrite(v, n.Doc).(*CommentGroup)
		}
		for i, s := range n.Specs {
			n.Specs[i] = Rewrite(v, s).(Spec)
		}

	case *FuncDecl:
		if n.Doc != nil {
			n.Doc = Rewrite(v, n.Doc).(*CommentGroup)
		}
		if n.Recv != nil {
			n.Recv = Rewrite(v, n.Recv).(*FieldList)
		}
		n.Name = Rewrite(v, n.Name).(*Ident)
		n.Type = Rewrite(v, n.Type).(*FuncType)
		if n.Body != nil {
			n.Body = Rewrite(v, n.Body).(*BlockStmt)
		}

	// Files and packages
	case *File:
		if n.Doc != nil {
			n.Doc = Rewrite(v, n.Doc).(*CommentGroup)
		}
		n.Name = Rewrite(v, n.Name).(*Ident)
		rewriteDeclList(v, n.Decls)
		// don't rewrite n.Comments - they have been
		// visited already through the individual
		// nodes

	case *Package:
		for i, f := range n.Files {
			n.Files[i] = Rewrite(v, f).(*File)
		}

	default:
		fmt.Printf("ast.Rewrite: unexpected node type %T", n)
		panic("ast.Rewrite")
	}

	return rewritten
}
