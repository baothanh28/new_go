package uuidv7

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGenerator(t *testing.T) {
	gen := NewGenerator()
	assert.NotNil(t, gen)
}

func TestGenerate(t *testing.T) {
	gen := NewGenerator()

	id, err := gen.Generate()
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, id)
	assert.Equal(t, uuid.Version(7), id.Version())
}

func TestGenerateString(t *testing.T) {
	gen := NewGenerator()

	idStr, err := gen.GenerateString()
	require.NoError(t, err)
	assert.NotEmpty(t, idStr)
	assert.Len(t, idStr, 36) // Standard UUID string length

	// Verify it's a valid UUID
	parsed, err := uuid.Parse(idStr)
	require.NoError(t, err)
	assert.Equal(t, uuid.Version(7), parsed.Version())
}

func TestGenerateUniqueness(t *testing.T) {
	gen := NewGenerator()

	// Generate multiple UUIDs rapidly
	ids := make(map[uuid.UUID]bool)
	for i := 0; i < 100; i++ {
		id, err := gen.Generate()
		require.NoError(t, err)
		assert.False(t, ids[id], "Duplicate UUID generated")
		ids[id] = true
	}
}

func TestGenerateMonotonicOrdering(t *testing.T) {
	gen := NewGenerator()

	// Generate multiple UUIDs and verify they're in ascending order
	var prevID uuid.UUID
	for i := 0; i < 50; i++ {
		id, err := gen.Generate()
		require.NoError(t, err)

		if i > 0 {
			// UUIDv7 should be lexicographically sortable
			assert.True(t, id.String() > prevID.String() || id.String() == prevID.String(),
				"UUIDs should be in ascending order or equal")
		}
		prevID = id
	}
}

func TestMustGenerate(t *testing.T) {
	gen := NewGenerator()

	id := gen.MustGenerate()
	assert.NotEqual(t, uuid.Nil, id)
	assert.Equal(t, uuid.Version(7), id.Version())
}

func TestMustGenerateString(t *testing.T) {
	gen := NewGenerator()

	idStr := gen.MustGenerateString()
	assert.NotEmpty(t, idStr)
	assert.Len(t, idStr, 36)

	// Verify it's a valid UUID
	parsed, err := uuid.Parse(idStr)
	require.NoError(t, err)
	assert.Equal(t, uuid.Version(7), parsed.Version())
}

func TestGenerateBatch(t *testing.T) {
	gen := NewGenerator()

	ids, err := gen.GenerateBatch(10)
	require.NoError(t, err)
	assert.Len(t, ids, 10)

	// Verify all are unique
	unique := make(map[uuid.UUID]bool)
	for _, id := range ids {
		assert.NotEqual(t, uuid.Nil, id)
		assert.Equal(t, uuid.Version(7), id.Version())
		assert.False(t, unique[id], "Duplicate UUID in batch")
		unique[id] = true
	}
}

func TestGenerateBatchZero(t *testing.T) {
	gen := NewGenerator()

	ids, err := gen.GenerateBatch(0)
	require.NoError(t, err)
	assert.Nil(t, ids)
}

func TestGenerateBatchNegative(t *testing.T) {
	gen := NewGenerator()

	ids, err := gen.GenerateBatch(-1)
	require.NoError(t, err)
	assert.Nil(t, ids)
}

func TestGenerateBatchStrings(t *testing.T) {
	gen := NewGenerator()

	idStrs, err := gen.GenerateBatchStrings(5)
	require.NoError(t, err)
	assert.Len(t, idStrs, 5)

	for _, idStr := range idStrs {
		assert.NotEmpty(t, idStr)
		assert.Len(t, idStr, 36)

		parsed, err := uuid.Parse(idStr)
		require.NoError(t, err)
		assert.Equal(t, uuid.Version(7), parsed.Version())
	}
}

func TestConcurrentGeneration(t *testing.T) {
	gen := NewGenerator()

	const goroutines = 10
	const idsPerGoroutine = 100

	results := make(chan uuid.UUID, goroutines*idsPerGoroutine)
	errors := make(chan error, goroutines*idsPerGoroutine)

	// Generate UUIDs concurrently
	for i := 0; i < goroutines; i++ {
		go func() {
			for j := 0; j < idsPerGoroutine; j++ {
				id, err := gen.Generate()
				if err != nil {
					errors <- err
					return
				}
				results <- id
			}
		}()
	}

	// Collect all results
	allIDs := make(map[uuid.UUID]bool)
	for i := 0; i < goroutines*idsPerGoroutine; i++ {
		select {
		case id := <-results:
			assert.False(t, allIDs[id], "Duplicate UUID in concurrent generation")
			allIDs[id] = true
		case err := <-errors:
			t.Fatalf("Error generating UUID: %v", err)
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for UUID generation")
		}
	}

	assert.Len(t, allIDs, goroutines*idsPerGoroutine)
}

func TestParse(t *testing.T) {
	gen := NewGenerator()
	id, err := gen.Generate()
	require.NoError(t, err)

	idStr := id.String()
	parsed, err := Parse(idStr)
	require.NoError(t, err)
	assert.Equal(t, id, parsed)
}

func TestParseInvalid(t *testing.T) {
	_, err := Parse("invalid-uuid")
	assert.Error(t, err)
}

func TestMustParse(t *testing.T) {
	gen := NewGenerator()
	id, err := gen.Generate()
	require.NoError(t, err)

	idStr := id.String()
	parsed := MustParse(idStr)
	assert.Equal(t, id, parsed)
}

func TestMustParsePanic(t *testing.T) {
	assert.Panics(t, func() {
		MustParse("invalid-uuid")
	})
}

func TestIsValid(t *testing.T) {
	gen := NewGenerator()
	id, err := gen.Generate()
	require.NoError(t, err)

	assert.True(t, IsValid(id.String()))
	assert.False(t, IsValid("invalid-uuid"))
	assert.False(t, IsValid(""))
	assert.False(t, IsValid("not-a-uuid"))
}

func TestHighFrequencyGeneration(t *testing.T) {
	gen := NewGenerator()

	// Generate many UUIDs rapidly to test google/uuid's internal counter handling
	ids := make(map[uuid.UUID]bool)
	for i := 0; i < 10000; i++ {
		id, err := gen.Generate()
		require.NoError(t, err)
		assert.False(t, ids[id], "Duplicate UUID generated at iteration %d", i)
		ids[id] = true
		assert.Equal(t, uuid.Version(7), id.Version())
	}
}

func BenchmarkGenerate(b *testing.B) {
	gen := NewGenerator()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := gen.Generate()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGenerateString(b *testing.B) {
	gen := NewGenerator()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := gen.GenerateString()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGenerateBatch(b *testing.B) {
	gen := NewGenerator()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := gen.GenerateBatch(100)
		if err != nil {
			b.Fatal(err)
		}
	}
}
