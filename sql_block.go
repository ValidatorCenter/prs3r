package main

import (
	"fmt"
	"time"

	ms "github.com/ValidatorCenter/minter-go-sdk"
	s "github.com/ValidatorCenter/prs3r/strc"

	"github.com/satori/go.uuid"

	// SQL (mail.ru)
	"github.com/mailru/dbr"
	_ "github.com/mailru/go-clickhouse"
)

// Добавить блок в SQL
func addBlockSql(db *dbr.Connection, dt *s.BlockResponse2) bool {
	var err error

	dt.UpdYCH = time.Now().Format("2006-01-02")

	sess := db.NewSession(nil)

	stmt := sess.InsertInto("blocks").Columns(
		"hash",
		"height_i32",
		"time",
		"num_txs_i32",
		"total_txs_i32",
		"transactions",
		"block_reward_f32",
		"size_i32",
		"proposer",
		"updated_date").Record(&dt)

	_, err = stmt.Exec()
	if err != nil {
		log("ERR", fmt.Sprint("[sql_block.go] addBlockSql(Exec) - ", err), "")
		return false
	}
	log("INF", "INSERT", fmt.Sprint("block ", dt.Height))

	return true
}

type OneBlockEvnt struct {
	ID              uuid.UUID `db:"_id"`
	Height          uint32    `db:"height_i32"`
	Type            string    `db:"type"`
	Role            string    `db:"role"`
	Address         string    `db:"address"`
	Amount          float32   `db:"amount_f32"`
	Coin            string    `db:"coin"`
	ValidatorPubKey string    `db:"validator_pub_key"`
	Upd             string    `db:"updated_date"`
	//Upd             time.Time `db:"updated_date"`
}

// Добавить о блоке события в SQL
func addBlcokEventSql(db *dbr.Connection, bHeight uint32, dt *ms.BlockEvResponse) {
	var err error
	sess := db.NewSession(nil)

	stmt := sess.InsertInto("block_event").Columns(
		"_id",
		"height_i32",
		"type",
		"role",
		"address",
		"amount_f32",
		"coin",
		"validator_pub_key",
		"updated_date")

	uTm := time.Now().Format("2006-01-02")

	for _, st1 := range dt.Events {
		// FIXME: реализовано по идее №2, жду когда будет №1 после исправление Даниила Лашина
		stmt.Record(OneBlockEvnt{
			ID:              uuid.Must(uuid.NewV4()), // FIXME: _id UUID - нужно отказаться см. №1
			Height:          bHeight,
			Type:            st1.Type,
			Role:            st1.Value.Role,
			Address:         st1.Value.Address,
			Amount:          st1.Value.Amount,
			Coin:            st1.Value.Coin,
			ValidatorPubKey: st1.Value.ValidatorPubKey,
			Upd:             uTm,
		})
	}

	_, err = stmt.Exec()
	if err != nil {
		log("ERR", fmt.Sprint("[sql_block.go] addBlcokEventSql(Exec) - ", err), "")
		return
	}
	log("INF", "INSERT", fmt.Sprint("block_event ", bHeight))
}

type OneBlockValid struct {
	Height uint32 `db:"height_i32"`
	PubKey string `db:"pub_key"`
	Signed bool   `db:"signed"`
	Upd    string `db:"updated_date"`
	//Upd    time.Time `db:"updated_date"`
}

// Добавить о блоке валидаторов участвующих в SQL
func addBlockValidSqlArr(db *dbr.Connection, dt *ms.BlockResponse) {
	var err error
	sess := db.NewSession(nil)
	stmt := sess.InsertInto("block_valid").Columns(
		"height_i32",
		"pub_key",
		"signed",
		"updated_date")

	uTm := time.Now().Format("2006-01-02")

	for _, oneVld := range dt.Validators {
		stmt.Record(OneBlockValid{
			Height: uint32(dt.Height),
			PubKey: oneVld.PubKey,
			Signed: oneVld.Signed,
			Upd:    uTm,
		})
	}

	_, err = stmt.Exec()
	if err != nil {
		log("ERR", fmt.Sprint("[sql_block.go] addBlockValidSqlArr(Exec) - ", err), "")
	}
}
