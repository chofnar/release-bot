package consts

const (
	SeeAllReposMessage = "See all repos"

	AddRepoMessage = "Add a repo"

	AboutMessage = "Find my source code at github.com/chofnar/release-bot\nConsider supporting this project :) https://www.buymeacoffee.com/chofnar\nThe bot is based on the Telegram API implementation in Go made by Artem Yadelskyi: https://github.com/mymmrac/telego\nFor suggestions and issues, contact the creator of the bot at catalin.hofnar@gmail.com"

	StartMessage = "Pick one of the options below"

	UnknownCommandMessage = "Sorry, I don't understand. Please pick one of the valid options."

	ShowingAddRepoMessage = "Send a message containing your repo in one of the following formats: user/repo, https://github.com/user/repo"

	ShowingAddRepoCancel = "Cancel"

	InvalidRepoMessage = "Error: Invalid repo. Send a message containing your repo in one of the following formats: user/repo, https://github.com/user/repo"

	ShowingAllReposMessage = "Here's all your added repos with their releases. The third button being active means you will be notified of prereleases for the repo."

	ShowingAllReposButNoneFoundMessage = "There are no watched repos. Add one?"

	DelteRepoEmoji = "🗑️"

	PreReleasesActive   = "Pre: ✔️"
	PreReleasesInactive = "Pre: ❌"

	// Very creative
	Yes = "Yes"

	No = "No"

	AddedRepoSuccesfully = "Repo added succesfully. Add another?"

	AddedRepoSuccesfullyNoReleases = "Repo added succesfully but it has no releases. I will ping you when there is one. Add another?"

	RepoExists = "Repo already exists in your watched list. Try another?"

	RepoNotFound = "I could not find the repo. Try again?"

	CheckRepo = "Check it out"

	OperationPrefix = "OP_"
)
