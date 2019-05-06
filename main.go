package main

import (
	"fmt"
	"time"

	ms "github.com/ValidatorCenter/minter-go-sdk"

	// SQL
	"github.com/jmoiron/sqlx"
	_ "github.com/kshvakov/clickhouse"

	// Redis
	"github.com/go-redis/redis"
)

var (
	CoinMinter       string // Основная монета Minter
	amntN_block      int    // всего блоков в сети
	amntBlocksLoad   uint   // количество загружаемых блоков за раз
	pauseBlocksLoad  uint   // паузе между загрузками блоков (сек)
	loadCorrection   uint   // на сколько блоков не дозагружать из блокчейна, если еще не синхронизировалось в блокчейне валидаторам подписантам
	sdk              ms.SDK
	worketInputBlock chan ms.BlockResponse
	worketInputTrx   chan TrxExt
	worketInputBNode chan B1NExt
	worketInputBEvnt chan uint32

	workerBlock uint
	chanBlock   uint
	workerBEvnt uint
	chanBEvnt   uint
	workerTrx   uint
	chanTrx     uint
	workerNode  uint
	chanNode    uint
	workerBNode uint
	chanBNode   uint

	dbSQL *sqlx.DB
	dbSys *redis.Client
)

func main() {

	initParser()
	defer dbSQL.Close()
	defer dbSys.Close()

	// создаём каналы
	worketInputBlock = make(chan ms.BlockResponse, chanBlock)
	worketInputTrx = make(chan TrxExt, chanTrx)
	worketInputBNode = make(chan B1NExt, chanBNode)
	worketInputBEvnt = make(chan uint32, chanBEvnt)

	// запуск воркеров-демонов
	for i := uint(0); i < workerBlock; i++ {
		go startWorkerBlock(i, worketInputBlock)
	}
	for i := uint(0); i < workerTrx; i++ {
		go startWorkerTrx(i, worketInputTrx)
	}
	for i := uint(0); i < workerBNode; i++ {
		go startWorkerBNode(i, worketInputBNode)
	}
	for i := uint(0); i < workerBEvnt; i++ {
		go startWorkerBEvnt(i, worketInputBEvnt)
	}

	// Обновление информации о нодах
	go appNodes_go()

	// Загрузка блока с блок-чейна
	for { // бесконечный цикл
		appBlocks()
		time.Sleep(time.Minute * 3) // пауза 3мин ....в этот момент лучше прерывать
	}

	close(worketInputBlock)
	close(worketInputTrx)
	close(worketInputBNode)

	time.Sleep(10 * time.Second) // ждём чтобы наверняка завершилась корректно запись в БД при закрытие каналов
	fmt.Println("конец, нажмите любую кнопку....")
	fmt.Scanln()
}
