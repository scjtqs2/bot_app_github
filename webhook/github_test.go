package webhook

import (
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	"github.com/tebeka/selenium/firefox"
	"net/http"
	"testing"
	"time"
)

func TestChrome(t *testing.T) {
	addr := "http://127.0.0.1:4444"
	selenium.HTTPClient = &http.Client{
		Timeout: time.Second * 30,
	}
	caps := selenium.Capabilities{"browserName": "chrome"}
	// chrome参数
	chromeCaps := chrome.Capabilities{
		Args: []string{
			"--headless", // 设置Chrome无头模式，在linux下运行，需要设置这个参数，否则会报错
			// "--no-sandbox",
			"--window-size=600,812",
			// fmt.Sprintf("--proxy-server=%s", "http://192.168.28.101:7890"), // --proxy-server=http://127.0.0.1:1234
		},
	}
	caps.AddChrome(chromeCaps)
	wd, err := selenium.NewRemote(caps, addr)
	defer func() {
		_ = wd.Close()
		_ = wd.Quit()
	}()
	url := "https://github.com/scjtqs2/bot_app_github/issues/1"
	if err := wd.Get(url); err != nil {
		t.Fatalf("wd error %v", err)
	}
	_ = wd.Wait(func(wd selenium.WebDriver) (bool, error) {
		_, err = wd.FindElement(selenium.ByCSSSelector, "#js-repo-pjax-container")
		return err == nil, nil
	})
	issue, err := wd.FindElement(selenium.ByCSSSelector, "#js-repo-pjax-container")
	if err != nil {
		t.Fatalf("issur error %v", err)
	}
	// issue, err := wd.FindElement(selenium.ByCSSSelector, fmt.Sprintf("#issue-%s > div", issueID))
	// if err != nil {
	// 	return nil, err
	// }
	sidebar, err := issue.FindElement(selenium.ByCSSSelector, "#partial-discussion-sidebar")
	if err == nil {
		_, _ = wd.ExecuteScript("arguments[0].remove();", []interface{}{sidebar})
	}
	signBar, err := issue.FindElement(selenium.ByCSSSelector, ".discussion-timeline-actions")
	if err == nil {
		_, _ = wd.ExecuteScript("arguments[0].remove();", []interface{}{signBar})
	}
	size, err := issue.Size()
	if err != nil {
		t.Fatalf("size error %v", err)
	}
	window, _ := wd.CurrentWindowHandle()
	_ = wd.ResizeWindow(window, size.Width, size.Height+100)
	pic, err := issue.Screenshot(false)
	if err != nil {
		t.Fatalf("pic err %v", err)
	}
	t.Logf("pic %v", pic)
}

func TestFirefox(t *testing.T) {
	addr := "http://127.0.0.1:4444"
	selenium.HTTPClient = &http.Client{
		Timeout: time.Second * 60,
	}
	caps := selenium.Capabilities{"browserName": "firefox"}
	// firefox 参数
	firefoxCaps := firefox.Capabilities{
		Args: []string{
			"--headless", // 设置Chrome无头模式，在linux下运行，需要设置这个参数，否则会报错
			"--disable-gpu",
			"window-size=600,812",
			// fmt.Sprintf("--proxy-server=%s", "http://192.168.28.101:7890"), // --proxy-server=http://127.0.0.1:1234
		},
	}
	caps.AddFirefox(firefoxCaps)
	wd, err := selenium.NewRemote(caps, addr)
	defer func() {
		_ = wd.Close()
		_ = wd.Quit()
	}()
	url := "https://github.com/scjtqs2/bot_app_github/issues/1"
	if err := wd.Get(url); err != nil {
		t.Fatalf("wd error %v", err)
	}
	_ = wd.Wait(func(wd selenium.WebDriver) (bool, error) {
		_, err = wd.FindElement(selenium.ByCSSSelector, "#js-repo-pjax-container")
		return err == nil, nil
	})
	issue, err := wd.FindElement(selenium.ByCSSSelector, "#js-repo-pjax-container")
	if err != nil {
		t.Fatalf("issur error %v", err)
	}
	// issue, err := wd.FindElement(selenium.ByCSSSelector, fmt.Sprintf("#issue-%s > div", issueID))
	// if err != nil {
	// 	return nil, err
	// }
	sidebar, err := issue.FindElement(selenium.ByCSSSelector, "#partial-discussion-sidebar")
	if err == nil {
		_, _ = wd.ExecuteScript("arguments[0].remove();", []interface{}{sidebar})
	}
	signBar, err := issue.FindElement(selenium.ByCSSSelector, ".discussion-timeline-actions")
	if err == nil {
		_, _ = wd.ExecuteScript("arguments[0].remove();", []interface{}{signBar})
	}
	size, err := issue.Size()
	if err != nil {
		t.Fatalf("size error %v", err)
	}
	window, _ := wd.CurrentWindowHandle()
	_ = wd.ResizeWindow(window, size.Width, size.Height+100)
	pic, err := issue.Screenshot(false)
	if err != nil {
		t.Fatalf("pic err %v", err)
	}
	t.Logf("pic %v", pic)
}
