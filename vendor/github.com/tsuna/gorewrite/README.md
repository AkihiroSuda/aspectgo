# gorewrite
Go has pretty amazing tooling to parse Go code and pretty-print the AST back
into the original code.  It also has a utility to visit an AST and modify
parts of it as you Go, using [`ast.Walk()`](http://golang.org/pkg/go/ast/#Walk)
with a [`Visitor`](http://golang.org/pkg/go/ast/#Visitor).

This is fine when you just to tweak the contents of existing AST nodes, or
change the sub-nodes of a given node, but it doesn't let you replace an existing
node with another one.  This is needed, for instance, if you want to go from one
type of Node to another (e.g. when you want to rewrite one type of expression as
another type of expression).

**gorewrite** gives you the exact same thing as `Walk()` with a `Visitor` except
that your `Visitor`, called a `Rewriter` can return a modified `ast.Node`, and
that you need to call `Rewrite()` instead of `Walk()`.
