package repox

import (
	"errors"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"gorm.io/gorm"
)

func wrapError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, mongo.ErrNoDocuments) {
		return DataNotFound
	}
	return err
}

func FromPtrSlice[T any](data []*T) []T {
	v := make([]T, len(data))
	for i := range data {
		v[i] = *data[i]
	}
	return v
}

func ToPtrSlice[T any](data []T) []*T {
	v := make([]*T, len(data))
	for i := range data {
		v[i] = &data[i]
	}
	return v
}

func ToAnySlice[T any](data []T) []any {
	v := make([]any, len(data))
	for i := range data {
		v[i] = data[i]
	}
	return v
}
