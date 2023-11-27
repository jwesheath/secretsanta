package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/smtp"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

type Participant struct {
	Name    string
	Address string
	Email   string
}

type Pair struct {
	Santa     Participant
	Recipient Participant
}

func main() {
	godotenv.Load()

	if len(os.Args) != 2 {
		panic("Usage: go run santa.go <participants>")
	}
	participantsFilepath := os.Args[1]

	participants := getParticipants(participantsFilepath)
	pairs := makeParticipantPairs(participants)

	var wg sync.WaitGroup
	wg.Add(len(pairs))
	for _, pair := range pairs {
		go func(pair Pair) {
			defer wg.Done()
			sendMail(pair)
		}(pair)
	}
	wg.Wait()
}

func getParticipants(participantsFilepath string) []Participant {
	data, readErr := os.ReadFile(participantsFilepath)
	if readErr != nil {
		panic("Couldn't read from participantsFilepath")
	}

	var participants []Participant
	unmarshalErr := json.Unmarshal(data, &participants)
	if unmarshalErr != nil {
		panic("Couldn't unmarshal participants in participantsFilepath")
	}

	return participants
}

func makeParticipantPairs(participants []Participant) []Pair {
	size := len(participants)
	perm := rand.Perm(size)
	pairs := make([]Pair, size)
	for i := range perm[:size-1] {
		pairs[i] = Pair{participants[perm[i]], participants[perm[i+1]]}
	}
	pairs[size-1] = Pair{participants[perm[size-1]], participants[perm[0]]}
	return pairs
}

func sendMail(pair Pair) {
	server := os.Getenv("SMTP_SERVER")
	port := os.Getenv("SMTP_PORT")
	from := os.Getenv("EMAIL")
	password := os.Getenv("PASSWORD")

	message := "From: " + from + "\n" +
		"To: " + pair.Santa.Email + "\n" +
		"Subject: Your Secret Santa assignment\n\n" +
		"Dear " + pair.Santa.Name + ",\n" +
		"You are most cordially invited to buy a lovely present for " + pair.Recipient.Name + ". " +
		"Their address is " + pair.Recipient.Address + ".\n" +
		"Merriest of Christmases,\n" +
		"Ham\n"

	auth := smtp.PlainAuth("", from, password, server)

	err := smtp.SendMail(
		server+":"+port,
		auth,
		from,
		[]string{pair.Santa.Email},
		[]byte(message),
	)

	if err != nil {
		fmt.Println("Error sending email to: " + pair.Santa.Email)
		fmt.Println(err)
		return
	}

	fmt.Println("Sent email to " + pair.Santa.Name + " at " + pair.Santa.Email)
}
