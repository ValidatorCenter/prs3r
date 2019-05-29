package main

import (
	"fmt"
	"runtime"

	//ms "github.com/ValidatorCenter/minter-go-sdk"
	s "github.com/ValidatorCenter/prs3r/strc"
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
			}
		}

		// Добавляем события блока в SQL
		addBlcokEventSql(dbSQL, bHeight, &retEv)

		runtime.Gosched()
	}
}
