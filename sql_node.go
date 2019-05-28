package main

import (
	"fmt"
	"time"

	s "github.com/ValidatorCenter/prs3r/strc"

	"github.com/jmoiron/sqlx"
	_ "github.com/kshvakov/clickhouse"
)

// Добавить ноду в SQL
func addNodeSql(db *sqlx.DB, dt *s.NodeExt) bool {
	var err error
	tx := db.MustBegin()

	dt.UpdYCH = time.Now()
	qPg := `
	INSERT INTO nodes (
		pub_key,
		pub_key_min,
		reward_address,
		owner_address,
		created,
		commission,
		created_at_block,
		updated_date
	) VALUES (
		:pub_key,
		:pub_key_min,
		:reward_address,
		:owner_address,
		:created,
		:commission,
		:created_at_block,
		:updated_date
	);
	`
	_, err = tx.NamedExec(qPg, &dt)
	if err != nil {
		log("ERR", fmt.Sprint("[sql_node.go] addNodeSql(NamedExec) - ", err), "")
		panic(err)
		return false
	}
	log("INF", "INSERT", fmt.Sprint("node ", dt.PubKey))

	err = tx.Commit()
	if err != nil {
		log("ERR", fmt.Sprint("[sql_node.go] addNodeSql(Commit) -", err), "")
		panic(err)
		return false
	}

	return true
}

// Добавить о ноде историю блоков в SQL
func addNodeBlockstorySql(db *sqlx.DB, dt *s.NodeExt) bool {
	chNAr := srchNodeBlockstory(db, dt.PubKey) // Для проверки на задвоенность блока для ноды

	for _, bs1 := range dt.Blocks {

		//+Проверка
		fndBS := false
		for _, tst1 := range chNAr {
			if tst1.ID == bs1.ID && tst1.Type == bs1.Type {
				fndBS = true
			}
		}

		if fndBS {
			log("ERR", fmt.Sprint("[sql_node.go] addNodeBlockstorySql(Find!) - [", dt.PubKey, " ", bs1.ID, " ", bs1.Type, "]"), "")
			//panic("!!!")
			continue
		}
		//-

		tx := db.MustBegin()
		qPg := `
		INSERT INTO node_blockstory (
			pub_key,
			block_id,
			block_type
		) VALUES (
			:pub_key,
			:block_id,
			:block_type
		)`

		m2 := map[string]interface{}{
			"pub_key":    dt.PubKey,
			"block_id":   bs1.ID,
			"block_type": bs1.Type,
		}

		_, err := tx.NamedExec(qPg, m2)
		if err != nil {
			//Возможно добавили раньше, поэтому не сильно надо обращать на эту ошибку, но за неё может не пройти другие транзакции
			log("WRN", fmt.Sprint("[sql_node.go] addNodeBlockstorySql(NamedExec) - [", dt.PubKey, " ", bs1.ID, " ", bs1.Type, "] ", err), "")
			continue
		}
		err = tx.Commit()
		if err != nil {
			log("ERR", fmt.Sprint("[sql_node.go] addNodeBlockstorySql(Commit) -", err), "")
			return false
		}
		log("INF", "INSERT", fmt.Sprint("node_blockstory ", dt.PubKey))
	}

	return true
}

// Добавить о ноде стэк в SQL
func addNodeStakeSql(db *sqlx.DB, dt *s.NodeExt) bool {
	for _, st1 := range dt.Stakes {
		tx := db.MustBegin()
		// Через tx.MustExec и "?" не сработало, выдаёт ошибку
		qPg := `
		INSERT INTO node_stakes (
			pub_key,
			owner_address,
			coin,
			value_f32,
			bip_value_f32,
			updated_date
		) VALUES (
			:pub_key,
			:owner_address,
			:coin,
			:value_f32,
			:bip_value_f32,
			:updated_date
		)`

		m2 := map[string]interface{}{
			"pub_key":       dt.PubKey,
			"owner_address": st1.Owner,
			"coin":          st1.Coin,
			"value_f32":     st1.Value,
			"bip_value_f32": st1.BipValue,
			"updated_date":  time.Now(),
		}

		_, err := tx.NamedExec(qPg, m2)
		if err != nil {
			log("ERR", fmt.Sprint("[sql_node.go] addNodeStakeSql(NamedExec) - [", dt.PubKey, "] ", err), "")
			panic(err) // dbg
			return false
		}

		err = tx.Commit()
		if err != nil {
			log("ERR", fmt.Sprint("[sql_node.go] addNodeStakeSql(Commit) -", err), "")
			panic(err)   // dbg
			return false //continue
		}
	}
	log("INF", "INSERT", fmt.Sprint("node_stakes ", dt.PubKey))
	return true
}

// Поиск ноды по публичному ключу и основному адресу кошелька в SQL
func srchNodeSql_oa(db *sqlx.DB, pub_key string, owner_address string) s.NodeExt {
	p := s.NodeExt{}
	err := db.Get(&p, fmt.Sprintf(`
		SELECT * 
		FROM nodes FINAL 
		WHERE pub_key = '%s' AND owner_address = '%s'
		LIMIT 1
		`, pub_key, owner_address))
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			log("WRN", fmt.Sprint("[sql_node.go] srchNodeSql_oa(Select) - ", err), "")
		} else {
			log("ERR", fmt.Sprint("[sql_node.go] srchNodeSql_oa(Select) - ", err), "")
		}
		return s.NodeExt{}
	}
	return p
}

// Поиск ноды по публичному ключу в SQL
func srchNodeSql_pk(db *sqlx.DB, pub_key string) s.NodeExt {
	p := s.NodeExt{}
	err := db.Get(&p, fmt.Sprintf(`
		SELECT * 
		FROM nodes FINAL 
		WHERE pub_key = '%s' 
		LIMIT 1
		`, pub_key))
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			log("WRN", fmt.Sprint("[sql_node.go] srchNodeSql_pk(Select) - ", err), "")
		} else {
			log("ERR", fmt.Sprint("[sql_node.go] srchNodeSql_pk(Select) - ", err), "")
			panic(err)
		}
		return s.NodeExt{}
	}
	return p
}

// Берем всех пользователей с новым % из SQL
func srchNodeUserXSql(db *sqlx.DB) []s.NodeUserX {
	// FIXME: надо как-то учитывать что период возврата Пользователю, может
	// быть уже закончился!
	pp := []s.NodeUserX{}
	err := db.Select(&pp, "SELECT * FROM node_userx")
	if err != nil {
		log("ERR", fmt.Sprint("[sql_node.go] srchNodeUserXSql(Select) - ", err), "")
		return []s.NodeUserX{}
	}
	return pp
}

// Берем все ноды из SQL
func srchNodeSql_all(db *sqlx.DB) []s.NodeExt {
	pp := []s.NodeExt{}
	err := db.Select(&pp, "SELECT * FROM nodes")
	if err != nil {
		log("ERR", fmt.Sprint("[sql_node.go] srchCoinSql_all(Select) -", err), "")
		return []s.NodeExt{}
	}
	return pp
}

// Добавить задачу для ноды в SQL
func addNodeTaskSql(db *sqlx.DB, dt *s.NodeTodo) bool {
	var err error
	tx := db.MustBegin()

	dt.UpdYCH = time.Now()

	qPg := `
		INSERT INTO node_tasks (
			_id,
			priority,
			done,
			created,
			donet,
			type,
			height_i32,
			pub_key,
			address,
			amount_f32,
			comment,
			tx_hash,
			updated_date
		) VALUES (
			:_id,
			:priority,
			:done,
			:created,
			:donet,
			:type,
			:height_i32,
			:pub_key,
			:address,
			:amount_f32,
			:comment,
			:tx_hash,
			:updated_date
		)`

	_, err = tx.NamedExec(qPg, &dt)
	if err != nil {
		log("ERR", fmt.Sprint("[sql_node.go] addNodeTaskSql(NamedExec) -", err), "")
		panic(err)
		return false
	}
	log("INF", "INSERT", fmt.Sprint("node-task ", dt.Address, " ", dt.PubKey))

	err = tx.Commit()
	if err != nil {
		log("ERR", fmt.Sprint("[sql_node.go] addNodeTaskSql(Commit() -", err), "")
		return false
	}
	return true
}

//Поиск блоков-истории ноды по паблику
func srchNodeBlockstory(db *sqlx.DB, pub_key string) []s.BlocksStory {
	v := []s.BlocksStory{}
	err := db.Select(&v, fmt.Sprintf(`
		SELECT block_id, block_type
		FROM node_blockstory
		WHERE pub_key = '%s'
	`, pub_key))
	if err != nil {
		log("ERR", fmt.Sprint("[sql_node.go] srchNodeBlockstory(Select) - [pub_key ", pub_key, "] ERR:", err), "")
		panic(err) //dbg
		return v
	}
	return v
}

// Добавляем условие акции для кошелька(пользователя)
func addNodeUserX(db *sqlx.DB, dt *s.NodeUserX) bool {
	var err error
	tx := db.MustBegin()

	dt.UpdYCH = time.Now()

	qPg := `
		INSERT INTO node_userx (
			pub_key,
			address,
			start,
			finish,
			commission,
			updated_date
		) VALUES (
			:pub_key,
			:address,
			:start,
			:finish,
			:commission,
			:updated_date
		)`

	_, err = tx.NamedExec(qPg, &dt)
	if err != nil {
		log("ERR", fmt.Sprint("[sql_node.go] addNodeUserX(NamedExec) -", err), "")
		return false
	}
	log("INF", "INSERT", fmt.Sprint("node-userx ", dt.Address, " ", dt.PubKey))

	err = tx.Commit()
	if err != nil {
		log("ERR", fmt.Sprint("[sql_node.go] addNodeUserX(Commit) -", err), "")
		return false
	}
	return true
}
