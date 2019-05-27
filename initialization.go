package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"time"

	"gopkg.in/ini.v1"

	ms "github.com/ValidatorCenter/minter-go-sdk"
	s "github.com/ValidatorCenter/prs3r/strc"

	// SQL
	"github.com/jmoiron/sqlx"
	_ "github.com/kshvakov/clickhouse"

	// Redis
	"github.com/go-redis/redis"
)

const mbchV = "1.0" // версия Minter

const c_workerBlock = 10 // количество воркеров для отработки Блоков
const c_chanBlock = 20   // размер канала-буфферизации для отработки Блоков
const c_workerBEvnt = 10 // количество воркеров для отработки Событий блока
const c_chanBEvnt = 20   // размер канала-буфферизации для отработки Событий блока
const c_workerTrx = 100  // количество воркеров для отработки Транзакций
const c_chanTrx = 200    // размер канала-буфферизации для отработки Транзакций
const c_workerNode = 10  // количество воркеров для отработки Нод блокчейна
const c_chanNode = 10    // размер канала-буфферизации для отработки Нод блокчейна
const c_workerBNode = 10 // количество воркеров для отработки Валидаторов(нод) блока
const c_chanBNode = 20   // размер канала-буфферизации для отработки Валидаторов(нод) блока

// Структура для файла "start.json"
type ConfigStart struct {
	AppUserX []s.NodeUserX         `json:"app_userx"`
	AppCoin  []s.CoinMarketCapData `json:"coins"`
	/*app_state
			|-candidates[]
				|-pub_key
				  reward_address
	        	  owner_address
			      total_bip_stake
	        	  commission

	*/
}

// загрузка файла genesis.json
func loadStartJSON() bool {
	file, _ := os.Open("start.json")
	decoder := json.NewDecoder(file)
	cfgStr := new(ConfigStart)
	err := decoder.Decode(&cfgStr)
	if err != nil {
		log("ERR", fmt.Sprint("Чтение файла JSON - ", err), "")
		return false
	}
	// Заносим в базу: индивидуальные % ставки для кошельков!
	for iStp, _ := range cfgStr.AppUserX {
		if !addNodeUserX(dbSQL, &cfgStr.AppUserX[iStp]) {
			log("ERR", fmt.Sprint("Запись в node_userx - ", cfgStr.AppUserX[iStp].Address), "")
			return false
		}
	}
	// TODO: данные по нодам, если после хардфорка

	// Данные по монетам, если после хардфорка
	for iCn, _ := range cfgStr.AppCoin {
		if !addCoinSql(dbSQL, &cfgStr.AppCoin[iCn]) {
			log("ERR", fmt.Sprint("Запись в coin - ", cfgStr.AppCoin[iCn].CoinSymbol), "")
			return false
		}
	}

	return true
}

// инициализация парсера, загрузка параметров
func initParser() {
	sdk.Debug = true
	ConfFileName := "config.ini"
	cmdClearDB := false
	cmdLoadJSON := false

	// проверяем есть ли входной параметр/аргумент
	if len(os.Args) == 2 {
		if os.Args[1] == "del" {
			cmdClearDB = true
		} else if os.Args[1] == "json" {
			cmdLoadJSON = true
		} else {
			ConfFileName = os.Args[1]
		}
	}
	log("", fmt.Sprintf("Config file => %s", ConfFileName), "")

	// INI
	cfg, err := ini.LoadSources(ini.LoadOptions{IgnoreInlineComment: true}, ConfFileName)
	if err != nil {
		log("ERR", fmt.Sprint("Чтение файла настроек-", err), "")
		os.Exit(0) //завершаем работу программы
	} else {
		log("", "...данные с INI файла = загружены!", "")
	}
	secMN := cfg.Section("masternode")
	sdk.MnAddress = secMN.Key("ADDRESS").String()
	_strChain := secMN.Key("CHAIN").String()
	if strings.ToLower(_strChain) == "main" {
		sdk.ChainMainnet = true
	} else {
		sdk.ChainMainnet = false
	}

	loadCorrection = 0
	loadCorrection, err = secMN.Key("BLOCK_CORRECTION").Uint()
	if err != nil {
		loadCorrection = 1000
	}

	secP3 := cfg.Section("parser")
	workerBlock, err = secP3.Key("WORKER_BLOCK").Uint()
	if err != nil || workerBlock == 0 {
		workerBlock = c_workerBlock
	}
	chanBlock, err = secP3.Key("CHAN_BLOCK").Uint()
	if err != nil || chanBlock == 0 {
		chanBlock = c_chanBlock
	}
	workerBEvnt, err = secP3.Key("WORKER_BEVNT").Uint()
	if err != nil || workerBEvnt == 0 {
		workerBEvnt = c_workerBEvnt
	}
	chanBEvnt, err = secP3.Key("CHAN_BEVNT").Uint()
	if err != nil || chanBEvnt == 0 {
		chanBEvnt = c_chanBEvnt
	}
	workerTrx, err = secP3.Key("WORKER_TRX").Uint()
	if err != nil || workerTrx == 0 {
		workerTrx = c_workerTrx
	}
	chanTrx, err = secP3.Key("CHAN_TRX").Uint()
	if err != nil || chanTrx == 0 {
		chanTrx = c_chanTrx
	}
	workerNode, err = secP3.Key("WORKER_NODE").Uint()
	if err != nil || workerNode == 0 {
		workerNode = c_workerNode
	}
	chanNode, err = secP3.Key("CHAN_NODE").Uint()
	if err != nil || chanNode == 0 {
		chanNode = c_chanNode
	}
	workerBNode, err = secP3.Key("WORKER_BNODE").Uint()
	if err != nil || workerBNode == 0 {
		workerBNode = c_workerBNode
	}
	chanBNode, err = secP3.Key("CHAN_BNODE").Uint()
	if err != nil || chanBNode == 0 {
		chanBNode = c_chanBNode
	}

	secDB := cfg.Section("database")
	CoinMinter = ms.GetBaseCoin()

	amntBlocksLoad, err = secDB.Key("BLOCKS_LOAD").Uint()
	if err != nil || amntBlocksLoad == 0 {
		amntBlocksLoad = 1000
	}
	pauseBlocksLoad, err = secDB.Key("BLOCKS_LOAD_PAUSE").Uint()
	if err != nil || pauseBlocksLoad == 0 {
		pauseBlocksLoad = 1
	}

	r_db, err := secDB.Key("REDIS_DB").Int()
	if err != nil {
		r_db = 0
	}
	dbSys = redis.NewClient(&redis.Options{
		Addr:     secDB.Key("REDIS_ADDRESS").String(),
		Password: secDB.Key("REDIS_PSWRD").String(), // no password set
		DB:       r_db,                              // use default DB
	})
	//defer dbSys.Close() // рано закрывается(в конце этой функции), в итоге перенес в main
	log("OK", "Подключились к БД - Redis", "")

	dbSQL, err = sqlx.Open("clickhouse", secDB.Key("CLICKHOUSE_ADDRESS").String())
	if err != nil {
		log("ERR", fmt.Sprint("Подключение к БД - ClickHouse ", err), "")
	}
	//defer dbSQL.Close() // рано закрывается(в конце этой функции), в итоге перенес в main
	log("OK", "Подключились к БД - ClickHouse", "")

	if cmdClearDB == true {
		// очистка и создание таблиц в базе ClickHouse-SQL
		ClearChSqlDB()
		// очистка в базе Redis
		ClearSysDB()
		// завершение
		dbSQL.Close()
		dbSys.Close()
		os.Exit(0)
	} else if cmdLoadJSON == true {
		// Загрузка первоночальных данных с файла JSON
		if !loadStartJSON() {
			// завершение
			dbSQL.Close()
			dbSys.Close()
			os.Exit(0)
		}

		/*
			// hardfork
			// Загрузка первоночальных данных о Нодах (УДАЛИТЬ: после загрузки нод с JSON)
			if !preStartNodeData() {
				// завершение по ошибке!
			}*/
		// завершение
		dbSQL.Close()
		dbSys.Close()
		os.Exit(0)
	}

	// Проверка версии Минтера
	mbch := ms.ResultNetwork{}
	for {
		mbch, err = sdk.GetStatus()
		if err != nil {
			log("ERR", fmt.Sprint("Подключение к Minter-", err.Error()), "")
			time.Sleep(10 * time.Second) // ждём до новой попытки
		} else {
			break
		}
	}

	if mbch.Version[0:len(mbchV)] != mbchV {
		log("ERR", fmt.Sprint("Парсер не для данной версии блокчейна Minter"), "")
		dbSQL.Close()
		dbSys.Close()
		os.Exit(0)
	} else {
		log("INF", "INIT", fmt.Sprintf("Minter v.%s", mbch.Version))
	}
}

// Загрузка первоночальных данных о Нодах с блокчейна
func preStartNodeData() bool {
	// получаем список всех Кандидатов в блокчейне
	allNodes, err := sdk.GetCandidates()
	if err != nil {
		log("ERR", fmt.Sprint("[init...go] preStartNodeData(GetCandidates) -", err), "")
		return false
	}
	if len(allNodes) == 0 {
		panic("sdk.GetCandidates() =>> 0")
	}
	for _, oneNode := range allNodes {
		// Бежим по всем нодам Кандидатов...
		oneNodeX := s.NodeExt{}
		oneNodeX.PubKey = oneNode.PubKey
		oneNodeX.PubKeyMin = getMinString(oneNode.PubKey)
		oneNodeX.OwnerAddress = oneNode.OwnerAddress
		oneNodeX.RewardAddress = oneNode.RewardAddress
		oneNodeX.Commission = oneNode.Commission
		oneNodeX.CreatedAtBlock = oneNode.CreatedAtBlock

		// Запрос по API № блока, что-бы узнать дату создания
		// НО! если после хардфорка, то будет не реальная дата создания
		// а дата создания первого блока в новом хардфорке, что не правильно :(
		blockDt1, err := sdk.GetBlock(oneNode.CreatedAtBlock)
		if err != nil {
			log("ERR", fmt.Sprint("[init...go] preStartNodeData(sdk.GetBlock) -", err), "")
			return false
		} else {
			oneNodeX.Created = blockDt1.Time
		}

		if !addNodeSql(dbSQL, &oneNodeX) {
			log("ERR", "[init...go] preStartNodeData(addNodeSql)", "")
			return false
		}
	}

	return true
}
