package builder

import (
	"gorm.io/gorm/clause"
)

// GormQueryBuilder GORM 查询构建器
type GormQueryBuilder struct {
	conditions *QueryConditions
	idField    string
}

// NewGormQueryBuilder 创建 GORM 查询构建器
func NewGormQueryBuilder() *GormQueryBuilder {
	return &GormQueryBuilder{
		conditions: NewQueryConditions(),
		idField:    "id",
	}
}

// Id 设置 ID 条件
func (b *GormQueryBuilder) Id(id any) QBuilder {
	b.conditions.AddCondition(b.idField, OpEq, id)
	return b
}

// Eq 等于条件
func (b *GormQueryBuilder) Eq(key string, value any) QBuilder {
	b.conditions.AddCondition(key, OpEq, value)
	return b
}

// Ne 不等于条件
func (b *GormQueryBuilder) Ne(key string, value any) QBuilder {
	b.conditions.AddCondition(key, OpNe, value)
	return b
}

// Gt 大于条件
func (b *GormQueryBuilder) Gt(key string, value any) QBuilder {
	b.conditions.AddCondition(key, OpGt, value)
	return b
}

// Gte 大于等于条件
func (b *GormQueryBuilder) Gte(key string, value any) QBuilder {
	b.conditions.AddCondition(key, OpGte, value)
	return b
}

// Lt 小于条件
func (b *GormQueryBuilder) Lt(key string, value any) QBuilder {
	b.conditions.AddCondition(key, OpLt, value)
	return b
}

// Lte 小于等于条件
func (b *GormQueryBuilder) Lte(key string, value any) QBuilder {
	b.conditions.AddCondition(key, OpLte, value)
	return b
}

// In 包含条件
func (b *GormQueryBuilder) In(key string, value ...any) QBuilder {
	b.conditions.AddCondition(key, OpIn, value)
	return b
}

// Nin 不包含条件
func (b *GormQueryBuilder) Nin(key string, value ...any) QBuilder {
	b.conditions.AddCondition(key, OpNin, value)
	return b
}

// Like 模糊匹配条件
func (b *GormQueryBuilder) Like(key string, value string, mode MatchMode) QBuilder {
	pattern := b.buildPattern(value, mode)
	b.conditions.AddCondition(key, OpLike, pattern)
	return b
}

// buildPattern 构建 MySQL LIKE 模式
func (b *GormQueryBuilder) buildPattern(value string, mode MatchMode) string {
	switch mode {
	case MatchStartsWith:
		return value + "%"
	case MatchEndsWith:
		return "%" + value
	case MatchContains:
		fallthrough
	default:
		return "%" + value + "%"
	}
}

// And 逻辑与
func (b *GormQueryBuilder) And(conditions ...any) QBuilder {
	b.conditions.AddLogicalGroup("and", conditions)
	return b
}

// Or 逻辑或
func (b *GormQueryBuilder) Or(conditions ...any) QBuilder {
	b.conditions.AddLogicalGroup("or", conditions)
	return b
}

// Build 构建 GORM 查询条件，返回 clause.Expression
func (b *GormQueryBuilder) Build() any {
	var exprs []clause.Expression

	// 处理字段条件
	for field, conditions := range b.conditions.Fields {
		col := clause.Column{Name: field}
		for _, cond := range conditions {
			exprs = append(exprs, b.buildConditionExpr(col, cond))
		}
	}

	// 处理逻辑组
	for _, group := range b.conditions.LogicalGroups {
		expr := b.buildLogicalGroupExpr(group)
		if expr != nil {
			exprs = append(exprs, expr)
		}
	}

	if len(exprs) == 0 {
		return nil
	}
	if len(exprs) == 1 {
		return exprs[0]
	}
	return clause.And(exprs...)
}

// buildConditionExpr 构建单个条件表达式
func (b *GormQueryBuilder) buildConditionExpr(col clause.Column, cond Condition) clause.Expression {
	switch cond.Op {
	case OpEq:
		return clause.Eq{Column: col, Value: cond.Value}
	case OpNe:
		return clause.Neq{Column: col, Value: cond.Value}
	case OpGt:
		return clause.Gt{Column: col, Value: cond.Value}
	case OpGte:
		return clause.Gte{Column: col, Value: cond.Value}
	case OpLt:
		return clause.Lt{Column: col, Value: cond.Value}
	case OpLte:
		return clause.Lte{Column: col, Value: cond.Value}
	case OpIn:
		values, _ := cond.Value.([]any)
		return clause.IN{Column: col, Values: values}
	case OpNin:
		values, _ := cond.Value.([]any)
		return clause.Not(clause.IN{Column: col, Values: values})
	case OpLike:
		pattern, _ := cond.Value.(string)
		return clause.Like{Column: col, Value: pattern}
	default:
		return clause.Eq{Column: col, Value: cond.Value}
	}
}

// buildLogicalGroupExpr 构建逻辑组表达式
func (b *GormQueryBuilder) buildLogicalGroupExpr(group LogicalGroup) clause.Expression {
	if len(group.Conditions) == 0 {
		return nil
	}

	var exprs []clause.Expression
	for _, cond := range group.Conditions {
		expr := b.convertToExpr(cond)
		if expr != nil {
			exprs = append(exprs, expr)
		}
	}

	if len(exprs) == 0 {
		return nil
	}

	if group.Type == "or" {
		return clause.Or(exprs...)
	}
	return clause.And(exprs...)
}

// convertToExpr 转换条件为表达式
func (b *GormQueryBuilder) convertToExpr(cond any) clause.Expression {
	switch v := cond.(type) {
	case clause.Expression:
		return v
	case map[string]any:
		var exprs []clause.Expression
		for key, value := range v {
			exprs = append(exprs, clause.Eq{Column: clause.Column{Name: key}, Value: value})
		}
		if len(exprs) == 1 {
			return exprs[0]
		}
		return clause.And(exprs...)
	default:
		if builder, ok := cond.(IBuilder); ok {
			if expr, ok := builder.Build().(clause.Expression); ok {
				return expr
			}
		}
		return nil
	}
}
