package main

import (
	"os"
	"fmt"
	"log"
	"bytes"
	"regexp"
	"context"
	"reflect"
	"strings"
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
		translateClient = getTranslateClient()
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
	return stringValue(res.TranslatedText), nil
}

func getTranslateClient() *translate.Client {
	if cfg.Region != os.Getenv("REGION") {
		cfg = getConfig()
	}
	return translate.NewFromConfig(cfg)
}

func getConfig() aws.Config {
	var err error
	newConfig, err := config.LoadDefaultConfig()
	newConfig.Region = os.Getenv("REGION")
	if err != nil {
		log.Print(err)
	}
	return newConfig
}

func stringValue(i interface{}) string {
	var buf bytes.Buffer
	strVal(reflect.ValueOf(i), 0, &buf)
	res := buf.String()
	return res[1:len(res) - 1]
}

func strVal(v reflect.Value, indent int, buf *bytes.Buffer) {
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Struct:
		buf.WriteString("{\n")
		for i := 0; i < v.Type().NumField(); i++ {
			ft := v.Type().Field(i)
			fv := v.Field(i)
			if ft.Name[0:1] == strings.ToLower(ft.Name[0:1]) {
				continue // ignore unexported fields
			}
			if (fv.Kind() == reflect.Ptr || fv.Kind() == reflect.Slice) && fv.IsNil() {
				continue // ignore unset fields
			}
			buf.WriteString(strings.Repeat(" ", indent+2))
			buf.WriteString(ft.Name + ": ")
			if tag := ft.Tag.Get("sensitive"); tag == "true" {
				buf.WriteString("<sensitive>")
			} else {
				strVal(fv, indent+2, buf)
			}
			buf.WriteString(",\n")
		}
		buf.WriteString("\n" + strings.Repeat(" ", indent) + "}")
	case reflect.Slice:
		nl, id, id2 := "", "", ""
		if v.Len() > 3 {
			nl, id, id2 = "\n", strings.Repeat(" ", indent), strings.Repeat(" ", indent+2)
		}
		buf.WriteString("[" + nl)
		for i := 0; i < v.Len(); i++ {
			buf.WriteString(id2)
			strVal(v.Index(i), indent+2, buf)
			if i < v.Len()-1 {
				buf.WriteString("," + nl)
			}
		}
		buf.WriteString(nl + id + "]")
	case reflect.Map:
		buf.WriteString("{\n")
		for i, k := range v.MapKeys() {
			buf.WriteString(strings.Repeat(" ", indent+2))
			buf.WriteString(k.String() + ": ")
			strVal(v.MapIndex(k), indent+2, buf)
			if i < v.Len()-1 {
				buf.WriteString(",\n")
			}
		}
		buf.WriteString("\n" + strings.Repeat(" ", indent) + "}")
	default:
		format := "%v"
		switch v.Interface().(type) {
		case string:
			format = "%q"
		}
		fmt.Fprintf(buf, format, v.Interface())
	}
}

func main() {
	lambda.Start(HandleRequest)
}
