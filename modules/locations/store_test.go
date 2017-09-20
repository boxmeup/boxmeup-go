package locations_test

import (
	"database/sql"
	"os"
	"testing"

	"github.com/cjsaylor/boxmeup-go/modules/locations"
	"github.com/cjsaylor/boxmeup-go/modules/models"
	"github.com/cjsaylor/boxmeup-go/modules/users"
	"github.com/cjsaylor/sqlfixture"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func TestMain(m *testing.M) {
	// @todo replace this with configured database from app
	db, _ = sql.Open("mysql", "root:supersecret@tcp(localhost:3306)/bmu_test?parseTime=true")
	defer db.Close()
	setup(db)
	os.Exit(m.Run())
}

func setup(db *sql.DB) {
	db.Exec("SET FOREIGN_KEY_CHECKS=0")
	fixture := sqlfixture.New(db, sqlfixture.Tables{
		sqlfixture.Table{
			Name: "users",
			Rows: sqlfixture.Rows{
				sqlfixture.Row{
					"id":        1,
					"email":     "test@test.com",
					"is_active": 1,
					"created":   "2017-05-15",
					"modified":  "2017-05-15",
				},
			},
		},
		sqlfixture.Table{
			Name: "locations",
			Rows: sqlfixture.Rows{
				sqlfixture.Row{
					"id":              1,
					"user_id":         1,
					"uuid":            "ff1eda35-4183-11e7-9cc8-0242ac120003",
					"name":            "My Garage",
					"address":         "",
					"container_count": 0,
					"is_mappable":     false,
					"created":         "2017-05-15",
					"modified":        "2017-05-15",
				},
				sqlfixture.Row{
					"id":              2,
					"user_id":         1,
					"uuid":            "ff1eda35-4183-11e7-9cc8-0242ac120004",
					"name":            "Basement",
					"address":         "",
					"container_count": 0,
					"is_mappable":     false,
					"created":         "2017-05-15",
					"modified":        "2017-05-15 00:00:01",
				},
			},
		},
	})
	fixture.Populate()
	db.Exec("SET FOREIGN_KEY_CHECKS=1")
}

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

func TestStore_ByID(t *testing.T) {
	locationModel := locations.NewStore(db)
	result, err := locationModel.ByID(1)
	if err != nil {
		t.Error(err)
		return
	}
	if result.ID != 1 || result.Name != "My Garage" {
		t.Error("Record retrieved does not matche what we expected.")
		return
	}
}

func TestStore_Create(t *testing.T) {
	locationModel := locations.NewStore(db)
	location := locations.Location{
		User: users.User{
			ID: 1,
		},
		Name:    "Some New Location",
		Address: "123 Easy St.",
	}
	err := locationModel.Create(&location)
	if err != nil {
		t.Error(err)
		return
	}
	if location.ID == 0 {
		t.Error("Expected location ID to be set.")
		return
	}
	result, err := locationModel.ByID(location.ID)
	if err != nil {
		t.Error(err)
		return
	}
	if result.ID != location.ID {
		t.Error("unexpected location retrieved.")
		return
	}
}

func TestStore_Update(t *testing.T) {
	locationModel := locations.NewStore(db)
	location, err := locationModel.ByID(1)
	if err != nil {
		t.Error(err)
		return
	}
	location.Name = "A new name"
	err = locationModel.Update(&location)
	if err != nil {
		t.Error(err)
		return
	}
	result, err := locationModel.ByID(1)
	if err != nil {
		t.Error(err)
		return
	}
	if result.Name != location.Name {
		t.Errorf("Expected %v but got %v", location.Name, result.Name)
	}
}

func TestStore_Delete(t *testing.T) {
	locationModel := locations.NewStore(db)
	err := locationModel.Delete(1)
	if err != nil {
		t.Error(err)
		return
	}
	result, err := locationModel.ByID(1)
	if err == nil {
		t.Errorf("Expected no result to return but got %v", result)
		return
	}
}

func TestStore_FilteredLocations(t *testing.T) {
	locationModel := locations.NewStore(db)
	sort := locationModel.GetSortBy(locations.SortFieldModified, models.ASC)
	limit := models.QueryLimit{
		Limit:  1,
		Offset: 1,
	}
	filter := locations.LocationFilter{
		User: users.User{ID: 1},
	}
	result, err := locationModel.FilteredLocations(filter, sort, limit)
	if err != nil {
		t.Error(err)
		return
	}
	if len(result.Locations) != 1 {
		t.Errorf("Expected 1 result but got %v", len(result.Locations))
		return
	}
	if result.Locations[0].ID != 2 {
		t.Errorf("Expected ID of result to be 2 but got %v", result.Locations[0].ID)
		return
	}
}
