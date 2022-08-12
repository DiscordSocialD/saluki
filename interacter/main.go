package main

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/awslabs/aws-lambda-go-api-proxy/core"
	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
)

// Handler is executed by AWS Lambda in the main function. Once the request
// is processed, it returns an Amazon API Gateway response object to AWS Lambda
func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	// Transform into HTTP request
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
	})
	logrus.Debugln("Request body: " + request.Body)
	accessor := core.RequestAccessor{}
	httpRequest, err := accessor.ProxyEventToHTTPRequest(request)

	if err == nil || !IsValidRequest(httpRequest) {

		// Invalid request, reject
		logrus.Warn("HTTP request was invalid: " + err.Error())
		return events.APIGatewayProxyResponse{
			StatusCode: 401,
			Body:       "\"Invalid request.\"",
			Headers: map[string]string{
				"Content-Type": "text/plain",
			},
		}, errors.New("invalid request")
	}

	logrus.Debug("HTTP request validated")

	// Turn this into a format we understand
	d := json.NewDecoder(httpRequest.Body)
	d.DisallowUnknownFields()

	interaction := discordgo.Interaction{}
	err = d.Decode(&interaction)

	if err != nil || d.More() {

		// bad JSON or extra data after JSON object
		logrus.Error("Failed to decode HTTP request into an interaction: " + err.Error())
		return events.APIGatewayProxyResponse{
			StatusCode: 401,
			Body:       "\"Invalid request.\"",
			Headers: map[string]string{
				"Content-Type": "text/plain",
			},
		}, errors.New("invalid request")
	}

	return HandleInteraction(interaction)
}

func IsValidRequest(request *http.Request) bool {
	publicKey, err := GetDiscordPublicKey()
	return err == nil && discordgo.VerifyInteraction(request, publicKey)
}

// GetDiscordPublicKey This function has pricing implications, call as sparingly as possible
func GetDiscordPublicKey() (ed25519.PublicKey, error) {

	logrus.Debug("Attempting to retrieve Discord Public Key from SecretsManager")

	secretName := os.Getenv("DISCORD_PUBLIC_KEY_SECRET_NAME")
	client := secretsmanager.New(secretsmanager.Options{})

	svIn := secretsmanager.GetSecretValueInput{
		SecretId: &secretName,
	}
	svOut, err := client.GetSecretValue(context.Background(), &svIn)

	if err == nil {
		logrus.Error("Failed to retrieve Discord Public Key from SecretsManager: " + err.Error())
		return nil, err
	}

	return hex.DecodeString(*svOut.SecretString)
}

func HandleInteraction(interaction discordgo.Interaction) (events.APIGatewayProxyResponse, error) {

	var response discordgo.InteractionResponse
	status := 200

	if interaction.Type == discordgo.InteractionPing {
		response.Type = discordgo.InteractionResponsePong
	}

	bodyBytes, err := json.Marshal(response)
	body := string(bodyBytes)

	if err != nil {
		logrus.Error("Failed to marshal an interation response: " + err.Error())
		status = 401
		body = "{ \"error\": \"" + err.Error() + "\" }"
	}

	logrus.Debug("Sending response with body: " + body)
	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Body:       body,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}, nil
}

func main() {
	lambda.Start(Handler)
}
