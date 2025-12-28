package repox

import (
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type SortOrder int

const (
	Asc SortOrder = iota + 1
	Desc
)

type SortField struct {
	Field string
	Order SortOrder
}

type Sort struct {
	fields []SortField
}

func NewSort() *Sort {
	return &Sort{}
}

func (s *Sort) Asc(field string) *Sort {
	s.fields = append(s.fields, SortField{
		Field: field,
		Order: Asc,
	})
	return s
}

func (s *Sort) Desc(field string) *Sort {
	s.fields = append(s.fields, SortField{
		Field: field,
		Order: Desc,
	})
	return s
}

func (s *Sort) ToBson() bson.D {
	d := bson.D{}
	for _, f := range s.fields {
		order := 1
		if f.Order == Desc {
			order = -1
		}
		d = append(d, bson.E{
			Key:   f.Field,
			Value: order,
		})
	}
	return d
}

func (s *Sort) ToSqlStr() string {
	parts := make([]string, 0, len(s.fields))
	for _, f := range s.fields {
		dir := "ASC"
		if f.Order == Desc {
			dir = "DESC"
		}
		parts = append(parts, fmt.Sprintf("%s %s", f.Field, dir))
	}
	return strings.Join(parts, ", ")
}
