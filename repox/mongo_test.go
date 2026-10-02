package repox

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type profile struct {
	Id     bson.ObjectID `bson:"_id,omitempty"`
	UserId int64         `bson:"user_id"`
	Name   string        `bson:"name"`
	Views  int64         `bson:"views"`
}

func TestUpsertUpdateKeepsPathsDisjoint(t *testing.T) {
	id := bson.NewObjectID()
	got, err := upsertUpdate(profile{Id: id, UserId: 1, Name: "a", Views: 9}, UpsertOptions{
		ConflictKvs: map[string]any{"user_id": 1},
		Inc:         map[string]int64{"views": 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := bson.M{
		"$set":         bson.M{"user_id": int64(1), "name": "a"},
		"$inc":         map[string]int64{"views": 1},
		"$setOnInsert": bson.M{"_id": id},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("update = %v, want %v", got, want)
	}
}

func TestUpsertUpdateSkipsZeroId(t *testing.T) {
	type withId struct {
		Id   bson.ObjectID `bson:"_id"`
		Name string        `bson:"name"`
	}
	got, err := upsertUpdate(withId{Name: "a"}, UpsertOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if want := (bson.M{"$set": bson.M{"name": "a"}}); !reflect.DeepEqual(got, want) {
		t.Fatalf("update = %v, want %v", got, want)
	}
}

func TestUpsertUpdateUsesExplicitSet(t *testing.T) {
	got, err := upsertUpdate(profile{Name: "ignored"}, UpsertOptions{Set: map[string]any{"name": "b"}})
	if err != nil {
		t.Fatal(err)
	}
	if want := (bson.M{"$set": map[string]any{"name": "b"}}); !reflect.DeepEqual(got, want) {
		t.Fatalf("update = %v, want %v", got, want)
	}
}
