package lint

import (
	"go/ast"
	"go/token"

	"github.com/go-toolsmith/astcopy"
	"github.com/go-toolsmith/astequal"
)

func init() {
	addChecker(&assignOpChecker{}, attrExperimental)
}

type assignOpChecker struct {
	checkerBase
}

func (c *assignOpChecker) InitDocumentation(d *Documentation) {
	d.Summary = "Detects assignments that can be simplified by using assignment operators"
	d.Before = `x = x * 2`
	d.After = `x *= 2`
}

func (c *assignOpChecker) VisitStmt(stmt ast.Stmt) {
	assign, ok := stmt.(*ast.AssignStmt)
	cond := ok &&
		assign.Tok == token.ASSIGN &&
		len(assign.Lhs) == 1 &&
		len(assign.Rhs) == 1 &&
		isSafeExpr(c.ctx.typesInfo, assign.Lhs[0])
	if !cond {
		return
	}

	// TODO(quasilyte): can take commutativity into account.
	expr, ok := assign.Rhs[0].(*ast.BinaryExpr)
	if !ok || !astequal.Expr(assign.Lhs[0], expr.X) {
		return
	}

	// TODO(quasilyte): perform unparen?
	switch expr.Op {
	case token.MUL:
		c.warn(assign, token.MUL_ASSIGN, expr.Y)
	case token.QUO:
		c.warn(assign, token.QUO_ASSIGN, expr.Y)
	case token.REM:
		c.warn(assign, token.REM_ASSIGN, expr.Y)
	case token.ADD:
		c.warn(assign, token.ADD_ASSIGN, expr.Y)
	case token.SUB:
		c.warn(assign, token.SUB_ASSIGN, expr.Y)
	case token.AND:
		c.warn(assign, token.AND_ASSIGN, expr.Y)
	case token.OR:
		c.warn(assign, token.OR_ASSIGN, expr.Y)
	case token.XOR:
		c.warn(assign, token.XOR_ASSIGN, expr.Y)
	case token.SHL:
		c.warn(assign, token.SHL_ASSIGN, expr.Y)
	case token.SHR:
		c.warn(assign, token.SHR_ASSIGN, expr.Y)
	case token.AND_NOT:
		c.warn(assign, token.AND_NOT_ASSIGN, expr.Y)
	}
}

func (c *assignOpChecker) warn(cause *ast.AssignStmt, op token.Token, rhs ast.Expr) {
	suggestion := astcopy.AssignStmt(cause)
	suggestion.Tok = op
	suggestion.Rhs[0] = rhs
	c.ctx.Warn(cause, "replace `%s` with `%s`", cause, suggestion)
}
