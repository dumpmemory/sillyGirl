package wx

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/astaxie/beego/logs"
	"github.com/cdle/sillyGirl/core"
	"github.com/gorilla/websocket"
)

var c *websocket.Conn

func enableVLW() {
	addr := wx.Get("vlw_addr")
	if addr == "" {
		return
	}
	defer func() {
		time.Sleep(time.Second * 2)
		enableVLW()
	}()
	u := url.URL{Scheme: "ws", Host: addr, Path: "/"}
	logs.Info("连接vlw %s", u.String())
	var err error
	c, _, err = websocket.DefaultDialer.Dial(u.String(), http.Header{})
	if err != nil {
		logs.Warn("连接vlw错误:", err)
		return
	}
	defer c.Close()
	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				logs.Info("read:", err)
				return
			}
			type AutoGenerated struct {
				SdkVer  int    `json:"sdkVer"`
				Event   string `json:"Event"`
				Content struct {
					RobotWxid     string `json:"robot_wxid"`
					Type          int    `json:"type"`
					FromGroup     string `json:"from_group"`
					FromGroupName string `json:"from_group_name"`
					FromWxid      string `json:"from_wxid"`
					FromName      string `json:"from_name"`
					Msg           string `json:"msg"`
					MsgSource     struct {
						Atuserlist []struct {
							Wxid         string `json:"wxid"`
							Nickname     string `json:"nickname"`
							PositionFrom int    `json:"position_from"`
							PositionTo   int    `json:"position_to"`
						} `json:"atuserlist"`
					} `json:"msg_source"`
					Clientid  int `json:"clientid"`
					RobotType int `json:"robot_type"`
				} `json:"content"`
				WsMCBreqID int `json:"wsMCBreqID"`
			}
			ag := &AutoGenerated{}
			json.Unmarshal(message, ag)
			// w, err := c.NextWriter(websocket.TextMessage)
			// if err != nil {
			// 	continue
			// }
			// w.Write([]byte(fmt.Sprintf(`{"wsMCBreqID": %d, "Code": -1}`, ag.WsMCBreqID)))
			// w.Close()
			c.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(`{"wsMCBreqID": %d, "Code": -1, "cao": "sbsbsbsbsbsbsbsbsbsbsbsbsbsbsbsbbsbsbsbssbsbsbsbsbbsbsbsbsbsbsbbsbsbsbsbsbsb"}`, ag.WsMCBreqID)))
			// c.WriteJSON(map[string]interface{}{
			// 	"wsMCBreqID": ag.WsMCBreqID,
			// 	"Code":       -1,
			// })
			if ag.Event == "EventPrivateChat" || ag.Event == "EventGroupChat" {
				wm := wxmsg{}
				wm.content = ag.Content.Msg
				wm.user_id = ag.Content.FromWxid
				wm.user_name = ag.Content.FromName
				if ag.Content.FromGroup != "" {
					wm.chat_id = core.Int(strings.Replace(ag.Content.FromGroup, "@chatroom", "", -1))
				}
				if robot_wxid != ag.Content.RobotWxid {
					robot_wxid = ag.Content.RobotWxid
					wx.Set("robot_wxid", ag.Content.RobotWxid)
				}
				core.Senders <- &Sender{
					value: wm,
				}
			}
			logs.Info("recv: %s", message)
		}
	}()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := c.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
			if err != nil {
				logs.Info("vlw:", err)
				c = nil
				return
			}
		}
	}
}
