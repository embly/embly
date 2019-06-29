package main

import (
	"database/sql"
	"embly/api/pkg/db"
)

// Up is executed when this migration is applied
func Up_20190623002235(txn *sql.Tx) {
	db.ExecQueries(txn,
		`CREATE TABLE users (
			id SERIAL PRIMARY KEY,
			username VARCHAR NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE functions (
			id SERIAL PRIMARY KEY,
			name VARCHAR NOT NULL,
			tag VARCHAR,
			hash VARCHAR NOT NULL,
			user_id INT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			FOREIGN KEY (user_id) REFERENCES users(id) 
		)`,
	)
}

// Down is executed when this migration is rolled back
func Down_20190623002235(txn *sql.Tx) {
	db.ExecQueries(txn,
		`DROP TABLE functions`,
		`DROP TABLE users`,
	)
}
