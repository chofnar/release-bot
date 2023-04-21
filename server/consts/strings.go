package consts

const (
	SeeAllReposMessage = "See all repos"

	AddRepoMessage = "Add a repo"

	AboutMessage = "The bot will ignore prereleases as of 21 April 2023\nFind my source code at github.com/chofnar/release-bot\nThe bot is based on the Telegram API implementation in Go made by Artem Yadelskyi: https://github.com/mymmrac/telego\nFor suggestions and issues, contact the creator of the bot at catalin.hofnar@gmail.com"

	StartMessage = "Pick one of the options below"

	UnknownCommandMessage = "Sorry, I don't understand. Please pick one of the valid options."

	ShowingAddRepoMessage = "Send a message containing your repo in one of the following formats: user/repo, https://github.com/user/repo"

	ShowingAddRepoCancel = "Cancel"

	InvalidRepoMessage = "Error: Invalid repo. Send a message containing your repo in one of the following formats: user/repo, https://github.com/user/repo"

	ShowingAllReposMessage = "Here's all your added repos with their releases"

	ShowingAllReposButNoneFoundMessage = "There are no watched repos. Add one?"

	DelteRepoEmoji = "🗑️"

	// Very creative
	Yes = "Yes"

	No = "No"

	AddedRepoSuccesfully = "Repo added succesfully. Add another?"

	AddedRepoSuccesfullyNoReleases = "Repo added succesfully but it has no releases. I will ping you when there is one. Add another?"

	RepoExists = "Repo already exists in your watched list. Try another?"

	RepoNotFound = "I could not find the repo. Try again?"

	CheckRepo = "Check it out"
)
