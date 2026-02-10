package database

// NOTE: These tests use SQLite which requires CGO to be enabled.
// To run these tests on Windows, you need to install a C compiler like MinGW-w64:
//   - Install MinGW-w64 from https://www.mingw-w64.org/
//   - Add MinGW-w64/bin to your PATH
//   - Run tests with: go test -tags=cgo
// On Linux/Mac, CGO is usually available by default.
//
// Alternatively, skip these tests and rely on integration tests with PostgreSQL.

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestEntity is a test model for repository operations
type TestEntity struct {
	ID        uint      `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Name      string `gorm:"size:255"`
	Status    string `gorm:"size:50"`
	Value     int
}

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto migrate the test entity
	err = db.AutoMigrate(&TestEntity{})
	require.NoError(t, err)

	return db
}

// TestBaseRepository_Insert tests inserting a single entity
func TestBaseRepository_Insert(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBaseRepository[TestEntity](db)
	ctx := context.Background()

	t.Run("insert valid entity", func(t *testing.T) {
		entity := &TestEntity{
			Name:   "Test Entity",
			Status: "active",
			Value:  100,
		}

		err := repo.Insert(ctx, entity)
		assert.NoError(t, err)
		assert.NotZero(t, entity.ID)
	})

	t.Run("insert multiple entities", func(t *testing.T) {
		entity1 := &TestEntity{Name: "Entity 1", Status: "active", Value: 1}
		entity2 := &TestEntity{Name: "Entity 2", Status: "inactive", Value: 2}

		err := repo.Insert(ctx, entity1)
		require.NoError(t, err)

		err = repo.Insert(ctx, entity2)
		require.NoError(t, err)

		assert.NotEqual(t, entity1.ID, entity2.ID)
	})
}

// TestBaseRepository_InsertBatch tests batch insert
func TestBaseRepository_InsertBatch(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBaseRepository[TestEntity](db)
	ctx := context.Background()

	t.Run("insert batch of entities", func(t *testing.T) {
		entities := []*TestEntity{
			{Name: "Batch 1", Status: "active", Value: 10},
			{Name: "Batch 2", Status: "active", Value: 20},
			{Name: "Batch 3", Status: "inactive", Value: 30},
		}

		err := repo.InsertBatch(ctx, entities)
		assert.NoError(t, err)

		for _, entity := range entities {
			assert.NotZero(t, entity.ID)
		}
	})

	t.Run("insert empty batch", func(t *testing.T) {
		entities := []*TestEntity{}
		err := repo.InsertBatch(ctx, entities)
		assert.NoError(t, err)
	})
}

// TestBaseRepository_GetByID tests retrieving entity by ID
func TestBaseRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBaseRepository[TestEntity](db)
	ctx := context.Background()

	t.Run("get existing entity", func(t *testing.T) {
		// Create entity
		entity := &TestEntity{Name: "Test", Status: "active", Value: 50}
		err := repo.Insert(ctx, entity)
		require.NoError(t, err)

		// Retrieve entity
		retrieved, err := repo.GetByID(ctx, entity.ID)
		assert.NoError(t, err)
		assert.Equal(t, entity.Name, retrieved.Name)
		assert.Equal(t, entity.Status, retrieved.Status)
		assert.Equal(t, entity.Value, retrieved.Value)
	})

	t.Run("get non-existent entity", func(t *testing.T) {
		retrieved, err := repo.GetByID(ctx, 99999)
		assert.Error(t, err)
		assert.Nil(t, retrieved)
		assert.Contains(t, err.Error(), "not found")
	})
}

// TestBaseRepository_GetAll tests retrieving all entities
func TestBaseRepository_GetAll(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBaseRepository[TestEntity](db)
	ctx := context.Background()

	// Setup test data
	entities := []*TestEntity{
		{Name: "Entity 1", Status: "active", Value: 1},
		{Name: "Entity 2", Status: "active", Value: 2},
		{Name: "Entity 3", Status: "inactive", Value: 3},
		{Name: "Entity 4", Status: "active", Value: 4},
		{Name: "Entity 5", Status: "inactive", Value: 5},
	}
	err := repo.InsertBatch(ctx, entities)
	require.NoError(t, err)

	t.Run("get all without pagination", func(t *testing.T) {
		results, err := repo.GetAll(ctx, 0, 0)
		assert.NoError(t, err)
		assert.Len(t, results, 5)
	})

	t.Run("get all with limit", func(t *testing.T) {
		results, err := repo.GetAll(ctx, 3, 0)
		assert.NoError(t, err)
		assert.Len(t, results, 3)
	})

	t.Run("get all with offset", func(t *testing.T) {
		results, err := repo.GetAll(ctx, 0, 2)
		assert.NoError(t, err)
		assert.Len(t, results, 3) // Total 5 - offset 2 = 3
	})

	t.Run("get all with limit and offset", func(t *testing.T) {
		results, err := repo.GetAll(ctx, 2, 1)
		assert.NoError(t, err)
		assert.Len(t, results, 2)
	})
}

// TestBaseRepository_GetWhere tests conditional retrieval
func TestBaseRepository_GetWhere(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBaseRepository[TestEntity](db)
	ctx := context.Background()

	// Setup test data
	entities := []*TestEntity{
		{Name: "Active 1", Status: "active", Value: 10},
		{Name: "Active 2", Status: "active", Value: 20},
		{Name: "Inactive 1", Status: "inactive", Value: 30},
	}
	err := repo.InsertBatch(ctx, entities)
	require.NoError(t, err)

	t.Run("get by status", func(t *testing.T) {
		results, err := repo.GetWhere(ctx, map[string]interface{}{
			"status": "active",
		})
		assert.NoError(t, err)
		assert.Len(t, results, 2)
		for _, entity := range results {
			assert.Equal(t, "active", entity.Status)
		}
	})

	t.Run("get by multiple conditions", func(t *testing.T) {
		results, err := repo.GetWhere(ctx, map[string]interface{}{
			"status": "active",
			"value":  10,
		})
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "Active 1", results[0].Name)
	})

	t.Run("get with no matches", func(t *testing.T) {
		results, err := repo.GetWhere(ctx, map[string]interface{}{
			"status": "pending",
		})
		assert.NoError(t, err)
		assert.Len(t, results, 0)
	})
}

// TestBaseRepository_UpdateByID tests updating entity by ID
func TestBaseRepository_UpdateByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBaseRepository[TestEntity](db)
	ctx := context.Background()

	t.Run("update existing entity", func(t *testing.T) {
		// Create entity
		entity := &TestEntity{Name: "Original", Status: "active", Value: 100}
		err := repo.Insert(ctx, entity)
		require.NoError(t, err)

		// Update entity
		entity.Name = "Updated"
		entity.Value = 200
		err = repo.UpdateByID(ctx, entity.ID, entity)
		assert.NoError(t, err)

		// Verify update
		retrieved, err := repo.GetByID(ctx, entity.ID)
		assert.NoError(t, err)
		assert.Equal(t, "Updated", retrieved.Name)
		assert.Equal(t, 200, retrieved.Value)
	})
}

// TestBaseRepository_UpdateWhere tests conditional update
func TestBaseRepository_UpdateWhere(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBaseRepository[TestEntity](db)
	ctx := context.Background()

	// Setup test data
	entities := []*TestEntity{
		{Name: "Entity 1", Status: "active", Value: 10},
		{Name: "Entity 2", Status: "active", Value: 20},
		{Name: "Entity 3", Status: "inactive", Value: 30},
	}
	err := repo.InsertBatch(ctx, entities)
	require.NoError(t, err)

	t.Run("update multiple entities", func(t *testing.T) {
		err := repo.UpdateWhere(ctx,
			map[string]interface{}{"status": "active"},
			map[string]interface{}{"value": 999},
		)
		assert.NoError(t, err)

		// Verify updates
		results, err := repo.GetWhere(ctx, map[string]interface{}{"status": "active"})
		assert.NoError(t, err)
		for _, entity := range results {
			assert.Equal(t, 999, entity.Value)
		}

		// Verify inactive entity not updated
		inactive, err := repo.GetWhere(ctx, map[string]interface{}{"status": "inactive"})
		assert.NoError(t, err)
		assert.Equal(t, 30, inactive[0].Value)
	})
}

// TestBaseRepository_DeleteByID tests deleting entity by ID
func TestBaseRepository_DeleteByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBaseRepository[TestEntity](db)
	ctx := context.Background()

	t.Run("delete existing entity", func(t *testing.T) {
		entity := &TestEntity{Name: "To Delete", Status: "active", Value: 50}
		err := repo.Insert(ctx, entity)
		require.NoError(t, err)

		err = repo.DeleteByID(ctx, entity.ID)
		assert.NoError(t, err)

		// Verify deletion
		_, err = repo.GetByID(ctx, entity.ID)
		assert.Error(t, err)
	})

	t.Run("delete non-existent entity", func(t *testing.T) {
		err := repo.DeleteByID(ctx, 99999)
		assert.NoError(t, err) // GORM doesn't error on delete of non-existent ID
	})
}

// TestBaseRepository_DeleteWhere tests conditional deletion
func TestBaseRepository_DeleteWhere(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBaseRepository[TestEntity](db)
	ctx := context.Background()

	// Setup test data
	entities := []*TestEntity{
		{Name: "Keep 1", Status: "active", Value: 10},
		{Name: "Delete 1", Status: "inactive", Value: 20},
		{Name: "Delete 2", Status: "inactive", Value: 30},
	}
	err := repo.InsertBatch(ctx, entities)
	require.NoError(t, err)

	t.Run("delete multiple entities", func(t *testing.T) {
		err := repo.DeleteWhere(ctx, map[string]interface{}{
			"status": "inactive",
		})
		assert.NoError(t, err)

		// Verify deletions
		remaining, err := repo.GetAll(ctx, 0, 0)
		assert.NoError(t, err)
		assert.Len(t, remaining, 1)
		assert.Equal(t, "active", remaining[0].Status)
	})
}

// TestBaseRepository_Count tests counting entities
func TestBaseRepository_Count(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBaseRepository[TestEntity](db)
	ctx := context.Background()

	// Setup test data
	entities := []*TestEntity{
		{Name: "Entity 1", Status: "active", Value: 10},
		{Name: "Entity 2", Status: "active", Value: 20},
		{Name: "Entity 3", Status: "inactive", Value: 30},
	}
	err := repo.InsertBatch(ctx, entities)
	require.NoError(t, err)

	t.Run("count all entities", func(t *testing.T) {
		count, err := repo.Count(ctx, map[string]interface{}{})
		assert.NoError(t, err)
		assert.Equal(t, int64(3), count)
	})

	t.Run("count with condition", func(t *testing.T) {
		count, err := repo.Count(ctx, map[string]interface{}{
			"status": "active",
		})
		assert.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})
}

// TestBaseRepository_Exists tests checking entity existence
func TestBaseRepository_Exists(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBaseRepository[TestEntity](db)
	ctx := context.Background()

	// Setup test data
	entity := &TestEntity{Name: "Test", Status: "active", Value: 50}
	err := repo.Insert(ctx, entity)
	require.NoError(t, err)

	t.Run("check existing entity", func(t *testing.T) {
		exists, err := repo.Exists(ctx, map[string]interface{}{
			"status": "active",
		})
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("check non-existent entity", func(t *testing.T) {
		exists, err := repo.Exists(ctx, map[string]interface{}{
			"status": "pending",
		})
		assert.NoError(t, err)
		assert.False(t, exists)
	})
}

// TestBaseRepository_WithTx tests transaction support
func TestBaseRepository_WithTx(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBaseRepository[TestEntity](db)
	ctx := context.Background()

	t.Run("use repository with transaction", func(t *testing.T) {
		tx := db.Begin()
		txRepo := repo.WithTx(tx)

		entity := &TestEntity{Name: "TX Test", Status: "active", Value: 100}
		err := txRepo.Insert(ctx, entity)
		assert.NoError(t, err)

		// Rollback transaction
		tx.Rollback()

		// Verify entity was not persisted
		_, err = repo.GetByID(ctx, entity.ID)
		assert.Error(t, err)
	})
}

// TestBaseRepository_GetDB tests getting underlying database connection
func TestBaseRepository_GetDB(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBaseRepository[TestEntity](db)

	t.Run("get database connection", func(t *testing.T) {
		dbConn := repo.GetDB()
		assert.NotNil(t, dbConn)
		assert.Equal(t, db, dbConn)
	})
}

// TestMasterRepo_Creation tests MasterRepo creation
func TestMasterRepo_Creation(t *testing.T) {
	t.Run("create master repo", func(t *testing.T) {
		db := setupTestDB(t)
		dbManager := &DatabaseManager{
			MasterDB: db,
		}

		repo := NewMasterRepo[TestEntity](dbManager)
		assert.NotNil(t, repo)
		assert.NotNil(t, repo.BaseRepository)
	})
}

// TestTenantRepo_Creation tests TenantRepo creation
func TestTenantRepo_Creation(t *testing.T) {
	t.Run("create tenant repo", func(t *testing.T) {
		db := setupTestDB(t)
		connManager := &TenantConnectionManager{
			masterDB: db,
		}

		repo := NewTenantRepo[TestEntity](connManager)
		assert.NotNil(t, repo)
		assert.NotNil(t, repo.connManager)
	})
}
