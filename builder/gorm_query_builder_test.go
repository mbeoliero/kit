package builder

import (
	"testing"

	"gorm.io/gorm/clause"
)

func TestGormQueryBuilder_Id(t *testing.T) {
	tests := []struct {
		name  string
		id    any
		check func(t *testing.T, expr clause.Expression)
	}{
		{
			name: "string id",
			id:   "123",
			check: func(t *testing.T, expr clause.Expression) {
				eq, ok := expr.(clause.Eq)
				if !ok {
					t.Errorf("expected clause.Eq, got %T", expr)
					return
				}
				col, ok := eq.Column.(clause.Column)
				if !ok || col.Name != "id" {
					t.Errorf("expected column name 'id', got %v", eq.Column)
				}
				if eq.Value != "123" {
					t.Errorf("expected value '123', got %v", eq.Value)
				}
			},
		},
		{
			name: "int id",
			id:   456,
			check: func(t *testing.T, expr clause.Expression) {
				eq, ok := expr.(clause.Eq)
				if !ok {
					t.Errorf("expected clause.Eq, got %T", expr)
					return
				}
				if eq.Value != 456 {
					t.Errorf("expected value 456, got %v", eq.Value)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewGormQueryBuilder()
			result := b.Id(tt.id).Build()
			if result == nil {
				t.Error("expected non-nil result")
				return
			}
			tt.check(t, result.(clause.Expression))
		})
	}
}

func TestGormQueryBuilder_Eq(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value any
	}{
		{"string value", "name", "test"},
		{"int value", "age", 18},
		{"bool value", "active", true},
		{"nil value", "deleted_at", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewGormQueryBuilder()
			result := b.Eq(tt.key, tt.value).Build()
			eq, ok := result.(clause.Eq)
			if !ok {
				t.Errorf("expected clause.Eq, got %T", result)
				return
			}
			col, ok := eq.Column.(clause.Column)
			if !ok || col.Name != tt.key {
				t.Errorf("expected column name '%s', got %v", tt.key, eq.Column)
			}
			if eq.Value != tt.value {
				t.Errorf("expected value %v, got %v", tt.value, eq.Value)
			}
		})
	}
}

func TestGormQueryBuilder_Ne(t *testing.T) {
	b := NewGormQueryBuilder()
	result := b.Ne("status", "deleted").Build()
	neq, ok := result.(clause.Neq)
	if !ok {
		t.Errorf("expected clause.Neq, got %T", result)
		return
	}
	col, ok := neq.Column.(clause.Column)
	if !ok || col.Name != "status" {
		t.Errorf("expected column name 'status', got %v", neq.Column)
	}
	if neq.Value != "deleted" {
		t.Errorf("expected value 'deleted', got %v", neq.Value)
	}
}

func TestGormQueryBuilder_Gt(t *testing.T) {
	b := NewGormQueryBuilder()
	result := b.Gt("age", 18).Build()
	gt, ok := result.(clause.Gt)
	if !ok {
		t.Errorf("expected clause.Gt, got %T", result)
		return
	}
	col, ok := gt.Column.(clause.Column)
	if !ok || col.Name != "age" {
		t.Errorf("expected column name 'age', got %v", gt.Column)
	}
	if gt.Value != 18 {
		t.Errorf("expected value 18, got %v", gt.Value)
	}
}

func TestGormQueryBuilder_Gte(t *testing.T) {
	b := NewGormQueryBuilder()
	result := b.Gte("score", 60).Build()
	gte, ok := result.(clause.Gte)
	if !ok {
		t.Errorf("expected clause.Gte, got %T", result)
		return
	}
	col, ok := gte.Column.(clause.Column)
	if !ok || col.Name != "score" {
		t.Errorf("expected column name 'score', got %v", gte.Column)
	}
	if gte.Value != 60 {
		t.Errorf("expected value 60, got %v", gte.Value)
	}
}

func TestGormQueryBuilder_Lt(t *testing.T) {
	b := NewGormQueryBuilder()
	result := b.Lt("price", 100).Build()
	lt, ok := result.(clause.Lt)
	if !ok {
		t.Errorf("expected clause.Lt, got %T", result)
		return
	}
	col, ok := lt.Column.(clause.Column)
	if !ok || col.Name != "price" {
		t.Errorf("expected column name 'price', got %v", lt.Column)
	}
	if lt.Value != 100 {
		t.Errorf("expected value 100, got %v", lt.Value)
	}
}

func TestGormQueryBuilder_Lte(t *testing.T) {
	b := NewGormQueryBuilder()
	result := b.Lte("quantity", 10).Build()
	lte, ok := result.(clause.Lte)
	if !ok {
		t.Errorf("expected clause.Lte, got %T", result)
		return
	}
	col, ok := lte.Column.(clause.Column)
	if !ok || col.Name != "quantity" {
		t.Errorf("expected column name 'quantity', got %v", lte.Column)
	}
	if lte.Value != 10 {
		t.Errorf("expected value 10, got %v", lte.Value)
	}
}

func TestGormQueryBuilder_In(t *testing.T) {
	tests := []struct {
		name   string
		key    string
		values []any
	}{
		{"string values", "status", []any{"active", "pending"}},
		{"int values", "type", []any{1, 2, 3}},
		{"single value", "category", []any{"tech"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewGormQueryBuilder()
			result := b.In(tt.key, tt.values...).Build()
			in, ok := result.(clause.IN)
			if !ok {
				t.Errorf("expected clause.IN, got %T", result)
				return
			}
			col, ok := in.Column.(clause.Column)
			if !ok || col.Name != tt.key {
				t.Errorf("expected column name '%s', got %v", tt.key, in.Column)
			}
			if len(in.Values) != len(tt.values) {
				t.Errorf("expected %d values, got %d", len(tt.values), len(in.Values))
			}
		})
	}
}

func TestGormQueryBuilder_Nin(t *testing.T) {
	b := NewGormQueryBuilder()
	result := b.Nin("status", "deleted", "archived").Build()

	// Nin 应该返回 NOT(IN(...))，clause.Not 是函数返回 clause.NotConditions
	notExpr, ok := result.(clause.NotConditions)
	if !ok {
		t.Errorf("expected clause.NotConditions, got %T", result)
		return
	}
	if len(notExpr.Exprs) != 1 {
		t.Errorf("expected 1 expression in Not, got %d", len(notExpr.Exprs))
		return
	}
	in, ok := notExpr.Exprs[0].(clause.IN)
	if !ok {
		t.Errorf("expected clause.IN inside Not, got %T", notExpr.Exprs[0])
		return
	}
	col, ok := in.Column.(clause.Column)
	if !ok || col.Name != "status" {
		t.Errorf("expected column name 'status', got %v", in.Column)
	}
	if len(in.Values) != 2 {
		t.Errorf("expected 2 values, got %d", len(in.Values))
	}
}

func TestGormQueryBuilder_MultipleConditionsSameField(t *testing.T) {
	tests := []struct {
		name         string
		builder      func() QBuilder
		expectedType string
		checkCount   int
	}{
		{
			name: "range query gt and lt",
			builder: func() QBuilder {
				return NewGormQueryBuilder().Gt("age", 18).Lt("age", 30)
			},
			expectedType: "And",
			checkCount:   2,
		},
		{
			name: "range query gte and lte",
			builder: func() QBuilder {
				return NewGormQueryBuilder().Gte("price", 100).Lte("price", 500)
			},
			expectedType: "And",
			checkCount:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.builder().Build()
			andExpr, ok := result.(clause.AndConditions)
			if !ok {
				t.Errorf("expected clause.AndConditions, got %T", result)
				return
			}
			if len(andExpr.Exprs) != tt.checkCount {
				t.Errorf("expected %d expressions, got %d", tt.checkCount, len(andExpr.Exprs))
			}
		})
	}
}

func TestGormQueryBuilder_MultipleFields(t *testing.T) {
	b := NewGormQueryBuilder()
	result := b.Eq("name", "test").Gt("age", 18).In("status", "active", "pending").Build()

	andExpr, ok := result.(clause.AndConditions)
	if !ok {
		t.Errorf("expected clause.AndConditions, got %T", result)
		return
	}
	if len(andExpr.Exprs) != 3 {
		t.Errorf("expected 3 expressions, got %d", len(andExpr.Exprs))
	}
}

func TestGormQueryBuilder_Or(t *testing.T) {
	tests := []struct {
		name    string
		builder func() QBuilder
		check   func(t *testing.T, result any)
	}{
		{
			name: "or with map conditions",
			builder: func() QBuilder {
				return NewGormQueryBuilder().Or(
					map[string]any{"status": "active"},
					map[string]any{"status": "pending"},
				)
			},
			check: func(t *testing.T, result any) {
				orExpr, ok := result.(clause.OrConditions)
				if !ok {
					t.Errorf("expected clause.OrConditions, got %T", result)
					return
				}
				if len(orExpr.Exprs) != 2 {
					t.Errorf("expected 2 or conditions, got %d", len(orExpr.Exprs))
				}
			},
		},
		{
			name: "or with nested builder",
			builder: func() QBuilder {
				sub1 := NewGormQueryBuilder().Eq("status", "active")
				sub2 := NewGormQueryBuilder().Gt("age", 18)
				return NewGormQueryBuilder().Or(sub1, sub2)
			},
			check: func(t *testing.T, result any) {
				orExpr, ok := result.(clause.OrConditions)
				if !ok {
					t.Errorf("expected clause.OrConditions, got %T", result)
					return
				}
				if len(orExpr.Exprs) != 2 {
					t.Errorf("expected 2 or conditions, got %d", len(orExpr.Exprs))
				}
			},
		},
		{
			name: "or with clause.Expression",
			builder: func() QBuilder {
				return NewGormQueryBuilder().Or(
					clause.Eq{Column: clause.Column{Name: "a"}, Value: 1},
					clause.Eq{Column: clause.Column{Name: "b"}, Value: 2},
				)
			},
			check: func(t *testing.T, result any) {
				orExpr, ok := result.(clause.OrConditions)
				if !ok {
					t.Errorf("expected clause.OrConditions, got %T", result)
					return
				}
				if len(orExpr.Exprs) != 2 {
					t.Errorf("expected 2 or conditions, got %d", len(orExpr.Exprs))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.builder().Build()
			tt.check(t, result)
		})
	}
}

func TestGormQueryBuilder_And(t *testing.T) {
	tests := []struct {
		name    string
		builder func() QBuilder
		check   func(t *testing.T, result any)
	}{
		{
			name: "and with map conditions",
			builder: func() QBuilder {
				return NewGormQueryBuilder().And(
					map[string]any{"status": "active"},
					map[string]any{"verified": true},
				)
			},
			check: func(t *testing.T, result any) {
				andExpr, ok := result.(clause.AndConditions)
				if !ok {
					t.Errorf("expected clause.AndConditions, got %T", result)
					return
				}
				if len(andExpr.Exprs) != 2 {
					t.Errorf("expected 2 and conditions, got %d", len(andExpr.Exprs))
				}
			},
		},
		{
			name: "and with nested builder",
			builder: func() QBuilder {
				sub1 := NewGormQueryBuilder().Eq("status", "active")
				sub2 := NewGormQueryBuilder().Eq("verified", true)
				return NewGormQueryBuilder().And(sub1, sub2)
			},
			check: func(t *testing.T, result any) {
				andExpr, ok := result.(clause.AndConditions)
				if !ok {
					t.Errorf("expected clause.AndConditions, got %T", result)
					return
				}
				if len(andExpr.Exprs) != 2 {
					t.Errorf("expected 2 and conditions, got %d", len(andExpr.Exprs))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.builder().Build()
			tt.check(t, result)
		})
	}
}

func TestGormQueryBuilder_ComplexQueries(t *testing.T) {
	tests := []struct {
		name    string
		builder func() QBuilder
		check   func(t *testing.T, result any)
	}{
		{
			name: "fields with or",
			builder: func() QBuilder {
				return NewGormQueryBuilder().
					Eq("tenant_id", "t1").
					Or(
						map[string]any{"status": "active"},
						map[string]any{"status": "pending"},
					)
			},
			check: func(t *testing.T, result any) {
				andExpr, ok := result.(clause.AndConditions)
				if !ok {
					t.Errorf("expected clause.AndConditions, got %T", result)
					return
				}
				// 应该有 Eq 和 Or 两个条件
				if len(andExpr.Exprs) != 2 {
					t.Errorf("expected 2 expressions, got %d", len(andExpr.Exprs))
				}
			},
		},
		{
			name: "nested and or",
			builder: func() QBuilder {
				return NewGormQueryBuilder().
					Eq("active", true).
					Or(
						NewGormQueryBuilder().Eq("type", "admin"),
						NewGormQueryBuilder().And(
							map[string]any{"type": "user"},
							map[string]any{"verified": true},
						),
					)
			},
			check: func(t *testing.T, result any) {
				andExpr, ok := result.(clause.AndConditions)
				if !ok {
					t.Errorf("expected clause.AndConditions, got %T", result)
					return
				}
				if len(andExpr.Exprs) != 2 {
					t.Errorf("expected 2 expressions, got %d", len(andExpr.Exprs))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.builder().Build()
			tt.check(t, result)
		})
	}
}

func TestGormQueryBuilder_ChainCalls(t *testing.T) {
	b := NewGormQueryBuilder()
	var q QBuilder = b

	q = q.Id("123")
	q = q.Eq("name", "test")
	q = q.Ne("status", "deleted")
	q = q.Gt("age", 18)
	q = q.Gte("score", 60)
	q = q.Lt("price", 100)
	q = q.Lte("quantity", 10)
	q = q.In("type", 1, 2, 3)
	q = q.Nin("category", "a", "b")
	q = q.And(map[string]any{"verified": true})
	q = q.Or(map[string]any{"admin": true})

	result := q.Build()
	if result == nil {
		t.Error("expected non-nil result")
		return
	}

	andExpr, ok := result.(clause.AndConditions)
	if !ok {
		t.Errorf("expected clause.AndConditions, got %T", result)
		return
	}

	// 应该有 11 个条件（9个字段条件 + 1个And + 1个Or）
	if len(andExpr.Exprs) != 11 {
		t.Errorf("expected 11 expressions, got %d", len(andExpr.Exprs))
	}
}

func TestGormQueryBuilder_EmptyBuild(t *testing.T) {
	b := NewGormQueryBuilder()
	result := b.Build()

	if result != nil {
		t.Errorf("expected nil for empty builder, got %v", result)
	}
}

func TestGormQueryBuilder_SingleCondition(t *testing.T) {
	// 单个条件不应该被包装在 And 中
	b := NewGormQueryBuilder()
	result := b.Eq("name", "test").Build()

	_, ok := result.(clause.Eq)
	if !ok {
		t.Errorf("expected clause.Eq for single condition, got %T", result)
	}
}

func TestGormQueryBuilder_MapWithMultipleKeys(t *testing.T) {
	b := NewGormQueryBuilder()
	result := b.Or(
		map[string]any{"a": 1, "b": 2},
		map[string]any{"c": 3},
	).Build()

	orExpr, ok := result.(clause.OrConditions)
	if !ok {
		t.Errorf("expected clause.OrConditions, got %T", result)
		return
	}
	if len(orExpr.Exprs) != 2 {
		t.Errorf("expected 2 or conditions, got %d", len(orExpr.Exprs))
	}

	// 第一个条件应该是 And（因为 map 有多个 key）
	firstExpr := orExpr.Exprs[0]
	_, isAnd := firstExpr.(clause.AndConditions)
	_, isEq := firstExpr.(clause.Eq)
	if !isAnd && !isEq {
		t.Errorf("expected first condition to be And or Eq, got %T", firstExpr)
	}
}

func TestGormQueryBuilder_InterfaceCompliance(t *testing.T) {
	// 确保 GormQueryBuilder 实现了 QBuilder 接口
	var _ QBuilder = (*GormQueryBuilder)(nil)
	var _ IBuilder = (*GormQueryBuilder)(nil)
}

func TestMongoQueryBuilder_InterfaceCompliance(t *testing.T) {
	// 确保 MongoQueryBuilder 实现了 QBuilder 接口
	var _ QBuilder = (*MongoQueryBuilder)(nil)
	var _ IBuilder = (*MongoQueryBuilder)(nil)
}

func TestGormQueryBuilder_Like(t *testing.T) {
	tests := []struct {
		name            string
		key             string
		value           string
		mode            MatchMode
		expectedPattern string
	}{
		{
			name:            "contains mode",
			key:             "name",
			value:           "test",
			mode:            MatchContains,
			expectedPattern: "%test%",
		},
		{
			name:            "starts with mode",
			key:             "name",
			value:           "test",
			mode:            MatchStartsWith,
			expectedPattern: "test%",
		},
		{
			name:            "ends with mode",
			key:             "name",
			value:           "test",
			mode:            MatchEndsWith,
			expectedPattern: "%test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewGormQueryBuilder()
			result := b.Like(tt.key, tt.value, tt.mode).Build()

			likeExpr, ok := result.(clause.Like)
			if !ok {
				t.Errorf("expected clause.Like, got %T", result)
				return
			}

			col, ok := likeExpr.Column.(clause.Column)
			if !ok || col.Name != tt.key {
				t.Errorf("expected column name '%s', got %v", tt.key, likeExpr.Column)
			}

			if likeExpr.Value != tt.expectedPattern {
				t.Errorf("expected pattern '%s', got '%v'", tt.expectedPattern, likeExpr.Value)
			}
		})
	}
}

func TestGormQueryBuilder_Like_WithOtherConditions(t *testing.T) {
	b := NewGormQueryBuilder()
	result := b.
		Eq("status", "active").
		Like("name", "test", MatchContains).
		Gt("age", 18).
		Build()

	andExpr, ok := result.(clause.AndConditions)
	if !ok {
		t.Errorf("expected clause.AndConditions, got %T", result)
		return
	}

	if len(andExpr.Exprs) != 3 {
		t.Errorf("expected 3 expressions, got %d", len(andExpr.Exprs))
		return
	}

	// 验证包含 Like 表达式
	hasLike := false
	for _, expr := range andExpr.Exprs {
		if _, ok := expr.(clause.Like); ok {
			hasLike = true
			break
		}
	}
	if !hasLike {
		t.Error("expected to find clause.Like in expressions")
	}
}

func TestGormQueryBuilder_Like_ChainCalls(t *testing.T) {
	b := NewGormQueryBuilder()
	var q QBuilder = b

	// 测试链式调用返回正确的接口类型
	q = q.Like("name", "test", MatchContains)
	q = q.Like("email", "example", MatchEndsWith)

	result := q.Build()
	andExpr, ok := result.(clause.AndConditions)
	if !ok {
		t.Errorf("expected clause.AndConditions, got %T", result)
		return
	}

	if len(andExpr.Exprs) != 2 {
		t.Errorf("expected 2 expressions, got %d", len(andExpr.Exprs))
	}
}

func TestGormQueryBuilder_Like_EmptyValue(t *testing.T) {
	b := NewGormQueryBuilder()
	result := b.Like("name", "", MatchContains).Build()

	likeExpr, ok := result.(clause.Like)
	if !ok {
		t.Errorf("expected clause.Like, got %T", result)
		return
	}

	// 空值应该生成 "%%"
	if likeExpr.Value != "%%" {
		t.Errorf("expected pattern '%%%%', got '%v'", likeExpr.Value)
	}
}
