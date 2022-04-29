package models

import (
	"database/sql"
	"fmt"
	"log"
)

const (
	host     = 	"localhost"
	port     = 	5432
	user     = 	"postgres"
	password = 	"pass"
	dbname   = 	"MortyGrabberDB"
)

func OpenConnection() *sql.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	//psqlInfo := "host=localhost port=5432 user=postgres password=pass dbname=MortyGRAB sslmode=disable"
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	return db
}