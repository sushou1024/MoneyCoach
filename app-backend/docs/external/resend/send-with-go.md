# Send emails with Go

> Learn how to send your first email using the Resend Go SDK.

## Prerequisites

To get the most out of this guide, you'll need to:

* [Create an API key](https://resend.com/api-keys)
* [Verify your domain](https://resend.com/domains)

## 1. Install

Get the Resend Go SDK.

<CodeGroup>
  ```bash bash theme={null}
  go get github.com/resend/resend-go/v3
  ```
</CodeGroup>

## 2. Send email using HTML

The easiest way to send an email is by using the `html` parameter.

```Go main.go theme={null}
package main

import "github.com/resend/resend-go/v3"

func main() {
    apiKey := "re_xxxxxxxxx"

    client := resend.NewClient(apiKey)

    params := &resend.SendEmailRequest{
        From:    "Acme <onboarding@resend.dev>",
        To:      []string{"delivered@resend.dev"},
        Html:    "<strong>hello world</strong>",
        Subject: "Hello from Golang",
        Cc:      []string{"cc@example.com"},
        Bcc:     []string{"bcc@example.com"},
        ReplyTo: "replyto@example.com",
    }

    sent, err := client.Emails.Send(params)
    if err != nil {
        fmt.Println(err.Error())
        return
    }
    fmt.Println(sent.Id)
}
```

## 3. Try it yourself

<Card title="Golang Examples" icon="arrow-up-right-from-square" href="https://github.com/resend/resend-go-example">
  See the full source code.
</Card>


---

> To find navigation and other pages in this documentation, fetch the llms.txt file at: https://resend.com/docs/llms.txt