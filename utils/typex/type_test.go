package typex

import "testing"

func TestToAnyERejectsOverflow(t *testing.T) {
	if _, err := ToAnyE[uint8]("300"); err == nil {
		t.Fatal("uint8 300 should overflow")
	}
	if _, err := ToAnyE[uint64]("-1"); err == nil {
		t.Fatal("negative uint64 should fail")
	}
	if v, err := ToAnyE[uint64]("18446744073709551615"); err != nil || v != 1<<64-1 {
		t.Fatalf("max uint64 = %d, %v", v, err)
	}
	if v, err := ToAnyE[int64]("-9223372036854775808"); err != nil || v != -1<<63 {
		t.Fatalf("min int64 = %d, %v", v, err)
	}
}

func TestToAnyEDecodesJson(t *testing.T) {
	type item struct {
		Id   int64  `json:"id"`
		Name string `json:"name"`
	}
	got, err := ToAnyE[item](`{"id":1234567890123456789,"name":"a"}`)
	if err != nil || got != (item{Id: 1234567890123456789, Name: "a"}) {
		t.Fatalf("got %+v, %v", got, err)
	}
	if v, err := ToAnyE[bool]("true"); err != nil || !v {
		t.Fatalf("bool = %v, %v", v, err)
	}
	if v, err := ToAnyE[int](""); err != nil || v != 0 {
		t.Fatalf("empty = %v, %v", v, err)
	}
}

func TestToStringFloat32(t *testing.T) {
	if got := ToString(float32(0.1)); got != "0.1" {
		t.Fatalf("float32 0.1 = %s", got)
	}
	if got := ToString(int64(-42)); got != "-42" {
		t.Fatalf("int64 = %s", got)
	}
}
