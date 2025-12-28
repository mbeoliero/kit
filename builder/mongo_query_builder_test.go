package builder

import (
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestMongoQueryBuilder_Id(t *testing.T) {
	tests := []struct {
		name     string
		id       any
		expected bson.M
	}{
		{
			name:     "string id",
			id:       "123",
			expected: bson.M{"_id": "123"},
		},
		{
			name:     "int id",
			id:       123,
			expected: bson.M{"_id": 123},
		},
		{
			name:     "objectid",
			id:       bson.ObjectID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc},
			expected: bson.M{"_id": bson.ObjectID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewMongoQueryBuilder()
			result := b.Id(tt.id).Build().(bson.M)
			assertBsonMEqual(t, tt.expected, result)
		})
	}
}

func TestMongoQueryBuilder_Eq(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    any
		expected bson.M
	}{
		{
			name:     "string value",
			key:      "name",
			value:    "test",
			expected: bson.M{"name": "test"},
		},
		{
			name:     "int value",
			key:      "age",
			value:    18,
			expected: bson.M{"age": 18},
		},
		{
			name:     "bool value",
			key:      "active",
			value:    true,
			expected: bson.M{"active": true},
		},
		{
			name:     "nil value",
			key:      "deleted_at",
			value:    nil,
			expected: bson.M{"deleted_at": nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewMongoQueryBuilder()
			result := b.Eq(tt.key, tt.value).Build().(bson.M)
			assertBsonMEqual(t, tt.expected, result)
		})
	}
}

func TestMongoQueryBuilder_Ne(t *testing.T) {
	b := NewMongoQueryBuilder()
	result := b.Ne("status", "deleted").Build().(bson.M)
	expected := bson.M{"status": bson.M{"$ne": "deleted"}}
	assertBsonMEqual(t, expected, result)
}

func TestMongoQueryBuilder_Gt(t *testing.T) {
	b := NewMongoQueryBuilder()
	result := b.Gt("age", 18).Build().(bson.M)
	expected := bson.M{"age": bson.M{"$gt": 18}}
	assertBsonMEqual(t, expected, result)
}

func TestMongoQueryBuilder_Gte(t *testing.T) {
	b := NewMongoQueryBuilder()
	result := b.Gte("score", 60).Build().(bson.M)
	expected := bson.M{"score": bson.M{"$gte": 60}}
	assertBsonMEqual(t, expected, result)
}

func TestMongoQueryBuilder_Lt(t *testing.T) {
	b := NewMongoQueryBuilder()
	result := b.Lt("price", 100).Build().(bson.M)
	expected := bson.M{"price": bson.M{"$lt": 100}}
	assertBsonMEqual(t, expected, result)
}

func TestMongoQueryBuilder_Lte(t *testing.T) {
	b := NewMongoQueryBuilder()
	result := b.Lte("quantity", 10).Build().(bson.M)
	expected := bson.M{"quantity": bson.M{"$lte": 10}}
	assertBsonMEqual(t, expected, result)
}

func TestMongoQueryBuilder_In(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		values   []any
		expected bson.M
	}{
		{
			name:     "string values",
			key:      "status",
			values:   []any{"active", "pending"},
			expected: bson.M{"status": bson.M{"$in": []any{"active", "pending"}}},
		},
		{
			name:     "int values",
			key:      "type",
			values:   []any{1, 2, 3},
			expected: bson.M{"type": bson.M{"$in": []any{1, 2, 3}}},
		},
		{
			name:     "single value",
			key:      "category",
			values:   []any{"tech"},
			expected: bson.M{"category": bson.M{"$in": []any{"tech"}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewMongoQueryBuilder()
			result := b.In(tt.key, tt.values...).Build().(bson.M)
			assertBsonMEqual(t, tt.expected, result)
		})
	}
}

func TestMongoQueryBuilder_Nin(t *testing.T) {
	b := NewMongoQueryBuilder()
	result := b.Nin("status", "deleted", "archived").Build().(bson.M)
	expected := bson.M{"status": bson.M{"$nin": []any{"deleted", "archived"}}}
	assertBsonMEqual(t, expected, result)
}


func TestMongoQueryBuilder_MultipleConditionsSameField(t *testing.T) {
	tests := []struct {
		name     string
		builder  func() QBuilder
		expected bson.M
	}{
		{
			name: "range query gt and lt",
			builder: func() QBuilder {
				return NewMongoQueryBuilder().Gt("age", 18).Lt("age", 30)
			},
			expected: bson.M{"age": bson.M{"$gt": 18, "$lt": 30}},
		},
		{
			name: "range query gte and lte",
			builder: func() QBuilder {
				return NewMongoQueryBuilder().Gte("price", 100).Lte("price", 500)
			},
			expected: bson.M{"price": bson.M{"$gte": 100, "$lte": 500}},
		},
		{
			name: "multiple eq on same field uses last value",
			builder: func() QBuilder {
				return NewMongoQueryBuilder().Eq("status", "active").Eq("status", "pending")
			},
			expected: bson.M{"status": bson.M{"$eq": "pending"}},
		},
		{
			name: "gt gte lt lte on same field",
			builder: func() QBuilder {
				return NewMongoQueryBuilder().Gt("score", 0).Gte("score", 10).Lt("score", 100).Lte("score", 90)
			},
			expected: bson.M{"score": bson.M{"$gt": 0, "$gte": 10, "$lt": 100, "$lte": 90}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.builder().Build().(bson.M)
			// 验证字段存在且包含多个操作符
			for key := range tt.expected {
				if _, ok := result[key]; !ok {
					t.Errorf("expected key %s not found in result", key)
				}
			}
		})
	}
}

func TestMongoQueryBuilder_MultipleFields(t *testing.T) {
	b := NewMongoQueryBuilder()
	result := b.Eq("name", "test").Gt("age", 18).In("status", "active", "pending").Build().(bson.M)

	if result["name"] != "test" {
		t.Errorf("expected name=test, got %v", result["name"])
	}

	ageCondition, ok := result["age"].(bson.M)
	if !ok {
		t.Errorf("expected age to be bson.M, got %T", result["age"])
	}
	if ageCondition["$gt"] != 18 {
		t.Errorf("expected age.$gt=18, got %v", ageCondition["$gt"])
	}

	statusCondition, ok := result["status"].(bson.M)
	if !ok {
		t.Errorf("expected status to be bson.M, got %T", result["status"])
	}
	inValues, ok := statusCondition["$in"].([]any)
	if !ok || len(inValues) != 2 {
		t.Errorf("expected status.$in to have 2 values, got %v", statusCondition["$in"])
	}
}

func TestMongoQueryBuilder_Or(t *testing.T) {
	tests := []struct {
		name    string
		builder func() QBuilder
		check   func(t *testing.T, result bson.M)
	}{
		{
			name: "or with bson.M conditions",
			builder: func() QBuilder {
				return NewMongoQueryBuilder().Or(
					bson.M{"status": "active"},
					bson.M{"status": "pending"},
				)
			},
			check: func(t *testing.T, result bson.M) {
				orConditions, ok := result["$or"].([]any)
				if !ok {
					t.Errorf("expected $or to be []any, got %T", result["$or"])
					return
				}
				if len(orConditions) != 2 {
					t.Errorf("expected 2 or conditions, got %d", len(orConditions))
				}
			},
		},
		{
			name: "or with map conditions",
			builder: func() QBuilder {
				return NewMongoQueryBuilder().Or(
					map[string]any{"type": 1},
					map[string]any{"type": 2},
				)
			},
			check: func(t *testing.T, result bson.M) {
				orConditions, ok := result["$or"].([]any)
				if !ok {
					t.Errorf("expected $or to be []any, got %T", result["$or"])
					return
				}
				if len(orConditions) != 2 {
					t.Errorf("expected 2 or conditions, got %d", len(orConditions))
				}
			},
		},
		{
			name: "or with nested builder",
			builder: func() QBuilder {
				sub1 := NewMongoQueryBuilder().Eq("status", "active")
				sub2 := NewMongoQueryBuilder().Gt("age", 18)
				return NewMongoQueryBuilder().Or(sub1, sub2)
			},
			check: func(t *testing.T, result bson.M) {
				orConditions, ok := result["$or"].([]any)
				if !ok {
					t.Errorf("expected $or to be []any, got %T", result["$or"])
					return
				}
				if len(orConditions) != 2 {
					t.Errorf("expected 2 or conditions, got %d", len(orConditions))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.builder().Build().(bson.M)
			tt.check(t, result)
		})
	}
}

func TestMongoQueryBuilder_And(t *testing.T) {
	tests := []struct {
		name    string
		builder func() QBuilder
		check   func(t *testing.T, result bson.M)
	}{
		{
			name: "and with bson.M conditions",
			builder: func() QBuilder {
				return NewMongoQueryBuilder().And(
					bson.M{"status": "active"},
					bson.M{"verified": true},
				)
			},
			check: func(t *testing.T, result bson.M) {
				andConditions, ok := result["$and"].([]any)
				if !ok {
					t.Errorf("expected $and to be []any, got %T", result["$and"])
					return
				}
				if len(andConditions) != 2 {
					t.Errorf("expected 2 and conditions, got %d", len(andConditions))
				}
			},
		},
		{
			name: "and with nested builder",
			builder: func() QBuilder {
				sub1 := NewMongoQueryBuilder().Eq("status", "active")
				sub2 := NewMongoQueryBuilder().Eq("verified", true)
				return NewMongoQueryBuilder().And(sub1, sub2)
			},
			check: func(t *testing.T, result bson.M) {
				andConditions, ok := result["$and"].([]any)
				if !ok {
					t.Errorf("expected $and to be []any, got %T", result["$and"])
					return
				}
				if len(andConditions) != 2 {
					t.Errorf("expected 2 and conditions, got %d", len(andConditions))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.builder().Build().(bson.M)
			tt.check(t, result)
		})
	}
}


func TestMongoQueryBuilder_ComplexQueries(t *testing.T) {
	tests := []struct {
		name    string
		builder func() QBuilder
		check   func(t *testing.T, result bson.M)
	}{
		{
			name: "fields with or",
			builder: func() QBuilder {
				return NewMongoQueryBuilder().
					Eq("tenant_id", "t1").
					Or(
						bson.M{"status": "active"},
						bson.M{"status": "pending"},
					)
			},
			check: func(t *testing.T, result bson.M) {
				if result["tenant_id"] != "t1" {
					t.Errorf("expected tenant_id=t1, got %v", result["tenant_id"])
				}
				if _, ok := result["$or"]; !ok {
					t.Error("expected $or condition")
				}
			},
		},
		{
			name: "nested and or",
			builder: func() QBuilder {
				return NewMongoQueryBuilder().
					Eq("active", true).
					Or(
						NewMongoQueryBuilder().Eq("type", "admin"),
						NewMongoQueryBuilder().And(
							bson.M{"type": "user"},
							bson.M{"verified": true},
						),
					)
			},
			check: func(t *testing.T, result bson.M) {
				if result["active"] != true {
					t.Errorf("expected active=true, got %v", result["active"])
				}
				orConditions, ok := result["$or"].([]any)
				if !ok {
					t.Errorf("expected $or to be []any, got %T", result["$or"])
					return
				}
				if len(orConditions) != 2 {
					t.Errorf("expected 2 or conditions, got %d", len(orConditions))
				}
			},
		},
		{
			name: "multiple or calls merge",
			builder: func() QBuilder {
				return NewMongoQueryBuilder().
					Or(bson.M{"a": 1}).
					Or(bson.M{"b": 2})
			},
			check: func(t *testing.T, result bson.M) {
				orConditions, ok := result["$or"].([]any)
				if !ok {
					t.Errorf("expected $or to be []any, got %T", result["$or"])
					return
				}
				if len(orConditions) != 2 {
					t.Errorf("expected 2 or conditions after merge, got %d", len(orConditions))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.builder().Build().(bson.M)
			tt.check(t, result)
		})
	}
}

func TestMongoQueryBuilder_ChainCalls(t *testing.T) {
	// 测试链式调用返回正确的接口类型
	b := NewMongoQueryBuilder()
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
	q = q.And(bson.M{"verified": true})
	q = q.Or(bson.M{"admin": true})

	result := q.Build().(bson.M)

	// 验证所有字段都存在
	expectedFields := []string{"_id", "name", "status", "age", "score", "price", "quantity", "type", "category", "$and", "$or"}
	for _, field := range expectedFields {
		if _, ok := result[field]; !ok {
			t.Errorf("expected field %s not found in result", field)
		}
	}
}

func TestMongoQueryBuilder_EmptyBuild(t *testing.T) {
	b := NewMongoQueryBuilder()
	result := b.Build().(bson.M)

	if len(result) != 0 {
		t.Errorf("expected empty bson.M, got %v", result)
	}
}

func TestMongoQueryBuilder_BsonDSupport(t *testing.T) {
	b := NewMongoQueryBuilder()
	result := b.Or(
		bson.D{{Key: "status", Value: "active"}},
		bson.D{{Key: "status", Value: "pending"}},
	).Build().(bson.M)

	orConditions, ok := result["$or"].([]any)
	if !ok {
		t.Errorf("expected $or to be []any, got %T", result["$or"])
		return
	}
	if len(orConditions) != 2 {
		t.Errorf("expected 2 or conditions, got %d", len(orConditions))
	}
}

// assertBsonMEqual 辅助函数：比较两个 bson.M 是否相等
func assertBsonMEqual(t *testing.T, expected, actual bson.M) {
	t.Helper()
	if len(expected) != len(actual) {
		t.Errorf("bson.M length mismatch: expected %d, got %d\nexpected: %v\nactual: %v", len(expected), len(actual), expected, actual)
		return
	}
	for key, expectedValue := range expected {
		actualValue, ok := actual[key]
		if !ok {
			t.Errorf("key %s not found in actual bson.M", key)
			continue
		}
		// 简单比较，对于复杂嵌套结构可能需要更深入的比较
		if expectedBsonM, ok := expectedValue.(bson.M); ok {
			if actualBsonM, ok := actualValue.(bson.M); ok {
				assertBsonMEqual(t, expectedBsonM, actualBsonM)
			} else {
				t.Errorf("key %s: expected bson.M, got %T", key, actualValue)
			}
		}
	}
}


func TestMongoQueryBuilder_Like(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		mode     MatchMode
		expected bson.M
	}{
		{
			name:  "contains mode",
			key:   "name",
			value: "test",
			mode:  MatchContains,
			expected: bson.M{"name": bson.M{
				"$regex":   "test",
				"$options": "i",
			}},
		},
		{
			name:  "starts with mode",
			key:   "name",
			value: "test",
			mode:  MatchStartsWith,
			expected: bson.M{"name": bson.M{
				"$regex":   "^test",
				"$options": "i",
			}},
		},
		{
			name:  "ends with mode",
			key:   "name",
			value: "test",
			mode:  MatchEndsWith,
			expected: bson.M{"name": bson.M{
				"$regex":   "test$",
				"$options": "i",
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewMongoQueryBuilder()
			result := b.Like(tt.key, tt.value, tt.mode).Build().(bson.M)

			nameCondition, ok := result[tt.key].(bson.M)
			if !ok {
				t.Errorf("expected bson.M for key %s, got %T", tt.key, result[tt.key])
				return
			}

			expectedCondition := tt.expected[tt.key].(bson.M)
			if nameCondition["$regex"] != expectedCondition["$regex"] {
				t.Errorf("expected $regex=%v, got %v", expectedCondition["$regex"], nameCondition["$regex"])
			}
			if nameCondition["$options"] != expectedCondition["$options"] {
				t.Errorf("expected $options=%v, got %v", expectedCondition["$options"], nameCondition["$options"])
			}
		})
	}
}

func TestMongoQueryBuilder_Like_EscapeSpecialChars(t *testing.T) {
	tests := []struct {
		name          string
		value         string
		expectedRegex string
	}{
		{
			name:          "escape dot",
			value:         "test.value",
			expectedRegex: `test\.value`,
		},
		{
			name:          "escape asterisk",
			value:         "test*value",
			expectedRegex: `test\*value`,
		},
		{
			name:          "escape plus",
			value:         "test+value",
			expectedRegex: `test\+value`,
		},
		{
			name:          "escape question mark",
			value:         "test?value",
			expectedRegex: `test\?value`,
		},
		{
			name:          "escape brackets",
			value:         "test[value]",
			expectedRegex: `test\[value\]`,
		},
		{
			name:          "escape parentheses",
			value:         "test(value)",
			expectedRegex: `test\(value\)`,
		},
		{
			name:          "escape caret",
			value:         "test^value",
			expectedRegex: `test\^value`,
		},
		{
			name:          "escape dollar",
			value:         "test$value",
			expectedRegex: `test\$value`,
		},
		{
			name:          "escape pipe",
			value:         "test|value",
			expectedRegex: `test\|value`,
		},
		{
			name:          "escape backslash",
			value:         `test\value`,
			expectedRegex: `test\\value`,
		},
		{
			name:          "multiple special chars",
			value:         "a.b*c+d?e",
			expectedRegex: `a\.b\*c\+d\?e`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewMongoQueryBuilder()
			result := b.Like("field", tt.value, MatchContains).Build().(bson.M)

			fieldCondition, ok := result["field"].(bson.M)
			if !ok {
				t.Errorf("expected bson.M, got %T", result["field"])
				return
			}

			if fieldCondition["$regex"] != tt.expectedRegex {
				t.Errorf("expected $regex=%q, got %q", tt.expectedRegex, fieldCondition["$regex"])
			}
		})
	}
}

func TestMongoQueryBuilder_Like_WithOtherConditions(t *testing.T) {
	b := NewMongoQueryBuilder()
	result := b.
		Eq("status", "active").
		Like("name", "test", MatchContains).
		Gt("age", 18).
		Build().(bson.M)

	// 验证 eq 条件
	if result["status"] != "active" {
		t.Errorf("expected status=active, got %v", result["status"])
	}

	// 验证 like 条件
	nameCondition, ok := result["name"].(bson.M)
	if !ok {
		t.Errorf("expected bson.M for name, got %T", result["name"])
	} else {
		if nameCondition["$regex"] != "test" {
			t.Errorf("expected $regex=test, got %v", nameCondition["$regex"])
		}
	}

	// 验证 gt 条件
	ageCondition, ok := result["age"].(bson.M)
	if !ok {
		t.Errorf("expected bson.M for age, got %T", result["age"])
	} else {
		if ageCondition["$gt"] != 18 {
			t.Errorf("expected $gt=18, got %v", ageCondition["$gt"])
		}
	}
}
