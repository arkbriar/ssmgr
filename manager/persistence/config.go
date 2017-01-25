// Package persistence defines ORM models and maintains
// the connection with database.
package persistence

import (
	"log"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"

	// import mysql, postgressql, sqlite drivers
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var (
	db *gorm.DB
	// Parsing from config files or command line arguments
	dialect string
	dbName  string
)

const (
	logFile = "/tmp/ssmgr-db.log"
)

// migrate(create) tables: users, admin_users, servers, services,
// products, orders.
func autoMigrate() {
	if dialect == "mysql" {
		db.InstantSet("gorm:table_options", "ENGINE=InnoDB")
	}
	tables := []interface{}{&Server{}, &User{}, &User{Role: "admin"}, &Service{}, &Product{}, &Order{}}
	errs := db.AutoMigrate(tables...).GetErrors()
	if errs != nil {
		logrus.Panicf("Tables' migration failed: %s\n", errs)
	}
}

func init() {
	var err error
	db, err = gorm.Open(dialect, dbName)
	if err != nil {
		logrus.Panicln(err)
	}
	db.LogMode(true).SetLogger(log.New(mustOpen(logFile), "\r\n", 0))
	// Set up connection pool
	db.DB().SetMaxIdleConns(20)
	db.DB().SetMaxOpenConns(50)
	autoMigrate()
}

func mustOpen(filename string) *os.File {
	file, err := os.Open(filename)
	if err != nil {
		logrus.Panicln(err)
	}
	return file
}

// GetDB returns a clone of the current DB connection.
func GetDB() *gorm.DB {
	return db.New()
}
