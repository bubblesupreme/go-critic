package lint

import (
	"go/ast"
	"go/token"

	"github.com/go-toolsmith/astcast"

	"github.com/go-toolsmith/astequal"
)

func init() {
	addChecker(&appendCombineChecker{})
}

type appendCombineChecker struct {
	checkerBase
}

func (c *appendCombineChecker) InitDocumentation(d *Documentation) {
	d.Summary = "Detects `append` chains to the same slice that can be done in a single `append` call"
	d.Before = `
xs = append(xs, 1)
xs = append(xs, 2)`
	d.After = `xs = append(xs, 1, 2)`
}

func (c *appendCombineChecker) VisitStmtList(list []ast.Stmt) {
	var cause ast.Node // First append
	var slice ast.Expr // Slice being appended to
	chain := 0         // How much appends in a row we've seen

	// Break the chain.
	// If enough appends are in chain, print warning.
	flush := func() {
		if chain > 1 {
			c.warn(cause, chain)
		}
		chain = 0
		slice = nil
	}

	for _, stmt := range list {
		call := c.matchAppend(stmt, slice)
		if call == nil {
			flush()
			continue
		}

		if chain == 0 {
			// First append in a chain.
			chain = 1
			slice = call.Args[0]
			cause = stmt
		} else {
			chain++
		}
	}

	// Required for printing chains that consist of trailing
	// statements from the list.
	flush()
}

func (c *appendCombineChecker) matchAppend(stmt ast.Stmt, slice ast.Expr) *ast.CallExpr {
	// Seeking for:
	//	slice = append(slice, xs...)
	// xs are 0-N append arguments, but not variadic argument,
	// because it makes append combining impossible.

	assign := astcast.ToAssignStmt(stmt)
	if len(assign.Lhs) != 1 || len(assign.Rhs) != 1 {
		return nil
	}

	call, ok := assign.Rhs[0].(*ast.CallExpr)
	{
		cond := ok &&
			qualifiedName(call.Fun) == "append" &&
			call.Ellipsis == token.NoPos &&
			astequal.Expr(assign.Lhs[0], call.Args[0])
		if !cond {
			return nil
		}
	}

	// Check that current append slice match previous append slice.
	// Otherwise we should break the chain.
	if slice == nil || astequal.Expr(slice, call.Args[0]) {
		return call
	}
	return nil
}

func (c *appendCombineChecker) warn(cause ast.Node, chain int) {
	c.ctx.Warn(cause, "can combine chain of %d appends into one", chain)
}
