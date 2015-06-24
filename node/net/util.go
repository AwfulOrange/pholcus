// socket通信
package net

import (
	"encoding/json"
	"github.com/henrylee2cn/pholcus/runtime/cache"
	"log"
)

// 备注：[]byte("^{}}\r\n") == []byte{94,123,125, 13, 10}

//编码通信数据
func marshal(data *cache.NetData) ([]byte, error) {
	b, err := json.Marshal(*data)
	b = append(b, byte(94))
	return b, err
}

//解码通信数据，包括粘包处理
func unmarshal(data []byte) (datas []*cache.NetData, err error) {
	offset := 0
	for k, v := range data {
		if v == byte(0) {
			offset++
			continue
		}
		if v == byte(94) {
			d := new(cache.NetData)
			err = json.Unmarshal(data[offset:k], d)
			datas = append(datas, d)
			offset = k + 1
		}
	}
	// log.Println("datas的长度:", len(datas), "错误：", err)
	return
}

func checkError(err error) {
	if err != nil {
		log.Printf("Fatal error: %s", err.Error())
	}
}
