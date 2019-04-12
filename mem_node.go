package main

import (
	"fmt"
	"strconv"
	"time"

	s "github.com/ValidatorCenter/prs3r/strc"
	"github.com/go-redis/redis"
)

// Обновить информационную запись о Ноде (сокращенная)
func updNodeInfoRds_ext(db *redis.Client, dt *NodeExtInfo) bool {
	var err error

	if dt.PubKey != "" {
		m2 := map[string]interface{}{
			//"validator_address":    dt.WWW,
			"validator_name":       dt.Name,
			"validator_url":        dt.WWW,
			"validator_logo_img":   dt.Logo,
			"validator_desciption": dt.Descr,
		}
		err = db.HMSet(fmt.Sprintf("%s_info", dt.PubKey), m2).Err()
		if err != nil {
			log("ERR", fmt.Sprint("[mem_node.go] updNodeInfoRds_ext(hmset...", dt.PubKey, ") - ", err), "")
			return false
		}
	} else {
		log("ERR", "[mem_node.go] updNodeInfoRds_ext(...) PubKey = 0", "")
		return false
	}

	return true
}

// Поиск информации о Ноде
func srchNodeInfoRds(db *redis.Client, dt *s.NodeExt) bool {
	if dt.PubKey != "" {
		//db.HGetAll("").Val()
		_lbRes, err := db.HGetAll(fmt.Sprintf("%s_info", dt.PubKey)).Result()
		if err != nil {
			log("ERR", fmt.Sprint("[mem_node.go] updNodeInfoRds(hgetall...", dt.PubKey, ") - ", err), "")
			return false
		}
		// Всё "хорошо", заносим в dt новые данные
		dt.ValidatorAddress = _lbRes["validator_address"]
		dt.ValidatorName = _lbRes["validator_name"]
		dt.ValidatorURL = _lbRes["validator_url"]
		dt.ValidatorLogoImg = _lbRes["validator_logo_img"]
		dt.ValidatorDesc = _lbRes["validator_desciption"]
		dt_Uptime, _ := strconv.ParseFloat(_lbRes["uptime"], 32)
		dt.Uptime = float32(dt_Uptime)
		dt.StatusInt, _ = strconv.Atoi(_lbRes["status"])
		dt.TimeUpdate, _ = time.Parse(time.RFC3339, _lbRes["time_update"])
		//dt.AmnNoBlocks, _ = strconv.Atoi(_lbRes["amnt_blocks"]) // TODO: или dt.AmnBlocks???
		dt_AmntBlocks, _ := strconv.Atoi(_lbRes["amnt_blocks"])
		dt.AmntBlocks = uint64(dt_AmntBlocks)                    // всего подписано блоков
		dt.AmntSlashed, _ = strconv.Atoi(_lbRes["amnt_slashed"]) // всего штрафов
		dt_TotalStake, _ := strconv.ParseFloat(_lbRes["total_stake_f32"], 32)
		dt.TotalStake = float32(dt_TotalStake)
	} else {
		log("ERR", "[mem_node.go] srchNodeInfoRds(...) PubKey = 0", "")
		return false
	}

	return true
}

// Создать/обновить информационную запись о Ноде
func updNodeInfoRds(db *redis.Client, dt *s.NodeExt) bool {
	var err error

	if dt.PubKey != "" {
		m2 := map[string]interface{}{
			/*"validator_name":       dt.ValidatorName,
			"validator_url":        dt.ValidatorURL,
			"validator_logo_img":   dt.ValidatorLogoImg,
			"validator_desciption": dt.ValidatorDesc,*/
			"validator_address": dt.ValidatorAddress,
			"uptime":            dt.Uptime,
			"status":            dt.StatusInt,
			"time_update":       dt.TimeUpdate.Format("2006-01-02T15:04:05.999999-07:00"),
			"amnt_blocks":       dt.AmntBlocks,  // всего подписано блоков
			"amnt_slashed":      dt.AmntSlashed, // всего получено штрафов
			"total_stake_f32":   dt.TotalStake,
		}

		err = db.HMSet(fmt.Sprintf("%s_info", dt.PubKey), m2).Err()
		if err != nil {
			log("ERR", fmt.Sprint("[mem_node.go] updNodeInfoRds(hmset...", dt.PubKey, ") - ", err), "")
			return false
		}
	} else {
		log("ERR", "[mem_node.go] updNodeInfoRds(...) PubKey = 0", "")
		return false
	}

	return true
}
