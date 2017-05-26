// Package sqlfixture is a go library that enables pre-populating a MySQL database with data to be used during testing.
package sqlfixture

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v2"
)

// Row represents an arbitrary key-value representation of a sql row to insert.
type Row map[string]interface{}

// Rows represents a collection of Row
type Rows []Row

// Table represents a database table with its correspoding rows to insert
type Table struct {
	Name string `json:"name" yaml:"name"`
	Rows Rows   `json:"rows" yaml:"rows"`
}

// Tables represents a collection of Table
type Tables []Table

// Fixture structure to house tables to be populated
type Fixture struct {
	db     *sql.DB
	Tables Tables
}

// New fixture instance that will work with the database
func New(db *sql.DB, tables Tables) Fixture {
	fixture := Fixture{db: db, Tables: tables}
	return fixture
}

// FromYAML allows a fixture to be created from yaml input
func FromYAML(db *sql.DB, yamlIn []byte) (Fixture, error) {
	var tables Tables
	err := yaml.Unmarshal(yamlIn, &tables)
	return New(db, tables), err
}

// FromJSON allows a fixture to be created from json input
func FromJSON(db *sql.DB, jsonIn []byte) (Fixture, error) {
	var tables Tables
	err := json.Unmarshal(jsonIn, &tables)
	return New(db, tables), err
}

// Populate the database tables within this fixture
// Warning: This will truncate any and all data for each table in the fixture
func (f *Fixture) Populate() {
	for _, t := range f.Tables {
		_, err := f.db.Exec(fmt.Sprintf("truncate %v", t.Name))
		if err != nil {
			panic(err)
		}
		for _, r := range t.Rows {
			q := "insert into %v (%v) values (?%v)"
			columns := ""
			values := make([]interface{}, len(r))
			i := 0
			for k, v := range r {
				columns += k + ","
				values[i] = v
				i++
			}
			columns = strings.Trim(columns, ",")
			q = fmt.Sprintf(q, t.Name, columns, strings.Repeat(",?", len(values)-1))

			_, err := f.db.Exec(q, values...)
			if err != nil {
				panic(err)
			}
		}
	}
}
