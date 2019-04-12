package main

import (
	"encoding/json"
	"fmt"
	"time"

	ms "github.com/ValidatorCenter/minter-go-sdk"
	s "github.com/ValidatorCenter/prs3r/strc"

	"github.com/jmoiron/sqlx"
	_ "github.com/kshvakov/clickhouse"
)

// Возвращает сумму в Bip которая прошла по транзакции
func returnAmntDataTx(txType int, txData interface{}) float32 {
	switch txType {
	case ms.TX_SendData: //1
		//t0.TypeTxt = "Send"
		dt0 := s.Tx1SendData{}
		jsonBytes, _ := json.Marshal(txData)
		json.Unmarshal(jsonBytes, &dt0)
		return dt0.Value
	case ms.TX_SellCoinData: //2
		//t0.TypeTxt = "SellCoin"
		dt0 := s.Tx2SellCoinData{}
		jsonBytes, _ := json.Marshal(txData)
		json.Unmarshal(jsonBytes, &dt0)
		return dt0.ValueToSell
	case ms.TX_SellAllCoinData: //3
		//t0.TypeTxt = "SellAllCoin"
		// TODO: рассчетная!
		return 0 //dt0....
	case ms.TX_BuyCoinData: //4
		//t0.TypeTxt = "BuyCoin"
		dt0 := s.Tx4BuyCoinData{}
		jsonBytes, _ := json.Marshal(txData)
		json.Unmarshal(jsonBytes, &dt0)
		return dt0.ValueToBuy
	case ms.TX_CreateCoinData: //5
		//t0.TypeTxt = "CreateCoin"
		dt0 := s.Tx5CreateCoinData{}
		jsonBytes, _ := json.Marshal(txData)
		json.Unmarshal(jsonBytes, &dt0)
		return dt0.InitialReserve
	case ms.TX_DeclareCandidacyData: //6
		//t0.TypeTxt = "DeclareCandidacy"
		dt0 := s.Tx6DeclareCandidacyData{}
		jsonBytes, _ := json.Marshal(txData)
		json.Unmarshal(jsonBytes, &dt0)
		return dt0.Stake
	case ms.TX_DelegateDate: //7
		//t0.TypeTxt = "Delegate"
		dt0 := s.Tx7DelegateDate{}
		jsonBytes, _ := json.Marshal(txData)
		json.Unmarshal(jsonBytes, &dt0)
		return dt0.Stake
	case ms.TX_UnbondData: //8
		//t0.TypeTxt = "Unbond"
		dt0 := s.Tx8UnbondData{}
		jsonBytes, _ := json.Marshal(txData)
		json.Unmarshal(jsonBytes, &dt0)
		return dt0.Value
	case ms.TX_RedeemCheckData: //9
		//t0.TypeTxt = "RedeemCheck"
		return 0 //dt0...
	case ms.TX_SetCandidateOnData: //10
		//t0.TypeTxt = "SetCandidateOn"
		return 0 //dt0...
	case ms.TX_SetCandidateOffData: //11
		//t0.TypeTxt = "SetCandidateOff"
		return 0 //dt0...
	case ms.TX_CreateMultisigData: //12
		//t0.TypeTxt = "CreateMultisig"
		return 0 //dt0...
	case ms.TX_MultisendData: //13
		//t0.TypeTxt = "MultiSend"
		dt0 := s.Tx13MultisendData{}
		jsonBytes, _ := json.Marshal(txData)
		json.Unmarshal(jsonBytes, &dt0)
		sumAmnt := float32(0.0)
		for _, itm := range dt0.List {
			// TODO: нужно учитывать что VALUE может быть монеты не BIP(MNT)
			sumAmnt += itm.Value
		}
		return sumAmnt
	case ms.TX_EditCandidateData: //14
		//t0.TypeTxt = "..EditCandidateData.."
		return 0 //dt0...
	}
	return 0

}

// Добавить транзакцию в SQL
func addTrxSql(db *sqlx.DB, dt *ms.TransResponse) bool {
	var err error

	// Транзакция в SQL: http://jmoiron.github.io/sqlx/#transactions
	tx := db.MustBegin()
	qPg_Tx := `
	INSERT INTO trx (
		hash,
		raw_tx,
		height_i32,
		index_i32,
		from_adrs,
		nonce_i32,
		gas_price_i32,
		gas_coin,
		gas_used_i32,
		type,
		amount_bip_f32,
		payload,
		tags_return,
		tags_sellamnt,
		code,
		log,
		updated_date
	) VALUES (
		:hash,
		:raw_tx,
		:height_i32,
		:index_i32,
		:from_adrs,
		:nonce_i32,
		:gas_price_i32,
		:gas_coin,
		:gas_used_i32,
		:type,
		:amount_bip_f32,
		:payload,
		:tags_return,
		:tags_sellamnt,
		:code,
		:log,
		:updated_date
	)
	`
	m1 := map[string]interface{}{
		"hash":           dt.Hash,
		"raw_tx":         dt.RawTx,
		"height_i32":     dt.Height,
		"index_i32":      dt.Index,
		"from_adrs":      dt.From,
		"nonce_i32":      dt.Nonce,
		"gas_price_i32":  dt.GasPrice,
		"gas_coin":       dt.GasCoin,
		"gas_used_i32":   dt.GasUsed,
		"type":           dt.Type,
		"amount_bip_f32": returnAmntDataTx(dt.Type, dt.Data),
		"payload":        dt.Payload,
		"tags_return":    dt.Tags.TxReturn,
		"tags_sellamnt":  dt.Tags.TxSellAmount,
		"code":           dt.Code,
		"log":            dt.Log,
		"updated_date":   time.Now(),
	}
	_, err = tx.NamedExec(qPg_Tx, m1)
	if err != nil {
		log("ERR", fmt.Sprint("[sql_trx.go] addTrxSql(NamedExec --> trx) - ", err, dt), "")
		return false
	}

	err = tx.Commit()
	if err != nil {
		log("ERR", fmt.Sprint("[sql_trx.go] addTrxSql(Commit --> trx) - ", err), "")
		return false
	}
	log("INF", "INSERT", fmt.Sprint("trx ", dt.Hash))

	return true
}

// Добавить данные транзакции в SQL
func addTrxDataSql(db *sqlx.DB, dt *ms.TransResponse) bool {
	var err error
	tx := db.MustBegin()
	qPg_TxDt := ""
	m2 := map[string]interface{}{}

	switch dt.Type {
	case ms.TX_SendData: //1
		dt0 := s.Tx1SendData{}
		jsonBytes, _ := json.Marshal(dt.Data)
		json.Unmarshal(jsonBytes, &dt0)
		qPg_TxDt = `
		INSERT INTO trx_data (
			hash,
			coin,
			to_adrs,
			value_f32,
			updated_date
		) VALUES (
			:hash,
			:coin,
			:to_adrs,
			:value_f32,
			:updated_date
		)`
		m2 = map[string]interface{}{
			"hash":         dt.Hash,
			"coin":         dt0.Coin,
			"to_adrs":      dt0.To,
			"value_f32":    dt0.Value,
			"updated_date": time.Now(),
		}
	case ms.TX_SellCoinData: //2
		dt0 := s.Tx2SellCoinData{}
		jsonBytes, _ := json.Marshal(dt.Data)
		json.Unmarshal(jsonBytes, &dt0)
		qPg_TxDt = `
		INSERT INTO trx_data (
			hash,
			coin_to_sell,
			coin_to_buy,
			value_to_sell_f32,
			updated_date
		) VALUES (
			:hash,
			:coin_to_sell,
			:coin_to_buy,
			:value_to_sell_f32,
			:updated_date
		)`
		m2 = map[string]interface{}{
			"hash":              dt.Hash,
			"coin_to_sell":      dt0.CoinToSell,
			"coin_to_buy":       dt0.CoinToBuy,
			"value_to_sell_f32": dt0.ValueToSell,
			"updated_date":      time.Now(),
		}
	case ms.TX_SellAllCoinData: //3
		dt0 := s.Tx3SellAllCoinData{}
		jsonBytes, _ := json.Marshal(dt.Data)
		json.Unmarshal(jsonBytes, &dt0)
		qPg_TxDt = `
		INSERT INTO trx_data (
			hash,
			coin_to_sell,
			coin_to_buy,
			updated_date
		) VALUES (
			:hash,
			:coin_to_sell,
			:coin_to_buy,
			:updated_date
		)`
		m2 = map[string]interface{}{
			"hash":         dt.Hash,
			"coin_to_sell": dt0.CoinToSell,
			"coin_to_buy":  dt0.CoinToBuy,
			"updated_date": time.Now(),
		}
	case ms.TX_BuyCoinData: //4
		dt0 := s.Tx4BuyCoinData{}
		jsonBytes, _ := json.Marshal(dt.Data)
		json.Unmarshal(jsonBytes, &dt0)
		qPg_TxDt = `
		INSERT INTO trx_data (
			hash,
			coin_to_sell,
			coin_to_buy,			
			value_to_buy_f32,
			updated_date
		) VALUES (
			:hash,
			:coin_to_sell,
			:coin_to_buy,
			:value_to_buy_f32,
			:updated_date
		)`
		m2 = map[string]interface{}{
			"hash":             dt.Hash,
			"coin_to_sell":     dt0.CoinToSell,
			"coin_to_buy":      dt0.CoinToBuy,
			"value_to_buy_f32": dt0.ValueToBuy,
			"updated_date":     time.Now(),
		}
	case ms.TX_CreateCoinData: //5
		dt0 := s.Tx5CreateCoinData{}
		jsonBytes, _ := json.Marshal(dt.Data)
		json.Unmarshal(jsonBytes, &dt0)
		qPg_TxDt = `
		INSERT INTO trx_data (
			hash,
			name,
			symbol,
			constant_reserve_ratio,
			initial_amount_f32,
			initial_reserve_f32,
			updated_date
		) VALUES (
			:hash,
			:name,
			:symbol,
			:constant_reserve_ratio,
			:initial_amount_f32,
			:initial_reserve_f32,
			:updated_date
		)`
		m2 = map[string]interface{}{
			"hash":                   dt.Hash,
			"name":                   dt0.Name,
			"symbol":                 dt0.CoinSymbol,
			"constant_reserve_ratio": dt0.ConstantReserveRatio,
			"initial_amount_f32":     dt0.InitialAmount,
			"initial_reserve_f32":    dt0.InitialReserve,
			"updated_date":           time.Now(),
		}
	case ms.TX_DeclareCandidacyData: //6
		dt0 := s.Tx6DeclareCandidacyData{}
		jsonBytes, _ := json.Marshal(dt.Data)
		json.Unmarshal(jsonBytes, &dt0)
		qPg_TxDt = `
		INSERT INTO trx_data (
			hash,
			coin,
			address,
			pub_key,
			commission,
			stake_f32,
			updated_date
		) VALUES (
			:hash,
			:coin,
			:address,
			:pub_key,
			:commission,
			:stake_f32,
			:updated_date
		)`
		m2 = map[string]interface{}{
			"hash":         dt.Hash,
			"coin":         dt0.Coin,
			"address":      dt0.Address,
			"pub_key":      dt0.PubKey,
			"commission":   dt0.Commission,
			"stake_f32":    dt0.Stake,
			"updated_date": time.Now(),
		}
	case ms.TX_DelegateDate: //7
		dt0 := s.Tx7DelegateDate{}
		jsonBytes, _ := json.Marshal(dt.Data)
		json.Unmarshal(jsonBytes, &dt0)
		qPg_TxDt = `
		INSERT INTO trx_data (
			hash,
			coin,
			pub_key,
			stake_f32,
			updated_date
		) VALUES (
			:hash,
			:coin,
			:pub_key,
			:stake_f32,
			:updated_date
		)`
		m2 = map[string]interface{}{
			"hash":         dt.Hash,
			"coin":         dt0.Coin,
			"pub_key":      dt0.PubKey,
			"stake_f32":    dt0.Stake,
			"updated_date": time.Now(),
		}
	case ms.TX_UnbondData: //8
		dt0 := s.Tx8UnbondData{}
		jsonBytes, _ := json.Marshal(dt.Data)
		json.Unmarshal(jsonBytes, &dt0)
		qPg_TxDt = `
		INSERT INTO trx_data (
			hash,
			coin,
			value_f32,
			pub_key,
			updated_date
		) VALUES (
			:hash,
			:coin,
			:value_f32,
			:pub_key,
			:updated_date
		)`
		m2 = map[string]interface{}{
			"hash":         dt.Hash,
			"coin":         dt0.Coin,
			"value_f32":    dt0.Value,
			"pub_key":      dt0.PubKey,
			"updated_date": time.Now(),
		}
	case ms.TX_RedeemCheckData: //9
		dt0 := s.Tx9RedeemCheckData{}
		jsonBytes, _ := json.Marshal(dt.Data)
		json.Unmarshal(jsonBytes, &dt0)
		qPg_TxDt = `
		INSERT INTO trx_data (
			hash,
			raw_check,
			proof,
			updated_date
		) VALUES (
			:hash,
			:raw_check,
			:proof,
			:updated_date
		)`
		m2 = map[string]interface{}{
			"hash":         dt.Hash,
			"raw_check":    dt0.RawCheck,
			"proof":        dt0.Proof,
			"updated_date": time.Now(),
		}
	case ms.TX_SetCandidateOnData: //10
		dt0 := s.Tx10SetCandidateOnData{}
		jsonBytes, _ := json.Marshal(dt.Data)
		json.Unmarshal(jsonBytes, &dt0)
		qPg_TxDt = `
		INSERT INTO trx_data (
			hash,
			pub_key,
			updated_date
		) VALUES (
			:hash,
			:pub_key,
			:updated_date
		)`
		m2 = map[string]interface{}{
			"hash":         dt.Hash,
			"pub_key":      dt0.PubKey,
			"updated_date": time.Now(),
		}
	case ms.TX_SetCandidateOffData: //11
		dt0 := s.Tx11SetCandidateOffData{}
		jsonBytes, _ := json.Marshal(dt.Data)
		json.Unmarshal(jsonBytes, &dt0)
		qPg_TxDt = `
		INSERT INTO trx_data (
			hash,
			pub_key,
			updated_date
		) VALUES (
			:hash,
			:pub_key,
			:updated_date
		)`
		m2 = map[string]interface{}{
			"hash":         dt.Hash,
			"pub_key":      dt0.PubKey,
			"updated_date": time.Now(),
		}
	case ms.TX_CreateMultisigData: //12
		// TODO: Реализовать
		/*dt0 := s.Tx12CreateMultisigData{}
		jsonBytes, _ := json.Marshal(dt.Data)
		json.Unmarshal(jsonBytes, &dt0)
		qPg_TxDt = ``
		m2 = map[string]interface{}{}*/
	case ms.TX_MultisendData: //13
		dt0 := s.Tx13MultisendData{}
		jsonBytes, _ := json.Marshal(dt.Data)
		json.Unmarshal(jsonBytes, &dt0)
		qPg_TxDt = `
		INSERT INTO trx_data (
			hash,
			coin_13a,
			to_13a,
			value_f32_13a,
			updated_date
		) VALUES (
			:hash,
			:coin_13a,
			:to_13a,
			:value_f32_13a,
			:updated_date
		)`

		c_a := []string{}
		t_a := []string{}
		v_a := []float32{}

		for _, dt1 := range dt0.List {
			c_a = append(c_a, dt1.Coin)
			t_a = append(t_a, dt1.To)
			v_a = append(v_a, dt1.Value)
		}

		m2 = map[string]interface{}{
			"hash":          dt.Hash,
			"coin_13a":      c_a,
			"to_13a":        t_a,
			"value_f32_13a": v_a,
			"updated_date":  time.Now(),
		}
	case ms.TX_EditCandidateData: //14
		// TODO: реализовать этап №14
		log("WRN", "[sql_trx.go] addTrxDataSql(ms.TX_EditCandidateData) - НЕ РЕАЛИЗОВАН!", "")
		return false
	default:
		log("ERR", fmt.Sprint("[sql_trx.go] addTrxDataSql(...) - неизвестный статус dt.Type - ", dt.Type), "")
		return false
	}

	_, err = tx.NamedExec(qPg_TxDt, m2)
	if err != nil {
		log("ERR", fmt.Sprint("[sql_trx.go] addTrxDataSql(NamedExec --> trx_data) - [type №", dt.Type, ", hash=", m2["hash"], "] - ", err), "")
		log("ERR", fmt.Sprint("%#v", m2), "")
		panic(err)
		return false
	}

	err = tx.Commit()
	if err != nil {
		log("ERR", fmt.Sprint("[sql_trx.go] addTrxDataSql(Commit --> trx_data) - ", err), "")
		return false
	}
	log("INF", "INSERT", fmt.Sprint("trx_data ", dt.Hash))
	return true
}
