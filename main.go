package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
)

// cache client object when service starts
var api *slack.Client

func main() {

	err := godotenv.Load(".env")

	// do not break code if .evn not found
	if err != nil {
		fmt.Printf("Error loading .env file %s", err)
	}

	// APP_TOKEN is mandatory as we cant interact with slack otherise
	appToken := os.Getenv("APP_TOKEN")
	if appToken == "" {
		log.Panicln("APP_TOKEN not found in env. Exiting..")
	}
	api = slack.New(appToken)

	helloHandler := func(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, "Hello, world!\n")
	}

	http.HandleFunc("/hello", helloHandler)
	http.HandleFunc("/shuffle", slashCommandHandler)

	log.Println("Server started..")
	log.Fatal(http.ListenAndServe(":8000", nil))
}

// Handler for the /shuffle request. Here we parse the slack slash command
func slashCommandHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("got shuffle req")
	s, err := slack.SlashCommandParse(r)
	if err != nil {
		log.Println("error parsing slash command", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !s.ValidateToken(os.Getenv("SLACK_VERIFICATION_TOKEN")) {
		log.Println("error validating token")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	switch s.Command {
	case "/shuffle":

		//todo: filter non human users like bots, apps etc
		memberIDs, _, err := api.GetUsersInConversation(&slack.GetUsersInConversationParameters{ChannelID: s.ChannelID})
		if err != nil {
			log.Printf("could not get users for channel %v, reason: %v", s.ChannelID, err)
			w.WriteHeader(http.StatusOK)
			response := fmt.Sprintf("Error: %v", err)
			w.Write([]byte(response))
			return
		}

		memberStringList := strings.Join(getUsersList(memberIDs), "\n")

		text := fmt.Sprintf("Shuffled users:\n%v", memberStringList)
		api.SendMessage(s.ChannelID, slack.MsgOptionText(text, false))
		w.WriteHeader(http.StatusOK)
	default:
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// function to get a randomly shuffled list of input users
func getUsersList(users []string) []string {

	// so weird but thanks SO gods
	rand.Shuffle(len(users), func(i, j int) {
		users[i], users[j] = users[j], users[i]
	})

	// format user for slack msg text
	for i := range users {
		users[i] = fmt.Sprintf("<@%s>", users[i])
	}

	return users
}
