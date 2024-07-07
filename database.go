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

func getConnection(config *DatabaseConfig) *sql.DB {
	log.Println("Creating new database connection")

	// Create a DSN (Data Source Name) string
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		config.Username, config.Password, config.Host, config.Port, config.DatabaseName)

	// Open the database connection
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(fmt.Sprintf("Failed to open a database connection: %v", err))
	}

	// Set connection settings
	db.SetMaxOpenConns(config.MaxConn)
	db.SetMaxIdleConns(1)
	db.SetConnMaxIdleTime(time.Minute * 10)
	db.SetConnMaxLifetime(time.Minute * 60)

	if err = db.Ping(); err != nil {
		panic(fmt.Sprintf("Failed to ping the database: %v", err))
	}

	return db
}

func GetInstance(dbConfig *DatabaseConfig) *MySQLConnPool {
	if instance == nil {
		instance = &MySQLConnPool{
			db: getConnection(dbConfig),
		}
	}

	return instance
}

func (pool *MySQLConnPool) Insert(query string, params []interface{}) {
	stmt, err := pool.db.Prepare(query)
	if err != nil {
		panic(fmt.Sprintf("Batch statement %v", err))
	}

	defer func(stmt *sql.Stmt) {
		if err := stmt.Close(); err != nil {
			panic(err)
		}
	}(stmt)

	_, err = stmt.Exec(params...)
	if err != nil {
		panic(fmt.Sprintf("Batch execution %v", err))
	}
}
