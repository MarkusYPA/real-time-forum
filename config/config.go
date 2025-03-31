package config

import (
	"html/template"
	forumModels "real-time-forum/modules/forumManagement/models"
	"sync"

	"github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3"
)

type Message struct {
	Post                    forumModels.Post             `json:"post"`
	Comment                 forumModels.Comment          `json:"comment"`
	MsgType                 string                       `json:"msgType"`
	Updated                 bool                         `json:"updated"`
	UserUUID                string                       `json:"uuid"`
	IsLikAction             bool                         `json:"isLikeAction"`
	NumberOfReplis          int                          `json:"numberOfReplies"`
	IsReplied               bool                         `json:"isReplied"`
	ChattedUsers            []forumModels.ChatUser       `json:"chattedUsers"`
	UnchattedUsers          []forumModels.ChatUser       `json:"unchattedUsers"`
	PrivateMessage          forumModels.PrivateMessage   `json:"privateMessage"`
	ReciverUserUUID         string                       `json:"reciverUserUUID"`
	ReceiverUserName        string                       `json:"receiverUserName"`
	Messages                []forumModels.PrivateMessage `json:"privateMessages"`
	SendNotification        bool                         `json:"notification"`
	GotAllMessagesRequested bool                         `json:"allMessagesGot"`
}

var (
	HomeTmpl  *template.Template
	Upgrader  = websocket.Upgrader{}
	Clients   = make(map[string]*websocket.Conn)
	Broadcast = make(chan Message)
	Mu        sync.Mutex
)

const (
	TitleMaxLen   int = 100
	ContentMaxLen int = 3000
)
