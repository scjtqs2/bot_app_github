package webhook

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/scjtqs2/bot_adapter/coolq"
	"github.com/scjtqs2/bot_adapter/pb/entity"

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
		var msg string
		switch event.Type {
		case "star":
			// Tim-Paik starred Mrs4s/go-cqhttp (total 3914 stargazers)
			var a string
			switch event.Action {
			case "created":
				a = "starred"
			case "deleted":
				a = "unstarred"
			}
			msg = fmt.Sprintf("%s %s %s/%s (total %d stargazers)", event.FromUser, a, event.Owner, event.Repo, event.Payload.Get("repository.stargazers_count").Int())
		case "fork":
			msg = fmt.Sprintf("%s forked %s/%s (total %d forks_count", event.FromUser, event.Owner, event.Repo, event.Payload.Get("repository.forks_count").Int())
		case "issues":
			var labels string
			for _, result := range event.Payload.Get("issue.labels").Array() {
				labels += fmt.Sprintf("[%s]", result.Get("name").String())
			}
			switch event.Action {
			case "opened":
				// wdvxdr1123 opened issue Mrs4s/go-cqhttp#1358
				msg = fmt.Sprintf("%s %s issue %s/%s #%d \n", event.FromUser, event.Action, event.Owner, event.Repo, event.Payload.Get("issue.number").Int()) +
					fmt.Sprintf("%s Title: %s \n", labels, event.Payload.Get("issue.title").String()) +
					fmt.Sprintf("Body: %s", event.Payload.Get("issue.body").String()) +
					fmt.Sprintf("jump: %s \n", event.Payload.Get("issue.html_url").String()) +
					coolq.EnImageCode(fmt.Sprintf("https://opengraph.githubassets.com/0/%s/%s/issues/%d", event.Owner, event.Repo, event.Payload.Get("issue.number").Int()), 0)
			case "closed":
				msg = fmt.Sprintf("%s %s issue %s/%s #%d \n", event.FromUser, event.Action, event.Owner, event.Repo, event.Payload.Get("issue.number").Int()) +
					fmt.Sprintf("%s Title: %s \n", labels, event.Payload.Get("issue.title").String()) +
					fmt.Sprintf("Body: %s \n", event.Payload.Get("issue.body").String()) +
					fmt.Sprintf("jump: %s \n", event.Payload.Get("issue.html_url").String()) +
					coolq.EnImageCode(fmt.Sprintf("https://opengraph.githubassets.com/0/%s/%s/issues/%d", event.Owner, event.Repo, event.Payload.Get("issue.number").Int()), 0)
			case "reopened":
				msg = fmt.Sprintf("%s %s issue %s/%s #%d \n", event.FromUser, event.Action, event.Owner, event.Repo, event.Payload.Get("issue.number").Int()) +
					fmt.Sprintf("%s Title: %s \n", labels, event.Payload.Get("issue.title").String()) +
					fmt.Sprintf("Body: %s", event.Payload.Get("issue.body").String()) +
					fmt.Sprintf("jump: %s \n", event.Payload.Get("issue.html_url").String()) +
					coolq.EnImageCode(fmt.Sprintf("https://opengraph.githubassets.com/0/%s/%s/issues/%d", event.Owner, event.Repo, event.Payload.Get("issue.number").Int()), 0)
			}
		case "issue_comment":
			var labels string
			for _, result := range event.Payload.Get("issue.labels").Array() {
				labels += fmt.Sprintf("[%s]", result.Get("name").String())
			}
			switch event.Action {
			case "created":
				msg = fmt.Sprintf("%s commented on %s/%s #%d", event.FromUser, event.Owner, event.Repo, event.Payload.Get("issue.number").Int()) +
					fmt.Sprintf("%s Title: %s \n", labels, event.Payload.Get("issue.title").String()) +
					fmt.Sprintf("Body: %s \n", event.Payload.Get("issue.body").String()) +
					fmt.Sprintf("Comment: %s \n", event.Payload.Get("comment.body").String()) +
					fmt.Sprintf("jump: %s \n", event.Payload.Get("comment.html_url").String())
			case "edited":
				msg = fmt.Sprintf("%s edited commente on %s/%s #%d", event.FromUser, event.Owner, event.Repo, event.Payload.Get("issue.number").Int()) +
					fmt.Sprintf("%s Title: %s \n", labels, event.Payload.Get("issue.title").String()) +
					fmt.Sprintf("Body: %s \n", event.Payload.Get("issue.body").String()) +
					fmt.Sprintf("Comment: %s \n", event.Payload.Get("comment.body").String()) +
					fmt.Sprintf("jump: %s \n", event.Payload.Get("comment.html_url").String())
			case "deleted":
				msg = fmt.Sprintf("%s deleted commente on %s/%s #%d", event.FromUser, event.Owner, event.Repo, event.Payload.Get("issue.number").Int()) +
					fmt.Sprintf("%s Title: %s \n", labels, event.Payload.Get("issue.title").String()) +
					fmt.Sprintf("Body: %s \n", event.Payload.Get("issue.body").String()) +
					fmt.Sprintf("Comment: %s \n", event.Payload.Get("comment.body").String()) +
					fmt.Sprintf("jump: %s \n", event.Payload.Get("comment.html_url").String())
			}
		case "pull_request":
			switch event.Action {
			case "opened":
				// wdvxdr1123 opened an pull request for Mrs4s/go-cqhttp#1356(dev<wdvxdr1123:test_pr_review)
				msg = fmt.Sprintf("%s opened an pull request for %s/%s #%d (%s<-%s:%s) \n", event.FromUser, event.Owner, event.Repo,
					event.Payload.Get("pull_request.number").Int(),
					event.BaseBranch, event.Owner, event.Branch) +
					fmt.Sprintf("jump: %s \n", event.Payload.Get("pull_request.html_url").String()) +
					coolq.EnImageCode(fmt.Sprintf("https://opengraph.githubassets.com/0/%s/%s/pull/%d", event.BaseOwner, event.BaseRepo, event.Payload.Get("pull_request.number").Int()), 0)
			default:
				continue
			}
		default:
			log.Warnf("unknow eventType:%s,action:%s from %s/%s", event.Type, event.Action, event.Owner, event.Repo)
			continue
		}
		if msg == "" {
			continue
		}
		if g.NotifyQQ != 0 {
			_, err := g.Cli.SendPrivateMsg(context.TODO(), &entity.SendPrivateMsgReq{
				UserId:  g.NotifyQQ,
				Message: []byte(msg),
			})
			if err != nil {
				log.Errorf("push to NotifyQQ %d err:%v", g.NotifyQQ, err)
			}
		}
		if g.NotifyQQGroup != 0 {
			_, err := g.Cli.SendGroupMsg(context.TODO(), &entity.SendGroupMsgReq{GroupId: g.NotifyQQGroup, Message: []byte(msg)})
			if err != nil {
				log.Errorf("push to NotifyQQGroup %d err:%v", g.NotifyQQGroup, err)
			}
		}
	}
}
