// Package search github搜索服务
package search

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/scjtqs2/bot_adapter/client"
	"github.com/scjtqs2/bot_adapter/coolq"
	"github.com/scjtqs2/bot_adapter/event"
	"github.com/scjtqs2/bot_adapter/pb/entity"
	"github.com/tidwall/gjson"
)

// GSearch github search 服务
type GSearch struct {
	Cli *client.AdapterService
}

// NewGSearch 初始化 gsearch服务
func NewGSearch(cli *client.AdapterService) *GSearch {
	return &GSearch{
		Cli: cli,
	}
}

// SearchPrivate 私聊处理
func (g *GSearch) SearchPrivate(req event.MessagePrivate) {
	searchType, keyword, ok := g.parseKeyword(req.RawMessage)
	if ok {
		msg := g.searchText(searchType, keyword)
		_, _ = g.Cli.SendPrivateMsg(context.TODO(), &entity.SendPrivateMsgReq{
			UserId:  req.UserID,
			Message: []byte(msg),
		})
	}
}

// SearchGroup 群聊处理
func (g *GSearch) SearchGroup(req event.MessageGroup) {
	searchType, keyword, ok := g.parseKeyword(req.RawMessage)
	if ok {
		msg := g.searchText(searchType, keyword)
		_, _ = g.Cli.SendGroupMsg(context.TODO(), &entity.SendGroupMsgReq{
			GroupId: req.GroupID,
			Message: []byte(msg),
		})
	}
}

// parseKeyword 提取keyword 并判断 是否是有效的命令
func (g *GSearch) parseKeyword(msg string) (string, string, bool) {
	req := regexp.MustCompile(`^#github\s(-.{1,10}? )?(.*)$`)
	if !req.MatchString(msg) {
		return "", "", false
	}
	matchs := req.FindStringSubmatch(msg)
	return strings.TrimSpace(matchs[1]), strings.TrimSpace(matchs[2]), true
}

// searchText 通过github搜索项目
func (g *GSearch) searchText(searchType, keyword string) string {
	// 发送请求
	header := http.Header{
		"User-Agent": []string{"User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.182 Safari/537.36"},
	}
	api, _ := url.Parse("https://api.github.com/search/repositories")
	api.RawQuery = url.Values{
		"q": []string{keyword},
	}.Encode()
	body, err := netGet(api.String(), header)
	if err != nil {
		return fmt.Sprintf("ERROR:%v", err)
	}
	// 解析请求
	info := gjson.ParseBytes(body)
	if info.Get("total_count").Int() == 0 {
		return "ERROR: 没有找到这样的仓库"
	}
	repo := info.Get("items.0")
	var msg string
	switch searchType {
	case "-p": // 图片模式
		msg = coolq.EnImageCode("https://opengraph.githubassets.com/0/"+repo.Get("full_name").String(), 0)
	case "-t":
		msg = fmt.Sprintf("%s\n"+
			"Description: "+
			"%s\n"+
			"Star/Fork/Issue: "+
			"%d/%d/%d\n"+
			"Language: "+
			"%s\n"+
			"License: "+
			"%s\n"+
			"Last pushed: "+
			"%s\n"+
			"Jump: "+
			"%s\n",
			repo.Get("full_name").String(),
			repo.Get("description").String(),
			repo.Get("watchers").Int(), repo.Get("forks").Int(), repo.Get("open_issues").Int(),
			notnull(repo.Get("language").String(), "None"),
			notnull(repo.Get("license.key").String(), "None"),
			repo.Get("pushed_at").String(),
			repo.Get("html_url").String())
	default:
		msg = fmt.Sprintf("%s\n"+
			"Description: "+
			"%s\n"+
			"Star/Fork/Issue: "+
			"%d/%d/%d\n"+
			"Language: "+
			"%s\n"+
			"License: "+
			"%s\n"+
			"Last pushed: "+
			"%s\n"+
			"Jump: "+
			"%s\n",
			repo.Get("full_name").String(),
			repo.Get("description").String(),
			repo.Get("watchers").Int(), repo.Get("forks").Int(), repo.Get("open_issues").Int(),
			notnull(repo.Get("language").String(), "None"),
			notnull(repo.Get("license.key").String(), "None"),
			repo.Get("pushed_at").String(),
			repo.Get("html_url").String()) + coolq.EnImageCode("https://opengraph.githubassets.com/0/"+repo.Get("full_name").String(), 0)
	}
	return msg
}

// notnull 如果传入文本为空，则返回默认值
//nolint: unparam
func notnull(text, defstr string) string {
	if text == "" {
		return defstr
	}
	return text
}

// netGet 返回请求结果
func netGet(dest string, header http.Header) ([]byte, error) {
	c := &http.Client{}

	req, err := http.NewRequest("GET", dest, nil)
	if err != nil {
		return nil, err
	}
	req.Header = header
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if code := resp.StatusCode; code != 200 {
		// 如果返回不是200则立刻抛出错误
		errmsg := fmt.Sprintf("code %d", code)
		return nil, errors.New(errmsg)
	}
	return body, nil
}
