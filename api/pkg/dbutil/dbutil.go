package dbutil

import (
	"database/sql"
	"fmt"
	"os"

	"embly/api/pkg/config"

	_ "github.com/lib/pq" // for the db
	"github.com/pkg/errors"
)

// Connect ...
func Connect() (db *sql.DB, err error) {
	host := config.Get("DB_HOST")
	port := config.Get("DB_PORT")
	user := config.Get("DB_USERNAME")
	dbname := config.Get("DB_DATABASE")
	pass := config.Get("DB_PASSWORD")
	disablessl := (os.Getenv("DB_DISABLE_SSL") != "")
	if host == "" {
		return nil, errors.New("no db host env var set")
	}
	var sslval string
	if disablessl {
		sslval = "sslmode=disable"
	}
	dbstring := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s password=%s %s",
		host, port, user, dbname, pass, sslval,
	)
	return sql.Open("postgres", dbstring)

}
