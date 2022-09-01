# Relay

A dead simple web server for forwarding GitHub webhook based on filters.

As of now, only relaying push events with ref filter to Lark is supported.

# Configure GitHub webhook

Make sure to set the webhook content type as `application/json`.

# Run

```sh
$ go run main.go --ref-prefix="refs/heads/release/" --lark-url="https://open.feishu.cn/open-apis/bot/v2/hook/foo"

# --lark-url can also be a comma separated list
$ go run main.go --ref-prefix="refs/heads/release/" --lark-url="https://open.feishu.cn/open-apis/bot/v2/hook/foo,https://open.feishu.cn/open-apis/bot/v2/hook/bar"

# Pass LARK_URL via environment variable
$ export LARK_URL=https://open.feishu.cn/open-apis/bot/v2/hook/xxxxxxxxxxxxxxxxx
$ go run main.go --ref-prefix="refs/heads/release/"

# By default, the server runs on localhost:2830, you may specify a custom host:port
$ export FLAMEGO_ADDR=localhost:8080
$ go run main.go --ref-prefix="refs/heads/release/" --lark-url="https://open.feishu.cn/open-apis/bot/v2/hook/xxxxxxxxxxxxxxxxx"
```
