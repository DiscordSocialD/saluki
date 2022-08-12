package main

import "fmt"

// // Handler is executed by AWS Lambda in the main function. Once the request
// // is processed, it returns an Amazon API Gateway response object to AWS Lambda
// func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
//
//	if IsValidRequest(request) {
//		intactn := interaction.Interaction{
//			Request: request,
//		}
//		return HandleInteraction(intactn)
//	}
//
//	// Invalid request, reject
//	return events.APIGatewayProxyResponse{
//		StatusCode: 401,
//		Body:       "\"Invalid request.\"",
//		Headers: map[string]string{
//			"Content-Type": "text/plain",
//		},
//	}, nil
//
// }
//
//	func IsValidRequest(request events.APIGatewayProxyRequest) bool {
//		pubkey, err := GetDiscordPublicKey()
//		return err == nil && interactions.Verify(request, pubkey)
//	}
//
//	func main() {
//		lambda.Start(Handler)
//	}
func main() {
	fmt.Println("TODO")
}
