package builder

// Op 操作符类型
type Op string

const (
	OpEq   Op = "eq"
	OpNe   Op = "ne"
	OpGt   Op = "gt"
	OpGte  Op = "gte"
	OpLt   Op = "lt"
	OpLte  Op = "lte"
	OpIn   Op = "in"
	OpNin  Op = "nin"
	OpLike Op = "like"
)

// MatchMode 模糊匹配模式
type MatchMode int

const (
	MatchContains   MatchMode = iota // 包含 %value%
	MatchStartsWith                  // 前缀 value%
	MatchEndsWith                    // 后缀 %value
)

// Condition 单个条件
type Condition struct {
	Op    Op
	Value any
}

// LogicalGroup 逻辑组（And/Or）
type LogicalGroup struct {
	Type       string // "and" or "or"
	Conditions []any
}

// QueryConditions 查询条件集合，作为中间结构
type QueryConditions struct {
	// Fields 存储字段条件，key 为字段名，value 为该字段的所有条件
	Fields map[string][]Condition
	// LogicalGroups 存储逻辑组
	LogicalGroups []LogicalGroup
}

// NewQueryConditions 创建新的查询条件集合
func NewQueryConditions() *QueryConditions {
	return &QueryConditions{
		Fields:        make(map[string][]Condition),
		LogicalGroups: make([]LogicalGroup, 0),
	}
}

// AddCondition 添加字段条件
func (qc *QueryConditions) AddCondition(key string, op Op, value any) {
	qc.Fields[key] = append(qc.Fields[key], Condition{Op: op, Value: value})
}

// AddLogicalGroup 添加逻辑组
func (qc *QueryConditions) AddLogicalGroup(groupType string, conditions []any) {
	qc.LogicalGroups = append(qc.LogicalGroups, LogicalGroup{
		Type:       groupType,
		Conditions: conditions,
	})
}
