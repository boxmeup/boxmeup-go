package locations_test

import (
	"database/sql"
	"testing"

	"github.com/cjsaylor/boxmeup-go/modules/locations"
	"github.com/cjsaylor/boxmeup-go/modules/models"
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

func TestSortableFieldByName(t *testing.T) {
	var sql sql.DB
	locationModel := locations.NewStore(&sql)
	result, err := locationModel.SortableFieldByName("modified")
	if err != nil {
		t.Error(err)
	}
	if result != locations.SortFieldModified {
		t.Errorf("Expected modified sort field but got %v", result)
	}
	result, err = locationModel.SortableFieldByName("not_existing")
	if err == nil {
		t.Error("Expected to fail with not found field.")
	}
}
