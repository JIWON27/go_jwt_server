package db

import (
	"database/sql"
	"fmt"
	"jwt/models"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var PostgreDB *gorm.DB
var Sqldb *sql.DB

func ConnectDB() (err error) {
	dbinfo := fmt.Sprintf("user=%s password=%s host=%s port=%s  dbname=%s sslmode=disable",
		"jiwon", "0512", "localhost", "5432", "UserDB")

	PostgreDB, err = gorm.Open(postgres.Open(dbinfo), &gorm.Config{})

	if err != nil {
		panic(err)
	}

	Sqldb, err = PostgreDB.DB()

	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	Sqldb.SetMaxIdleConns(10)

	// SetMaxOpenConns sets the maximum number of open connections to the database.
	Sqldb.SetMaxOpenConns(100)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	Sqldb.SetConnMaxLifetime(time.Hour)

	if err != nil {
		log.Println("models/setup.go : Setup func error : " + err.Error())
		return err
	}

	// User 테이블 자동 생성
	err = PostgreDB.AutoMigrate(&models.User{})
	if err != nil {
		return err
	}

	log.Println("Connected DB")
	return nil

}
