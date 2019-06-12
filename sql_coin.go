package main

import (
	"fmt"
	"time"

	s "github.com/ValidatorCenter/prs3r/strc"

	// SQL (mail.ru)
	"github.com/mailru/dbr"
	_ "github.com/mailru/go-clickhouse"
)

// Поиск по coins в SQL
func srchCoinSql(db *dbr.Connection, symbol string) s.CoinMarketCapData {
	p := s.CoinMarketCapData{}
	sess := db.NewSession(nil)
	query := sess.Select("*").From("coins")
	query.Where(dbr.Eq("symbol", symbol))
	if _, err := query.Load(&p); err != nil {
		log("ERR", fmt.Sprint("[sql_coin.go] srchCoinSql(Load) - ", symbol, " ERR:", err), "")
		return s.CoinMarketCapData{}
	}
	return p
}

// Добавить монету в SQL
func addCoinSql(db *dbr.Connection, dt *s.CoinMarketCapData) bool {
	var err error

	dt.UpdYCH = time.Now().Format("2006-01-02")

	sess := db.NewSession(nil)

	stmt := sess.InsertInto("coins").Columns(
		"name",
		"symbol",
		"time",
		"initial_amount_f32",
		"initial_reserve_f32",
		"constant_reserve_ratio",
		"creator",
		"updated_date").Record(dt)

	_, err = stmt.Exec()
	if err != nil {
		log("ERR", fmt.Sprint("[sql_coin.go] addCoinSql(Exec) -", err), "")
		return false
	}

	log("INF", "INSERT", fmt.Sprint("coins ", dt.CoinSymbol))
	return true
}

// Добавить информацию о транзакциях монеты в SQL (ОСТАВИТЬ: т.к. есть расчетные поля, напрямую не очень удобно брать для DEX)
func addCoinTrxSql(db *dbr.Connection, dt *s.CoinActionpData) bool {
	var err error

	dt.UpdYCH = time.Now().Format("2006-01-02")

	sess := db.NewSession(nil)

	stmt := sess.InsertInto("coin_trx").Columns(
		"hash",
		"time",
		"type",
		"coin_to_buy",
		"coin_to_sell",
		"value_to_buy_f32",
		"value_to_sell_f32",
		"price_f32",
		"volume_f32",
		"updated_date").Record(dt)

	_, err = stmt.Exec()
	if err != nil {
		log("ERR", fmt.Sprint("[sql_coin.go] addCoinTrxSql(Exec) -", err), "")
		return false
	}
	log("INF", "INSERT", fmt.Sprint("coin_trx -", dt.Hash))
	return true
}

// Получить транзакции для пары монет за определенный период
func srchCoin2CoinTrxSql(db *dbr.Connection, tkS, tkB string, tmSt, tmFn time.Time) ([]s.CoinActionpData, error) {
	var err error
	coinTrxs := []s.CoinActionpData{}
	sess := db.NewSession(nil)
	query := sess.SelectBySql(fmt.Sprintf(`
		SELECT *
		FROM coin_trx
		WHERE (coin_to_sell = '%s' AND coin_to_buy = '%s') AND (time >= '%s' AND time < '%s')
		ORDER BY time ASC
	`, tkS, tkB, tmSt.Format("2006-01-02 15:04:05"), tmFn.Format("2006-01-02 15:04:05")))
	if _, err := query.Load(&coinTrxs); err != nil {
		log("ERR", fmt.Sprint("[sql_coin.go] srchCoin2CoinTrxSql(Load) - ", err), "")
	}
	return coinTrxs, err
}
