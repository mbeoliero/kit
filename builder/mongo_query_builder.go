package builder

import (
	"regexp"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// regexSpecialChars 正则表达式特殊字符
var regexSpecialChars = regexp.MustCompile(`([.*+?^${}()|\[\]\\])`)

// escapeRegex 转义正则表达式特殊字符
func escapeRegex(s string) string {
	return regexSpecialChars.ReplaceAllString(s, `\$1`)
}

// mongoOpMap MongoDB 操作符映射
var mongoOpMap = map[Op]string{
	OpEq:  "$eq",
	OpNe:  "$ne",
	OpGt:  "$gt",
	OpGte: "$gte",
	OpLt:  "$lt",
	OpLte: "$lte",
	OpIn:  "$in",
	OpNin: "$nin",
}

// MongoQueryBuilder MongoDB 查询构建器
type MongoQueryBuilder struct {
	conditions *QueryConditions
	idField    string
}

// NewMongoQueryBuilder 创建 MongoDB 查询构建器
func NewMongoQueryBuilder() *MongoQueryBuilder {
	return &MongoQueryBuilder{
		conditions: NewQueryConditions(),
		idField:    "_id",
	}
}

// Id 设置 ID 条件
func (b *MongoQueryBuilder) Id(id any) QBuilder {
	b.conditions.AddCondition(b.idField, OpEq, id)
	return b
}

// Eq 等于条件
func (b *MongoQueryBuilder) Eq(key string, value any) QBuilder {
	b.conditions.AddCondition(key, OpEq, value)
	return b
}

// Ne 不等于条件
func (b *MongoQueryBuilder) Ne(key string, value any) QBuilder {
	b.conditions.AddCondition(key, OpNe, value)
	return b
}

// Gt 大于条件
func (b *MongoQueryBuilder) Gt(key string, value any) QBuilder {
	b.conditions.AddCondition(key, OpGt, value)
	return b
}

// Gte 大于等于条件
func (b *MongoQueryBuilder) Gte(key string, value any) QBuilder {
	b.conditions.AddCondition(key, OpGte, value)
	return b
}

// Lt 小于条件
func (b *MongoQueryBuilder) Lt(key string, value any) QBuilder {
	b.conditions.AddCondition(key, OpLt, value)
	return b
}

// Lte 小于等于条件
func (b *MongoQueryBuilder) Lte(key string, value any) QBuilder {
	b.conditions.AddCondition(key, OpLte, value)
	return b
}

// In 包含条件
func (b *MongoQueryBuilder) In(key string, value ...any) QBuilder {
	b.conditions.AddCondition(key, OpIn, value)
	return b
}

// Nin 不包含条件
func (b *MongoQueryBuilder) Nin(key string, value ...any) QBuilder {
	b.conditions.AddCondition(key, OpNin, value)
	return b
}

// Like 模糊匹配条件
func (b *MongoQueryBuilder) Like(key string, value string, mode MatchMode) QBuilder {
	pattern := b.buildPattern(value, mode)
	b.conditions.AddCondition(key, OpLike, pattern)
	return b
}

// buildPattern 构建 MongoDB 正则表达式模式
func (b *MongoQueryBuilder) buildPattern(value string, mode MatchMode) string {
	// 转义正则特殊字符
	escaped := escapeRegex(value)
	switch mode {
	case MatchStartsWith:
		return "^" + escaped
	case MatchEndsWith:
		return escaped + "$"
	case MatchContains:
		fallthrough
	default:
		return escaped
	}
}

// And 逻辑与
func (b *MongoQueryBuilder) And(conditions ...any) QBuilder {
	b.conditions.AddLogicalGroup("and", conditions)
	return b
}

// Or 逻辑或
func (b *MongoQueryBuilder) Or(conditions ...any) QBuilder {
	b.conditions.AddLogicalGroup("or", conditions)
	return b
}

// Build 构建 MongoDB 查询条件，返回 bson.M
func (b *MongoQueryBuilder) Build() any {
	result := bson.M{}

	// 处理字段条件
	for field, conditions := range b.conditions.Fields {
		if len(conditions) == 1 {
			cond := conditions[0]
			if cond.Op == OpEq {
				// 单个 eq 条件直接赋值
				result[field] = cond.Value
			} else if cond.Op == OpLike {
				// Like 条件使用 $regex
				result[field] = bson.M{"$regex": cond.Value, "$options": "i"}
			} else {
				// 单个非 eq 条件
				result[field] = bson.M{mongoOpMap[cond.Op]: cond.Value}
			}
		} else {
			// 多个条件合并到同一字段
			fieldConditions := bson.M{}
			for _, cond := range conditions {
				if cond.Op == OpLike {
					fieldConditions["$regex"] = cond.Value
					fieldConditions["$options"] = "i"
				} else if cond.Op == OpEq {
					fieldConditions[mongoOpMap[OpEq]] = cond.Value
				} else {
					fieldConditions[mongoOpMap[cond.Op]] = cond.Value
				}
			}
			result[field] = fieldConditions
		}
	}

	// 处理逻辑组
	for _, group := range b.conditions.LogicalGroups {
		groupConditions := make([]any, 0, len(group.Conditions))
		for _, cond := range group.Conditions {
			groupConditions = append(groupConditions, b.convertCondition(cond))
		}

		opKey := "$and"
		if group.Type == "or" {
			opKey = "$or"
		}

		if existing, ok := result[opKey]; ok {
			// 合并已存在的逻辑组
			if existingSlice, ok := existing.([]any); ok {
				result[opKey] = append(existingSlice, groupConditions...)
			}
		} else {
			result[opKey] = groupConditions
		}
	}

	return result
}

// convertCondition 转换条件为 bson.M
func (b *MongoQueryBuilder) convertCondition(cond any) any {
	switch v := cond.(type) {
	case bson.M:
		return v
	case bson.D:
		return v
	case map[string]any:
		return bson.M(v)
	default:
		// 支持传入其他 QBuilder 的 Build 结果
		if builder, ok := cond.(IBuilder); ok {
			return builder.Build()
		}
		return cond
	}
}
