package main

import (
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	"go.bryk.io/pkg/ulid"
	"log"
	"math/rand"
	"os"
	"path"
	"sync"
	"time"
)

type ConcurrentCache struct {
	sync.RWMutex
	UniqueIndexes map[string]map[string]int
	Items         map[string]map[string][]interface{}
}

func NewConcurrentCache() *ConcurrentCache {
	return &ConcurrentCache{
		UniqueIndexes: make(map[string]map[string]int),
		Items:         make(map[string]map[string][]interface{}),
	}
}

func (cc *ConcurrentCache) Append(fd *ForeignKey, value interface{}) {
	cc.Lock()
	defer cc.Unlock()

	_, ok := cc.Items[fd.Table]
	if !ok {
		cc.Items[fd.Table] = make(map[string][]interface{})
		cc.UniqueIndexes[fd.Table] = make(map[string]int)
	}

	_, ok = cc.Items[fd.Table][fd.Column]
	if !ok {
		cc.Items[fd.Table][fd.Column] = make([]interface{}, 0)
		cc.UniqueIndexes[fd.Table][fd.Column] = -1
	}

	if value == nil {
		return
	}

	cc.Items[fd.Table][fd.Column] = append(cc.Items[fd.Table][fd.Column], value)
}

func (cc *ConcurrentCache) KeyExists(fd *ForeignKey) bool {
	cc.Lock()
	defer cc.Unlock()

	_, ok := cc.Items[fd.Table][fd.Column]
	return ok
}

func (cc *ConcurrentCache) GetRandom(fk *ForeignKey) interface{} {
	cc.Lock()
	defer cc.Unlock()

	cache, ok := cc.Items[fk.Table][fk.Column]
	if !ok {
		log.Panicf("Cache doesn't have key for table %s column %s", fk.Table, fk.Column)
	}

	cacheSize := len(cache)
	if cacheSize == 0 {
		log.Panicf("Cache doesn't have value for table %s column %s", fk.Table, fk.Column)
	}

	return cache[rand.Intn(cacheSize)]
}

func (cc *ConcurrentCache) Pull(fk *ForeignKey) interface{} {
	cc.Lock()
	defer cc.Unlock()

	cache, ok := cc.Items[fk.Table][fk.Column]
	if !ok {
		log.Panicf("Cache doesn't have key for table %s column %s", fk.Table, fk.Column)
	}

	cacheSize := len(cache)
	if cacheSize == 0 {
		log.Panicf("Cache doesn't have value for table %s column %s", fk.Table, fk.Column)
	}

	lastIndex, _ := cc.UniqueIndexes[fk.Table][fk.Column]
	cc.UniqueIndexes[fk.Table][fk.Column] += 1
	return cache[lastIndex]
}

func (cc *ConcurrentCache) ResetUniqueIndexes() {
	cc.Lock()
	defer cc.Unlock()

	for table, columns := range cc.UniqueIndexes {
		for column := range columns {
			cc.UniqueIndexes[table][column] = 0
		}
	}
}

func NewULID() string {
	id, err := ulid.New()
	if err != nil {
		log.Panicf(err.Error())
	}

	return id.String()
}

func PrettyPrint(data interface{}) {
	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("%s \n", content)
}

func main() {
	if len(os.Args) == 1 {
		panic("Can't run program, no file provided")
	}

	fileName := os.Args[1]
	execDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	fileBytes, err := os.ReadFile(path.Join(execDir, fileName))
	if err != nil {
		log.Fatal(err)
	}

	var config Config
	err = json.Unmarshal(fileBytes, &config)
	if err != nil {
		log.Fatalln(err)
	}

	go log.Println("Running database filler with config:")
	go PrettyPrint(config)

	cache := NewConcurrentCache()
	dbPool := GetInstance(&config.Database)

	for _, table := range config.Tables {
		for _, schema := range table.GetAllSchema() {
			foreign := schema.ForeignKey
			if foreign != nil {
				cache.Append(foreign, nil)
			}
		}
	}

	t := time.Now()
	totalCount := 0

	go log.Println("Start filling all tables")
	for _, table := range config.Tables {
		log.Printf("Filling %s table with %d items", table.Name, table.Count)
		totalCount += table.Count

		tableTime := time.Now()
		table.Fill(dbPool.db, cache)
		log.Printf("%s table fill complete with duration %v", table.Name, time.Since(tableTime))

		cache.ResetUniqueIndexes()
	}

	log.Printf("Tables filled with total %d items, duration %v ", totalCount, time.Since(t))
}
