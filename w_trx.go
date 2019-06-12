package main

import (
	"encoding/json"
	"fmt"
	"runtime"

	"time"

	ms "github.com/ValidatorCenter/minter-go-sdk"
	s "github.com/ValidatorCenter/prs3r/strc"
)

// Структура расширяющая функционал структуры из SDK - ms.TransResponse
type TrxExt struct {
	ms.TransResponse
	Created time.Time // создана time
}

// Воркер для обработки Транзакций и записи в БД
func startWorkerTrx(workerNum uint, in <-chan ms.BlockResponse) {
	var err error
	//var buffTrx []ms.TransResponse

	for retBlck := range in {

		// Цикл по транзакциям, обработка индивидуально определенных условий (например: создание монеты и т.п.)
		for _, retTrns := range retBlck.Transactions {
			// нужно передовать ДатуБлока(+),НомерБлока(+) и тело одной транзакции(+)
			if retTrns.Hash != "" {
				oneTrxX := TrxExt{TransResponse: retTrns}
				oneTrxX.Height = retBlck.Height // высота блока
				oneTrxX.Created = retBlck.Time

				// Обработка определенного типа транзакций:
				if retTrns.Code == 0 { // без ошибочная транзакция!

					bHeight := uint32(retBlck.Height) // высота блока

					// При загрузке транзакций - прогружать данные по поводу Валидаторов,
					// создание и т.п.
					switch retTrns.Type {

					////////////////////////////////////////////////////////////////////
					// SendCoin - перевод монет
					case ms.TX_SendData: //1
						dt0 := s.Tx1SendData{}
						jsonBytes, _ := json.Marshal(retTrns.Data)
						json.Unmarshal(jsonBytes, &dt0)

						// Заносим данные о валидаторе по протоколу Monsternode
						if retTrns.Payload != "" && dt0.To == "Mxa62da2d2714f23738a4d1658909eb6c920669b0e" {
							// Разбираем JSON
							neInf := NodeExtInfo{}
							err = json.Unmarshal([]byte(retTrns.Payload), &neInf)
							if err != nil {
								log("ERR", fmt.Sprint("[w_trx.go] json.Unmarshal -", err.Error()), "")
							} else {
								if neInf.PubKey != "" {
									//#=> Мастернода
									oneNodeX := srchNodeSql_oa(dbSQL, neInf.PubKey, retTrns.From) // Проверяем адрес отправителя From владеет pubkey

									if oneNodeX.PubKey != "" {
										// обновляем данные о ноде в MEM
										if !updNodeInfoRds_ext(dbSys, &neInf) {
											log("ERR", "[w_trx.go] startWorkerTrx(updNodeInfoRds_ext)", "")
										}
									}
								} else if neInf.Ticker != "" {
									//#=> Монета
									oneCoinX := srchCoinSql(dbSQL, neInf.Ticker)
									srchCoinInfoRds(dbSys, &oneCoinX)                                  // получаем доп.инфу о Монете
									if oneCoinX.CoinSymbol != "" && oneCoinX.Creator == retTrns.From { // Проверяем адрес отправителя, что он создатель монеты
										oneCoinX.CoinURL = neInf.WWW
										oneCoinX.CoinLogoImg = neInf.Logo
										oneCoinX.CoinDesc = neInf.Descr

										// обновляем данные о монете в MEM
										if !updCoinInfoRds_3v(dbSys, &oneCoinX) {
											log("ERR", "[w_trx.go] startWorkerTrx(updCoinInfoRds_3v)", "")
										}
									}
								}
							}
						}

					////////////////////////////////////////////////////////////////////
					// CreateCoin - создание монеты
					case ms.TX_CreateCoinData: //5
						trxCreateCoin(&oneTrxX)

					////////////////////////////////////////////////////////////////////
					// SellCoin - продажа монет [2]
					// SellAllCoin - продажа всех монет [3]
					// BuyCoin - покупка монет [4]
					case ms.TX_SellCoinData, ms.TX_SellAllCoinData, ms.TX_BuyCoinData: // 2,3,4
						trxSellBuyCoin(&oneTrxX)

					////////////////////////////////////////////////////////////////////
					// DeclareCandidacy - декларирование ноды в Кандидаты
					case ms.TX_DeclareCandidacyData: //6
						if retTrns.DataTx.PubKey != "" {
							// Создаём в базах
							trxCreateNode(retTrns.DataTx.PubKey, retBlck.Time)
						} else {
							log("ERR", "[w_trx.go] startWorkerTrx(TX_DeclareCandidacyData) PubKey =0", "")
						}

					////////////////////////////////////////////////////////////////////
					// Delegate - делегирование
					case ms.TX_DelegateDate: //7

					////////////////////////////////////////////////////////////////////
					// Unbond - отмена делегирования
					case ms.TX_UnbondData: //8

					////////////////////////////////////////////////////////////////////
					// SetCandidateOn - запуск Ноды-кадидата
					case ms.TX_SetCandidateOnData: //10
						trxStartNode(bHeight, &oneTrxX)

					////////////////////////////////////////////////////////////////////
					// SetCandidateOff - остановка Ноды-кадидата
					case ms.TX_SetCandidateOffData: //11
						trxStopNode(bHeight, &oneTrxX)

					////////////////////////////////////////////////////////////////////
					// MultisendCoin - мульти перевод монет
					case ms.TX_MultisendData: //13
						// TODO: проверка, если тут будет комментарий для обработки

					////////////////////////////////////////////////////////////////////
					// EditCandidate - изменение адреса в Ноде-кандидате
					case ms.TX_EditCandidateData: //14
						// TODO: единственный случай когда могут меняться данные в таблице "nodes"

					}
				}

			}
		}

		// добавляем одну транзакцию в БД
		if !addTrxSql(dbSQL, &retBlck) {
			log("ERR", "[w_trx.go] startWorkerTrx(addTrxSql)", "")
		}
		// добавляем данные для транзакции в БД
		if !addTrxDataSql(dbSQL, &retBlck) {
			log("ERR", "[w_trx.go] startWorkerTrx(addTrxDataSql)", "")
		}

		runtime.Gosched() // попробуйте закомментировать
	}
}
