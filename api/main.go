package main

import (
	"os"
	"fmt"
	"log"
	"regexp"
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/translate"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type APIResponse struct {
	Message  string `json:"message"`
}

type Response events.APIGatewayProxyResponse

const layout              string = "2006-01-02 15:04"
const languageCodeJa      string = "ja"
const languageCodeEn      string = "en"

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (Response, error) {
	var jsonBytes []byte
	var err error
	d := make(map[string]string)
	json.Unmarshal([]byte(request.Body), &d)
	if v, ok := d["action"]; ok {
		switch v {
		case "sendmessage" :
			if m, ok := d["message"]; ok {
				res, e := sendMessage(m)
				if e != nil {
					err = e
				} else {
					jsonBytes, _ = json.Marshal(APIResponse{Message: res})
				}
			}
		}
	}
	log.Print(request.RequestContext.Identity.SourceIP)
	if err != nil {
		log.Print(err)
		jsonBytes, _ = json.Marshal(APIResponse{Message: fmt.Sprint(err)})
		return Response{
			StatusCode: 500,
			Body: string(jsonBytes),
		}, nil
	}
	return Response {
		StatusCode: 200,
		Body: string(jsonBytes),
	}, nil
}

func sendMessage(message string)(string, error) {
	if regexp.MustCompile(`[a-zA-Z]`).Match([]byte(message)) {
		return translateText(message, languageCodeEn, languageCodeJa)
	}
	return translateText(message, languageCodeJa, languageCodeEn)
}

func translateText(message string, sourceLanguageCode string, targetLanguageCode string)(string, error) {
	svc := translate.New(session.New(), &aws.Config{
		Region: aws.String(os.Getenv("REGION")),
	})
	params := &translate.TextInput{
		Text: aws.String(message),
		SourceLanguageCode: aws.String(sourceLanguageCode),
		TargetLanguageCode: aws.String(targetLanguageCode),
	}
	res, err := svc.Text(params)
	if err != nil {
		log.Print(err)
		return "", err
	}
	return aws.StringValue(res.TranslatedText), nil
}

func main() {
	lambda.Start(HandleRequest)
}
