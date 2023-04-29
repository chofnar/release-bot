package server

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"time"

	"github.com/chofnar/release-bot/database"
	databaseLoader "github.com/chofnar/release-bot/database/loader"
	"github.com/chofnar/release-bot/errors"
	"github.com/chofnar/release-bot/server/behaviors"
	botConfig "github.com/chofnar/release-bot/server/config"
	"github.com/chofnar/release-bot/server/consts"
	myHandlers "github.com/chofnar/release-bot/server/telegohandlers"
	th "github.com/mymmrac/telego/telegohandler"

	// "github.com/fasthttp/router"
	"github.com/hasura/go-graphql-client"
	"github.com/mymmrac/telego"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

func Initialize(logger zap.SugaredLogger) (*botConfig.BotConfig, database.Database) {
	conf := botConfig.LoadBotConfig()
	db := databaseLoader.GetDatabase(logger)
	return conf, db
}

func Start() {
	unsugared, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}

	logger := unsugared.Sugar()
	botConf, db := Initialize(*logger)

	bot, err := telego.NewBot(botConf.TelegramToken, telego.WithLogger(logger))
	if err != nil {
		logger.Error(err)

		panic(err)
	}

	// Create Github GraphQL token
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: botConf.GithubGQLToken},
	)

	httpClient := oauth2.NewClient(context.Background(), src)
	githubGQLClient := graphql.NewClient("https://api.github.com/graphql", httpClient)

	linkRegex, _ := regexp.Compile("(?:https://)github.com[:/](.*)[:/](.*)")
	directRegex, _ := regexp.Compile("(.*)[/](.*)")

	behaviorHandler := behaviors.BehaviorHandler{
		Bot:         bot,
		LinkRegex:   linkRegex,
		DirectRegex: directRegex,
		GQLClient:   githubGQLClient,
		DB:          db,
	}

	awaitingAddRepo := map[int64]struct{}{}

	handler := myHandlers.Handler{
		BehaviorHandler: behaviorHandler,
		Logger:          *logger,
		AwaitingAddRepo: awaitingAddRepo,
	}

	if botConf.ResetWebhookUrl != "" {
		err = bot.SetWebhook(&telego.SetWebhookParams{
			URL: botConf.WebhookSite + "/bot/" + botConf.TelegramToken,
		})
		if err != nil {
			panic(err)
		}
	}

	mux := http.NewServeMux()
	tp := TimePath{}
	mux.Handle("/time", tp)
	up := UpdatePath{}
	mux.Handle("/updateRepos", up.UpdateRepos(&behaviorHandler, *logger))

	updates, err := bot.UpdatesViaWebhook("/bot/"+bot.Token(), telego.WithWebhookServer(telego.HTTPWebhookServer{
		Logger:   logger,
		Server:   &http.Server{},
		ServeMux: mux,
	}))
	if err != nil {
		panic(err)
	}

	botHandler, err := th.NewBotHandler(bot, updates)
	if err != nil {
		panic(err)
	}

	if botConf.UseUpdateLoop != "" {
		go updateLoop(context.Background(), updates, &behaviorHandler, *logger)
	} else {
		// register handlers
		botHandler.Handle(handler.Start(), th.CommandEqual("start"))
		botHandler.Handle(handler.About(), th.CommandEqual("about"))
		botHandler.Handle(handler.UnknownOrSent(), th.AnyMessageWithText())

		// Callback queries
		botHandler.HandleCallbackQuery(handler.SeeAll(), th.CallbackDataEqual(consts.SeeAllCallback))
		botHandler.HandleCallbackQuery(handler.Add(), th.CallbackDataEqual(consts.AddCallback))
		botHandler.HandleCallbackQuery(handler.Menu(), th.CallbackDataEqual(consts.MenuCallback))
		botHandler.HandleCallbackQuery(handler.AnyCallbackRouter(), th.AnyCallbackQuery())

		// start listening

		go func() {
			botHandler.Start()
		}()
	}

	go func() {
		err = bot.StartWebhook("0.0.0.0:" + botConf.Port)
		if err != nil {
			panic(err)
		}
	}()

	defer func() {
		_ = bot.StopWebhook()
		botHandler.Stop()
	}()

	ctx := context.Background()
	nctx, stop := signal.NotifyContext(ctx, os.Interrupt, os.Kill)

	defer stop()
	<-nctx.Done()
}

type TimePath struct{}

func (hp TimePath) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tm := time.Now().Format(time.UnixDate)
	w.Write([]byte("The time is: " + tm))
}

type UpdatePath struct{}

func (up UpdatePath) UpdateRepos(behaviorHandler *behaviors.BehaviorHandler, logger zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			msg := "could not read request body"
			logger.Error(msg)
			w.Write([]byte(msg))
			return
		}

		bodyStr := string(b)

		if bodyStr != os.Getenv("SUPER_SECRET_TOKEN") {
			w.Write([]byte(errors.ErrUpdateIncorrectToken.Error()))
			logger.Error(errors.ErrUpdateIncorrectToken)
			return
		}

		failedRepoErrors := behaviorHandler.UpdateRepos()
		marshaledErrors, err := json.Marshal(failedRepoErrors)
		if err != nil {
			w.Write([]byte(err.Error()))
			logger.Error(err)
			return
		}

		if len(failedRepoErrors) != 0 {
			w.Write(marshaledErrors)
			logger.Error(marshaledErrors)
			return
		}

		msg := "Repos updated succesfully with no funky business"
		logger.Info(msg)
		w.Write([]byte(msg))
	}
}

func updateLoop(ctx context.Context, updates <-chan telego.Update, behaviorHandler *behaviors.BehaviorHandler, logger zap.SugaredLogger) {
	awaitingAddRepo := map[int64]struct{}{}
	type void struct{}
	var set void

	for {
		select {
		case <-ctx.Done():
			return
		case update := <-updates:
			var err error

			// command
			if update.Message != nil {
				chatID := update.Message.Chat.ID

				switch update.Message.Text {
				case "/start":
					err = behaviorHandler.Start(chatID)
					if err != nil {
						logger.Error(err)
					}

				case "/about":
					err = behaviorHandler.About(chatID)
					if err != nil {
						logger.Error(err)
					}

				default:
					if _, ok := awaitingAddRepo[chatID]; !ok {
						err = behaviorHandler.UnknownCommand(chatID)
						if err != nil {
							logger.Error(err)
						}
					} else {
						err = behaviorHandler.SentRepo(update.Message.Text, update.Message.MessageID, chatID)
						if err != nil {
							logger.Error(err)
						}
					}
				}
			} else {
				chatID := update.CallbackQuery.Message.Chat.ID
				messageID := update.CallbackQuery.Message.MessageID

				switch update.CallbackQuery.Data {
				case consts.SeeAllCallback:
					err := behaviorHandler.SeeAll(chatID, messageID)
					if err != nil {
						logger.Error(err)
					}

				case consts.AddCallback:
					err = behaviorHandler.Add(chatID, messageID)
					if err != nil {
						logger.Error(err)
					}
					awaitingAddRepo[chatID] = set

				case consts.MenuCallback:
					err := behaviorHandler.Menu(chatID, messageID)
					if err != nil {
						logger.Error(err)
					}
					delete(awaitingAddRepo, chatID)

				//TODO: can probably just use the name now
				// name hash callback, delete
				default:
					err = behaviorHandler.DeleteRepo(chatID, messageID, update.CallbackQuery.Data)
					if err != nil {
						logger.Error(err)
					}
				}
			}
		}
	}
}
