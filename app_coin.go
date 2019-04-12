package main

import (
	"encoding/json"
	"fmt"
	"time"

	ms "github.com/ValidatorCenter/minter-go-sdk"
	s "github.com/ValidatorCenter/prs3r/strc"
)

// Обработка транзакции создания Монеты
func trxCreateCoin(retTrns *TrxExt) {
	dt5CrtCoin := s.Tx5CreateCoinData{}
	jsonBytes, _ := json.Marshal(retTrns.Data)
	json.Unmarshal(jsonBytes, &dt5CrtCoin)

	coinMCD := s.CoinMarketCapData{
		Name:                 dt5CrtCoin.Name,
		CoinSymbol:           dt5CrtCoin.CoinSymbol,
		Time:                 retTrns.Created,
		Creator:              retTrns.From,
		InitialAmount:        dt5CrtCoin.InitialAmount,
		InitialReserve:       dt5CrtCoin.InitialReserve,
		ConstantReserveRatio: dt5CrtCoin.ConstantReserveRatio,
	}

	// добавляем монету в SQL
	if !addCoinSql(dbSQL, &coinMCD) {
		log("ERR", "[app_coin.go] startWorkerTrx(addCoinSql)", "")
	}
	// добавляем монету в Redis
	if !updCoinInfoRds(dbSys, &coinMCD) {
		log("ERR", "[app_coin.go] startWorkerTrx(updCoinInfoRds)", "")
	}
	log("OK", fmt.Sprintf("Монета: %s (%s)", dt5CrtCoin.CoinSymbol, dt5CrtCoin.Name), "")
}

// Обработка транзакций покупки/продажи Монет
func trxSellBuyCoin(retTrns *TrxExt) {
	addCoin := s.CoinActionpData{
		Hash: retTrns.Hash,    // Хэш транзакции
		Time: retTrns.Created, // дата движения
		Type: retTrns.Type,    // type: продажа или покупка
	}

	// разбор движения для пары: монета и фиат(mnt)
	if addCoin.Type == ms.TX_SellCoinData { //SELL
		dtS := s.Tx2SellCoinData{}
		jsonBytes, _ := json.Marshal(retTrns.Data)
		json.Unmarshal(jsonBytes, &dtS)

		addCoin.CoinToBuy = dtS.CoinToBuy   // Монета покупки
		addCoin.CoinToSell = dtS.CoinToSell // Монета продажи
		addCoin.ValueToSell = dtS.ValueToSell
		addCoin.ValueToBuy = retTrns.Tags.TxReturn

	} else if addCoin.Type == ms.TX_BuyCoinData { //BUY
		dtS := s.Tx4BuyCoinData{}
		jsonBytes, _ := json.Marshal(retTrns.Data)
		json.Unmarshal(jsonBytes, &dtS)

		addCoin.CoinToBuy = dtS.CoinToBuy   // Монета покупки
		addCoin.CoinToSell = dtS.CoinToSell // Монета продажи
		addCoin.ValueToBuy = dtS.ValueToBuy
		addCoin.ValueToSell = retTrns.Tags.TxReturn

	} else { //ms.TX_SellAllCoinData (как ms.TX_SellCoinData) SELL_ALL
		dtS := s.Tx3SellAllCoinData{}
		jsonBytes, _ := json.Marshal(retTrns.Data)
		json.Unmarshal(jsonBytes, &dtS)

		addCoin.CoinToBuy = dtS.CoinToBuy   // Монета покупки
		addCoin.CoinToSell = dtS.CoinToSell // Монета продажи
		addCoin.ValueToBuy = retTrns.Tags.TxReturn
		addCoin.ValueToSell = retTrns.Tags.TxSellAmount
	}

	calcCoin(&addCoin)

	if !addCoinTrxSql(dbSQL, &addCoin) {
		log("ERR", "[app_coin.go] startWorkerTrx(addCoinTrxSql)", "")
	}

	// Пересчет с2с
	calcCoin2Coin(addCoin.CoinToSell, addCoin.CoinToBuy)
}

// Расчет движения за 24ч для пары монет
func calcCoin2Coin(tickerS, tickerB string) {
	var mPCL s.PairCoins
	// 1) Запос с базы coin_trx движения за последние 24ч пары монет, не
	// зависемо от направления пары
	now := time.Now()
	lt24 := now.Add(-time.Duration(24) * time.Hour)
	allC2cTrx, err := srchCoin2CoinTrxSql(dbSQL, tickerS, tickerB, lt24, now)
	if err != nil {
		log("ERR", fmt.Sprint("[w_trx.go] calcCoin2Coin(srchCoin2CoinTrxSql) -", err), "")
		return
	}

	// 2) Расчет
	mPCL.CoinToSell = tickerS
	mPCL.CoinToBuy = tickerB
	mPCL.TimeUpdate = time.Now()
	mPCL.PriceBuy, mPCL.PriceSell = CoinPriceNow(mPCL.CoinToBuy, mPCL.CoinToSell)

	for _, tr1 := range allC2cTrx {
		mPCL.Volume24 += tr1.Volume
		// т.к. отсортированы по дате, то первая цена нам и нужна
		if mPCL.PriceBuyOld == 0.0 {
			mPCL.PriceBuyOld = tr1.Price
		}
	}

	// Как регулируется первая продажа??? старая 0, новая 20 => 100%??? НЕТ,
	// будет считаться как установка цены значит должен быть 0%
	if mPCL.PriceBuyOld != 0.0 {
		mPCL.Change24 = ((mPCL.PriceBuy - mPCL.PriceBuyOld) / mPCL.PriceBuyOld) * 100
	} else {
		mPCL.Change24 = 0
	}

	// 3) Заносим в MEM
	if !updCoin2Redis(dbSys, &mPCL) {
		log("ERR", "[w_trx.go] calcCoin2Coin(updCoin2Redis)", "")
	}

	////////////////////////////////////////////////////////////////////////////
	// получаем информацию о монете с блокчейна - а точнее нас интересует
	// текущий из БЧейна: объем и баланс резерва
	for _, tickerX := range []string{tickerB, tickerS} {
		if tickerX != CoinMinter { // за исключением дефолтной монеты системы
			coin1 := s.CoinMarketCapData{CoinSymbol: tickerX}
			iCn, err := sdk.GetCoinInfo(coin1.CoinSymbol)
			if err != nil {
				log("ERR", fmt.Sprint("[w_trx.go] calcCoin2Coin(GetCoinInfo)", coin1.CoinSymbol, " - ERR:", err), "")
				continue
			}

			coin1.TimeUpdate = time.Now()
			coin1.VolumeNow = iCn.Volume
			coin1.ReserveBalanceNow = iCn.ReserveBalance

			if !updCoinInfoRds_3(dbSys, &coin1) {
				log("ERR", "[w_trx.go] calcCoin2Coin(updCoinInfoRds_3)", "")
			}
		}
	}
}

// получение "текущей" стоимости монеты: покупки, продажи (TODO: относительно другой монеты)
func CoinPriceNow(coinSmbl string, coinSmbl2 string) (float32, float32) {
	dataB, err := sdk.EstimateCoinBuy(coinSmbl, coinSmbl2, 1)
	if err != nil {
		log("ERR", fmt.Sprint("[coin2coin_mdb.go] sdk.EstimateCoinBuy ", coinSmbl, "/", coinSmbl2, " - ", err), "")
		//panic(err)
	}
	dataS, err := sdk.EstimateCoinSell(coinSmbl, coinSmbl2, 1)
	if err != nil {
		log("ERR", fmt.Sprint("[coin2coin_mdb.go] sdk.EstimateCoinSell ", coinSmbl, "/", coinSmbl2, " - ", err), "")
		//panic(err)
	}
	return dataB.WillPay, dataS.WillGet
}

// Расчет стоимости монеты
func calcCoin(addCoin *s.CoinActionpData) {
	// FIXME: реализовано движение для монета и фиат(mnt) А НУЖНО ДЛЯ ВСЕХ!
	// в коде идет проверка для кастомной-монеты к MNT/BIP(CoinMinter)
	// хотя код не совсем корректен, т.к. если первая не MNT то она будет
	// второй монетой точно, хотя это не так, может быть ETH/BTC, а не ETH/MNT
	var fAmntMNT float32 = 0.0
	var fAmntToken float32 = 0.0

	if addCoin.Type == ms.TX_SellCoinData { //SELL
		if addCoin.CoinToBuy == CoinMinter {
			fAmntMNT = addCoin.ValueToBuy
			fAmntToken = addCoin.ValueToSell
		} else {
			fAmntMNT = addCoin.ValueToSell
			fAmntToken = addCoin.ValueToBuy
		}

	} else if addCoin.Type == ms.TX_BuyCoinData { //BUY
		fAmntMNT = addCoin.ValueToSell
		fAmntToken = addCoin.ValueToBuy

	} else { //ms.TX_SellAllCoinData (как ms.TX_SellCoinData) SELL_ALL
		if addCoin.CoinToBuy == CoinMinter {
			fAmntMNT = addCoin.ValueToBuy
			fAmntToken = addCoin.ValueToSell
		} else {
			fAmntMNT = addCoin.ValueToSell
			fAmntToken = addCoin.ValueToBuy
		}
	}
	// вычесляем цену одной монеты
	addCoin.Price = fAmntMNT / fAmntToken // сумма/количесво
	addCoin.Volume = fAmntMNT
}
