package strc

import (
	"time"

	ms "github.com/ValidatorCenter/minter-go-sdk"
	//"github.com/satori/go.uuid"
)

/*
Причина не использования стандартного с SDK:
-транзакции не хранить в блоке
-в блоке хранить только ссылку в виде хэша на транзакцию
*/

type BlockResponse2 struct {
	Hash          string                       `json:"hash" bson:"hash" gorm:"hash" db:"hash"`
	Height        int                          `json:"height_i32" bson:"height_i32" gorm:"height_i32" db:"height_i32"`
	Time          time.Time                    `json:"time" bson:"time" gorm:"time" db:"time"`
	NumTxs        int                          `json:"num_txs_i32" bson:"num_txs_i32" gorm:"num_txs_i32" db:"num_txs_i32"`
	TotalTxs      int                          `json:"total_txs_i32" bson:"total_txs_i32" gorm:"total_txs_i32" db:"total_txs_i32"`
	Transactions  []TransResponseMin           `json:"transactions" bson:"transactions" gorm:"-" db:"-"` // НЕ для SQL
	TransHashArr  []string                     `json:"-" bson:"-" gorm:"transactions" db:"transactions"` // для SQL
	BlockReward   float32                      `json:"block_reward_f32" bson:"block_reward_f32" gorm:"block_reward_f32" db:"block_reward_f32"`
	Size          int                          `json:"size_i32" bson:"size_i32" gorm:"size_i32" db:"size_i32"`
	Validators    []ms.BlockValidatorsResponse `json:"validators" bson:"validators" gorm:"-" db:"-"`
	ValidatorAmnt int                          `json:"valid_amnt" db:"valid_amnt"`
	Proposer      string                       `json:"proposer" bson:"proposer" gorm:"proposer" db:"proposer"` // PubKey пропозер блока
	Events        []ms.BlockEventsResponse     `json:"events" bson:"events" gorm:"-" db:"-"`
	UpdYCH        string                       `json:"-" bson:"-" db:"updated_date"` // ClickHouse::UpdateDate
}

// для списка содержимого блока сойдет, можно еще меньше!
type TransResponseMin struct {
	Hash    string  `json:"hash" bson:"hash" db:"hash"`
	From    string  `json:"from" bson:"from" db:"from_adrs"`
	Type    int     `json:"type" bson:"type" db:"type"` // тип транзакции
	TypeTxt string  `-`                                 // для вывода в вебе webd
	Code    int     `json:"code" bson:"code" db:"code"` // если не 0, то ОШИБКА, читаем лог(Log)
	Amount  float32 `json:"amount_bip_f32" bson:"amount_bip_f32" db:"amount_bip_f32"`
}

// Структура - одного события в блоке (для SQL)
type BlockEvent struct {
	ID              string  `db:"_id"` // TODO: добавлен по необходимости
	Height          int     `db:"height_i32"`
	Type            string  `db:"type"`
	Role            string  `db:"role"` //DAO,Developers,Validator,Delegator
	Address         string  `db:"address"`
	Amount          float32 `db:"amount_f32"`
	Coin            string  `db:"coin"`
	ValidatorPubKey string  `db:"validator_pub_key"`
	UpdYCH          string  `db:"updated_date"` // ClickHouse::UpdateDate
}

// Пользователи валидаторов с новой комиссией
type NodeUserX struct {
	PubKey     string    `json:"pub_key" db:"pub_key"`       // мастернода
	Address    string    `json:"address" db:"address"`       // адрес кошелька X
	Start      time.Time `json:"start" db:"start"`           // дата старта
	Finish     time.Time `json:"finish" db:"finish"`         // дата финиша
	Commission int       `json:"commission" db:"commission"` // новая ставка комиссии
	UpdYCH     string    `json:"-" db:"updated_date"`        // ClickHouse::UpdateDate
}

// Задачи для исполнения ноде
type NodeTodo struct {
	//ID         uuid.UUID `json:"_id" db:"_id"`
	ID         string    `json:"_id" db:"_id"`
	Priority   int       `json:"priority" db:"priority"`     // от 0 до макс! главные:(0)?, (1)возврат делегатам,(2) на возмещение штрафов,(3) оплата сервера, на развитие, (4) распределние между соучредителями
	Done       bool      `json:"done" db:"done"`             // выполнено
	Created    time.Time `json:"created" db:"created"`       // создана time
	DoneT      time.Time `json:"donet" db:"donet"`           // выполнено time
	Type       string    `json:"type" db:"type"`             // тип задачи: SEND-CASHBACK,...
	Height     uint32    `json:"height_i32" db:"height_i32"` // блок
	PubKey     string    `json:"pub_key" db:"pub_key"`       // мастернода
	PubKeyMin  string    `-`                                 // мастернода (сокращенная)
	Address    string    `json:"address" db:"address"`       // адрес кошелька X
	AddressMin string    `-`                                 // адрес кошелька X (сокращенная)
	Amount     float32   `json:"amount_f32" db:"amount_f32"` // сумма
	Comment    string    `json:"comment" db:"comment"`       // комментарий
	TxHash     string    `json:"tx_hash" db:"tx_hash"`       // транзакция исполнения
	TxHashMin  string    `-`                                 // транзакция исполнения (сокращенная)
	UpdYCH     string    `json:"-" db:"updated_date"`        // ClickHouse::UpdateDate
}
