package main

import (
	"fmt"
	"time"

	s "github.com/ValidatorCenter/prs3r/strc"

	// SQL (mail.ru)
	"github.com/mailru/dbr"
	_ "github.com/mailru/go-clickhouse"
)

// Добавить ноду в SQL
func addNodeSql(db *dbr.Connection, dt *s.NodeExt) bool {
	var err error

	sess := db.NewSession(nil)
	stmt := sess.InsertInto("nodes").Columns(
		"pub_key",
		"pub_key_min",
		"reward_address",
		"owner_address",
		"created",
		"commission",
		"created_at_block",
		"updated_date").Record(&dt)

	_, err = stmt.Exec()
	if err != nil {
		log("ERR", fmt.Sprint("[sql_node.go] addNodeSql(Exec) -", err), "")
		panic(err)
		return false
	}

	return true
}

type OneBlockStory struct {
	BlockID uint32 `db:"block_id"`
	PubKey  string `db:"pub_key"`
	Type    string `db:"block_type"`
}

// Добавить о ноде историю блоков в SQL
func addNodeBlockstorySql(db *dbr.Connection, dt *s.NodeExt) bool {
	var err error
	chNAr := srchNodeBlockstory(db, dt.PubKey) // Для проверки на задвоенность блока для ноды
	sess := db.NewSession(nil)
	stmt := sess.InsertInto("node_blockstory").Columns(
		"pub_key",
		"block_id",
		"block_type")

	amntAdd := 0
	for _, bs1 := range dt.Blocks {

		//+Проверка
		fndBS := false
		for _, tst1 := range chNAr {
			if tst1.ID == bs1.ID && tst1.Type == bs1.Type {
				fndBS = true
			}
		}

		if fndBS {
			log("WRN", fmt.Sprint("[sql_node.go] addNodeBlockstorySql(Find!) - [", dt.PubKey, " ", bs1.ID, " ", bs1.Type, "]"), "")
			continue
		}
		//-

		stmt.Record(OneBlockStory{
			BlockID: bs1.ID,
			PubKey:  dt.PubKey,
			Type:    bs1.Type,
		})
		amntAdd++
	}

	if amntAdd > 0 {
		_, err = stmt.Exec()
		if err != nil {
			log("ERR", fmt.Sprint("[sql_node.go] addNodeBlockstorySql(Exec) -", err), "")
			return false
		}
		log("INF", "INSERT", fmt.Sprint("node_blockstory ", dt.PubKey))
	} else {
		log("WRN", fmt.Sprint("[sql_node.go] addNodeBlockstorySql() - node_blockstory amount new ADD=0"), "")
	}

	return true
}

type OneNodeStake struct {
	PubKey   string    `db:"pub_key"`
	Owner    string    `db:"owner_address"`
	Coin     string    `db:"coin"`
	Value    float32   `db:"value_f32"`
	ValueBip float32   `db:"bip_value_f32"`
	Upd      time.Time `db:"updated_date"`
}

// Добавить о ноде стэк в SQL
func addNodeStakeSql(db *dbr.Connection, dt *s.NodeExt) bool {
	var err error
	sess := db.NewSession(nil)
	stmt := sess.InsertInto("node_stakes").Columns(
		"pub_key",
		"owner_address",
		"coin",
		"value_f32",
		"bip_value_f32",
		"updated_date")

	for _, st1 := range dt.Stakes {
		stmt.Record(OneNodeStake{
			PubKey:   dt.PubKey,
			Owner:    st1.Owner,
			Coin:     st1.Coin,
			Value:    st1.Value,
			ValueBip: st1.BipValue,
			Upd:      time.Now(),
		})
	}

	_, err = stmt.Exec()
	if err != nil {
		log("ERR", fmt.Sprint("[sql_node.go] addNodeStakeSql(Exec) -", err), "")
		panic(err)   // dbg
		return false //continue
	}

	log("INF", "INSERT", fmt.Sprint("node_stakes ", dt.PubKey))
	return true
}

// Поиск ноды по публичному ключу и основному адресу кошелька в SQL
func srchNodeSql_oa(db *dbr.Connection, pub_key string, owner_address string) s.NodeExt {
	p := s.NodeExt{}
	sess := db.NewSession(nil)
	query := sess.SelectBySql(fmt.Sprintf(`
		SELECT * 
		FROM nodes FINAL 
		WHERE pub_key = '%s' AND owner_address = '%s'
		LIMIT 1
		`, pub_key, owner_address))
	if _, err := query.Load(&p); err != nil {
		if err.Error() == "sql: no rows in result set" {
			log("WRN", fmt.Sprint("[sql_node.go] srchNodeSql_oa(Load) - ", err), "")
		} else {
			log("ERR", fmt.Sprint("[sql_node.go] srchNodeSql_oa(Load) - ", err), "")
		}
		return s.NodeExt{}
	}
	return p
}

// Поиск ноды по публичному ключу в SQL
func srchNodeSql_pk(db *dbr.Connection, pub_key string) s.NodeExt {
	p := s.NodeExt{}
	sess := db.NewSession(nil)
	query := sess.SelectBySql(fmt.Sprintf(`
		SELECT * 
		FROM nodes FINAL 
		WHERE pub_key = '%s' 
		LIMIT 1
		`, pub_key))
	if _, err := query.Load(&p); err != nil {
		if err.Error() == "sql: no rows in result set" {
			log("WRN", fmt.Sprint("[sql_node.go] srchNodeSql_pk(Load) - ", err), "")
		} else {
			log("ERR", fmt.Sprint("[sql_node.go] srchNodeSql_pk(Load) - ", err), "")
			panic(err)
		}
		return s.NodeExt{}
	}
	return p
}

// Берем все ноды из SQL
func srchNodeSql_all(db *dbr.Connection) []s.NodeExt {
	pp := []s.NodeExt{}
	sess := db.NewSession(nil)
	query := sess.SelectBySql("SELECT * FROM nodes FINAL")
	//query := sess.Select("*").From("nodes")
	if _, err := query.Load(&pp); err != nil {
		log("ERR", fmt.Sprint("[sql_node.go] srchCoinSql_all(Load) -", err), "")
		return []s.NodeExt{}
	}
	return pp
}

//Поиск блоков-истории ноды по паблику
func srchNodeBlockstory(db *dbr.Connection, pub_key string) []s.BlocksStory {
	v := []s.BlocksStory{}
	sess := db.NewSession(nil)
	query := sess.SelectBySql(fmt.Sprintf(`
		SELECT block_id, block_type
		FROM node_blockstory
		WHERE pub_key = '%s'
	`, pub_key))
	if _, err := query.Load(&v); err != nil {
		log("ERR", fmt.Sprint("[sql_node.go] srchNodeBlockstory(Load) - [pub_key ", pub_key, "] ERR:", err), "")
		panic(err) //dbg
		return v
	}
	return v
}
