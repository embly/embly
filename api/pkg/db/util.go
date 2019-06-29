package db

import (
	"context"
	"database/sql"
	"log"
)

var ctx = context.Background()

// ExecQueries is used to execute migrations with a log.Fatal error state with rollbacks.
// can be used to easily generate migrations files with an array of strings
func ExecQueries(txn *sql.Tx, qs ...string) {
	for _, toExec := range qs {
		_, err := txn.ExecContext(ctx, toExec)
		if err != nil {
			if rollbackErr := txn.Rollback(); rollbackErr != nil {
				log.Fatalf("update drivers: unable to rollback: %v", rollbackErr)
			}
			log.Fatal(toExec, err)
		}
	}
}
