package main

import (
	"fmt"
	"time"

	s "github.com/ValidatorCenter/prs3r/strc"
)

// Обвязка для модуля обработки Нод
func appNodes_go() {
	for { // бесконечный цикл
		// Шаг 4, загрузка данных о валидаторах/кандидатах
		time.Sleep(time.Minute * 1)

		appNodes()

		time.Sleep(time.Minute * 1) // пауза 1мин
	}
}

// Модуль обработки НОД
func appNodes() {
	// получаем список всех Кандидатов в блокчейне
	allNodes, err := sdk.GetCandidates()
	if err != nil {
		log("ERR", fmt.Sprint("[app_node.go] appNodes(sdk.GetCandidates) - ", err), "")
		return
	}

	// получаем список всех текущих Валидаторов
	allValid, err := sdk.GetValidators()
	if err != nil {
		log("ERR", fmt.Sprint("[app_node.go] appNodes(sdk.GetValidators) - ", err), "")
		return
	}

	for _, oneNode := range allNodes {
		// Бежим по всем нодам Кандидатов...
		oneNodeX := s.NodeExt{}
		oneNodeX.PubKey = oneNode.PubKey
		srchNodeInfoRds(dbSys, &oneNodeX) // заполняем доп.инфой из MEM
		oneNodeX.TimeUpdate = time.Now()
		oneNodeX.TotalStake = oneNode.TotalStake
		oneNodeX.StatusInt = oneNode.StatusInt

		// Ставим статус Валидатор[77] - исключаем тех кто готов стать
		// валидатором, но не проходит по стэку!
		for _, itmV := range allValid {
			if itmV.PubKey == oneNode.PubKey && oneNode.StatusInt == 2 {
				oneNodeX.StatusInt = 77
			}
		}

		/*
			РЕАЛИЗОВАНО!!! в воркере startWorkerBNode(), файл [w_node.go]
			oneNodeX.NoBlocks = ... []
			oneNodeX.AmntBlocks = ...
			oneNodeX.AmntSlashed = ...
		*/

		if !updNodeInfoRds(dbSys, &oneNodeX) {
			log("ERR", "[app_node.go] appNodes(updNodeInfoRds)", "")
		}

		//............................................................
		// запрос API по паблику, что-бы узнать стэк валидатора
		nodeDt1, err := sdk.GetCandidate(oneNode.PubKey)
		if err != nil {
			log("ERR", fmt.Sprint("[app_node.go] appNodes(sdk.GetCandidate) - ", err), "")
			panic(err)
		} else {
			newStakes := []s.StakesInfo{}
			for _, valSt := range nodeDt1.Stakes {
				oneStake := s.StakesInfo{}
				oneStake.BipValue = valSt.BipValue
				oneStake.Coin = valSt.Coin
				oneStake.Owner = valSt.Owner
				oneStake.Value = valSt.Value
				newStakes = append(newStakes, oneStake)
			}
			oneNodeX.Stakes = newStakes // заодно и обнуляем, а не сразу в цикле в oneNodeX.Stakes!
			addNodeStakeSql(dbSQL, &oneNodeX)
		}
		//............................................................
	}
}

// Обработка транзакции создания Ноды
func trxCreateNode(pubkey string, blockDate time.Time) {
	// Получаем данные с блокчейна
	oneNode, err := sdk.GetCandidate(pubkey)
	if err != nil {
		log("ERR", fmt.Sprint("[app_node.go] trxCreateNode(GetCandidate) - ", err), "")
		panic(err) //dbg
		return
	}

	if oneNode.PubKey == "" {
		log("ERR", fmt.Sprint("[app_node.go] trxCreateNode(PubKey = 0) pubkey=", pubkey), "")
		panic("STOP") //dbg
		return
	}

	oneNodeX := s.NodeExt{}
	oneNodeX.PubKey = oneNode.PubKey
	oneNodeX.PubKeyMin = getMinString(oneNode.PubKey)
	oneNodeX.OwnerAddress = oneNode.OwnerAddress
	oneNodeX.RewardAddress = oneNode.RewardAddress
	oneNodeX.Commission = oneNode.Commission
	oneNodeX.CreatedAtBlock = oneNode.CreatedAtBlock
	//oneNodeX.ValidatorAddress = ms.GetVAddressPubKey(oneNode.PubKey) // TODO: используется еще?
	oneNodeX.Created = blockDate

	if !addNodeSql(dbSQL, &oneNodeX) {
		log("ERR", "[app_node.go] trxCreateNode(addNodeSql)", "")
	}

	oneNodeX.Blocks = append(oneNodeX.Blocks, s.BlocksStory{ID: uint32(oneNode.CreatedAtBlock), Type: "declareCandidacy"})
	if !addNodeBlockstorySql(dbSQL, &oneNodeX) {
		log("ERR", "[app_node.go] trxCreateNode(addNodeBlockstorySql) declareCandidacy", "")
	}
}

// Обработка транзакции запуска Ноды
func trxStartNode(bHeight uint32, retTrns *TrxExt) {
	if retTrns.DataTx.PubKey != "" {
		a := s.NodeExt{
			PubKey: retTrns.DataTx.PubKey,
		}
		a.Blocks = append(a.Blocks, s.BlocksStory{ID: bHeight, Type: "setCandidateOnData"})
		if !addNodeBlockstorySql(dbSQL, &a) {
			log("ERR", "[app_node.go] startWorkerTrx(addNodeBlockstorySql) setCandidateOnData", "")
		}
	} else {
		log("ERR", "[app_node.go] startWorkerTrx(TX_SetCandidateOnData) PubKey =0", "")
	}
}

// Обработка транзакции остановки Ноды
func trxStopNode(bHeight uint32, retTrns *TrxExt) {
	if retTrns.DataTx.PubKey != "" {
		a := s.NodeExt{
			PubKey: retTrns.DataTx.PubKey,
		}
		a.Blocks = append(a.Blocks, s.BlocksStory{ID: bHeight, Type: "setCandidateOffData"})
		if !addNodeBlockstorySql(dbSQL, &a) {
			log("ERR", "[app_node.go] startWorkerTrx(addNodeBlockstorySql) setCandidateOffData", "")
		}
	} else {
		log("ERR", "[app_node.go] startWorkerTrx(TX_SetCandidateOffData) PubKey =0", "")
	}
}
