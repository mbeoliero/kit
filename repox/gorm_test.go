package repox

import (
	"strings"
	"sync"
	"testing"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type user struct {
	Id    int64
	Name  string
	Token string
}

// BeforeCreate stands in for anything the database fills in, such as generated keys.
func (u *user) BeforeCreate(*gorm.DB) error {
	u.Token = "generated-" + u.Name
	return nil
}

func dryRunDB(t *testing.T) (*gorm.DB, func() []string) {
	t.Helper()
	db, err := gorm.Open(mysql.New(mysql.Config{DSN: "u:p@tcp(127.0.0.1:1)/db", SkipInitializeWithVersion: true}),
		&gorm.Config{DryRun: true, DisableAutomaticPing: true, SkipDefaultTransaction: true})
	if err != nil {
		t.Fatal(err)
	}
	var mu sync.Mutex
	var statements []string
	record := func(tx *gorm.DB) {
		mu.Lock()
		statements = append(statements, tx.Statement.SQL.String())
		mu.Unlock()
	}
	_ = db.Callback().Update().After("gorm:update").Register("test:record", record)
	_ = db.Callback().Create().After("gorm:create").Register("test:record", record)
	return db, func() []string {
		mu.Lock()
		defer mu.Unlock()
		return append([]string(nil), statements...)
	}
}

func TestGormUpdateOneLimitsToOneRow(t *testing.T) {
	db, statements := dryRunDB(t)
	repo := NewGormRepo[user](db)
	if _, err := repo.UpdateOne(t.Context(), map[string]any{"name": "a"}, map[string]any{"token": "x"}); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.UpdateMany(t.Context(), map[string]any{"name": "a"}, map[string]any{"token": "x"}); err != nil {
		t.Fatal(err)
	}
	got := statements()
	if len(got) != 2 || !strings.HasSuffix(got[0], "LIMIT ?") || strings.Contains(got[1], "LIMIT") {
		t.Fatalf("statements = %q", got)
	}
}

func TestGormCreateManyWritesBack(t *testing.T) {
	db, _ := dryRunDB(t)
	repo := NewGormRepo[user](db)
	users := []*user{{Name: "a"}, {Name: "b"}}
	if err := repo.CreateMany(t.Context(), users); err != nil {
		t.Fatal(err)
	}
	if users[0].Token != "generated-a" || users[1].Token != "generated-b" {
		t.Fatalf("users = %+v %+v", *users[0], *users[1])
	}
}
