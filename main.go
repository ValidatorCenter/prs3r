package main

import (
	"time"

	ms "github.com/ValidatorCenter/minter-go-sdk"

	//Gin web-framework
	"github.com/gin-gonic/gin"

	// SQL (mail.ru)
	"github.com/mailru/dbr"
	_ "github.com/mailru/go-clickhouse"

	// Redis
	"github.com/go-redis/redis"
)

var (
	CoinMinter       string // Основная монета Minter
	amntN_block      int    // всего блоков в сети
	amntL_block      int    // синхронизировано блоков из сети
	amntBlocksLoad   uint   // количество загружаемых блоков за раз
	pauseBlocksLoad  uint   // пауза между загрузками блоков (сек)
	pauseSystem      uint   // пауза между циклами и попытками при ошибках (сек)
	pauseNodeUpd     uint   // пауза между обновлением информации о нодах (сек)
	loadCorrection   uint   // на сколько блоков не дозагружать из блокчейна, если еще не синхронизировалось в блокчейне валидаторам подписантам
	ParserIsActive   bool   // активен парсер Да/нет
	sdk              ms.SDK
	worketInputBlock chan ms.BlockResponse
	worketInputTrx   chan ms.BlockResponse
	worketInputBNode chan ms.BlockResponse
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

	dbSQL *dbr.Connection
	dbSys *redis.Client
)

func main() {
	// запуск GiN
	r := gin.Default()

	initParser()
	defer dbSQL.Close()
	defer dbSys.Close()

	// создаём каналы
	worketInputBlock = make(chan ms.BlockResponse, chanBlock)
	worketInputTrx = make(chan ms.BlockResponse, chanTrx)
	worketInputBNode = make(chan ms.BlockResponse, chanBNode)
	worketInputBEvnt = make(chan uint32, chanBEvnt)

	///////////////////////////////////////
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

	///////////////////////////////////////
	// запуск демонов основных служб
	go appNodes_go()  // Обновление информации о нодах
	go appBlocks_go() // Обновление информации о блоках

	///////////////////////////////////////
	// запуск веб-службы упревления

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.GET("/start", func(c *gin.Context) {
		ParserIsActive = true
		c.JSON(200, gin.H{
			"message":       "start",
			"is_active":     ParserIsActive,
			"current_block": amntN_block, // всего блоков в сети
			"sync_block":    amntL_block, // синхронизировано блоков из сети
		})
	})
	r.GET("/stop", func(c *gin.Context) {
		ParserIsActive = false
		c.JSON(200, gin.H{
			"message":       "stop",
			"is_active":     ParserIsActive,
			"current_block": amntN_block, // всего блоков в сети
			"sync_block":    amntL_block, // синхронизировано блоков из сети
		})
	})
	r.GET("/exit", func(c *gin.Context) {
		ParserIsActive = false
		time.Sleep(60 * time.Second) // ждём завершения работы горутин

		close(worketInputBlock)
		close(worketInputTrx)
		close(worketInputBNode)

		time.Sleep(20 * time.Second) // ждём чтобы наверняка завершилась корректно запись в БД при закрытие каналов
		c.JSON(200, gin.H{
			"message": "exit",
		})
	})
	r.GET("/status", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message":       "status",
			"is_active":     ParserIsActive,
			"current_block": amntN_block, // всего блоков в сети
			"sync_block":    amntL_block, // синхронизировано блоков из сети
		})
	})

	r.Run(":8018")
}
