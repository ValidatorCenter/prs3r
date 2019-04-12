package main

import (
	"fmt"
	"time"

	ms "github.com/ValidatorCenter/minter-go-sdk"
	s "github.com/ValidatorCenter/prs3r/strc"

	"github.com/satori/go.uuid"

	"github.com/jmoiron/sqlx"
	_ "github.com/kshvakov/clickhouse"
)

// Добавить блок в SQL
func addBlockSql(db *sqlx.DB, dt *s.BlockResponse2) bool {
	var err error

	dt.UpdYCH = time.Now()

	// Транзакция в SQL: http://jmoiron.github.io/sqlx/#transactions
	tx := db.MustBegin()
	qPg := `
		INSERT INTO blocks (
			hash,
			height_i32,
			time,
			num_txs_i32,
			total_txs_i32,
			transactions,
			block_reward_f32,
			size_i32,
			proposer,
			updated_date
		) VALUES (
			:hash,
			:height_i32,
			:time,
			:num_txs_i32,
			:total_txs_i32,
			:transactions,
			:block_reward_f32,
			:size_i32,
			:proposer,
			:updated_date
		)`
	_, err = tx.NamedExec(qPg, &dt)
	if err != nil {
		log("ERR", fmt.Sprint("[sql_block.go] addBlockSql(NamedExec) - ", err), "")
		return false
	}

	err = tx.Commit()
	if err != nil {
		log("ERR", fmt.Sprint("[sql_block.go] addBlockSql(Commit) - ", err), "")
		return false
	}
	log("INF", "INSERT", fmt.Sprint("block ", dt.Height))

	return true
}

// Добавить о блоке события в SQL
func addBlcokEventSql(db *sqlx.DB, bHeight uint32, dt *ms.BlockEvResponse) {
	for _, st1 := range dt.Events {
		tx := db.MustBegin()
		// FIXME: реализовано по идее №2, жду когда будет №1 после исправление Даниила Лашина
		qPg := `
		INSERT INTO block_event (
			_id,
			height_i32,
			type,
			role,
			address,
			amount_f32,
			coin,
			validator_pub_key,
			updated_date
		) VALUES (
			:_id,
			:height_i32,
			:type,
			:role,
			:address,
			:amount_f32,
			:coin,
			:validator_pub_key,
			:updated_date
		)`

		m2 := map[string]interface{}{
			//"_id": uuid.New(), // FIXME: _id UUID - нужно отказаться см. №1
			"_id":               uuid.Must(uuid.NewV4()), // FIXME: _id UUID - нужно отказаться см. №1
			"height_i32":        bHeight,
			"type":              st1.Type,
			"role":              st1.Value.Role,
			"address":           st1.Value.Address,
			"amount_f32":        st1.Value.Amount,
			"coin":              st1.Value.Coin,
			"validator_pub_key": st1.Value.ValidatorPubKey,
			"updated_date":      time.Now(),
		}

		_, err := tx.NamedExec(qPg, m2)
		if err != nil {
			log("ERR", fmt.Sprint("[sql_block.go] addBlcokEventSql(NamedExec) - [block №", bHeight, "] ", err), "")
			log("ERR", fmt.Sprint("%#v", st1), "")
			panic(err) // dbg
			continue   //был	break
		}

		err = tx.Commit()
		if err != nil {
			log("ERR", fmt.Sprint("[sql_block.go] addBlcokEventSql(Commit) - ", err), "")
			continue
		}
	}
	log("INF", "INSERT", fmt.Sprint("block_event ", bHeight))
}

// Добавить о блоке валидаторов участвующих в SQL
func addBlockValidSql(db *sqlx.DB, bHeight uint32, dt *ms.BlockValidatorsResponse) {
	tx := db.MustBegin()
	qPg := `
		INSERT INTO block_valid (
			height_i32,
			pub_key,
			signed,
			updated_date
		) VALUES (
			:height_i32,
			:pub_key,
			:signed,
			:updated_date
		)`

	m2 := map[string]interface{}{
		"height_i32":   bHeight,
		"pub_key":      dt.PubKey,
		"signed":       dt.Signed,
		"updated_date": time.Now(),
	}

	_, err := tx.NamedExec(qPg, m2)
	if err != nil {
		log("ERR", fmt.Sprint("[sql_block.go] addBlockValidSql(NamedExec) - [block №", bHeight, "] ", err), "")
		log("ERR", fmt.Sprint("%#v", m2), "")
		panic(err)
		return
	}
	err = tx.Commit()
	if err != nil {
		log("ERR", fmt.Sprint("[sql_block.go] addBlockValidSql(Commit) - ", err), "")
		return
	}

	log("INF", "INSERT", fmt.Sprint("block_valid ", bHeight, "-", dt.PubKey))
}
