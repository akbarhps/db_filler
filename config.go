package main

import (
	"database/sql"
	"fmt"
	"github.com/jaswdr/faker"
	"log"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ForeignKey struct {
	Table  string
	Column string
}

type TableSchema struct {
	Column        string
	Value         string
	IsUnique      bool
	IsUserDefined bool
	OneOfValues   []string
	ForeignKey    *ForeignKey
}

func (s *TableSchema) GetValue(fake faker.Faker) interface{} {
	if s.IsUserDefined {
		return s.OneOfValues[rand.Intn(len(s.OneOfValues))]
	}

	switch s.Value {
	case "address":
		return fake.Address().Address()
	case "age":
		return strconv.Itoa(rand.Intn(100))
	case "city":
		return fake.Address().City()
	case "color":
		return fake.Color().ColorName()
	case "color_hex":
		return fake.Color().Hex()
	case "color_css":
		return fake.Color().CSS()
	case "company_name":
		return fake.Company().Name()
	case "date":
		return fake.Time().Time(time.Now()).Format("2006-01-02")
	case "datetime":
		return fake.Time().Time(time.Now()).Format("2006-01-02 15:04:05")
	case "email":
		id := NewULID()
		return fmt.Sprintf("%s@%s.mail",
			fake.Person().FirstName()+fake.Person().LastName(),
			strings.ToLower(id[len(id)-7:]))
	case "name":
		return fake.Person().Name()
	case "password":
		return fake.Internet().Password()
	case "phone_number":
		return fake.Phone().Number()
	case "sentence":
		return fake.Lorem().Sentence(10)
	case "ulid":
		return NewULID()
	case "word":
		if s.IsUnique {
			id := NewULID()
			return strings.ToLower(id[len(id)-7:])
		}
		return fake.Lorem().Word()
	case "year":
		return strconv.Itoa(fake.Time().Year())
	default:
		return ""
	}
}

type Table struct {
	Name   string            `json:"name,omitempty"`
	Count  int               `json:"count,omitempty"`
	Schema map[string]string `json:"schema,omitempty"`
}

func (t *Table) getOrderedSchema() []string {
	keys := make([]string, 0, len(t.Schema))
	for k := range t.Schema {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (t *Table) GetAllSchema() []TableSchema {
	// sort the schema key first to maintain order
	orderedSchema := t.getOrderedSchema()
	var schemas []TableSchema

	for index := range orderedSchema {
		column := orderedSchema[index]
		rawValue := t.Schema[column]
		schema := TableSchema{
			Column: column,
		}

		if strings.Contains(rawValue, "oneof:") {
			vs := strings.ReplaceAll(rawValue, "oneof:", "")
			oneOfValues := make([]string, 0)
			for _, v := range strings.Split(vs, ",") {
				oneOfValues = append(oneOfValues, strings.TrimSpace(v))
			}

			schema.IsUserDefined = true
			schema.OneOfValues = oneOfValues
			schemas = append(schemas, schema)
			continue
		}

		for _, value := range strings.Split(rawValue, ",") {
			value = strings.Trim(value, " ")
			if value == "unique" {
				schema.IsUnique = true
				continue
			}

			if strings.Contains(value, "->") {
				v := strings.Split(value, "->")
				if len(v) != 2 {
					log.Panicf("Invalid foreign key definition in table %s schema %s", t.Name, column)
				}
				schema.ForeignKey = &ForeignKey{
					Table:  v[0],
					Column: v[1],
				}
				continue
			}
			schema.Value = value
		}
		schemas = append(schemas, schema)
	}
	return schemas
}

// generateInsertQuery will generate
// INSERT INTO table_name (column1, column2, column3) VALUES
func (t *Table) generateInsertQuery() string {
	// sort schema keys first to get consistent query
	keys := t.getOrderedSchema()

	query := "INSERT INTO " + t.Name + " ("
	for i := 0; i < len(keys); i++ {
		query += keys[i]
		if i < len(t.Schema)-1 {
			query += ","
		}
	}
	return query + ") VALUES"
}

// generatePrepareQuery will generate (?, ?, ?) for prepare statement
func (t *Table) generatePrepareQuery() string {
	query := "("
	for i := 0; i < len(t.Schema); i++ {
		query += "?"
		if i < len(t.Schema)-1 {
			query += ","
		}
	}
	return query + ")"
}

func (t *Table) Fill(db *sql.DB, cache *ConcurrentCache) {
	var wg sync.WaitGroup
	n := int(math.Ceil(float64(t.Count) / 5_000))

	for iter := 0; iter < n; iter++ {
		wg.Add(1)

		go func(iter int) {
			defer wg.Done()
			fake := faker.New()
			schemas := t.GetAllSchema()

			query := t.generateInsertQuery()
			perBatch := int(math.Min(5_000, float64(t.Count-(iter*5_000))))
			values := make([]interface{}, 0, perBatch)

			for i := 0; i < perBatch; i++ {
				for _, schema := range schemas {
					value := schema.GetValue(fake)

					if schema.ForeignKey == nil {
						foreign := &ForeignKey{
							Table:  t.Name,
							Column: schema.Column,
						}
						if cache.KeyExists(foreign) {
							cache.Append(foreign, value)
						}
					} else {
						foreign := *schema.ForeignKey
						if schema.IsUnique {
							value = cache.Pull(&foreign)
						} else {
							value = cache.GetRandom(&foreign)
						}
					}

					values = append(values, value)
				}

				query += " " + t.generatePrepareQuery()
				if i < perBatch-1 {
					query += ","
				}
			}

			InsertBatch(db, query, values)
		}(iter)
	}

	wg.Wait()
}

type Config struct {
	Database DatabaseConfig `json:"database"`
	Tables   []Table        `json:"tables,omitempty"`
}
