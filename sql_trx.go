package main

import (
	"encoding/json"
	"fmt"
	"time"

	ms "github.com/ValidatorCenter/minter-go-sdk"
	s "github.com/ValidatorCenter/prs3r/strc"

	// SQL (mail.ru)
	"github.com/mailru/dbr"
	_ "github.com/mailru/go-clickhouse"
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

type TrxEx struct {
	ms.TransResponse
	FromAdrs     string  `json:"from_adrs" db:"from_adrs"`
	TxReturn     float32 `json:"tags_return" db:"tags_return"`
	TxSellAmount float32 `json:"tags_sellamnt" db:"tags_sellamnt"`
	AmountBip    float32 `json:"amount_bip_f32" db:"amount_bip_f32"`
	Upd          string  `json:"updated_date" db:"updated_date"`
	//Upd          time.Time `json:"updated_date" db:"updated_date"`
}

// Добавить транзакции блока в SQL
func addTrxSql(db *dbr.Connection, dt *ms.BlockResponse) bool {
	var err error

	if len(dt.Transactions) == 0 {
		log("INF", "INSERT [0]!", fmt.Sprint("trx ", dt.Hash))
		return true
	}

	sess := db.NewSession(nil)

	stmt := sess.InsertInto("trx").Columns(
		"hash",
		"raw_tx",
		"height_i32",
		"index_i32",
		"from_adrs",
		"nonce_i32",
		"gas_price_i32",
		"gas_coin",
		"gas_used_i32",
		"type",
		"amount_bip_f32",
		"payload",
		"tags_return",
		"tags_sellamnt",
		"code",
		"log",
		"updated_date")

	for _, oneTrx := range dt.Transactions {
		o1 := TrxEx{}
		o1.Hash = oneTrx.Hash
		o1.RawTx = oneTrx.RawTx
		o1.Height = oneTrx.Height
		o1.Index = oneTrx.Index
		o1.FromAdrs = oneTrx.From
		o1.Nonce = oneTrx.Nonce
		o1.GasPrice = oneTrx.GasPrice
		o1.GasCoin = oneTrx.GasCoin
		o1.GasUsed = oneTrx.GasUsed
		o1.Type = oneTrx.Type
		o1.Payload = oneTrx.Payload
		o1.TxReturn = oneTrx.Tags.TxReturn
		o1.TxSellAmount = oneTrx.Tags.TxSellAmount
		o1.Code = oneTrx.Code
		o1.Log = oneTrx.Log
		o1.AmountBip = returnAmntDataTx(oneTrx.Type, oneTrx.Data)
		o1.Upd = time.Now().Format("2006-01-02")

		stmt.Record(o1)
	}

	_, err = stmt.Exec()
	if err != nil {
		log("ERR", fmt.Sprint("[sql_trx.go] addTrxSql(Exec) - ", err), "")
		panic(err)
		return false
	}

	log("INF", "INSERT", fmt.Sprint("trx ", dt.Hash))
	return true
}

type OneTrxData struct {
	Hash        string    `db:"hash"`
	Coin        string    `db:"coin"`
	ToAddress   string    `db:"to_adrs"`
	Value       float32   `db:"value_f32"`
	CoinToSell  string    `db:"coin_to_sell"`
	CoinToBuy   string    `db:"coin_to_buy"`
	ValueToSell float32   `db:"value_to_sell_f32"`
	ValueToBuy  float32   `db:"value_to_buy_f32"`
	Name        string    `db:"name"`
	Symbol      string    `db:"symbol"`
	CRR         uint32    `db:"constant_reserve_ratio"`
	InitAmount  float32   `db:"initial_amount_f32"`
	InitReserve float32   `db:"initial_reserve_f32"`
	Address     string    `db:"address"`
	PubKey      string    `db:"pub_key"`
	Commission  uint32    `db:"commission"`
	Stake       float32   `db:"stake_f32"`
	RawCheck    string    `db:"raw_check"`
	Proof       string    `db:"proof"`
	CoinArr     []string  `db:"coin_13a"`
	ToArr       []string  `db:"to_13a"`
	ValueArr    []float32 `db:"value_f32_13a"`
	Upd         string    `db:"updated_date"`
	//Upd         time.Time `db:"updated_date"`
}

// Добавить данные транзакций в SQL
func addTrxDataSql(db *dbr.Connection, dtSlc *ms.BlockResponse) bool {
	var err error

	if len(dtSlc.Transactions) == 0 {
		log("INF", "INSERT [0]!", fmt.Sprint("trx_data ", dtSlc.Hash))
		return true
	}

	sess := db.NewSession(nil)

	stmt := sess.InsertInto("trx_data").Columns(
		"hash",
		"coin",
		"to_adrs",
		"value_f32",
		"coin_to_sell",
		"coin_to_buy",
		"value_to_sell_f32",
		"value_to_buy_f32",
		"name",
		"symbol",
		"constant_reserve_ratio",
		"initial_amount_f32",
		"initial_reserve_f32",
		"address",
		"pub_key",
		"commission",
		"stake_f32",
		"raw_check",
		"proof",
		"coin_13a",
		"to_13a",
		"value_f32_13a",
		"updated_date")

	cntrlAmntTx := 0
	for _, dt := range dtSlc.Transactions {
		oneTrxDt := OneTrxData{}

		oneTrxDt.Upd = time.Now().Format("2006-01-02")

		switch dt.Type {
		case ms.TX_SendData: //1
			dt0 := s.Tx1SendData{}
			jsonBytes, _ := json.Marshal(dt.Data)
			json.Unmarshal(jsonBytes, &dt0)

			oneTrxDt.Hash = dt.Hash
			oneTrxDt.Coin = dt0.Coin
			oneTrxDt.ToAddress = dt0.To
			oneTrxDt.Value = dt0.Value

			cntrlAmntTx++

		case ms.TX_SellCoinData: //2
			dt0 := s.Tx2SellCoinData{}
			jsonBytes, _ := json.Marshal(dt.Data)
			json.Unmarshal(jsonBytes, &dt0)

			oneTrxDt.Hash = dt.Hash
			oneTrxDt.CoinToSell = dt0.CoinToSell
			oneTrxDt.CoinToBuy = dt0.CoinToBuy
			oneTrxDt.ValueToSell = dt0.ValueToSell

			cntrlAmntTx++

		case ms.TX_SellAllCoinData: //3
			dt0 := s.Tx3SellAllCoinData{}
			jsonBytes, _ := json.Marshal(dt.Data)
			json.Unmarshal(jsonBytes, &dt0)

			oneTrxDt.Hash = dt.Hash
			oneTrxDt.CoinToSell = dt0.CoinToSell
			oneTrxDt.CoinToBuy = dt0.CoinToBuy

			cntrlAmntTx++

		case ms.TX_BuyCoinData: //4
			dt0 := s.Tx4BuyCoinData{}
			jsonBytes, _ := json.Marshal(dt.Data)
			json.Unmarshal(jsonBytes, &dt0)

			oneTrxDt.Hash = dt.Hash
			oneTrxDt.CoinToSell = dt0.CoinToSell
			oneTrxDt.CoinToBuy = dt0.CoinToBuy
			oneTrxDt.ValueToBuy = dt0.ValueToBuy

			cntrlAmntTx++

		case ms.TX_CreateCoinData: //5
			dt0 := s.Tx5CreateCoinData{}
			jsonBytes, _ := json.Marshal(dt.Data)
			json.Unmarshal(jsonBytes, &dt0)

			oneTrxDt.Hash = dt.Hash
			oneTrxDt.Name = dt0.Name
			oneTrxDt.Symbol = dt0.CoinSymbol
			oneTrxDt.CRR = uint32(dt0.ConstantReserveRatio)
			oneTrxDt.InitAmount = dt0.InitialAmount
			oneTrxDt.InitReserve = dt0.InitialReserve

			cntrlAmntTx++

		case ms.TX_DeclareCandidacyData: //6
			dt0 := s.Tx6DeclareCandidacyData{}
			jsonBytes, _ := json.Marshal(dt.Data)
			json.Unmarshal(jsonBytes, &dt0)

			oneTrxDt.Hash = dt.Hash
			oneTrxDt.Coin = dt0.Coin
			oneTrxDt.Address = dt0.Address
			oneTrxDt.PubKey = dt0.PubKey
			oneTrxDt.Commission = uint32(dt0.Commission)
			oneTrxDt.Stake = dt0.Stake

			cntrlAmntTx++

		case ms.TX_DelegateDate: //7
			dt0 := s.Tx7DelegateDate{}
			jsonBytes, _ := json.Marshal(dt.Data)
			json.Unmarshal(jsonBytes, &dt0)

			oneTrxDt.Hash = dt.Hash
			oneTrxDt.Coin = dt0.Coin
			oneTrxDt.PubKey = dt0.PubKey
			oneTrxDt.Stake = dt0.Stake

			cntrlAmntTx++

		case ms.TX_UnbondData: //8
			dt0 := s.Tx8UnbondData{}
			jsonBytes, _ := json.Marshal(dt.Data)
			json.Unmarshal(jsonBytes, &dt0)

			oneTrxDt.Hash = dt.Hash
			oneTrxDt.Coin = dt0.Coin
			oneTrxDt.Value = dt0.Value
			oneTrxDt.PubKey = dt0.PubKey

			cntrlAmntTx++

		case ms.TX_RedeemCheckData: //9
			dt0 := s.Tx9RedeemCheckData{}
			jsonBytes, _ := json.Marshal(dt.Data)
			json.Unmarshal(jsonBytes, &dt0)

			oneTrxDt.Hash = dt.Hash
			oneTrxDt.RawCheck = dt0.RawCheck
			oneTrxDt.Proof = dt0.Proof

			cntrlAmntTx++

		case ms.TX_SetCandidateOnData: //10
			dt0 := s.Tx10SetCandidateOnData{}
			jsonBytes, _ := json.Marshal(dt.Data)
			json.Unmarshal(jsonBytes, &dt0)

			oneTrxDt.Hash = dt.Hash
			oneTrxDt.PubKey = dt0.PubKey

			cntrlAmntTx++

		case ms.TX_SetCandidateOffData: //11
			dt0 := s.Tx11SetCandidateOffData{}
			jsonBytes, _ := json.Marshal(dt.Data)
			json.Unmarshal(jsonBytes, &dt0)

			oneTrxDt.Hash = dt.Hash
			oneTrxDt.PubKey = dt0.PubKey

			cntrlAmntTx++

		case ms.TX_CreateMultisigData: //12
			// TODO: Реализовать
			/*dt0 := s.Tx12CreateMultisigData{}
			jsonBytes, _ := json.Marshal(dt.Data)
			json.Unmarshal(jsonBytes, &dt0)
			qPg_TxDt = ``
			m2 = map[string]interface{}{}*/
			log("WRN", "[sql_trx.go] addTrxDataSql(ms.TX_CreateMultisigData) - НЕ РЕАЛИЗОВАН!", "")
			continue
		case ms.TX_MultisendData: //13
			dt0 := s.Tx13MultisendData{}
			jsonBytes, _ := json.Marshal(dt.Data)
			json.Unmarshal(jsonBytes, &dt0)

			c_a := []string{}
			t_a := []string{}
			v_a := []float32{}

			for _, dt1 := range dt0.List {
				c_a = append(c_a, dt1.Coin)
				t_a = append(t_a, dt1.To)
				v_a = append(v_a, dt1.Value)
			}

			oneTrxDt.Hash = dt.Hash
			oneTrxDt.CoinArr = c_a
			oneTrxDt.ToArr = t_a
			oneTrxDt.ValueArr = v_a

			cntrlAmntTx++

		case ms.TX_EditCandidateData: //14
			// TODO: реализовать этап №14
			log("WRN", "[sql_trx.go] addTrxDataSql(ms.TX_EditCandidateData) - НЕ РЕАЛИЗОВАН!", "")
			continue
		default:
			log("ERR", fmt.Sprint("[sql_trx.go] addTrxDataSql(...) - неизвестный статус dt.Type - ", dt.Type), "")
			continue
		}

		stmt.Record(oneTrxDt)

	}

	if cntrlAmntTx > 0 {

		_, err = stmt.Exec()
		if err != nil {
			log("ERR", fmt.Sprint("[sql_trx.go] addTrxDataSql(Exec) - ", err), "")
			return false
		}
		log("INF", "INSERT", fmt.Sprint("trx_data ", dtSlc.Hash))
		return true
	} else {
		log("INF", "INSERT [0]!!", fmt.Sprint("trx_data ", dtSlc.Hash))
		return true
	}
}
