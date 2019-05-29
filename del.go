package main

/*import (
	s "github.com/ValidatorCenter/prs3r/strc"
)*/

// Redis
func ClearSysDB() {
	//удаляем данные из 0 базы данных
	dbSys.FlushDB()
	log("OK", "...всё очищено - MEM.DB", "")
}

// Yandex.ClickHouse (SQL)
func ClearChSqlDB() {
	/**************************************************************************/
	/**************************************************************************/
	/**************************************************************************/

	////////////////////////////////////////////////////////////////////////////
	// Таблица содержит основные данные о Блоке
	delCh_blocks := `DROP TABLE IF EXISTS blocks`
	dbSQL.MustExec(delCh_blocks)
	schemaCh_blocks := `
			CREATE TABLE blocks (
				hash String,
				height_i32 UInt32,
				time DateTime,
				num_txs_i32 UInt32,
				total_txs_i32 UInt32,
				transactions Array(String),
				block_reward_f32 Float32,
				size_i32 UInt32,
				proposer String,
				updated_date Date
			) ENGINE = MergeTree(updated_date,(height_i32),8192)
			`
	dbSQL.MustExec(schemaCh_blocks)
	log("OK", "...очищена - blocks", "")

	////////////////////////////////////////////////////////////////////////////
	// Таблица содержит информацию о подписантах блока (связка: паблик
	// валидатора и номер блока)
	delCh_block_valid := `DROP TABLE IF EXISTS block_valid`
	dbSQL.MustExec(delCh_block_valid)
	schemaCh_block_valid := `
			CREATE TABLE block_valid (
				height_i32 UInt32,
				pub_key String,
				signed UInt8,
				updated_date Date
			) ENGINE = MergeTree(updated_date,(height_i32,pub_key),8192)
			`
	dbSQL.MustExec(schemaCh_block_valid)
	log("OK", "...очищена - block_valid", "")

	////////////////////////////////////////////////////////////////////////////
	//Daniil Lashin, [04.03.19 14:23] Могут повторяться, если монеты разные
	// делегированы от одного человека
	//
	// Хотя вся награда приходит в базовой монете!
	// FIXME: поставил уникальность и по сумме amount_f32, но нужно будет убрать если Даниил реализует 1-строкой-уникальной
	// есть случаи когда и сумма одинаковая, вот и возникает "Ж"! Поэтому придется ID
	//
	// Таблица содержит информацию о Наградах и Штрафах в блоке.
	/*schemaCh_block_event := `
	CREATE TABLE block_event (
		height_i32 UInt32,
		type String,
		role String,
		address String,
		amount_f32 Float32,
		coin String,
		validator_pub_key String,
		updated_date Date
	) ENGINE = MergeTree(updated_date,(height_i32,type,role,address,validator_pub_key),8192)
	`*/
	delCh_block_event := `DROP TABLE IF EXISTS block_event`
	dbSQL.MustExec(delCh_block_event)
	schemaCh_block_event := `
			CREATE TABLE block_event (
				_id UUID,
				height_i32 UInt32,
				type String,
				role String,
				address String,
				amount_f32 Float32,
				coin String,
				validator_pub_key String,
				updated_date Date
			) ENGINE = MergeTree(updated_date,(_id),8192)
			`

	dbSQL.MustExec(schemaCh_block_event)
	log("OK", "...очищена - block_event", "")

	/**************************************************************************/
	/**************************************************************************/
	/**************************************************************************/

	////////////////////////////////////////////////////////////////////////////
	// Таблица содержит основные данные о Транзакции
	delCh_trx := `DROP TABLE IF EXISTS trx`
	dbSQL.MustExec(delCh_trx)
	schemaCh_trx := `
			CREATE TABLE trx (
				hash String,
				raw_tx String,
				height_i32 UInt32,
				index_i32 UInt32,
				from_adrs String,
				nonce_i32 UInt32,
				gas_price_i32 UInt32,
				gas_coin String,
				gas_used_i32 UInt32,
				type UInt32,
				amount_bip_f32 Float32,
				payload String,
				tags_return Float32,
				tags_sellamnt Float32,
				code Int32,
				log String,
				updated_date Date
			) ENGINE = MergeTree(updated_date,(hash),8192)
			`
	dbSQL.MustExec(schemaCh_trx)
	log("OK", "...очищена - trx", "")

	////////////////////////////////////////////////////////////////////////////
	// Таблица специфичных данных Транзакции, зависит от типа транзакци
	delCh_trx_data := `DROP TABLE IF EXISTS trx_data`
	dbSQL.MustExec(delCh_trx_data)
	schemaCh_trx_data := `
			CREATE TABLE trx_data (
				hash String,
				coin String,
				to_adrs String,
				value_f32 Float32,
				coin_to_sell String,
				coin_to_buy String,
				value_to_sell_f32 Float32,
				value_to_buy_f32 Float32,
				name String,
				symbol String,
				constant_reserve_ratio UInt32,
				initial_amount_f32 Float32,
				initial_reserve_f32 Float32,
				address String,
				pub_key String,
				commission UInt32,
				stake_f32 Float32,
				raw_check String,
				proof String,
				coin_13a Array(String),
				to_13a Array(String),
				value_f32_13a Array(Float32),
				updated_date Date
			) ENGINE = MergeTree(updated_date,(hash),8192)
			`
	dbSQL.MustExec(schemaCh_trx_data)
	log("OK", "...очищена - trx_data", "")

	/**************************************************************************/
	/**************************************************************************/
	/**************************************************************************/

	////////////////////////////////////////////////////////////////////////////
	// Таблица содержит основные данные о Валидаторе/ноде
	//TODO: (CH: таблица имеет обновляемые реквизиты... информацию о ноде)
	// Оставляем: ReplacingMergeTree т.к. reward_address или owner_address
	// могут поменяться Tx
	delCh_nodes := `DROP TABLE IF EXISTS nodes`
	dbSQL.MustExec(delCh_nodes)
	schemaCh_nodes := `
			CREATE TABLE nodes (
				pub_key String,
				pub_key_min String,
				reward_address String,
				owner_address String,
				created DateTime,
				total_stake_f32 Float32,
				commission UInt32,
				created_at_block UInt32,
				updated_date Date
			) ENGINE = ReplacingMergeTree(updated_date, (pub_key), 8192)
			` // использовать: SELECT * FROM nodes FINAL;
	//FIXME: удалить total_stake_f32, переехал в MEM.DB
	dbSQL.MustExec(schemaCh_nodes)
	log("OK", "...очищена - nodes", "")

	////////////////////////////////////////////////////////////////////////////
	// УДАЛИТЬ: Брать с block_event (при db -> ClickHouse)
	// Таблица содержит историческую информацию о Валидаторе/ноде для него
	// значимых блоках (создание, пропуск блока, штраф, вкл/выкл)
	delCh_node_blockstory := `DROP TABLE IF EXISTS node_blockstory`
	dbSQL.MustExec(delCh_node_blockstory)
	schemaCh_node_blockstory := `
			CREATE TABLE node_blockstory (
				pub_key String,
				block_id UInt32,
				block_type String,
				updated_date Date
			) ENGINE = MergeTree(updated_date,(pub_key,block_id,block_type),8192)
			`
	dbSQL.MustExec(schemaCh_node_blockstory)
	log("OK", "...очищена - node_blockstory", "")

	////////////////////////////////////////////////////////////////////////////
	// Таблица стэйка валидатора в разрезе Делегатов и их Монет (так-же в
	// эквиваленте от базовой монеты)
	//(CH: таблица имеет обновляемые реквизиты... почти все параметры)
	delCh_node_stakes := `DROP TABLE IF EXISTS node_stakes`
	dbSQL.MustExec(delCh_node_stakes)
	schemaCh_node_stakes := `
			CREATE TABLE node_stakes (
				pub_key String,
				owner_address String,
				coin String,
				value_f32 Float32,
				bip_value_f32 Float32,
				updated_date Date
			) ENGINE = ReplacingMergeTree(updated_date, (pub_key,owner_address,coin), 8192)
			` // использовать: SELECT * FROM node_stakes FINAL;
	dbSQL.MustExec(schemaCh_node_stakes)
	log("OK", "...очищена - node_stakes", "")

	/**************************************************************************/
	/**************************************************************************/
	/**************************************************************************/

	////////////////////////////////////////////////////////////////////////////
	// Таблица содержит основные данные о кастомной Монете
	delCh_coin := `DROP TABLE IF EXISTS coins`
	dbSQL.MustExec(delCh_coin)
	schemaCh_coins := `
			CREATE TABLE coins (
				name String,
				symbol String,
				time DateTime,
				initial_amount_f32 Float32,
				initial_reserve_f32 Float32,
				constant_reserve_ratio UInt32,
				creator String,
				updated_date Date
			) ENGINE = MergeTree(updated_date, (symbol), 8192)
			`
	dbSQL.MustExec(schemaCh_coins)
	log("OK", "...очищена - coin", "")

	////////////////////////////////////////////////////////////////////////////
	// Таблица содержит данные о движение кастомных Монет {из транзакции} (ОСТАВИТЬ!!!)
	// некоторые поля расчетные и в дальнейшем облегчено получение данных для DEX
	delCh_coin_trx := `DROP TABLE IF EXISTS coin_trx`
	dbSQL.MustExec(delCh_coin_trx)
	schemaCh_coin_trx := `
			CREATE TABLE coin_trx (
				hash String,
				time DateTime,
				type UInt32,
				coin_to_buy String,
				coin_to_sell String,
				value_to_buy_f32 Float32,
				value_to_sell_f32 Float32,
				price_f32 Float32,
				volume_f32 Float32,
				updated_date Date
			) ENGINE=MergeTree(updated_date,(hash),8192)
			`
	dbSQL.MustExec(schemaCh_coin_trx)
	log("OK", "...очищена - coin_trx", "")

	/**************************************************************************/
	/**************************************************************************/
	/**************************************************************************/
}
