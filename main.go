package main

import (
	"flag"
	"fmt"
	"net/http"
	"reverse_proxy/common"
	"reverse_proxy/logic"
)

func main() {
	//日志系统初始化
	str, err := common.Conf.GetValue("log", "path")
	if nil != err {
		fmt.Printf("fail to get log path")
		return
	}
	defer common.Start(common.LogFilePath(str), common.EveryHour).Stop()
	h := flag.Int("http", 0, "0:don't Start http  Server; 1:Start http  Server")
	ws := flag.Int("websocket", 0, "0:don't Start websocket  Server; 1:Start websocket  Server")
	flag.Parse()

	if 1 == *h && 1 == *ws {
		go StartHttpProxy()
		StartWebsocketProxy()
	} else if 1 == *h {
		StartHttpProxy()
	} else if 1 == *ws {
		StartWebsocketProxy()
	} else {
		fmt.Printf("Use parameters -h to get more help\n")
		common.Errorf("Use parameters -h to get more help")
	}
}

func StartHttpProxy() error {
	//反向代理处理器
	h := &logic.HandleProxy{}
	//监听端口
	addr, _ := common.Conf.GetValue("outhttp", "addr")
	err := http.ListenAndServe(addr, h)

	if err != nil {
		common.Errorf("ListenAndServe:%s ", err.Error())
	}
	return nil
}

func StartWebsocketProxy() {
	addr, _ := common.Conf.GetValue("outwebsocket", "addr")
	err := http.ListenAndServe(addr, logic.NewProxy())
	if err != nil {
		common.Errorf("%v", err)
	}
}
