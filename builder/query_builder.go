package builder

// IBuilder 通用构建器接口
type IBuilder interface {
	Build() any
}

// QBuilder 查询构建器接口
type QBuilder interface {
	IBuilder
	Id(id any) QBuilder
	Eq(key string, value any) QBuilder
	Ne(key string, value any) QBuilder
	Gt(key string, value any) QBuilder
	Gte(key string, value any) QBuilder
	Lt(key string, value any) QBuilder
	Lte(key string, value any) QBuilder
	In(key string, value ...any) QBuilder
	Nin(key string, value ...any) QBuilder
	Like(key string, value string, mode MatchMode) QBuilder
	And(conditions ...any) QBuilder
	Or(conditions ...any) QBuilder
}

// BuilderType 构建器类型
type BuilderType int

const (
	BuilderTypeMongo BuilderType = iota
	BuilderTypeGorm
)

var (
	defaultBuilderType = BuilderTypeMongo
)

// SetQueryBuilder 设置默认查询构建器类型
func SetQueryBuilder(builderType BuilderType) {
	defaultBuilderType = builderType
}

// NewQueryBuilder 创建查询构建器
func NewQueryBuilder() QBuilder {
	switch defaultBuilderType {
	case BuilderTypeGorm:
		return NewGormQueryBuilder()
	default:
		return NewMongoQueryBuilder()
	}
}

// Eq 等于条件
func Eq(key string, value any) any {
	return NewQueryBuilder().Eq(key, value).Build()
}

// Ne 不等于条件
func Ne(key string, value any) any {
	return NewQueryBuilder().Ne(key, value).Build()
}

// Gt 大于条件
func Gt(key string, value any) any {
	return NewQueryBuilder().Gt(key, value).Build()
}

// Gte 大于等于条件
func Gte(key string, value any) any {
	return NewQueryBuilder().Gte(key, value).Build()
}

// Lt 小于条件
func Lt(key string, value any) any {
	return NewQueryBuilder().Lt(key, value).Build()
}

// Lte 小于等于条件
func Lte(key string, value any) any {
	return NewQueryBuilder().Lte(key, value).Build()
}

// In 包含条件
func In[T any](key string, value ...T) any {
	return NewQueryBuilder().In(key, ToAnySlice(value)...).Build()
}

// Nin 不包含条件
func Nin[T any](key string, value ...T) any {
	return NewQueryBuilder().Nin(key, ToAnySlice(value)...).Build()
}

// Like 模糊匹配条件
func Like(key string, value string, mode MatchMode) any {
	return NewQueryBuilder().Like(key, value, mode).Build()
}

// Id ID 条件
func Id(id any) any {
	return NewQueryBuilder().Id(id).Build()
}

// And 逻辑与
func And(conditions ...any) any {
	return NewQueryBuilder().And(conditions...).Build()
}

// Or 逻辑或
func Or(conditions ...any) any {
	return NewQueryBuilder().Or(conditions...).Build()
}
