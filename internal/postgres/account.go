package postgres

import (
	"context"
	"fmt"
	"github.com/Zzarin/transaction_system/internal"
	"github.com/jmoiron/sqlx"
	"time"
)

type Db struct {
	db *sqlx.DB
}

type Dbstructure struct {
	Id         int     `db:"id"`
	ClientName string  `db:"name"`
	Balance    float64 `db:"balance"`
}

func NewAccountSource(postgres *sqlx.DB) *Db {
	return &Db{
		db: postgres,
	}
}

func (db *Db) Select(ctx context.Context, accountID int) (*internal.StorageRespond, error) {
	ctxDbTimeout, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()
	dbRespond := &Dbstructure{}
	err := db.db.GetContext(
		ctxDbTimeout,
		dbRespond,
		"SELECT * FROM account WHERE id = $1;",
		accountID,
	)
	if dbRespond == nil {
		return nil, fmt.Errorf("find data by accountID:#%d, %w", accountID, err)
	}

	respond := toStorageRespond(dbRespond)

	return respond, nil
}

func (db *Db) Update(ctx context.Context, tasks chan internal.IncStruct, finishedTask chan bool) error {
	ctxDbTimeout, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	for task := range tasks {
		if task.TypeTXN == "withdraw" {
			_, err := db.db.ExecContext(
				ctxDbTimeout,
				"UPDATE account SET balance = balance - $1 WHERE id = $2",
				task.Amount,
				task.ClientID,
			)
			if err != nil {
				return fmt.Errorf("withdraw from client account:%d, %w", task.ClientID, err)
			}

		} else {
			_, err := db.db.ExecContext(
				ctxDbTimeout,
				"UPDATE account SET balance = balance + $1 WHERE id = $2",
				task.Amount,
				task.ClientID,
			)
			if err != nil {
				return fmt.Errorf("deposit to client account:%d, %w", task.ClientID, err)
			}
		}
		finishedTask <- true
	}
	return nil
}

func toStorageRespond(dbModel *Dbstructure) *internal.StorageRespond {
	return &internal.StorageRespond{
		Id:      fmt.Sprintf("%v", dbModel.Id),
		Name:    dbModel.ClientName,
		Balance: fmt.Sprintf("%.3f", dbModel.Balance),
	}
}
