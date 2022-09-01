# relay

A dead simple web server for forwarding GitHub webhook based on filters.

As of now, only relaying push events with ref filter to Lark is supported.

```
$ go run main.go --ref-prefix="refs/heads/release/" --lark-url="https://open.feishu.cn/open-apis/bot/v2/hook/xxxxxxxxxxxxxxxxx"

# Pass LARK_URL via environment variable

$ export LARK_URL=https://open.feishu.cn/open-apis/bot/v2/hook/xxxxxxxxxxxxxxxxx
$ go run main.go --ref-prefix="refs/heads/release/"

# Specify custom host:port

# By default, the server runs on localhost:2830

$ export FLAMEGO_ADDR=localhost:8080
$ go run main.go --ref-prefix="refs/heads/release/" --lark-url="https://open.feishu.cn/open-apis/bot/v2/hook/xxxxxxxxxxxxxxxxx"

```
