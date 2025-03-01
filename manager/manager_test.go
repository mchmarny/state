package manager

import (
	"reflect"
	"testing"
)

func TestMarshalUnmarshal(t *testing.T) {
	type TestStruct struct {
		Name  string
		Age   int
		Score float64
		Flag  bool
	}

	original := TestStruct{Name: "Alice", Age: 30, Score: 95.5, Flag: true}

	// Test Marshal
	data, err := binaryMarshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Test Unmarshal
	var decoded TestStruct
	if err := binaryUnmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Validate results
	if !reflect.DeepEqual(original, decoded) {
		t.Errorf("Expected %+v, got %+v", original, decoded)
	}
}

func TestComplexMarshalUnmarshal(t *testing.T) {
	type TestStruct struct {
		Map     map[string]int
		Content interface{}
	}

	original := TestStruct{
		Map: map[string]int{
			"one":   1,
			"two":   2,
			"three": 3,
		},
		Content: TestStruct{
			Map: map[string]int{
				"four": 4,
				"five": 5,
				"six":  6,
			},
		},
	}

	// Register custom types
	RegisterTypes(TestStruct{})

	// Test Marshal
	data, err := binaryMarshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Test Unmarshal
	var decoded TestStruct
	if err := binaryUnmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Validate results
	if !reflect.DeepEqual(original, decoded) {
		t.Errorf("Expected %+v, got %+v", original, decoded)
	}
}

func TestMarshalUnmarshalUsingStateAnnotation(t *testing.T) {
	type TestStruct struct {
		Name  string  `state:"name"`
		Age   int     `state:"age"`
		Score float64 `state:"score"`
		Flag  bool    `state:"flag"`
		Other string
	}

	original := TestStruct{
		Name:  "Alice",
		Age:   30,
		Score: 95.5,
		Flag:  true,
		Other: "This will not be saved",
	}

	// Test Marshal
	data, err := stateMarshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Test Unmarshal
	var decoded TestStruct
	if err := stateUnmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Validate results
	if original.Name != decoded.Name {
		t.Errorf("Expected %+v, got %+v", original.Name, decoded.Name)
	}

	if original.Age != decoded.Age {
		t.Errorf("Expected %+v, got %+v", original.Age, decoded.Age)
	}

	if original.Score != decoded.Score {
		t.Errorf("Expected %+v, got %+v", original.Score, decoded.Score)
	}

	if original.Flag != decoded.Flag {
		t.Errorf("Expected %+v, got %+v", original.Flag, decoded.Flag)
	}

	if decoded.Other != "" {
		t.Errorf("Expected '', got %+v", decoded.Other)
	}
}
