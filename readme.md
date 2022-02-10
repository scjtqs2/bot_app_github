# github插件 demo

## 群聊

+ `#github [-t] [xxx]` 文字搜索
+ `#github -p [xxx]`   图片搜索

## 私聊

+ `#github [-t] [xxx]`  文字搜索
+ `#github -p [xxx]`  图片搜索

## github webhook 推送通知

环境变量：

+ `GITHUB_WEBHOOK_ENABLE` 默认"false" 关闭。要开启，填"true"
+ `GITHUB_WEBHOOK_SECRET` github中配置webhook的时候填的secret,用于校验
+ `GITHUB_WEBHOOK_NOTIFY_QQ` 推送给哪个qq，不推送留空
+ `GITHUB_WEBHOOK_NOTIFY_GROUP` 推送给哪个群，不推送留空

推送接受地址 `http://ip:80/postreceive` 实际端口，请根据路由端口映射、docker端口映射做相应的调整

通知的event类型：

+ star
+ pull_request
+ fork
+ issue