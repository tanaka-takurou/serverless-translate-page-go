package main

import (
	"os"
	"fmt"
	"log"
	"regexp"
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/translate"
)

type APIResponse struct {
	Message  string `json:"message"`
}

type Response events.APIGatewayProxyResponse

var cfg aws.Config
var translateClient *translate.Client

const languageCodeJa string = "ja"
const languageCodeEn string = "en"

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (Response, error) {
	var jsonBytes []byte
	var err error
	d := make(map[string]string)
	json.Unmarshal([]byte(request.Body), &d)
	if v, ok := d["action"]; ok {
		switch v {
		case "sendmessage" :
			if m, ok := d["message"]; ok {
				res, e := sendMessage(ctx, m)
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

func sendMessage(ctx context.Context, message string)(string, error) {
	if regexp.MustCompile(`[a-zA-Z]`).Match([]byte(message)) {
		return translateText(ctx, message, languageCodeEn, languageCodeJa)
	}
	return translateText(ctx, message, languageCodeJa, languageCodeEn)
}

func translateText(ctx context.Context, message string, sourceLanguageCode string, targetLanguageCode string)(string, error) {
	if translateClient == nil {
		translateClient = getTranslateClient(ctx)
	}
	params := &translate.TranslateTextInput{
		Text: aws.String(message),
		SourceLanguageCode: aws.String(sourceLanguageCode),
		TargetLanguageCode: aws.String(targetLanguageCode),
	}
	res, err := translateClient.TranslateText(ctx, params)
	if err != nil {
		log.Print(err)
		return "", err
	}
	return aws.ToString(res.TranslatedText), nil
}

func getTranslateClient(ctx context.Context) *translate.Client {
	return translate.NewFromConfig(getConfig(ctx))
}

func getConfig(ctx context.Context) aws.Config {
	var err error
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("REGION")))
	if err != nil {
		log.Print(err)
	}
	return cfg
}

func main() {
	lambda.Start(HandleRequest)
}
