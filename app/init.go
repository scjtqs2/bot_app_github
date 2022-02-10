// Package app 面向对象的应用类
package app

import (
	"os"

	"github.com/kataras/iris/v12"
	"github.com/scjtqs2/bot_adapter/client"
	"github.com/scjtqs2/bot_adapter/sha256"
	"github.com/scjtqs2/bot_app_chat/bot"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"

	"github.com/scjtqs2/bot_app_github/search"
	"github.com/scjtqs2/bot_app_github/webhook"
)

// App 结构体
type App struct {
	appID            string
	appSecret        string
	appEncryptKey    string
	botAdapterAddr   string
	botAdapterClient *client.AdapterService
	search           *search.GSearch
	hook             *webhook.GHook
}

// NewApp 初始化app
func NewApp() *App {
	return &App{
		appID:          os.Getenv("APP_ID"),
		appSecret:      os.Getenv("APP_SECRET"),
		appEncryptKey:  os.Getenv("APP_ENCRYPT_KEY"),
		botAdapterAddr: os.Getenv("ADAPTER_ADDR"),
	}
}

// Init 初始化监听
func (a *App) Init() {
	var err error
	a.botAdapterClient, err = client.NewAdapterServiceClient(a.botAdapterAddr, a.appID, a.appSecret)
	if err != nil {
		log.Fatalf("faild to init grpc client err:%v", err)
	}
	app := iris.New()
	app.Post("/", a.msginput)
	go func() {
		port := "8080"
		if os.Getenv("HTTP_PORT") != "" {
			port = os.Getenv("HTTP_PORT")
		}
		err = app.Run(iris.Addr(":" + port))
		if err != nil {
			log.Fatalf("error init http listen port %s err:%v", port, err)
		}
	}()
	a.search = search.NewGSearch(a.botAdapterClient)
	a.hook = webhook.NewGHook(a.botAdapterClient)
	a.hook.Init()
}

func (a *App) msginput(ctx iris.Context) {
	raw, _ := ctx.GetBody()
	enc := gjson.ParseBytes(raw).Get("encrypt").String()
	// 解密推送数据
	msg, err := sha256.Decrypt(enc, a.appEncryptKey)
	if err != nil {
		log.Errorf("解密失败：enc:%s err:%s", enc, err.Error())
	}
	go a.parseMsg(msg)
	_, _ = ctx.JSON(bot.MSG{
		"code": 200,
		"msg":  "received",
	})
}
