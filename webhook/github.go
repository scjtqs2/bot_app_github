package webhook

import (
	"os"
	"strconv"

	"github.com/scjtqs2/bot_adapter/client"
	log "github.com/sirupsen/logrus"
)

// GHook github推送类
type GHook struct {
	Cli           *client.AdapterService
	Enable        bool    // 是否启用webhook
	NotifyQQ      int64   // 接收推送的qq
	NotifyQQGroup int64   // 接收推送的群
	GithubSecret  string  // github的hook的secret
	Server        *Server // http监听地址
}

// NewGHook 初始化 ghook
func NewGHook(cli *client.AdapterService) *GHook {
	qq, _ := strconv.ParseInt(os.Getenv("GITHUB_WEBHOOK_NOTIFY_QQ"), 10, 64)
	group, _ := strconv.ParseInt(os.Getenv("GITHUB_WEBHOOK_NOTIFY_GROUP"), 10, 64)
	return &GHook{
		Cli:           cli,
		Enable:        os.Getenv("GITHUB_WEBHOOK_ENABLE") == "true",
		NotifyQQ:      qq,
		NotifyQQGroup: group,
		GithubSecret:  os.Getenv("GITHUB_WEBHOOK_SECRET"),
	}
}

// Init 初始化
func (g *GHook) Init() {
	if !g.Enable {
		log.Warn("未开启github webhook")
		return
	}
	log.Infof("github webhook 开启中 notifyqq:%d ,notifyGroup:%d,secret:%s", g.NotifyQQ, g.NotifyQQGroup, g.GithubSecret)
	g.Server = NewServer()
	g.Server.Port = 80
	g.Server.Secret = g.GithubSecret
	g.Server.GoListenAndServe() // 开启监听
	go g.parseEvents()
}

// parseEvents 处理收到的events TODO 推送
func (g *GHook) parseEvents() {
	for event := range g.Server.Events {
		log.Infof("resived event %+v", event)
	}
}
