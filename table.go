package main

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
)

type ForeignKey struct {
	Table  string
	Column string
}

func (fk *ForeignKey) GetCacheKey() string {
	return GetCacheKey(fk.Table, fk.Column)
}

type Schema struct {
	Column      string
	FakerKey    string
	IsUnique    bool
	OneOfValues []string
	ForeignKey  *ForeignKey
}

func (s *Schema) Fake(faker *Faker) string {
	if len(s.OneOfValues) > 0 {
		return s.OneOfValues[rand.Intn(len(s.OneOfValues))]
	}

	if s.ForeignKey != nil {
		return ""
	}

	if s.IsUnique {
		return faker.GetUnique(s.FakerKey)
	}

	return faker.Get(s.FakerKey)
}

type Table struct {
	Name   string            `json:"name,omitempty"`
	Count  int               `json:"count,omitempty"`
	Schema map[string]string `json:"schema,omitempty"`
}

func (t *Table) sortedSchemaKeys() []string {
	keys := make([]string, 0, len(t.Schema))

	for k := range t.Schema {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}

func (t *Table) parseSchema(column string, rawValue string) *Schema {
	schema := &Schema{
		Column: column,
	}

	if strings.Contains(rawValue, "oneof:") {
		oneOfRawValues := strings.ReplaceAll(rawValue, "oneof:", "")
		if len(oneOfRawValues) == 0 {
			panic(fmt.Sprintf("Invalid oneof definition in table %s schema %s", t.Name, column))
		}

		oneOfValues := make([]string, 0)
		for _, value := range strings.Split(oneOfRawValues, ",") {
			oneOfValues = append(oneOfValues, strings.TrimSpace(value))
		}

		schema.OneOfValues = oneOfValues
		return schema
	}

	for _, value := range strings.Split(rawValue, ",") {
		value = strings.TrimSpace(value)

		if value == "unique" {
			schema.IsUnique = true
			continue
		}

		if strings.Contains(value, "->") {
			foreignValues := strings.Split(value, "->")

			if len(foreignValues) != 2 {
				panic(fmt.Sprintf(
					"Invalid foreign column definition in table %s schema %s\nformat: table_name->column_name",
					t.Name, column,
				))
			}

			schema.ForeignKey = &ForeignKey{
				Table:  foreignValues[0],
				Column: foreignValues[1],
			}
			continue
		}

		schema.FakerKey = value
	}

	return schema
}

func (t *Table) GetAllSchema() []Schema {
	var schemas []Schema
	// sort the schema key to maintain order
	keys := t.sortedSchemaKeys()

	for index := range keys {
		column := keys[index]
		value := t.Schema[column]

		schema := t.parseSchema(column, value)
		schemas = append(schemas, *schema)
	}

	return schemas
}

// GenerateInsertQuery will generate insert query using prepare statement for n rows
// INSERT INTO table_name (column1, column2, column3) VALUES (?, ?, ?), (?, ?, ?), ...
func (t *Table) GenerateInsertQuery(nRows int) string {
	// sort schema keys first to get consistent query
	keys := t.sortedSchemaKeys()

	// generate (?, ?, ?) for prepare statement
	colsValue := make([]string, 0, len(keys))
	for i := 0; i < len(keys); i++ {
		colsValue = append(colsValue, "?")
	}

	prepareRows := make([]string, 0, nRows)
	prepareColumns := fmt.Sprintf("(%s)", strings.Join(colsValue, ","))
	for j := 0; j < nRows; j++ {
		prepareRows = append(prepareRows, prepareColumns)
	}

	return fmt.Sprintf("INSERT INTO %s (%s) VALUES %s",
		t.Name,
		strings.Join(keys, ","),
		strings.Join(prepareRows, ","),
	)
}

func (t *Table) GenerateInsertRows(cache *ConcurrentCache, faker *Faker, nRow int) []interface{} {
	schemas := t.GetAllSchema()
	values := make([]interface{}, 0)

	for i := 0; i < nRow; i++ {
		for _, schema := range schemas {
			value := schema.Fake(faker)

			if schema.ForeignKey != nil {
				cacheKey := schema.ForeignKey.GetCacheKey()
				if schema.IsUnique {
					value = cache.Pull(cacheKey)
				} else {
					value = cache.GetRandom(cacheKey)
				}
			} else {
				foreign := &ForeignKey{
					Table:  t.Name,
					Column: schema.Column,
				}
				cacheKey := foreign.GetCacheKey()
				if cache.ItemExists(cacheKey) {
					cache.Add(cacheKey, &value)
				}
			}

			values = append(values, value)
		}
	}

	return values
}
