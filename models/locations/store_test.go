package locations_test

import (
	"database/sql"
	"testing"

	"github.com/cjsaylor/boxmeup-go/models"
	"github.com/cjsaylor/boxmeup-go/models/locations"
)

func TestSortableField_String(t *testing.T) {
	field := locations.SortFieldModified
	if field.String() != "modified" {
		t.Errorf("Expected 'modified' got %v", field.String())
	}
}

func TestGetSortBy(t *testing.T) {
	var sql sql.DB
	result := locations.NewStore(&sql).GetSortBy(locations.SortFieldName, models.ASC)
	if result.Direction != models.ASC {
		t.Errorf("Expected direction to be ascending but got %v", result.Direction)
	}
	if result.Field != locations.SortFieldName.String() {
		t.Errorf("Expected field to be 'name' but got %v", result.Field)
	}
}
