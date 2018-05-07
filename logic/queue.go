package logic

import (
	"encoding/json"
	"fmt"
	"reverse_proxy/common"
	"time"
	"user/models/myredis"

	"github.com/gomodule/redigo/redis"
)

const (
	MQ_CMD_RELATION = 20001 //代理心跳
	WAIT_TIME       = 60
	INVAILD         = 0
	VAILD           = 1
)

const (
	MQ_OBJ_USER     = 1 //用户系统
	MQ_OBJ_RELATION = 2 //关系服务系统
	MQ_OBJ_MESSAGE  = 3 //消息系统
	MQ_OBJ_RELAY    = 4 //转发系统
	MQ_OBJ_PROXY    = 5 //代理系统
)

var G_IsProxyHttp int
var G_IsProxyWebsocket int

//初始化队列消息
func InitQueue(isProxyHttp int, isProxyWebsocket int) {
	G_IsProxyHttp = isProxyHttp
	G_IsProxyWebsocket = isProxyWebsocket
	consumer()
}

//tcp协议结构体
const (
	OP_PROXY_HEART_REQ = int32(5001)
	OP_PROXY_HEART_RSP = int32(5002)
)

type StMessageHead struct {
	Version int16           `json:"ver" valid:"Required"`
	Cmd     int32           `json:"cmd" valid:"Required"`
	Seq     int32           `json:"seq" valid:"Required"`
	Body    json.RawMessage `json:"body"`
}

type StHeartReq struct {
	Ip   string `json:"ip" valid:"Required"`
	Port int32  `json:"port" valid:"Required"`
}

//消费redis队列消息
func consumer() {
	strKey := fmt.Sprintf("%d:%d:%d:%d", myredis.REDIS_T_LIST, MQ_CMD_RELATION, MQ_OBJ_RELATION, MQ_OBJ_MESSAGE)
	for {
		conn := myredis.GetQueueConn()
		if conn.Err() != nil {
			common.Errorf(conn.Err().Error())
			continue
		}

		strMember, err := redis.Values(conn.Do("BRPOP", strKey, WAIT_TIME))
		if err != nil {
			if err != redis.ErrNil {
				common.Errorf("error: %v", err)
			}
			conn.Close()
			continue
		}

		if len(strMember) != 2 {
			common.Errorf("error: param wrong")
			conn.Close()
			continue
		}
		conn.Close()

		p := StMessageHead{}
		if err := unpack(strMember[1].([]byte), &p); err != nil {
			continue
		}

		switch p.Cmd {
		case OP_PROXY_HEART_REQ:
			ProcessHeart(&p)
		}
	}
}

//解包消息包
func unpack(sData []byte, p *StMessageHead) (err error) {
	if p == nil || len(sData) == 0 {
		common.Errorf("msg queue data invalid")
		return
	}

	if err = json.Unmarshal(sData, p); err != nil {
		common.Errorf("err=%s", err.Error())
		return
	}
	return
}

func ProcessHeart(p *StMessageHead) {
	msg := StHeartReq{}
	var code int
	var err error
	if err = json.Unmarshal(p.Body, &msg); err != nil {
		common.Errorf("err=%s code=%d", err.Error(), code)
		return
	}

	if "" == msg.Ip || msg.Port < 1 {
		common.Errorf("parameter error")
		return
	}

	var str string
	str = fmt.Sprintf("%s::%d", msg.Ip, msg.Port)
	if VAILD == G_IsProxyHttp {
		if strTemp, _ := g_http_consistent.Get(str); "" == strTemp {
			g_http_consistent.Add(str)
		}
		g_mapTimerTaskHttp[str] = time.Now().Unix()
	}

	if VAILD == G_IsProxyWebsocket {
		if strTemp, _ := g_websocket_consistent.Get(str); "" == strTemp {
			g_websocket_consistent.Add(str)
		}
		g_mapTimerTaskWebsocket[str] = time.Now().Unix()
	}
}
