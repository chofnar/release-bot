package consts

import (
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

var StartKeyboard *telego.InlineKeyboardMarkup = tu.InlineKeyboard(
	tu.InlineKeyboardRow(
		telego.InlineKeyboardButton{
			Text: SeeAllReposMessage,
			CallbackData: SeeAllCallback,
		},
	),
	tu.InlineKeyboardRow(
		telego.InlineKeyboardButton{
			Text: AddRepoMessage,
			CallbackData: AddCallback,
		},
	),
)
