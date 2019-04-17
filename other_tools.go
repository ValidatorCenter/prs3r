package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
)

// Вывод служебного сообщения
func log(tp string, msg1 string, msg2 interface{}) {
	timeClr := fmt.Sprintf(color.MagentaString("[%s]"), time.Now().Format("2006-01-02 15:04:05"))
	msg0 := ""
	if tp == "ERR" {
		msg0 = fmt.Sprintf(color.RedString("ERROR: %s"), msg1)
	} else if tp == "WRN" {
		msg0 = fmt.Sprintf(color.HiYellowString("WARRNING: %s"), msg1)
	} else if tp == "INF" {
		infTag := fmt.Sprintf(color.YellowString("%s"), msg1)
		msg0 = fmt.Sprintf("%s: %#v", infTag, msg2)
	} else if tp == "OK" {
		msg0 = fmt.Sprintf(color.GreenString("%s"), msg1)
	} else if tp == "STR" {
		msg0 = fmt.Sprintf(color.CyanString("%s"), msg1)
	} else {
		msg0 = msg1
	}
	fmt.Printf("%s %s\n", timeClr, msg0)
}

// Сокращение длинных строк
func getMinString(bigStr string) string {
	if len(bigStr) > 8 {
		return fmt.Sprintf("%s...%s", bigStr[:6], bigStr[len(bigStr)-4:len(bigStr)])
	} else {
		fmt.Println("WARRNING: getMinString(", bigStr, ")")
		return bigStr
	}
}

// Замена в строке по мап-интерфейсу
func mapReplace(str1 string, map1 map[string]interface{}) (string, error) {
	for iS, dM := range map1 {
		switch t := dM.(type) {
		case bool:
			str1 = strings.Replace(str1, fmt.Sprintf(":%s,", iS), fmt.Sprintf("%t,", dM), -1)
			str1 = strings.Replace(str1, fmt.Sprintf(":%s\n", iS), fmt.Sprintf("%t\n", dM), -1)
		case int, int8, int32, int64, uint, uint32, uint64:
			str1 = strings.Replace(str1, fmt.Sprintf(":%s,", iS), fmt.Sprintf("%d,", dM), -1)
			str1 = strings.Replace(str1, fmt.Sprintf(":%s\n", iS), fmt.Sprintf("%d\n", dM), -1)
		case string:
			str1 = strings.Replace(str1, fmt.Sprintf(":%s,", iS), fmt.Sprintf("'%s',", dM), -1)
			str1 = strings.Replace(str1, fmt.Sprintf(":%s\n", iS), fmt.Sprintf("'%s'\n", dM), -1)
		case float32, float64:
			str1 = strings.Replace(str1, fmt.Sprintf(":%s,", iS), fmt.Sprintf("%f,", dM), -1)
			str1 = strings.Replace(str1, fmt.Sprintf(":%s\n", iS), fmt.Sprintf("%f\n", dM), -1)
		default:
			return "", errors.New(fmt.Sprintf("Don't know type %T", t))
		}
	}

	return str1, nil
}
