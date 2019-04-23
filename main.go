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

const worherBlock = 10 // количество воркеров для отработки Блоков
const chanBlock = 20   // размер канала-буфферизации для отработки Блоков
const worherBEvnt = 10 // количество воркеров для отработки Событий блока
const chanBEvnt = 20   // размер канала-буфферизации для отработки Событий блока
const worherTrx = 100  // количество воркеров для отработки Транзакций
const chanTrx = 200    // размер канала-буфферизации для отработки Транзакций
const worherNode = 10  // количество воркеров для отработки Нод блокчейна
const chanNode = 10    // размер канала-буфферизации для отработки Нод блокчейна
const worherBNode = 10 // количество воркеров для отработки Валидаторов(нод) блока
const chanBNode = 20   // размер канала-буфферизации для отработки Валидаторов(нод) блока
const mbchV = "0.20"   // версия Minter

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
	for i := 0; i < worherBlock; i++ {
		go startWorkerBlock(i, worketInputBlock)
	}
	for i := 0; i < worherTrx; i++ {
		go startWorkerTrx(i, worketInputTrx)
	}
	for i := 0; i < worherBNode; i++ {
		go startWorkerBNode(i, worketInputBNode)
	}
	for i := 0; i < worherBEvnt; i++ {
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
