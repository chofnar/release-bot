package telegohandlers

import (
	"strconv"
	"strings"

	"github.com/chofnar/release-bot/internal/server/behaviors"
	"github.com/chofnar/release-bot/internal/server/consts"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"go.uber.org/zap"
)

type Handler struct {
	BehaviorHandler behaviors.BehaviorHandler
	Logger          zap.SugaredLogger
	AwaitingAddRepo map[int64]struct{}
	Limit           int
}

type void struct{}

var set void

func (hc *Handler) Start() telegohandler.Handler {
	return func(bot *telego.Bot, update telego.Update) {
		err := hc.BehaviorHandler.Start(update.Message.Chat.ID)
		if err != nil {
			hc.Logger.Error(err)
		}
	}
}

func (hc *Handler) About() telegohandler.Handler {
	return func(bot *telego.Bot, update telego.Update) {
		err := hc.BehaviorHandler.About(update.Message.Chat.ID)
		if err != nil {
			hc.Logger.Error(err)
		}
	}
}

func (hc *Handler) UnknownOrSent() telegohandler.Handler {
	return func(bot *telego.Bot, update telego.Update) {
		if _, ok := hc.AwaitingAddRepo[update.Message.Chat.ID]; !ok {
			err := hc.BehaviorHandler.UnknownCommand(update.Message.Chat.ID)
			if err != nil {
				hc.Logger.Error(err)
			}
		} else {
			err := hc.BehaviorHandler.SentRepo(update.Message.Text, update.Message.MessageID, update.Message.Chat.ID)
			if err != nil {
				hc.Logger.Error(err)
			}
		}
	}
}

func (hc *Handler) SeeRepos(limit, page int) telegohandler.CallbackQueryHandler {
	return func(bot *telego.Bot, query telego.CallbackQuery) {
		err := hc.BehaviorHandler.SeeRepos(query.Message.Chat.ID, query.Message.MessageID, limit, page)
		if err != nil {
			hc.Logger.Error(err)
		}
	}
}

func (hc *Handler) Menu() telegohandler.CallbackQueryHandler {
	return func(bot *telego.Bot, query telego.CallbackQuery) {
		err := hc.BehaviorHandler.Menu(query.Message.Chat.ID, query.Message.MessageID)
		if err != nil {
			hc.Logger.Error(err)
		}
		delete(hc.AwaitingAddRepo, query.Message.Chat.ID)
	}
}

func (hc *Handler) Add() telegohandler.CallbackQueryHandler {
	return func(bot *telego.Bot, query telego.CallbackQuery) {
		err := hc.BehaviorHandler.Add(query.Message.Chat.ID, query.Message.MessageID)
		if err != nil {
			hc.Logger.Error(err)
		}
		hc.AwaitingAddRepo[query.Message.Chat.ID] = set
	}
}

func (hc *Handler) AnyCallbackRouter() telegohandler.CallbackQueryHandler {
	return func(bot *telego.Bot, query telego.CallbackQuery) {
		if strings.HasPrefix(query.Data, consts.FlipOperationPrefix) {
			err := hc.BehaviorHandler.FlipPreRelease(query.Message.Chat.ID, query.Message.MessageID, query.Data)
			if err != nil {
				hc.Logger.Error(err)
			}
		} else if strings.HasPrefix(query.Data, consts.PreviousOperationPrefix) {
			page, err := strconv.Atoi(strings.TrimPrefix(query.Data, consts.PreviousOperationPrefix))
			if err != nil {
				hc.Logger.Error(err)
				return
			}

			err = hc.BehaviorHandler.SeeRepos(query.Message.Chat.ID, query.Message.MessageID, hc.Limit, page)
			if err != nil {
				hc.Logger.Error(err)
			}
		} else if strings.HasPrefix(query.Data, consts.ForwardOperationPrefix) {
			page, err := strconv.Atoi(strings.TrimPrefix(query.Data, consts.ForwardOperationPrefix))
			if err != nil {
				hc.Logger.Error(err)
				return
			}

			err = hc.BehaviorHandler.SeeRepos(query.Message.Chat.ID, query.Message.MessageID, hc.Limit, page)
			if err != nil {
				hc.Logger.Error(err)
			}
		} else {
			err := hc.BehaviorHandler.DeleteRepo(query.Message.Chat.ID, query.Message.MessageID, query.Data)
			if err != nil {
				hc.Logger.Error(err)
			}
		}
	}
}
