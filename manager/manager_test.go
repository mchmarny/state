package manager

import (
	"path/filepath"
	"reflect"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestStruct is a sample struct used for testing.
type TestStruct struct {
	Name        string  `json:"name" yaml:"name" state:"name"`
	Age         int     `json:"age" yaml:"age" state:"age"`
	Temperature float64 `json:"temp" yaml:"temp" state:"temp"`
	Flag        bool    `json:"flag" yaml:"flag" state:"flag"`
}

func setupTempStateManager(t *testing.T, serializationType SerializationType) *StateManager {
	t.Helper()

	tempDir := t.TempDir()
	stateFile := filepath.Join(tempDir, "test_state")

	sm, err := NewStateManager(
		WithFilePath(stateFile),
		WithSerializationType(serializationType),
	)
	assert.NoError(t, err)

	return sm
}

// TestSaveAndLoadJSON ensures JSON serialization and deserialization works correctly.
func TestSaveAndLoadJSON(t *testing.T) {
	sm := setupTempStateManager(t, JSON)
	data := &TestStruct{"Alice", 30, 98.6, true}

	// Save data
	err := sm.Save(data)
	assert.NoError(t, err)
	assert.FileExists(t, sm.FilePath)

	// Load data
	loadedData := &TestStruct{}
	err = sm.Load(loadedData)
	assert.NoError(t, err)
	assert.Equal(t, data, loadedData)
}

// TestSaveAndLoadYAML ensures YAML serialization and deserialization works correctly.
func TestSaveAndLoadYAML(t *testing.T) {
	sm := setupTempStateManager(t, YAML)
	data := &TestStruct{"Bob", 40, 36.5, false}

	err := sm.Save(data)
	assert.NoError(t, err)

	loadedData := &TestStruct{}
	err = sm.Load(loadedData)
	assert.NoError(t, err)
	assert.Equal(t, data, loadedData)
}

// TestSaveAndLoadBinary ensures binary (gob) serialization works correctly.
func TestSaveAndLoadBinary(t *testing.T) {
	sm := setupTempStateManager(t, BIN)
	data := &TestStruct{"Charlie", 25, 99.1, true}

	err := sm.Save(data)
	assert.NoError(t, err)

	loadedData := &TestStruct{}
	err = sm.Load(loadedData)
	assert.NoError(t, err)
	assert.Equal(t, data, loadedData)
}

// TestSaveAndLoadStateFormat tests custom STATE serialization format.
func TestSaveAndLoadStateFormat(t *testing.T) {
	sm := setupTempStateManager(t, STATE)
	data := &TestStruct{"David", 35, 97.7, false}

	err := sm.Save(data)
	assert.NoError(t, err)

	loadedData := &TestStruct{}
	err = sm.Load(loadedData)
	assert.NoError(t, err)
	assert.Equal(t, data, loadedData)
}

// TestStateMarshal ensures correct encoding for struct with `state` tags.
func TestStateMarshal(t *testing.T) {
	data := &TestStruct{"Ivan", 50, 96.4, true}
	encoded, err := stateMarshal(data)
	assert.NoError(t, err)
	assert.Contains(t, string(encoded), "name: Ivan")
	assert.Contains(t, string(encoded), "age: 50")
	assert.Contains(t, string(encoded), "temp: 96.4")
	assert.Contains(t, string(encoded), "flag: true")
}

// TestStateUnmarshal ensures correct decoding for struct with `state` tags.
func TestStateUnmarshal(t *testing.T) {
	yamlData := `
name: Jake
age: 31
temp: 36.8
flag: false
`
	var data TestStruct
	err := stateUnmarshal([]byte(yamlData), &data)
	assert.NoError(t, err)
	assert.Equal(t, "Jake", data.Name)
	assert.Equal(t, 31, data.Age)
	assert.Equal(t, 36.8, data.Temperature)
	assert.Equal(t, false, data.Flag)
}

// TestBinaryMarshal ensures gob serialization works correctly.
func TestBinaryMarshal(t *testing.T) {
	data := &TestStruct{"Kelly", 27, 98.2, true}
	encoded, err := binaryMarshal(data)
	assert.NoError(t, err)
	assert.NotEmpty(t, encoded)
}

// TestBinaryUnmarshal ensures gob deserialization works correctly.
func TestBinaryUnmarshal(t *testing.T) {
	data := &TestStruct{"Liam", 32, 97.9, false}
	encoded, err := binaryMarshal(data)
	assert.NoError(t, err)

	var decodedData TestStruct
	err = binaryUnmarshal(encoded, &decodedData)
	assert.NoError(t, err)
	assert.Equal(t, data, &decodedData)
}

// TestConcurrentAccess ensures that concurrent Save and Load operations do not cause race conditions.
func TestConcurrentAccess(t *testing.T) {
	sm := setupTempStateManager(t, JSON)
	data := &TestStruct{"Helen", 45, 99.5, true}

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent Save operations
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = sm.Save(data)
		}()
	}

	wg.Wait()

	// Load the data after concurrent writes
	loadedData := &TestStruct{}
	err := sm.Load(loadedData)
	assert.NoError(t, err)
	assert.Equal(t, data, loadedData)
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
