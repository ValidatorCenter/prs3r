package strc

// Статус базы данных
type StatusDB struct {
	LatestBlockSave int `bson:"latest_block_save" db:"latest_block_save"` // Загруженный блок с базы Minter
	LatestBlockCMC  int `bson:"latest_block_cmc" db:"latest_block_cmc"`   // Обработано блоков для CoinMarketCap info
	LatestBlockVld  int `bson:"latest_block_vld" db:"latest_block_vld"`   // Обработано блоков для Validators info
	LatestBlockRwd  int `bson:"latest_block_rwd" db:"latest_block_rwd"`   // Обработано блоков для Reward-Events
}

// Настройки Автоделегирования для кошелька
type AutodelegCfg struct {
	Address   string `json:"address" db:"address"`
	PubKey    string `json:"pub_key" db:"pub_key"`
	Coin      string `json:"coin" db:"coin"`
	WalletPrc int    `json:"wallet_prc" db:"wallet_prc"`
}
