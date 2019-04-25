package main

import (
	"fmt"
	"runtime"

	"time"

	ms "github.com/ValidatorCenter/minter-go-sdk"
	s "github.com/ValidatorCenter/prs3r/strc"
)

// Структура расширяющая функционал структуры из SDK - BlockValidatorsResponse
type B1NExt struct {
	ms.BlockValidatorsResponse
	Height uint32 `json:"height_i32" db:"height_i32"`
}

// Доп.информация о ноде (для визуализации)
type NodeExtInfo struct {
	PubKey string `json:"PK" db:"pub_key"`
	Ticker string `json:"ticker" db:"-"`
	Name   string `json:"title" db:"validator_name"`
	WWW    string `json:"www" db:"validator_url"`
	Logo   string `json:"icon" db:"validator_logo_img"`
	Descr  string `json:"description" db:"validator_desciption"`
}

// Воркер для обработки Валидаторов блока и записи в БД (node_story) и MEM
func startWorkerBNode(workerNum int, in <-chan B1NExt) {
	for pV1 := range in {
		if pV1.Signed == false {
			// пропустили блок!

			a := s.NodeExt{
				PubKey: pV1.PubKey,
			}
			a.Blocks = append(a.Blocks, s.BlocksStory{ID: pV1.Height, Type: "AbsentBlock"})
			if !addNodeBlockstorySql(dbSQL, &a) {
				log("ERR", "[w_node.go] startWorkerBNode(addNodeBlockstorySql) AbsentBlock", "")
			}
			log("STR", fmt.Sprintf("AbsentBlock:: %d - %s", pV1.Height, pV1.PubKey), "")
		}

		/////////////////////////////////////////////////////
		// Подсчет итогов по пропускам и Uptime!
		oneNodeX := s.NodeExt{}
		oneNodeX.PubKey = pV1.PubKey
		srchNodeInfoRds(dbSys, &oneNodeX) // заполняем доп.инфой из MEM
		oneNodeX.TimeUpdate = time.Now()
		oneNodeX.AmntBlocks += 1
		oneNodeX.Blocks = srchNodeBlockstory(dbSQL, pV1.PubKey) // node_blockstory
		AbsentBlock := 0                                        // количество пропущенных блоков
		oneNodeX.AmntSlashed = 0                                // количество штрафов
		for _, b1 := range oneNodeX.Blocks {
			if b1.Type == "AbsentBlock" {
				// пропущен блок
				AbsentBlock++
			} else if b1.Type == "SlashEvent" {
				// штраф: прожиг
				oneNodeX.AmntSlashed++
			}
		}
		oneNodeX.Uptime = 100 - float32(AbsentBlock)/float32(oneNodeX.AmntBlocks)*100
		if oneNodeX.Uptime < 0 {
			oneNodeX.Uptime = 0
		}
		if !updNodeInfoRds(dbSys, &oneNodeX) {
			log("ERR", "[w_node.go] startWorkerBNode(updNodeInfoRds)", "")
		}

		// Добавляем валидатора блока в SQL
		addBlockValidSql(dbSQL, pV1.Height, &pV1.BlockValidatorsResponse)

		runtime.Gosched() // попробуйте закомментировать
	}
}
