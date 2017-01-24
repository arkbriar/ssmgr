// Package persistence defines ORM models and maintains
// the connection with database.
package persistence

import (
	"log"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"
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

func init() {
	var err error
	db, err = gorm.Open(dialect, dbName)
	if err != nil {
		logrus.Panicln(err)
	}
	initConfig()
	createTablesIfNotExsit()
}

func mustOpen(filename string) *os.File {
	file, err := os.Open(filename)
	if err != nil {
		logrus.Panicln(err)
	}
	return file
}

func initConfig() {
	db.LogMode(true).SetLogger(log.New(mustOpen(logFile), "\r\n", 0))
	// Set up connection pool
	db.DB().SetMaxIdleConns(20)
	db.DB().SetMaxOpenConns(50)
}

// create tables: users, admin_users, servers, services,
// products, orders.
func createTablesIfNotExsit() {
	if dialect == "mysql" {
		db.InstantSet("gorm:table_options", "ENGINE=InnoDB")
	}
	defaults := []interface{}{&Server{}, &User{}, &User{Role: "admin"}, &Service{}, &Product{}, &Order{}}
	for table := range defaults {
		if !db.HasTable(table) {
			if err := db.CreateTable(table).Error; err != nil {
				logrus.Panicln(err)
			}
		}
	}
}

func GetDB() *gorm.DB {
	return db.New()
}