package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

type MySQLConnPool struct {
	db *sql.DB
}

type DatabaseConfig struct {
	Username     string `json:"username,omitempty"`
	Password     string `json:"password,omitempty"`
	Host         string `json:"host,omitempty"`
	Port         string `json:"port,omitempty"`
	DatabaseName string `json:"name,omitempty"`
	MaxConn      int    `json:"max_conn,omitempty"`
}

var instance *MySQLConnPool

func getConnection(dbConfig *DatabaseConfig) *sql.DB {
	log.Println("Creating a new database connection with config:")
	PrettyPrint(dbConfig)

	// Create a DSN (Data Source Name) string
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		dbConfig.Username, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.DatabaseName)

	// Open the database connection
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Panicf("Failed to open a database connection: %v", err)
	}

	// Set connection settings
	db.SetMaxOpenConns(dbConfig.MaxConn)
	db.SetMaxIdleConns(1)
	db.SetConnMaxIdleTime(time.Minute * 10)
	db.SetConnMaxLifetime(time.Minute * 60)

	err = db.Ping()
	if err != nil {
		log.Panicf("Failed to ping the database: %v", err)
	}

	return db
}

func GetInstance(dbConfig *DatabaseConfig) *MySQLConnPool {
	if instance == nil {
		instance = &MySQLConnPool{}
		instance.db = getConnection(dbConfig)
	}
	return instance
}

func InsertBatch(db *sql.DB, query string, params []interface{}) {
	stmt, err := db.Prepare(query)
	if err != nil {
		log.Panicf("Batch statement %v", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(params...)
	if err != nil {
		log.Panicf("Batch execution %v", err)
	}
}
