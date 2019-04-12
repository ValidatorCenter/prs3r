package main

import (
	"fmt"
	"strconv"
	"time"

	s "github.com/ValidatorCenter/prs3r/strc"
	"github.com/go-redis/redis"
)

// Создать/обновить информационную запись о Монете
func updCoinInfoRds(db *redis.Client, dt *s.CoinMarketCapData) bool {
	if dt.CoinSymbol != "" {
		m2 := map[string]interface{}{
			"coin_url":                dt.CoinURL,
			"coin_logo_img":           dt.CoinLogoImg,
			"coin_desciption":         dt.CoinDesc,
			"time_update":             dt.TimeUpdate.Format("2006-01-02T15:04:05.999999-07:00"),
			"volume_now_f32":          dt.VolumeNow,
			"reserve_balance_now_f32": dt.ReserveBalanceNow,
			"amnt_trans_24x7":         dt.AmntTrans24x7,
		}

		err := db.HMSet(fmt.Sprintf("%s_info", dt.CoinSymbol), m2).Err()
		if err != nil {
			log("ERR", fmt.Sprint("[mem_coin.go] updCoinInfoRds(hmset...", dt.CoinSymbol, ") - ", err), "")
			return false
		}
	} else {
		log("ERR", "[mem_coin.go] updCoinInfoRds(...) Coin = '???'", "")
		return false
	}

	log("INF", "INSERT/UPDATE", fmt.Sprintf("%s_info", dt.CoinSymbol))
	return true
}

// Обновить информацию о 3-х View записях в Монете (coin_url, coin_logo_img, coin_desciption)
func updCoinInfoRds_3v(db *redis.Client, dt *s.CoinMarketCapData) bool {
	if dt.CoinSymbol != "" {
		m2 := map[string]interface{}{
			"coin_url":        dt.CoinURL,
			"coin_logo_img":   dt.CoinLogoImg,
			"coin_desciption": dt.CoinDesc,
		}

		err := db.HMSet(fmt.Sprintf("%s_info", dt.CoinSymbol), m2).Err()
		if err != nil {
			log("ERR", fmt.Sprint("[mem_coin.go] updCoinInfoRds_3v(hmset...", dt.CoinSymbol, ") - ", err), "")
			return false
		}
	} else {
		log("ERR", "[mem_coin.go] updCoinInfoRds_3v(...) Coin = '???'", "")
		return false
	}
	log("INF", "UPDATE", fmt.Sprintf("%s_info(3v)", dt.CoinSymbol))
	return true
}

// Поиск информации о Монете
func srchCoinInfoRds(db *redis.Client, dt *s.CoinMarketCapData) bool {
	if dt.CoinSymbol != "" {

		_lbRes, err := db.HGetAll(fmt.Sprintf("%s_info", dt.CoinSymbol)).Result()
		if err != nil {
			log("ERR", fmt.Sprint("[mem_coin.go] srchCoinInfoRds(hgetall...", dt.CoinSymbol, ") - ", err), "")
			return false
		}
		// Всё "хорошо", заносим в dt новые данные
		dt.CoinURL = _lbRes["coin_url"]
		dt.CoinLogoImg = _lbRes["coin_logo_img"]
		dt.CoinDesc = _lbRes["coin_desciption"]
		dt.TimeUpdate, _ = time.Parse(time.RFC3339, _lbRes["time_update"])
		dt_VolumeNow, _ := strconv.ParseFloat(_lbRes["volume_now_f32"], 32)
		dt.VolumeNow = float32(dt_VolumeNow)
		dt_ReserveBalanceNow, _ := strconv.ParseFloat(_lbRes["reserve_balance_now_f32"], 32)
		dt.ReserveBalanceNow = float32(dt_ReserveBalanceNow)
		dt.AmntTrans24x7, _ = strconv.Atoi(_lbRes["amnt_trans_24x7"])

	} else {
		log("ERR", "[mem_coin.go] srchCoinInfoRds(...)  = '???'", "")
		return false
	}

	return true
}

// Обновить данные о паре-монет (redis)
func updCoin2Redis(db *redis.Client, dt *s.PairCoins) bool {
	if dt.CoinToBuy != "" && dt.CoinToSell != "" {
		m2 := map[string]interface{}{
			"coin_to_buy":    dt.CoinToBuy,
			"coin_to_sell":   dt.CoinToSell,
			"price_buy_f32":  dt.PriceBuy,
			"price_sell_f32": dt.PriceSell,
			"volume_24_f32":  dt.Volume24,
			"change_24_f32":  dt.Change24,
			"time_update":    dt.TimeUpdate.Format("2006-01-02T15:04:05.999999-07:00"),
		}
		err := db.HMSet(fmt.Sprintf("%s%s_c2c", dt.CoinToBuy, dt.CoinToSell), m2).Err()
		if err != nil {
			log("ERR", fmt.Sprint("[mem_coin.go] updCoin2Redis(hmset...", dt.CoinToBuy, ") - ", err), "")
			return false
		}
	} else {
		log("ERR", "[mem_coin.go] updCoin2Redis(...) Coin_1 = '???' & Coin_2 = '???'", "")
		return false
	}

	log("INF", "INSERT/UPDATE", fmt.Sprintf("%s%s_c2c", dt.CoinToBuy, dt.CoinToSell))
	return true
}

// Обновить информацию о 3-х записях в Монете (time_update, volume_now_f32, reserve_balance_now_f32)
func updCoinInfoRds_3(db *redis.Client, dt *s.CoinMarketCapData) bool {
	if dt.CoinSymbol != "" {
		m2 := map[string]interface{}{
			"time_update":             dt.TimeUpdate.Format("2006-01-02T15:04:05.999999-07:00"),
			"volume_now_f32":          dt.VolumeNow,
			"reserve_balance_now_f32": dt.ReserveBalanceNow,
		}

		err := db.HMSet(fmt.Sprintf("%s_info", dt.CoinSymbol), m2).Err()
		if err != nil {
			log("ERR", fmt.Sprint("[mem_coin.go] updCoinInfoRds_3(hmset...", dt.CoinSymbol, ") - ", err), "")
			return false
		}
	} else {
		log("ERR", "[mem_coin.go] updCoinInfoRds_3(...) Coin = '???'", "")
		return false
	}
	log("INF", "UPDATE", fmt.Sprintf("%s_info(3)", dt.CoinSymbol))
	return true
}
