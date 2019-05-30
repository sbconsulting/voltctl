package filter

import (
	"testing"
)

type TestFilterStruct struct {
	One   string
	Two   string
	Three string
}

func TestFilterList(t *testing.T) {
	f, err := Parse("One=a,Two=b")
	if err != nil {
		t.Errorf("Unable to parse filter: %s", err.Error())
	}

	data := []interface{}{
		TestFilterStruct{
			One:   "a",
			Two:   "b",
			Three: "c",
		},
		TestFilterStruct{
			One:   "1",
			Two:   "2",
			Three: "3",
		},
		TestFilterStruct{
			One:   "a",
			Two:   "b",
			Three: "z",
		},
	}

	r, _ := f.Process(data)

	if _, ok := r.([]interface{}); !ok {
		t.Errorf("Expected list, but didn't get one")
	}

	if len(r.([]interface{})) != 2 {
		t.Errorf("Expected %d got %d", 2, len(r.([]interface{})))
	}

	if r.([]interface{})[0] != data[0] {
		t.Errorf("Filtered list did not match, item %d", 0)
	}
	if r.([]interface{})[1] != data[2] {
		t.Errorf("Filtered list did not match, item %d", 1)
	}
}

func TestFilterItem(t *testing.T) {
	f, err := Parse("One=a,Two=b")
	if err != nil {
		t.Errorf("Unable to parse filter: %s", err.Error())
	}

	data := TestFilterStruct{
		One:   "a",
		Two:   "b",
		Three: "c",
	}

	r, _ := f.Process(data)

	if r == nil {
		t.Errorf("Expected item, got nil")
	}

	if _, ok := r.([]interface{}); ok {
		t.Errorf("Expected item, but got list")
	}
}

func TestGoodFilters(t *testing.T) {
	var f Filter
	var err error
	f, err = Parse("One=a,Two=b")
	if err != nil {
		t.Errorf("1. Unable to parse filter: %s", err.Error())
	}
	if len(f) != 2 ||
		f["One"].Value != "a" ||
		f["One"].Op != EQ ||
		f["Two"].Value != "b" ||
		f["Two"].Op != EQ {
		t.Errorf("1. Filter did not parse correctly")
	}

	f, err = Parse("One=a")
	if err != nil {
		t.Errorf("2. Unable to parse filter: %s", err.Error())
	}
	if len(f) != 1 ||
		f["One"].Value != "a" ||
		f["One"].Op != EQ {
		t.Errorf("2. Filter did not parse correctly")
	}

	f, err = Parse("One<a")
	if err != nil {
		t.Errorf("3. Unable to parse filter: %s", err.Error())
	}
	if len(f) != 1 ||
		f["One"].Value != "a" ||
		f["One"].Op != LT {
		t.Errorf("3. Filter did not parse correctly")
	}

	f, err = Parse("One!=a")
	if err != nil {
		t.Errorf("4. Unable to parse filter: %s", err.Error())
	}
	if len(f) != 1 ||
		f["One"].Value != "a" ||
		f["One"].Op != NE {
		t.Errorf("4. Filter did not parse correctly")
	}
}

func TestBadFilters(t *testing.T) {
	_, err := Parse("One%a")
	if err == nil {
		t.Errorf("Parsed filter when it shouldn't have")
	}
}
