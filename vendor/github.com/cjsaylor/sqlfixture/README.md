# sqlfixture

[![GoDoc](https://godoc.org/github.com/cjsaylor/sqlfixture?status.png)](https://godoc.org/github.com/cjsaylor/sqlfixture)
[![Build Status](https://travis-ci.org/cjsaylor/sqlfixture.svg?branch=master)](https://travis-ci.org/cjsaylor/sqlfixture)

sqlfixture is a go library that enables simple pre-populating a MySQL database with data
to be used during testing.

Fixtures are supported via:

* Native `go` code
* `json` files
* `yaml` files

# Example

This example will truncate `my_table` and `my_other_table` and insert data into them
before each test is run.

> In foo_test.go

```go
package foo_test

import (
	"database/sql"
	"os"
	"testing"

	"github.com/cjsaylor/sqlfixture"
	_ "github.com/go-sql-driver/mysql"
)

func TestMain(m *testing.M) {
	db, _ := sql.Open("mysql", "tcp(localhost:3306)/test")
	err := db.Ping()
	if err != nil {
		panic(err)
	}
	setup(db)
	code := m.Run()
	os.Exit(code)
}

func setup(db *sql.DB) {
	// Setup your table schema here
	fixture := sqlfixture.New(db, sqlfixture.Tables{
		sqlfixture.Table{
			Name: "my_table",
			Rows: sqlfixture.Rows{
				sqlfixture.Row{
					"id": "1",
					"name": "Some value",
					"slug": "some-value",
					"date": "2017-05-15 00:00:00",
				},
			},
		},
		sqlfixture.Table{
			Name: "my_other_table",
			Rows: sqlfixture.Rows{
				sqlfixture.Row{
					"id": "1",
					"item": "Some item",
					"quantity": 9,
				},
				sqlfixture.Row{
					"id": "2",
					"item": "Some other item",
					"quantity": 3,
				},
			},
		},
	})
	fixture.Populate()
}
```

# Installation

```
go get github.com/cjsaylor/sqlfixture
```

## YAML support

You can import fixtures from YAML files.

We can rewrite the example above:

> In foo_test.go

```go
package foo_test

import (
	"database/sql"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/cjsaylor/sqlfixture"
	_ "github.com/go-sql-driver/mysql"
)

func TestMain(m *testing.M) {
	db, _ := sql.Open("mysql", "tcp(localhost:3306)/test")
	err := db.Ping()
	if err != nil {
		panic(err)
	}
	setup(db)
	code := m.Run()
	os.Exit(code)
}

func setup(db *sql.DB) {
	// Setup your table schema here
	filename, _ := filepath.Abs("./fixtures/test.yaml")
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Error("Unable to find fixture file.")
		return
	}
	fixture := sqlfixture.FromYAML(db, yamlFile)
	fixture.Populate()
}
```

> in fixtures/test.yaml

```yaml
- name: my_table
  rows:
    - id: 1
      name: "some value"
      slug: "some-value"
      date: "2017-05-15 00:00:00"
- name: my_other_table
  rows:
    - id: 1
      item: some item
      quantity: 9
    - id: 2
      item: some other item
      quantity: 3
```

## JSON Support

Similar to the `YAML` support, sqlfixture also support importing data via `JSON`:

> `foo_test.go`

```go
package foo_test

import (
	"database/sql"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/cjsaylor/sqlfixture"
	_ "github.com/go-sql-driver/mysql"
)

func TestMain(m *testing.M) {
	db, _ := sql.Open("mysql", "tcp(localhost:3306)/test")
	err := db.Ping()
	if err != nil {
		panic(err)
	}
	setup(db)
	code := m.Run()
	os.Exit(code)
}

func setup(db *sql.DB) {
	// Setup your table schema here
	filename, _ := filepath.Abs("./fixtures/test.json")
	jsonFile, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Error("Unable to find fixture file.")
		return
	}
	fixture := sqlfixture.FromJSON(db, jsonFile)
	fixture.Populate()
}
```

> `fixtures/test.json`
```json
[
  {
    "name": "my_table",
    "rows": [
      {
        "id": 1,
        "name": "some value",
        "slug": "some-value",
        "date": "2017-05-15 00:00:00"
      }
    ]
  },
  {
    "name": "my_other_table",
    "rows": [
      {
        "id": 1,
        "item": "some item",
        "quantity": 9
      },
      {
        "id": 2,
        "item": "some other item",
        "quantity": 3
      }
    ]
  }
]
```