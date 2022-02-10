// Package webhook github的webhook推送通知处理
package webhook

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
)

// ErrInvalidEventFormat 错误信息
var ErrInvalidEventFormat = errors.New("unable to parse event string. Invalid Format")

// Event 类
type Event struct {
	Owner      string       // The username of the owner of the repository
	Repo       string       // The name of the repository
	Branch     string       // The branch the event took place on
	FromUser   string       // 谁fork、start、pr
	Tag        string       //
	Commit     string       // The head commit hash attached to the event
	Type       string       // Can be either "pull_request" or "push"
	Action     string       // For Pull Requests, contains "assigned", "unassigned", "labeled", "unlabeled", "opened", "closed", "reopened", or "synchronize".
	BaseOwner  string       // For Pull Requests, contains the base owner
	BaseRepo   string       // For Pull Requests, contains the base repo
	BaseBranch string       // For Pull Requests, contains the base branch
	Payload    gjson.Result // 对象化的json数据
}

// NewEvent Create a new event from a string, the string format being the same as the one produced by event.String()
func NewEvent(e string) (*Event, error) {
	// Trim whitespace
	e = strings.Trim(e, "\n\t ")

	// Split into lines
	parts := strings.Split(e, "\n")

	// Sanity checking
	if len(parts) != 5 || len(parts) != 8 {
		return nil, ErrInvalidEventFormat
	}
	for _, item := range parts {
		if len(item) < 8 {
			return nil, ErrInvalidEventFormat
		}
	}

	// Fill in values for the event
	event := Event{}
	event.Type = parts[0][8:]
	event.Owner = parts[1][8:]
	event.Repo = parts[2][8:]
	event.Branch = parts[3][8:]
	event.Commit = parts[4][8:]

	// Fill in extra values if it's a pull_request
	if event.Type == "pull_request" {
		switch len(parts) {
		case 9: // New format
			event.Action = parts[5][8:]
			event.BaseOwner = parts[6][8:]
			event.BaseRepo = parts[7][8:]
			event.BaseBranch = parts[8][8:]
		case 8: // Old Format
			event.BaseOwner = parts[5][8:]
			event.BaseRepo = parts[6][8:]
			event.BaseBranch = parts[7][8:]
		default:
			return nil, ErrInvalidEventFormat
		}
	}

	return &event, nil
}

// String 字符串输出
func (e *Event) String() (output string) {
	output += "type:   " + e.Type + "\n"
	output += "owner:  " + e.Owner + "\n"
	output += "repo:   " + e.Repo + "\n"
	output += "branch: " + e.Branch + "\n"
	output += "commit: " + e.Commit + "\n"

	if e.Type == "pull_request" {
		output += "action: " + e.Action + "\n"
		output += "bowner: " + e.BaseOwner + "\n"
		output += "brepo:  " + e.BaseRepo + "\n"
		output += "bbranch:" + e.BaseBranch + "\n"
	}

	return
}

// Server 服务类
type Server struct {
	Port       int        // Port to listen on. Defaults to 80
	Path       string     // Path to receive on. Defaults to "/postreceive"
	Secret     string     // Option secret key for authenticating via HMAC
	IgnoreTags bool       // If set to false, also execute command if tag is pushed
	Events     chan Event // Channel of events. Read from this channel to get push events as they happen.
}

// NewServer Create a new server with sensible defaults.
// By default the Port is set to 80 and the Path is set to `/postreceive`
func NewServer() *Server {
	return &Server{
		Port:       80,
		Path:       "/postreceive",
		IgnoreTags: true,
		Events:     make(chan Event, 10), // buffered to 10 items
	}
}

// ListenAndServe Spin up the server and listen for github webhook push events. The events will be passed to Server.Events channel.
func (s *Server) ListenAndServe() error {
	return http.ListenAndServe(":"+strconv.Itoa(s.Port), s)
}

// GoListenAndServe Inside a go-routine, spin up the server and listen for github webhook push events. The events will be passed to Server.Events channel.
func (s *Server) GoListenAndServe() {
	go func() {
		err := s.ListenAndServe()
		if err != nil {
			panic(err)
		}
	}()
}

// ignoreRef Checks if the given ref should be ignored
func (s *Server) ignoreRef(rawRef string) bool {
	if rawRef[:10] == "refs/tags/" && !s.IgnoreTags {
		return false
	}
	return rawRef[:11] != "refs/heads/"
}

// ServeHTTP Satisfies the http.Handler interface.
// Instead of calling Server.ListenAndServe you can integrate hookserve.Server inside your own http server.
// If you are using hookserve.Server in his way Server.Path should be set to match your mux pattern and Server.Port will be ignored.
func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	if req.Method != "POST" {
		http.Error(w, "405 Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if req.URL.Path != s.Path {
		http.Error(w, "404 Not found", http.StatusNotFound)
		return
	}

	eventType := req.Header.Get("X-GitHub-Event")
	if eventType == "" {
		http.Error(w, "400 Bad Request - Missing X-GitHub-Event Header", http.StatusBadRequest)
		return
	}
	if !allowEvent(eventType) {
		http.Error(w, "400 Bad Request - Unknown Event Type "+eventType, http.StatusBadRequest)
		return
	}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// If we have a Secret set, we should check the MAC
	if s.Secret != "" {
		sig := req.Header.Get("X-Hub-Signature")

		if sig == "" {
			http.Error(w, "403 Forbidden - Missing X-Hub-Signature required for HMAC verification", http.StatusForbidden)
			return
		}

		mac := hmac.New(sha1.New, []byte(s.Secret))
		mac.Write(body)
		expectedMAC := mac.Sum(nil)
		expectedSig := "sha1=" + hex.EncodeToString(expectedMAC)
		if !hmac.Equal([]byte(expectedSig), []byte(sig)) {
			http.Error(w, "403 Forbidden - HMAC verification failed", http.StatusForbidden)
			return
		}
	}

	request := gjson.ParseBytes(body)
	if !gjson.ValidBytes(body) {
		http.Error(w, "error json request", http.StatusInternalServerError)
		return
	}

	// Parse the request and build the Event
	event := Event{}
	event.Payload = request
	event.Type = eventType
	switch eventType {
	case "push":
		rawRef := request.Get("ref").String()
		// If the ref is not a branch, we don't care about it
		if s.ignoreRef(rawRef) || request.Get("head_commit").Exists() {
			return
		}
		event.Branch = rawRef[11:]
		event.Repo = request.Get("repository.name").String()
		event.Commit = request.Get("head_commit.id").String()
		event.Owner = request.Get("repository.owner.name").String()
	case "pull_request":
		event.Action = request.Get("action").String()
		event.Owner = request.Get("pull_request.head.repo.owner.login").String()
		event.Repo = request.Get("pull_request.head.repo.name").String()
		event.Branch = request.Get("pull_request.head.ref").String()
		event.Commit = request.Get("pull_request.head.sha").String()
		event.BaseOwner = request.Get("pull_request.base.repo.owner.login").String()
		event.BaseRepo = request.Get("pull_request.base.repo.name").String()
		event.BaseBranch = request.Get("pull_request.base.ref").String()
	case "pull_request_review":
		event.Action = request.Get("action").String()
		event.Owner = request.Get("pull_request.head.repo.owner.login").String()
		event.Repo = request.Get("pull_request.head.repo.name").String()
		event.Branch = request.Get("pull_request.head.ref").String()
		event.Commit = request.Get("pull_request.head.sha").String()
		event.BaseOwner = request.Get("pull_request.base.repo.owner.login").String()
		event.BaseRepo = request.Get("pull_request.base.repo.name").String()
		event.BaseBranch = request.Get("pull_request.base.ref").String()
	case "pull_request_review_comment":
		event.Action = request.Get("action").String()
		event.Owner = request.Get("pull_request.head.repo.owner.login").String()
		event.Repo = request.Get("pull_request.head.repo.name").String()
		event.Branch = request.Get("pull_request.head.ref").String()
		event.Commit = request.Get("pull_request.head.sha").String()
		event.BaseOwner = request.Get("pull_request.base.repo.owner.login").String()
		event.BaseRepo = request.Get("pull_request.base.repo.name").String()
		event.BaseBranch = request.Get("pull_request.base.ref").String()
	case "release":
		event.Action = request.Get("action").String()
		event.FromUser = request.Get("sender.login").String()
		event.Repo = request.Get("repository.name").String()
		event.Commit = request.Get("head_commit.id").String()
		event.Owner = request.Get("repository.owner.login").String()
		event.Branch = request.Get("release.target_commitish").String()
		event.Tag = request.Get("release.tag_name").String()
	case "fork":
		event.Action = request.Get("action").String()
		event.FromUser = request.Get("sender.login").String()
		event.Repo = request.Get("repository.name").String()
		event.Owner = request.Get("repository.owner.login").String()
	case "star":
		event.Action = request.Get("action").String()
		event.FromUser = request.Get("sender.login").String()
		event.Repo = request.Get("repository.name").String()
		event.Owner = request.Get("repository.owner.login").String()
	case "issue_comment":
		event.Action = request.Get("action").String()
		event.FromUser = request.Get("sender.login").String()
		event.Repo = request.Get("repository.name").String()
		event.Owner = request.Get("repository.owner.login").String()
	case "issues":
		event.Action = request.Get("action").String()
		event.FromUser = request.Get("sender.login").String()
		event.Repo = request.Get("repository.name").String()
		event.Owner = request.Get("repository.owner.login").String()
	case "create":
		event.FromUser = request.Get("sender.login").String()
		event.Repo = request.Get("repository.name").String()
		event.Owner = request.Get("repository.owner.login").String()
	default:
		http.Error(w, "Unknown Event Type "+eventType, http.StatusInternalServerError)
		return
	}

	// We've built our Event - put it into the channel and we're done
	go func() {
		s.Events <- event
	}()

	_, _ = w.Write([]byte(event.String()))
}

// allowEvent 支持解析的event类型
func allowEvent(eventType string) bool {
	allow := []string{
		/**
		The action that was performed. Can be one of:
		assigned
		auto_merge_disabled
		auto_merge_enabled
		closed: If the action is closed and the merged key is false, the pull request was closed with unmerged commits. If the action is closed and the merged key is true, the pull request was merged.
		converted_to_draft
		edited
		labeled
		locked
		opened
		ready_for_review
		reopened
		review_request_removed
		review_requested
		synchronize: Triggered when a pull request's head branch is updated. For example, when the head branch is updated from the base branch, when new commits are pushed to the head branch, or when the base branch is changed.
		unassigned
		unlabeled
		unlocked
		*/
		"pull_request", // pr
		/**
		The action that was performed. Can be one of:
		submitted - A pull request review is submitted into a non-pending state.
		edited - The body of a review has been edited.
		dismissed - A review has been dismissed.
		*/
		"pull_request_review", // PR 的 review
		/**
		The action that was performed on the comment. Can be one of created, edited, or deleted.
		*/
		"pull_request_review_comment", // PR review的回复
		/**
		The action that was performed. Can be one of:
		published: a release, pre-release, or draft of a release is published
		unpublished: a release or pre-release is deleted
		created: a draft is saved, or a release or pre-release is published without previously being saved as a draft
		edited: a release, pre-release, or draft release is edited
		deleted: a release, pre-release, or draft release is deleted
		prereleased: a pre-release is created
		released: a release or draft of a release is published, or a pre-release is changed to a release
		*/
		"release", // 发版

		"fork", // A user forks a repository. For more information, see the "forks" REST API.
		"push", // 提交commit
		/**
		The action performed. Can be created or deleted.
		*/
		"star", // 有人点星星
		/**
		action The action that was performed on the comment. Can be one of created, edited, or deleted.
		*/
		"issue_comment", //
		/**
		action The action that was performed. Can be one of opened, edited, deleted, pinned, unpinned, closed, reopened, assigned, unassigned, labeled, unlabeled, locked, unlocked, transferred, milestoned, or demilestoned.
		*/
		"issues", //

		"create", // A Git branch or tag is created. For more information, see the "Git database" REST API.
	}
	for _, s := range allow {
		if s == eventType {
			return true
		}
	}
	return false
}
