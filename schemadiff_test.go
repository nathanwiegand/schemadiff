package schemadiff_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/nathanwiegand/schemadiff"
)

type A struct {
	String           string
	B                B
	BPtr             *B
	UnexportedString string `json:"-"`
	StringPtr        *string
	StringPtrPtr     ****string
	Map              map[string]string
	InvalidMap       map[bool]string
}

type B struct {
	String       string
	DSlice       []D      `json:"d"`
	DPtrSlice    []*D     `json:"dptr"`
	DPtrPtrSlice []*****D `json:"dptrptr"`
}

type D struct {
	String          string
	Int             int
	FloatSlice      []float32
	FloatSliceSlice [][]float32
	Time            time.Time
	TimePtr         *time.Time
}

func TestGetUnmappedJSONFields(t *testing.T) {
	var a A
	data := `
	{
		"String": "name",
		"UnexportedString": "should be unknown",
		"StringPtr": "this is a ptr",
		"StringPtrPtr": "this is a ptr",
		"MissingKey": "value",
		"Map": {"foo":"bar", "baz":"quux", "notstring": 3.4},
		"MapNotWorking": {"foo":"bar", "baz":"quux"},
		"B": {
			"String": "b's name",
			"MissingKey": "value",
			"d": [
				{
					"String": "d_name",
					"Int": "d_value"
				},
				{
					"String": "d_name"
				},
				{
					"UnknownString": "extra"
				},
				{
					"FloatSlice": [3.1, 1.4, "string"]
				},
				{
					"FloatSliceSlice": [[3.1, 1.4],[1.5]]
				},
				{
					"Time":"2020-02-03"
				},
				{
					"TimePtr":"2020-02-03"
				}
			]
		},
		"BPtr": {
			"String": "b ptr name",
			"dptr": [{"String": "name"}],
			"dptrptr": [{"String": "name"}],
			"UnknownSlice": ["unknown"]
		}
	}
	`

	want := map[string]string{
		"B.MissingKey":         `"value"`,
		"B.d[0].Int":           `"d_value"`,                                    // The value is a string, but we tried to write it to an int.
		"B.d[2].UnknownString": `"extra"`,                                      // Unknown field.
		"B.d[3].FloatSlice[2]": `"string"`,                                     // Slice value has wrong type.
		"BPtr.UnknownSlice":    "[\n \"unknown\"\n]",                           // Unknown field pointed.
		"UnexportedString":     `"should be unknown"`,                          // This field was explicitly ignored from json.
		"MapNotWorking":        "{\n \"baz\": \"quux\",\n \"foo\": \"bar\"\n}", // The key type of the map was wrong.
		`Map["notstring"]`:     "3.4",                                          // The value type of the json map was wrong.
		"MissingKey":           `"value"`,
	}

	fields, err := schemadiff.UnmappedJSONFields(reflect.TypeOf(a), []byte(data))
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(fields, want); diff != "" {
		t.Error(diff)
	}
}
