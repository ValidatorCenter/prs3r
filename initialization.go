package main

import (
	"encoding/json"
	"fmt"
	"os"

	//"time"

	"gopkg.in/ini.v1"

	ms "github.com/ValidatorCenter/minter-go-sdk"
	s "github.com/ValidatorCenter/prs3r/strc"

	// SQL
	"github.com/jmoiron/sqlx"
	_ "github.com/kshvakov/clickhouse"

	// Redis
	"github.com/go-redis/redis"
)

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
	secDB := cfg.Section("database")
	CoinMinter = ms.GetBaseCoin()

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
	mbch, err := sdk.GetStatus()
	if err != nil {
		log("ERR", fmt.Sprint("Подключение к Minter-", err.Error()), "")
		dbSQL.Close()
		dbSys.Close()
		os.Exit(0)
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
