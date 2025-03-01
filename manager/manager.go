package manager

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// SerializationType defines the available serialization formats
type SerializationType string

const (
	// Serialization types
	JSON  SerializationType = "json"
	YAML  SerializationType = "yaml"
	BIN   SerializationType = "bin"
	STATE SerializationType = "state"

	// StateAnnotationKey is the key used to define custom field names
	StateAnnotationKey = "state"

	// Default values
	SerializationTypeDefault = BIN
	DefaultStateFileName     = ".state"
)

// StateManager handles persisting state to a file.
type StateManager struct {
	FilePath          string
	SerializationType SerializationType

	mutex sync.Mutex
}

// StateOption defines a functional option for configuring StateManager
type StateOption func(*StateManager)

// WithSerializationType sets the serialization type for the State
func WithSerializationType(serializationType SerializationType) StateOption {
	return func(s *StateManager) {
		s.SerializationType = serializationType
	}
}

// WithFilePath sets a custom file path for the State
func WithFilePath(filePath string) StateOption {
	return func(s *StateManager) {
		s.FilePath = filePath
	}
}

// NewStateManager initializes a new State with functional options.
func NewStateManager(options ...StateOption) (*StateManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	s := &StateManager{
		FilePath:          filepath.Join(homeDir, DefaultStateFileName),
		SerializationType: SerializationTypeDefault,
	}

	for _, option := range options {
		option(s)
	}

	return s, nil
}

// Save persists the given struct to the file.
func (s *StateManager) Save(data interface{}) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	file, err := os.Create(s.FilePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	var b []byte

	switch s.SerializationType {
	case BIN:
		b, err = binaryMarshal(data)
	case JSON:
		b, err = json.MarshalIndent(data, "", "  ")
	case YAML:
		b, err = yaml.Marshal(data)
	case STATE:
		b, err = stateMarshal(data)
	default:
		err = fmt.Errorf("unsupported serialization format")
	}

	if err != nil {
		return fmt.Errorf("failed to encode data: %w", err)
	}

	// Ensure something is written to file
	if len(b) == 0 {
		return fmt.Errorf("no data was encoded")
	}

	// Write to a temporary file first
	tempFile := s.FilePath + ".tmp"
	if err := os.WriteFile(tempFile, b, 0600); err != nil {
		return fmt.Errorf("failed to write to temp file: %w", err)
	}

	// Atomically move temp file to actual file
	if err := os.Rename(tempFile, s.FilePath); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// Load reads the struct from the file.
func (s *StateManager) Load(data interface{}) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	c, err := os.ReadFile(s.FilePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	switch s.SerializationType {
	case BIN:
		err = binaryUnmarshal(c, data)
	case JSON:
		err = json.Unmarshal(c, data)
	case YAML:
		err = yaml.Unmarshal(c, data)
	case STATE:
		err = stateUnmarshal(c, data)
	default:
		err = fmt.Errorf("unsupported serialization format")
	}

	if err != nil {
		return fmt.Errorf("failed to decode data: %w", err)
	}

	return nil
}

// Exists checks if the file exists.
func (s *StateManager) Exists() bool {
	if _, err := os.Stat(s.FilePath); os.IsNotExist(err) {
		return false
	}
	return true
}

// stateMarshal handles struct serialization using field tags
func stateMarshal(data interface{}) ([]byte, error) {
	values := make(map[string]interface{})
	t := reflect.TypeOf(data)
	v := reflect.ValueOf(data)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		key := field.Tag.Get(StateAnnotationKey)

		// Only include fields that have the state tag
		if key == "" {
			continue
		}

		values[key] = v.Field(i).Interface() // Preserve original types
	}

	return yaml.Marshal(values)
}

func stateUnmarshal(data []byte, v interface{}) error {
	if reflect.TypeOf(v).Kind() != reflect.Ptr {
		return fmt.Errorf("unmarshal target must be a pointer to a struct")
	}

	values := make(map[string]interface{})
	if err := yaml.Unmarshal(data, &values); err != nil {
		return fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	vt := reflect.TypeOf(v).Elem()
	vv := reflect.ValueOf(v).Elem()

	for i := 0; i < vt.NumField(); i++ {
		field := vt.Field(i)
		key := field.Tag.Get(StateAnnotationKey)
		if key == "" {
			key = strings.ToLower(field.Name)
		}

		value, ok := values[key]
		if !ok {
			continue
		}

		fieldValue := vv.Field(i)
		if !fieldValue.CanSet() {
			continue
		}

		// Handle pointer fields
		if fieldValue.Kind() == reflect.Ptr {
			elemType := fieldValue.Type().Elem()
			newVal := reflect.New(elemType)
			if err := setReflectValue(newVal.Elem(), value); err == nil {
				fieldValue.Set(newVal)
			}
		} else {
			_ = setReflectValue(fieldValue, value)
		}
	}

	return nil
}

func setReflectValue(field reflect.Value, value interface{}) error {
	switch field.Kind() {
	case reflect.String:
		if str, ok := value.(string); ok {
			field.SetString(str)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if num, err := strconv.ParseInt(fmt.Sprintf("%v", value), 10, 64); err == nil {
			field.SetInt(num)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if num, err := strconv.ParseUint(fmt.Sprintf("%v", value), 10, 64); err == nil {
			field.SetUint(num)
		}
	case reflect.Float32, reflect.Float64:
		if num, err := strconv.ParseFloat(fmt.Sprintf("%v", value), 64); err == nil {
			field.SetFloat(num)
		}
	case reflect.Bool:
		if boolean, err := strconv.ParseBool(fmt.Sprintf("%v", value)); err == nil {
			field.SetBool(boolean)
		}
	default:
		return fmt.Errorf("unsupported field type: %s", field.Kind())
	}
	return nil
}

// binaryMarshal handles struct serialization using binary encoding
func binaryMarshal(data interface{}) ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(data); err != nil {
		return nil, fmt.Errorf("failed to encode data: %w", err)
	}
	return buf.Bytes(), nil
}

// binaryUnmarshal handles struct deserialization using binary encoding
func binaryUnmarshal(data []byte, v interface{}) error {
	if reflect.TypeOf(v).Kind() != reflect.Ptr {
		return errors.New("unmarshal target must be a pointer to a struct")
	}

	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)
	if err := decoder.Decode(v); err != nil {
		return fmt.Errorf("failed to decode binary data: %w", err)
	}
	return nil
}

// RegisterTypes pre-registers types for gob encoding.
// Required for Interfaces & Custom Types
func RegisterTypes(types ...interface{}) {
	for _, t := range types {
		gob.Register(t)
	}
}
