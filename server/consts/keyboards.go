package consts

import (
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

var StartKeyboard *telego.InlineKeyboardMarkup = tu.InlineKeyboard(
	tu.InlineKeyboardRow(
		telego.InlineKeyboardButton{
			Text:         SeeAllReposMessage,
			CallbackData: SeeAllCallback,
		},
	),
	tu.InlineKeyboardRow(
		telego.InlineKeyboardButton{
			Text:         AddRepoMessage,
			CallbackData: AddCallback,
		},
	),
)

var AddAnotherRepoKeyboard *telego.InlineKeyboardMarkup = tu.InlineKeyboard(
	tu.InlineKeyboardRow(
		telego.InlineKeyboardButton{
			Text:         Yes,
			CallbackData: AddCallback,
		},
		telego.InlineKeyboardButton{
			Text:         No,
			CallbackData: MenuCallback,
		},
	),
)

var CancelAddKeyboard *telego.InlineKeyboardMarkup = tu.InlineKeyboard(
	tu.InlineKeyboardRow(
		telego.InlineKeyboardButton{
			Text:         ShowingAddRepoCancel,
			CallbackData: MenuCallback,
		},
	),
)
