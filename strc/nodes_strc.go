package strc

import (
	"time"
	//ms "github.com/ValidatorCenter/minter-go-sdk"
)

/*
// ИСПОЛЬЗОВАНИЕ: oneNodeX := s.NodeExt{CandidateInfo: oneNode}
// Структура расширяющая функционал структуры из SDK - CandidateInfo
type NodeExt struct {
	ms.CandidateInfo
	TimeUpdate time.Time `bson:"time_update"` // UPDATE дата последнего обновления
}*/

// стэк кандидата/валидатора в каких монетах
type StakesInfo struct {
	Owner    string  `json:"owner" bson:"owner" gorm:"owner" db:"owner"`
	Coin     string  `json:"coin" bson:"coin" gorm:"coin" db:"coin"`
	Value    float32 `json:"value_f32" bson:"value_f32" gorm:"value_f32" db:"value_f32"`
	BipValue float32 `json:"bip_value_f32" bson:"bip_value_f32" gorm:"bip_value_f32" db:"bip_value_f32"`
}

// Информация о валидаторе (доработанная, расширенная)
type NodeExt struct {
	RatingID         int           `-`
	ValidatorName    string        `json:"validator_name" db:"validator_name"`
	ValidatorURL     string        `json:"validator_url" db:"validator_url"`
	ValidatorLogoImg string        `json:"validator_logo_img" db:"validator_logo_img"`
	ValidatorDesc    string        `json:"validator_desciption" db:"validator_desciption"`
	Uptime           float32       `json:"uptime" db:"uptime"`
	Created          time.Time     `json:"created" db:"created"`                     // дата создания
	Age              string        `-`                                               // дельта текущая и когда создан в виде строки
	RewardAddress    string        `json:"reward_address" db:"reward_address"`       // Адрес кошелька "Mx..." вознаграждения
	OwnerAddress     string        `json:"owner_address" db:"owner_address"`         // Адрес кошелька "Mx..." основной
	TotalStake       float32       `json:"total_stake_f32" db:"total_stake_f32"`     // Полный стэк
	PubKey           string        `json:"pub_key" db:"pub_key"`                     // паблик ноды "Mp..."
	PubKeyMin        string        `json:"pub_key_min" db:"pub_key_min"`             // короткий паблик ноды "Mp..."
	ValidatorAddress string        `json:"validator_address" db:"validator_address"` // Адрес ноды, в основном в транзакциях
	Commission       int           `json:"commission" db:"commission"`               // комиссия
	CreatedAtBlock   int           `json:"created_at_block" db:"created_at_block"`   // блок создания
	StatusInt        int           `json:"status" db:"status"`                       // числовое значение статуса: 1 - Offline, 2 - Online, 77 - Validator
	TimeUpdate       time.Time     `json:"time_update" db:"time_update"`             // UPDATE дата последнего обновления
	Stakes           []StakesInfo  `json:"stakes" db:"-"`                            // Только у: Candidate(по PubKey)
	Blocks           []BlocksStory `json:"blocks_story" db:"-"`                      // важные исторические блоки
	AmntBlocks       uint64        `json:"amnt_blocks" db:"amnt_blocks"`             // количество "всего"(с пропущенными) блоков в которых является подписантом
	AmntSlashed      int           `json:"amnt_slashed" db:"amnt_slashed"`           // количество штрафов, полный список по запросу из транзакций
	AmnNoBlocks      int           `-`                                               // количество пропущенных блоков
	AmntSlots        int           `json:"amnt_slots" db:"amnt_slots"`               // количество занятых слотов это количество уникальных объектов в Stakes
	UpdYCH           string        `json:"-" db:"updated_date"`                      // ClickHouse::UpdateDate
}

type BlocksStory struct {
	ID   uint32 `json:"block_id" db:"block_id"`
	Type string `json:"block_type" db:"block_type"` //создание, пропуск, штраф, старт, стоп //TODO: оформить типом!
}
