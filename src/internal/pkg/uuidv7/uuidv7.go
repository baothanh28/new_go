package uuidv7

import (
	"github.com/google/uuid"
)

// Generator provides a wrapper around google/uuid's UUIDv7 generation
// The underlying google/uuid library handles thread-safety and monotonic ordering internally
type Generator struct{}

// NewGenerator creates a new UUIDv7 generator instance
// This is a lightweight wrapper around google/uuid's UUIDv7 implementation
func NewGenerator() *Generator {
	return &Generator{}
}

// Generate creates a new UUIDv7 using google/uuid's implementation
// UUIDv7 includes a Unix timestamp in milliseconds (48 bits) and ensures
// monotonic ordering even when multiple UUIDs are generated in the same millisecond.
// The google/uuid library handles thread-safety and counter management internally.
func (g *Generator) Generate() (uuid.UUID, error) {
	return uuid.NewV7()
}

// GenerateString generates a UUIDv7 and returns it as a string
func (g *Generator) GenerateString() (string, error) {
	id, err := g.Generate()
	if err != nil {
		return "", err
	}
	return id.String(), nil
}

// MustGenerate generates a UUIDv7 and panics if an error occurs
// Use this only when you're certain generation cannot fail
func (g *Generator) MustGenerate() uuid.UUID {
	id, err := g.Generate()
	if err != nil {
		panic("uuidv7: failed to generate UUID: " + err.Error())
	}
	return id
}

// MustGenerateString generates a UUIDv7 string and panics if an error occurs
func (g *Generator) MustGenerateString() string {
	return g.MustGenerate().String()
}

// GenerateBatch generates multiple UUIDv7s in a single call
// Useful for bulk operations where you need multiple unique IDs
func (g *Generator) GenerateBatch(count int) ([]uuid.UUID, error) {
	if count <= 0 {
		return nil, nil
	}

	ids := make([]uuid.UUID, 0, count)
	for i := 0; i < count; i++ {
		id, err := g.Generate()
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, nil
}

// GenerateBatchStrings generates multiple UUIDv7 strings in a single call
func (g *Generator) GenerateBatchStrings(count int) ([]string, error) {
	ids, err := g.GenerateBatch(count)
	if err != nil {
		return nil, err
	}

	strings := make([]string, len(ids))
	for i, id := range ids {
		strings[i] = id.String()
	}

	return strings, nil
}

// Parse parses a UUID string and validates it
func Parse(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

// MustParse parses a UUID string and panics if parsing fails
func MustParse(s string) uuid.UUID {
	return uuid.MustParse(s)
}

// IsValid checks if a string is a valid UUID
func IsValid(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}
