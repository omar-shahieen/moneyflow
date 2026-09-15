package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateSchema_ValidCSV(t *testing.T) {
	data := []byte("CategoryName,Amount,Note,Date\nFood,1000,lunch,2024-01-15\nTransport,500,bus,2024-01-16\n")
	err := ValidateSchema(data, 2)
	assert.NoError(t, err)
}

func TestValidateSchema_MissingRequiredColumn(t *testing.T) {
	data := []byte("Category,Amount,Note,Date\nFood,1000,lunch,2024-01-15\n")
	err := ValidateSchema(data, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required column: CategoryName")
}

func TestValidateSchema_MissingAmountColumn(t *testing.T) {
	data := []byte("CategoryName,Note,Date\nFood,lunch,2024-01-15\n")
	err := ValidateSchema(data, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required column: Amount")
}

func TestValidateSchema_EmptyCSV(t *testing.T) {
	data := []byte("")
	err := ValidateSchema(data, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must have a header row and at least one data row")
}

func TestValidateSchema_HeaderOnly(t *testing.T) {
	data := []byte("CategoryName,Amount,Note,Date\n")
	err := ValidateSchema(data, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must have a header row and at least one data row")
}

func TestValidateSchema_RowCountMismatch(t *testing.T) {
	data := []byte("CategoryName,Amount,Note,Date\nFood,1000,lunch,2024-01-15\nTransport,500,bus,2024-01-16\n")
	err := ValidateSchema(data, 5)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "row count mismatch: expected 5 rows, got 2")
}

func TestValidateSchema_RowCountZeroSkipsCheck(t *testing.T) {
	data := []byte("CategoryName,Amount,Note,Date\nFood,1000,lunch,2024-01-15\n")
	err := ValidateSchema(data, 0)
	assert.NoError(t, err)
}

func TestValidateSchema_ExtraColumnsAllowed(t *testing.T) {
	data := []byte("CategoryName,Amount,Note,Date,ExtraCol\nFood,1000,lunch,2024-01-15,extra\n")
	err := ValidateSchema(data, 1)
	assert.NoError(t, err)
}

func TestValidateSchema_CaseSensitiveColumns(t *testing.T) {
	data := []byte("categoryname,amount,note,date\nFood,1000,lunch,2024-01-15\n")
	err := ValidateSchema(data, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required column: CategoryName")
}
