<h1 align="center">Joe Bot - Telegram Adapter</h1>
<p align="center">Connecting joe with the Telegram chat application. https://github.com/go-joe/joe</p>
<p align="center">
	<a href="https://github.com/robertgzr/joe-telegram-adapter/releases"><img src="https://img.shields.io/github/tag/robertgzr/joe-telegram-adapter.svg?label=version&color=brightgreen"></a>
	<a href="https://godoc.org/github.com/robertgzr/joe-telegram-adapter"><img src="https://img.shields.io/badge/godoc-reference-blue.svg?color=blue"></a>
</p>

---

This repository contains a module for the [Joe Bot library][joe]. Built using 
[telegram-bot-api][tgbotapi].

## Getting Started

This library is packaged using [Go modules][go-modules]. You can get it via:

```
go get github.com/robertgzr/joe-telegram-adapter
```

### Example usage

In order to connect your bot to telegram you can simply pass it as module when
creating a new bot:

```go
package main

import (
	"github.com/go-joe/joe"
	"github.com/robertgzr/joe-telegram-adapter"
)

func main() {
	b := joe.New("example-bot",
		telegram.Adapter(os.Getenv("TELEGRAM_BOT_TOKEN")),
		…
	)
	
	b.Respond("ping", Pong)

	err := b.Run()
	if err != nil {
		b.Logger.Fatal(err.Error())
	}
}
```

For how to create a telegram bot and connect to it, [see here](https://core.telegram.org/bots#3-how-do-i-create-a-bot).

This adapter will emit the following events to the robot brain:

- `joe.ReceiveMessageEvent`
- `ReceiveCommandEvent`

A common use-case is handling Telegram bot commands, `/command`. To make this 
easy a custom event type is emitted:

```go
package main

import (
	"github.com/go-joe/joe"
	"github.com/robertgzr/joe-telegram-adapter"
)

type ExampleBot {
    *joe.Bot
}

func main() {
	b := &ExampleBot{
	    Bot: joe.New("example-bot",
		    telegram.Adapter(os.Getenv("TELEGRAM_BOT_TOKEN")),
		    …
	    ),
	}

	b.Brain.RegisterHandler(b.HandleTelegramCommands)
	b.Respond("ping", Pong)

	err := b.Run()
	if err != nil {
		b.Logger.Fatal(err.Error())
	}
}

func (b *ExampleBot) HandleTelegramCommands(ev telegram.ReceiveCommandEvent) error {
	switch ev.Arg0 {
	case "command":
		b.Say(ev.Channel(), "Hello, world!")
		return nil
	default:
		return errors.New("unknown command")
	}
}
```

## License

[BSD-3-Clause](LICENSE)

[joe]: https://github.com/go-joe/joe
[tgbotapi]: https://github.com/go-telegram-bot-api/telegram-bot-api
[go-modules]: https://github.com/golang/go/wiki/Modules
