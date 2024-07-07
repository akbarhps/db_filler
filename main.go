package main

import (
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"math"
	"os"
	"path"
	"sync"
	"time"
)

const BatchSize = 5_000

func PrettyPrint(data interface{}) {
	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("%s \n", content)
}

type AppConfig struct {
	Database DatabaseConfig `json:"database"`
	Tables   []Table        `json:"tables,omitempty"`
}

func main() {
	if len(os.Args) == 1 {
		panic("Can't run program, no config file provided")
	}

	fileName := os.Args[1]
	execDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	fileBytes, err := os.ReadFile(path.Join(execDir, fileName))
	if err != nil {
		panic(err)
	}

	var config AppConfig
	if err = json.Unmarshal(fileBytes, &config); err != nil {
		panic(err)
	}

	go log.Println("Running database filler with config:")
	go PrettyPrint(config)

	// initialize foreign key cache
	go log.Println("Initializing cache")
	cache := NewConcurrentCache()
	for _, table := range config.Tables {
		for _, schema := range table.GetAllSchema() {
			foreign := schema.ForeignKey
			if foreign != nil {
				cache.Add(foreign.GetCacheKey(), nil)
			}
		}
	}

	totalCount := 0
	timer := time.Now()
	wg := sync.WaitGroup{}
	pool := GetInstance(&config.Database)

	log.Println("Filling tables")
	for _, t := range config.Tables {
		faker := NewFaker()
		totalCount += t.Count
		tableTimer := time.Now()
		totalIteration := int(math.Ceil(float64(t.Count) / BatchSize))

		log.Printf("Filling %s table with %d rows", t.Name, t.Count)
		for iteration := 0; iteration < totalIteration; iteration++ {
			wg.Add(1)
			go func(iter int, t *Table) {
				defer wg.Done()

				currentTotalSize := float64(t.Count - (iter * BatchSize))
				batchSize := int(math.Min(BatchSize, currentTotalSize))

				query := t.GenerateInsertQuery(batchSize)
				rows := t.GenerateInsertRows(cache, faker, batchSize)
				pool.Insert(query, rows)
			}(iteration, &t)
		}
		wg.Wait()

		log.Printf("%s table filled with duration %v", t.Name, time.Since(tableTimer))
		faker.ClearCache()
		cache.ClearCacheIndex()
	}

	log.Printf("Finish fill data with total %d records, duration %v ", totalCount, time.Since(timer))
}
