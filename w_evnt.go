package main

import (
	"fmt"
	"runtime"
	"time"

	//ms "github.com/ValidatorCenter/minter-go-sdk"
	s "github.com/ValidatorCenter/prs3r/strc"
)

// Воркер для обработки Событий блока и записи в БД (node_story)
func startWorkerBEvnt(workerNum int, in <-chan uint32) {
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

			// разбираем события для блока
			for _, retEv1 := range retEv.Events {

				// ШТРАФ:
				if retEv1.Type == "minter/SlashEvent" {
					oneNodeX1 := s.NodeExt{}
					oneNodeX1 = srchNodeSql_pk(dbSQL, retEv1.Value.ValidatorPubKey) // [*]
					if oneNodeX1.PubKey != "" {
						// TODO: возложить на SQL проверку, исправить не на полную прогрузку и анализ тут в коде
						// а на поиск в БД
						oneNodeX1.Blocks = srchNodeBlockstory(dbSQL, retEv1.Value.ValidatorPubKey) // прогружаем блоки

						// только перебором и сравнивать id и тип в blocks_story! запрос не корректно отрабатывается при поиске во вложениях
						findY := false
						for _, el1E := range oneNodeX1.Blocks {
							// проверяем - записан-ли уже штраф в базу:
							if el1E.ID == bHeight && el1E.Type == "SlashEvent" {
								findY = true
							}
						}

						if !findY {
							// нету, добавляем!
							oneNodeX1.Blocks = append(oneNodeX1.Blocks, s.BlocksStory{ID: bHeight, Type: "SlashEvent"})

							if !addNodeBlockstorySql(dbSQL, &oneNodeX1) {
								log("ERR", "[w_evnt.go] startWorkerBEvnt(addNodeBlockstorySql) SlashEvent", "")
							}
							log("STR", fmt.Sprintf("SlashEvent:: %d - %s", bHeight, oneNodeX1.PubKey), "")

						}
					}
				}

				// CASHBACK система:
				// 1 - нужно знать сколько комиссия у валидатора!
				// 2 - делаем для кагдого делегата возврат в соответствие условий уникальных
				// 		Пример: получил 21мнт в соответствии 17%
				// 			по условию стоит 5%, итого возврат = 21мнт-(21*5%)/17% - комиссияЗаПеревод
				// 3 - заносим в список задач на исполнение
				dateNow := time.Now()
				if retEv1.Type == "minter/RewardEvent" && retEv1.Value.Role == "Delegator" {
					for _, nux1 := range addrsXRet {
						if nux1.PubKey == retEv1.Value.ValidatorPubKey &&
							nux1.Address == retEv1.Value.Address &&
							nux1.Start.Unix() <= dateNow.Unix() && nux1.Finish.Unix() >= dateNow.Unix() {

							//..расчет возврата...............................
							// Период устраивает (хотя нужно сразу в запросе SQL это надо сделать)
							oneToDoMn := s.NodeTodo{}
							oneToDoMn.Priority = 1 // возврат делегатам
							oneToDoMn.Comment = "CashBack delegate masternode"
							oneToDoMn.Type = "SendCashback"
							oneToDoMn.Done = false // не исполнен пока
							oneToDoMn.Created = dateNow
							oneToDoMn.Height = bHeight
							oneToDoMn.PubKey = retEv1.Value.ValidatorPubKey
							oneToDoMn.Address = retEv1.Value.Address
							prcMnValid := 1
							for _, onMn := range nodeDelgateRet {
								if onMn.PubKey == retEv1.Value.ValidatorPubKey {
									prcMnValid = onMn.Commission
								}
							}
							oneToDoMn.Amount = retEv1.Value.Amount - (retEv1.Value.Amount*float32(nux1.Commission))/float32(prcMnValid) - 0.01 //комиссия 0.01 платит делегатор
							if oneToDoMn.Amount < 0 {
								oneToDoMn.Amount = 0
							}

							// Заносим задачу в базу SQL
							if !addNodeTaskSql(dbSQL, &oneToDoMn) {
								log("ERR", "[blocks_mdb.go] addNodeTaskSql(dbSQL, ...)", "")
							}
							//...............................................
						}
					}
				}
			}
		}

		// Добавляем события блока в SQL
		addBlcokEventSql(dbSQL, bHeight, &retEv)

		runtime.Gosched()
	}
}
