package logic

import (
	"net/url"
	"reverse_proxy/common"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const HEARTTIME = 10000

var g_http_consistent *Consistent
var g_websocket_consistent *Consistent
var g_mapUrl map[string]*url.URL
var g_mapTimerTaskHttp map[string]int64
var g_mapTimerTaskWebsocket map[string]int64

const (
	Select_all_user = "查找全部用户"
	secret          = "test"
)

type Claims struct {
	//Appid string `json:"Appid"`
	// recommended having
	Userid int `json:"Userid"`
	jwt.StandardClaims
}

func init() {
	g_mapTimerTaskHttp = make(map[string]int64)
	g_mapTimerTaskWebsocket = make(map[string]int64)
	g_http_consistent = NewConsistent()
	strValues := common.Conf.GetKeyList("innerhttp")
	for _, value := range strValues {
		if strAddr, err := common.Conf.GetValue("innerhttp", value); nil == err {
			g_http_consistent.Add(strAddr)
			g_mapTimerTaskHttp[strAddr] = time.Now().Unix()
		}
	}

	g_mapUrl = make(map[string]*url.URL)
	g_websocket_consistent = NewConsistent()
	strValues = common.Conf.GetKeyList("innerwebsocket")
	for _, value := range strValues {
		if strAddr, err := common.Conf.GetValue("innerwebsocket", value); nil == err {
			u, err1 := url.Parse("ws://" + strAddr)
			if err1 != nil {
				common.Errorf("%v", err1)
				return
			}
			g_websocket_consistent.Add(strAddr)
			g_mapUrl[strAddr] = u
			g_mapTimerTaskWebsocket[strAddr] = time.Now().Unix()
		}
	}

	go TimerTask()
}

func Token_auth(signedToken string) (int, error) {
	token, err := jwt.ParseWithClaims(signedToken, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		//fmt.Printf("%v %v", claims.Username, claims.StandardClaims.ExpiresAt)
		//fmt.Println(reflect.TypeOf(claims.StandardClaims.ExpiresAt))
		//return claims.Appid, err
		return claims.Userid, err
	}
	return 0, err
}

//处理请求的host、路径和配置，返回实际请求host和ip
func GetHttpRouteServer(strRouteKey string) (strInner string, err error) {
	return g_http_consistent.Get(strRouteKey)
}

//处理请求的host、路径和配置，返回实际请求host和ip
func GetWebsocketRouteServer(strRouteKey string) (strInner string, err error) {
	return g_websocket_consistent.Get(strRouteKey)
}

func TimerTask() {
	t1 := time.NewTimer(time.Second * HEARTTIME)

	for {
		select {
		case <-t1.C:
			tempTime := time.Now().Unix()
			for key := range g_mapTimerTaskHttp {
				if (tempTime - g_mapTimerTaskHttp[key]) > HEARTTIME {
					g_http_consistent.Remove(key)
				}
			}

			for key := range g_mapTimerTaskWebsocket {
				if (tempTime - g_mapTimerTaskWebsocket[key]) > HEARTTIME {
					g_websocket_consistent.Remove(key)
				}
			}
			t1.Reset(time.Second * HEARTTIME)
		}
	}
}
