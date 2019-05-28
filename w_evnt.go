package main

import (
	"fmt"
	"runtime"
	"time"

	//ms "github.com/ValidatorCenter/minter-go-sdk"
	s "github.com/ValidatorCenter/prs3r/strc"
	"github.com/satori/go.uuid"
)

const PRC100 = 100 // 100%

// Воркер для обработки Событий блока и записи в БД (node_story)
func startWorkerBEvnt(workerNum uint, in <-chan uint32) {
	for bHeight := range in {
		// Разбор Событий(event's) блока
		retEv, err := sdk.GetEvents(int(bHeight))
		if err != nil {
			log("ERR", fmt.Sprint("[blocks_mdb.go] sdk.GetEvents(i) -", err.Error()), "")
		} else {
			// список возвратов X
			addrsXRet := []s.NodeUserX{}
			addrsXRet = srchNodeUserXSql(dbSQL) // TODO: период!!!

			// список нод (pubkey и %)
			nodeDelgateRet := []s.NodeExt{}
			nodeDelgateRet = srchNodeSql_all(dbSQL)

			// разбираем события для блока, прочие события в других местах (декларирование{trxCreateNode}, запуск{trxStartNode}/остановка{trxStopNode} и пропуск{startWorkerBNode} - в других местах!)
			for _, retEv1 := range retEv.Events {

				// ШТРАФ:
				if retEv1.Type == "minter/SlashEvent" {
					oneNodeX1 := s.NodeExt{}
					oneNodeX1 = srchNodeSql_pk(dbSQL, retEv1.Value.ValidatorPubKey) // [*]
					if oneNodeX1.PubKey != "" {
						// нету, добавляем!
						oneNodeX1.Blocks = append(oneNodeX1.Blocks, s.BlocksStory{ID: bHeight, Type: "SlashEvent"})

						if !addNodeBlockstorySql(dbSQL, &oneNodeX1) {
							log("ERR", "[w_evnt.go] startWorkerBEvnt(addNodeBlockstorySql) SlashEvent", "")
						}
						log("STR", fmt.Sprintf("SlashEvent:: %d - %s", bHeight, oneNodeX1.PubKey), "")
					}
				}

				// CASHBACK система:
				// 1 - нужно знать сколько комиссия у валидатора!
				// 2 - делаем для кагдого делегата возврат в соответствие с его индивидуальным условием
				// 		Пример: получил 21мнт в соответствии 17%
				// 			по условию стоит 5%, итого возврат = 21мнт-(21*5%)/17% - комиссияЗаПеревод
				// 3 - заносим в список задач на исполнение
				dateNow := time.Now()
				if retEv1.Type == "minter/RewardEvent" && retEv1.Value.Role == "Delegator" {
					prcMnValid := PRC100          // процент мастер ноды (100%-максимум)
					prcIndividualWallet := PRC100 // процент индивидуальный для кошелька (100%-максимум)
					// Комиссия ноды - по умолчанию (установленная при создание ноды)
					for _, onMn := range nodeDelgateRet {
						if onMn.PubKey == retEv1.Value.ValidatorPubKey {
							prcMnValid = onMn.Commission
							prcIndividualWallet = onMn.Commission
						}
					}

					// Установка индивидуального условия, если имеются
					for _, nux1 := range addrsXRet { // смотрим все условия по всем на данный момент
						// Условия для всех адресов кошельков
						if nux1.PubKey == retEv1.Value.ValidatorPubKey &&
							nux1.Address == "Mx----------------------------------------" &&
							nux1.Start.Unix() <= dateNow.Unix() && nux1.Finish.Unix() >= dateNow.Unix() {
							// Есть индивидуальные условия для кошелька, и меньше чем уже установленная!
							if prcIndividualWallet > nux1.Commission {
								prcIndividualWallet = nux1.Commission
							}
						}

						// Проверка по ноде, адресу кошелька и периоду (текущей дате)
						if nux1.PubKey == retEv1.Value.ValidatorPubKey &&
							nux1.Address == retEv1.Value.Address &&
							nux1.Start.Unix() <= dateNow.Unix() && nux1.Finish.Unix() >= dateNow.Unix() {
							// Есть индивидуальные условия для кошелька, и меньше чем уже установленная!
							if prcIndividualWallet > nux1.Commission {
								prcIndividualWallet = nux1.Commission
							}
						}
					}

					// Создание записи возврата, если есть необходимость
					if prcIndividualWallet < prcMnValid {
						//..расчет возврата...............................
						// Период устраивает (хотя нужно сразу в запросе SQL это надо сделать)
						oneToDoMn := s.NodeTodo{}
						oneToDoMn.ID = uuid.Must(uuid.NewV4())
						oneToDoMn.Priority = 1 // возврат делегатам
						oneToDoMn.Comment = "CashBack delegate masternode"
						oneToDoMn.Type = "SendCashback"
						oneToDoMn.Done = false // не исполнен пока
						oneToDoMn.Created = dateNow
						oneToDoMn.Height = bHeight
						oneToDoMn.PubKey = retEv1.Value.ValidatorPubKey
						oneToDoMn.Address = retEv1.Value.Address

						//oneToDoMn.Amount = retEv1.Value.Amount - (retEv1.Value.Amount*float32(prcIndividualWallet))/float32(prcMnValid) //- 0.01 //комиссия 0.01 платит делегатор
						oneToDoMn.Amount = retEv1.Value.Amount * float32(prcMnValid-prcIndividualWallet) / float32(PRC100-prcMnValid)
						if oneToDoMn.Amount < 0 {
							oneToDoMn.Amount = 0
						}

						// Заносим задачу в базу SQL
						if !addNodeTaskSql(dbSQL, &oneToDoMn) {
							log("ERR", "[w_evnt.go] startWorkerBEvnt(addNodeTaskSql)", "")
						}
						//...............................................
					}
				}
			}
		}

		// Добавляем события блока в SQL
		addBlcokEventSql(dbSQL, bHeight, &retEv)

		runtime.Gosched()
	}
}
