package strc

// TODO: https://github.com/globalsign/mgo

import (
	"time"
)

////////////////////////////////////////////////////////////////////////////////
type Tx1SendData struct {
	Coin  string  `json:"coin" bson:"coin" gorm:"coin" db:"coin"`
	To    string  `json:"to" bson:"to" gorm:"to" db:"to"`
	Value float32 `json:"value_f32" bson:"value_f32" gorm:"value_f32" db:"value_f32"`
}

type Tx2SellCoinData struct {
	CoinToSell  string  `json:"coin_to_sell" bson:"coin_to_sell" gorm:"coin_to_sell" db:"coin_to_sell"`
	CoinToBuy   string  `json:"coin_to_buy" bson:"coin_to_buy" gorm:"coin_to_buy" db:"coin_to_buy"`
	ValueToSell float32 `json:"value_to_sell_f32" bson:"value_to_sell_f32" gorm:"value_to_sell_f32" db:"value_to_sell_f32"`
}

type Tx3SellAllCoinData struct {
	CoinToSell string `json:"coin_to_sell" bson:"coin_to_sell" gorm:"coin_to_sell" db:"coin_to_sell"`
	CoinToBuy  string `json:"coin_to_buy" bson:"coin_to_buy" gorm:"coin_to_buy" db:"coin_to_buy"`
}

type Tx4BuyCoinData struct {
	CoinToBuy  string  `json:"coin_to_buy" bson:"coin_to_buy" gorm:"coin_to_buy" db:"coin_to_buy"`
	CoinToSell string  `json:"coin_to_sell" bson:"coin_to_sell" gorm:"coin_to_sell" db:"coin_to_sell"`
	ValueToBuy float32 `json:"value_to_buy_f32" bson:"value_to_buy_f32" gorm:"value_to_buy_f32" db:"value_to_buy_f32"`
}

type Tx5CreateCoinData struct {
	Name                 string  `json:"name" bson:"name" gorm:"name" db:"name"`
	CoinSymbol           string  `json:"symbol" bson:"symbol" gorm:"symbol" db:"symbol"`
	ConstantReserveRatio int     `json:"constant_reserve_ratio" bson:"constant_reserve_ratio" gorm:"constant_reserve_ratio" db:"constant_reserve_ratio"`
	InitialAmount        float32 `json:"initial_amount_f32" bson:"initial_amount_f32" gorm:"initial_amount_f32" db:"initial_amount_f32"`
	InitialReserve       float32 `json:"initial_reserve_f32" bson:"initial_reserve_f32" gorm:"initial_reserve_f32" db:"initial_reserve_f32"`
}

type Tx6DeclareCandidacyData struct {
	Address    string  `json:"address" bson:"address" gorm:"address" db:"address"`
	PubKey     string  `json:"pub_key" bson:"pub_key" gorm:"pub_key" db:"pub_key"`
	Commission int     `json:"commission" bson:"commission" gorm:"commission" db:"commission"`
	Coin       string  `json:"coin" bson:"coin" gorm:"coin" db:"coin"`
	Stake      float32 `json:"stake_f32" bson:"stake_f32" gorm:"stake_f32" db:"stake_f32"`
}

type Tx7DelegateDate struct {
	PubKey string  `json:"pub_key" bson:"pub_key" gorm:"pub_key" db:"pub_key"`
	Coin   string  `json:"coin" bson:"coin" gorm:"coin" db:"coin"`
	Stake  float32 `json:"stake_f32" bson:"stake_f32" gorm:"stake_f32" db:"stake_f32"`
}

type Tx8UnbondData struct {
	PubKey string  `json:"pub_key" bson:"pub_key" gorm:"pub_key" db:"pub_key"`
	Coin   string  `json:"coin" bson:"coin" gorm:"coin" db:"coin"`
	Value  float32 `json:"value_f32" bson:"value_f32" gorm:"value_f32" db:"value_f32"`
}

type Tx9RedeemCheckData struct {
	RawCheck string `json:"raw_check" bson:"raw_check" gorm:"raw_check" db:"raw_check"`
	Proof    string `json:"proof" bson:"proof" gorm:"proof" db:"proof"`
}

type Tx10SetCandidateOnData struct {
	PubKey string `json:"pub_key" bson:"pub_key" gorm:"pub_key" db:"pub_key"`
}

type Tx11SetCandidateOffData struct {
	PubKey string `json:"pub_key" bson:"pub_key" gorm:"pub_key" db:"pub_key"`
}

type Tx12CreateMultisigData struct {
	/**
	Threshold uint
	Weights   []uint
	Addresses [][20]byte
	**/
}

type Tx13MultisendData struct {
	// FIXME: Нет кода и идей по хранению в базе DB!!! Исправить это!
	List []Tx1SendData `json:"list" bson:"list" gorm:"-" db:"-"`
}

////////////////////////////////////////////////////////////////////////////////
// Информация о монете
type CoinMarketCapData struct {
	Name                 string    `json:"name" db:"name"`     // название монеты
	CoinSymbol           string    `json:"symbol" db:"symbol"` // символ монеты
	CoinURL              string    `json:"coin_url" db:"coin_url"`
	CoinLogoImg          string    `json:"coin_logo_img" db:"coin_logo_img"`
	CoinDesc             string    `json:"coin_desciption" db:"coin_desciption"`
	Time                 time.Time `json:"time" db:"time"`               // дата создания
	TimeUpdate           time.Time `json:"time_update" db:"time_update"` // UPDATE дата последнего обновления
	InitialAmount        float32   `json:"initial_amount_f32" db:"initial_amount_f32"`
	InitialReserve       float32   `json:"initial_reserve_f32" db:"initial_reserve_f32"`
	ConstantReserveRatio int       `json:"constant_reserve_ratio" db:"constant_reserve_ratio"`   // uint, should be from 10 to 100 (в %).
	VolumeNow            float32   `json:"volume_now_f32" db:"volume_now_f32"`                   // UPDATE
	ReserveBalanceNow    float32   `json:"reserve_balance_now_f32" db:"reserve_balance_now_f32"` // UPDATE
	Creator              string    `json:"creator" db:"creator"`
	AmntTrans24x7        int       `json:"amnt_trans_24x7" db:"amnt_trans_24x7"` // UPDATE количество транзакций за последние 7дней (TODO: относительно MNT)
	// UPDATE капитализация
	// всего монет выпущено
	// UPDATE ликвидность движения всего купли/продажи монет за сутки
	Transactions []CoinActionpData `json:"transactions" db:"-"` // движения
	UpdYCH       string            `json:"-" db:"updated_date"` // ClickHouse::UpdateDate
}

// Движение монеты
type CoinActionpData struct {
	Hash        string    `json:"hash" db:"hash"`                           // Хэш транзакции
	Time        time.Time `json:"time" db:"time"`                           // дата движения
	Type        int       `json:"type" db:"type"`                           // type: продажа или покупка
	CoinToBuy   string    `json:"coin_to_buy" db:"coin_to_buy"`             // Монета покупки
	CoinToSell  string    `json:"coin_to_sell" db:"coin_to_sell"`           // Монета продажи
	ValueToBuy  float32   `json:"value_to_buy_f32" db:"value_to_buy_f32"`   // Количество покупки
	ValueToSell float32   `json:"value_to_sell_f32" db:"value_to_sell_f32"` // Количество продано
	Price       float32   `json:"price_f32" db:"price_f32"`                 // Цена
	Volume      float32   `json:"volume_f32" db:"volume_f32"`               // Объем
	UpdYCH      string    `json:"-" db:"updated_date"`                      // ClickHouse::UpdateDate
}

// Информация о паре монет
type PairCoins struct {
	CoinToBuy  string  `json:"coin_to_buy" db:"coin_to_buy"`
	CoinToSell string  `json:"coin_to_sell" db:"coin_to_sell"`
	PriceBuy   float32 `json:"price_buy_f32" db:"price_buy_f32"`
	PriceSell  float32 `json:"price_sell_f32" db:"price_sell_f32"`
	Volume24   float32 `json:"volume_24_f32" db:"volume_24_f32"` // Объем за 24ч
	Change24   float32 `json:"change_24_f32" db:"change_24_f32"` // Изменение за 24ч
	//TODO: добавить Динамика курса {За день | За месяц | За 6 месяцев | За год | За всё время}
	TimeUpdate time.Time `json:"time_update" db:"time_update"` // дата последнего обновления
	// то что заносится в базу не будет:
	PriceBuyOld  float32 `-`
	NewPairCoins bool    `-`
}
