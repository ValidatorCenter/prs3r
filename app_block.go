package main

import (
	"fmt"
	"time"

	ms "github.com/ValidatorCenter/minter-go-sdk"
)

// получение сколько всего блоков в блокчейне
func MinterLatestBlock() (int, error) {
	sts, err := sdk.GetStatus()
	if err != nil {
		return 0, err
	}
	log("INF", "INIT", fmt.Sprintf("Последний блок в Minter.Network: %d", sts.LatestBlockHeight))

	return sts.LatestBlockHeight, nil
}

// Модуль обработки БЛОКОВ
func appBlocks() {
	var err error
	amntN_block, err = MinterLatestBlock()
	if err != nil {
		log("ERR", fmt.Sprint("[app_block.go] appBlocks(MinterLatestBlock) - ", err), "")
		// Прерывать, скорее всего количество запросов к API превысили!!!
		// FIXME: надо как-то более корректно обрабатывать, чтобы не вылетало а просто ждало!
		return
	}

	// Корректировка по последнему загружаемому блоку
	if amntN_block > int(loadCorrection) {
		amntN_block -= int(loadCorrection)
	}

	// получаем системуную коллекцию
	st0 := srchSysSql(dbSys)
	log("INF", "INIT", fmt.Sprintf("Последний блок в БД: %d", st0.LatestBlockSave))
	actN_block := st0.LatestBlockSave + 1 // загружаем следующий блок!

	step_amntBlocksLoad := uint(0) // считает сколько загрузили за раз
	for i := actN_block; i <= amntN_block; i++ {
		log("INF", "LOAD", fmt.Sprintf("=== БЛОК %d из %d", i, amntN_block))

		// получаем блок по номеру i с блокчейна
		retBlck := ms.BlockResponse{}

		for {
			retBlck, err = sdk.GetBlock(i)
			if err != nil {
				//Возможно не доступна нода блокчейна, надо подождать а не паниковать
				log("ERR", fmt.Sprint("[app_block.go] appBlocks(sdk.GetBlock) - ", err), "")
				time.Sleep(10 * time.Second) // ждём до новой попытки
			} else {
				break
			}
		}

		if retBlck.Hash == "" {
			log("ERR", fmt.Sprint("[app_block.go] appBlocks(retBlck.Hash) => ПУСТО!"), "")
			//return false
			continue
		}

		// Отправляем воркеру на отработку Блока (там - же: Транзакция и Валидаторы-пропуск)
		worketInputBlock <- retBlck

		// Цикл по транзакциям...
		for _, retTrx := range retBlck.Transactions {
			// нужно передовать ДатуБлока(+),НомерБлока(+) и тело одной транзакции(+)
			if retTrx.Hash != "" {
				oneTrxX := TrxExt{TransResponse: retTrx}
				oneTrxX.Height = retBlck.Height // высота блока
				oneTrxX.Created = retBlck.Time

				worketInputTrx <- oneTrxX
			}
		}

		// Цикл по валидаторам, проверка пропуска и запись в node_blockstory
		for _, retB1node := range retBlck.Validators {
			oneBlockNodeX := B1NExt{BlockValidatorsResponse: retB1node}
			oneBlockNodeX.Height = uint32(retBlck.Height) // высота блока
			worketInputBNode <- oneBlockNodeX
		}

		// Обработка событий по номеру блока
		worketInputBEvnt <- uint32(retBlck.Height)

		if (amntN_block - actN_block) > int(amntBlocksLoad) {
			// Если количество блоков, которые нужно загрузить больше
			// максимального количества блоков разрешенных для загрузки за раз,
			// тогда будем делать паузы!

			if step_amntBlocksLoad < amntBlocksLoad {
				step_amntBlocksLoad++
			} else {
				// пауза, дадим записаться всему в БД
				step_amntBlocksLoad = 0
				time.Sleep(time.Second * time.Duration(pauseBlocksLoad))
			}
		}

	}
}
