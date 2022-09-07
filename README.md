# Relay

<img alt="Relay Logo" src="assets/logo.png" /></a>

<p align="center">
  <a href="https://goreportcard.com/report/github.com/bytebase/relay">
    <img alt="go report" src="https://goreportcard.com/badge/github.com/bytebase/relay" />
  </a>
</p>

Relay is a web server for forwarding event from service A to service B. While forwarding the event, it allows to inject arbitrary logic to process the event such as filtering, augmenting.

You may think Relay as a very very simplified version of self hosted [Zapier](https://zapier.com).

# Overview

<img alt="Overview" src="assets/overview.png" /></a>

Relay Contains 2 components Hooker and Sinker.

Hooker - Receive the webhook event from upstream services such as GitHub, GitLab, do some processing, and pass the processed payload to Sinker.

Sinker - Receive the payload from Hooker, do some processing, and send the event to downstream services such as Slack, Lark.

To relay an event from Service A to Service B, you would
1. Implement a Hooker to receive event from service A.
1. Implement a Sinker to process payload from that Hooker and send the processed message to Service B.
1. Register the Hooker to listen to event from service A and register the Sinker with the Hooker.

# Common Flags

#### --address (Env RELAY_ADDR)

The address where Relay runs. Default `localhost:5678`.

# Supported Hookers

## GitHub

*When configuring GitHub Webhook, make sure to set the webhook content type as `application/json`.*

### Flags

#### --github-ref-prefix

The prefix for the GitHub ref. GitHub Webhook iteself doesn't allow to specify a particular branch or branch filter. You can use `--github-ref-prefix` to only observe the events from the interested branch(es).

# Supported Sinkers

## Lark

### Flags

#### --lark-urls

A comma separated list of Lark message group webhook URLs

# Quickstart

```sh
$ go run main.go --github-ref-prefix="refs/heads/release/" --lark-urls="https://open.feishu.cn/open-apis/bot/v2/hook/foo"

# --lark-url can also be a comma separated list
$ go run main.go --github-ref-prefix="refs/heads/release/" --lark-urls="https://open.feishu.cn/open-apis/bot/v2/hook/foo,https://open.feishu.cn/open-apis/bot/v2/hook/bar"

# Runs on localhost:8080
$ go run main.go --address=localhost:8080 --github-ref-prefix="refs/heads/release/" --lark-urls="https://open.feishu.cn/open-apis/bot/v2/hook/foo"
```
