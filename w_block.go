package main

import (
	//"fmt"
	"runtime"

	ms "github.com/ValidatorCenter/minter-go-sdk"
	s "github.com/ValidatorCenter/prs3r/strc"
)

// Воркер для обработки Блока и записи его в БД
func startWorkerBlock(workerNum uint, in <-chan ms.BlockResponse) {
	for retBlck := range in {
		// Данные нового блока (в структуре подобной что в SDK)
		retBlckEx := s.BlockResponse2{}
		retBlckEx.Hash = retBlck.Hash
		retBlckEx.Height = retBlck.Height
		retBlckEx.Time = retBlck.Time
		retBlckEx.NumTxs = retBlck.NumTxs
		retBlckEx.TotalTxs = retBlck.TotalTxs
		retBlckEx.BlockReward = retBlck.BlockReward
		retBlckEx.Size = retBlck.Size
		retBlckEx.Proposer = retBlck.Proposer

		for _, retTrns := range retBlck.Transactions {
			// для sql-варианта, массив хэшей транзакций входящих в данный блок!
			retBlckEx.TransHashArr = append(retBlckEx.TransHashArr, retTrns.Hash)
		}

		if !addBlockSql(dbSQL, &retBlckEx) {
			log("ERR", "[w_block.go] startWorkerBlock(addBlockSql)", "")
		} else {
			// обноляем счетчик в БД
			if !updSystemDB_Save(dbSys, retBlck.Height) {
				log("ERR", "[w_block.go] startWorkerBlock(updSystemDB_Save)", "")
			}
		}

		runtime.Gosched()
	}
}
