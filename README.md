# webhook-transmitter

A dead simple web server for forwarding GitHub webhook with filters.

As of now, only transmitting push events to Lark with ref filter is supported.

```
$ go run main.go --ref-prefix="refs/heads/release/" --lark-url="https://open.feishu.cn/open-apis/bot/v2/hook/xxxxxxxxxxxxxxxxx"

# or

$ export LARK_URL=https://open.feishu.cn/open-apis/bot/v2/hook/xxxxxxxxxxxxxxxxx
$ go run main.go --ref-prefix="refs/heads/release/" 
```
