package main

import (
	"fmt"
	"time"

	s "github.com/ValidatorCenter/prs3r/strc"

	"github.com/jmoiron/sqlx"
	_ "github.com/kshvakov/clickhouse"
)

// Поиск по coins в SQL
func srchCoinSql(db *sqlx.DB, symbol string) s.CoinMarketCapData {
	p := s.CoinMarketCapData{}
	err := db.Get(&p, fmt.Sprintf(`
		SELECT * 
		FROM coins 
		WHERE symbol = '%s'
	`, symbol))
	if err != nil {
		log("ERR", fmt.Sprint("[sql_coin.go] srchCoinSql(Select) - ", symbol, " ERR:", err), "")
		return s.CoinMarketCapData{}
	}
	return p
}

// Добавить монету в SQL
func addCoinSql(db *sqlx.DB, dt *s.CoinMarketCapData) bool {
	var err error
	tx := db.MustBegin()

	qPg := `
		INSERT INTO coins (
			name,
			symbol,
			time,
			initial_amount_f32,
			initial_reserve_f32,
			constant_reserve_ratio,
			creator,
			updated_date
		) VALUES (
			:name,
			:symbol,
			:time,
			:initial_amount_f32,
			:initial_reserve_f32,
			:constant_reserve_ratio,
			:creator,
			:updated_date
		)`

	_, err = tx.NamedExec(qPg, &dt)
	if err != nil {
		log("ERR", fmt.Sprint("[sql_coin.go] addCoinSql(NamedExec) -", err), "")
		return false
	}
	log("INF", "INSERT", fmt.Sprint("coins ", dt.CoinSymbol))

	err = tx.Commit()
	if err != nil {
		log("ERR", fmt.Sprint("[sql_coin.go] addCoinSql(Commit) -", err), "")
		return false
	}
	return true
}

// Добавить информацию о транзакциях монеты в SQL (ОСТАВИТЬ: т.к. есть расчетные поля, напрямую не очень удобно брать для DEX)
func addCoinTrxSql(db *sqlx.DB, dt *s.CoinActionpData) bool {
	var err error
	tx := db.MustBegin()

	qPg := `
		INSERT INTO coin_trx (
			hash,
			time,
			type,
			coin_to_buy,
			coin_to_sell,
			value_to_buy_f32,
			value_to_sell_f32,
			price_f32,
			volume_f32,
			updated_date
		) VALUES (
			:hash,
			:time,
			:type,
			:coin_to_buy,
			:coin_to_sell,
			:value_to_buy_f32,
			:value_to_sell_f32,
			:price_f32,
			:volume_f32,
			:updated_date
		)`

	_, err = tx.NamedExec(qPg, dt)
	if err != nil {
		log("ERR", fmt.Sprint("[sql_coin.go] addCoinTrxSql(NamedExec) - ", dt.Hash, " err:", err), "")
		panic(err) // dbg
		return false
	}
	err = tx.Commit()
	if err != nil {
		log("ERR", fmt.Sprint("[sql_coin.go] addCoinTrxSql(Commit) -", err), "")
		return false
	}
	log("INF", "INSERT", fmt.Sprint("coin_trx -", dt.Hash))

	return true
}

// Получить транзакции для пары монет за определенный период
func srchCoin2CoinTrxSql(db *sqlx.DB, tkS, tkB string, tmSt, tmFn time.Time) ([]s.CoinActionpData, error) {
	var err error
	coinTrxs := []s.CoinActionpData{}

	err = db.Select(&coinTrxs, fmt.Sprintf(`
		SELECT *
		FROM coin_trx
		WHERE (coin_to_sell = '%s' AND coin_to_buy = '%s') AND (time >= '%s' AND time < '%s')
		ORDER BY time ASC
	`, tkS, tkB, tmSt.Format("2006-01-02 15:04:05"), tmFn.Format("2006-01-02 15:04:05")))

	/*if err != nil {
		return err
	}*/

	return coinTrxs, err
}
