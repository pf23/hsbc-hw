package main

import (
	"context"
	"time"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/option"
)

var (
	TestTopic                = "test-topic"
	ProjectID                = "openhabour-11223"
	WebPushRegistrationToken = "fwZLrPYwBGSMgXpF3MAAN9:APA91bFaDsF0ACUbKhYlweJVuB9sd3xaueUpA-tFE7t5UO8VS3FLCJs8kOeiAG8fVFJ5G5zn33nBxgfiTo1pCfnIu6URlEn7DLsBX_1h3YpWZayXP0sPL6dM4TUGAUan2itJyBcmoH8M"
	ServerKey                = "AAAAFDxzSXA:APA91bH-gWTcDFC905f2VdfwB66D8PbY9Pecg2NhsxHnzO9_I9WfieThAKnoJrZHTeLCIug57lDg2YYcHTtHT-SbNYR4GM-o9LEJbwve0o5dk6ClP_3Hc1dA6BURIUEUNZU_vb_QeZ9Y"
	CredentialFile           = "/Users/pf/Desktop/fcm/openhabour-11223-firebase-adminsdk-bvhyz-dcf9f45c36.json"
)

func initClient() *messaging.Client {
	config := &firebase.Config{ProjectID: ProjectID}
	opt := option.WithCredentialsFile(CredentialFile)
	app, err := firebase.NewApp(context.Background(), config, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}
	log.Info("new app created")
	/*
		client, err := app.Auth(context.Background())
		if err != nil {
			log.Fatalf("error getting Auth client: %v\n", err)
		}
	*/

	// TODO: in internal environment, need to set up proxy for HTTP using messaging.NewClient()
	mClient, err := app.Messaging(context.Background())
	if err != nil {
		log.Fatalf("error getting Messaging client: %v\n", err)
	}
	log.Info("messaginbg client created")
	return mClient
}

func getCurrentDatetime() string {
	return time.Now().Format(time.RFC3339)
}

func runServer() {
	client := initClient()
	msg := &messaging.Message{
		Data: map[string]string{
			"name":    "fei.pang",
			"subject": "test-notify-subject-pf",
		},
		Notification: &messaging.Notification{
			Title: "test-notify-from-pf",
			Body:  getCurrentDatetime(),
			//ImageURL: ,
		},
		//Android:    &messaging.AndroidConfig{},
		Webpush: &messaging.WebpushConfig{
			Notification: &messaging.WebpushNotification{
				Title: "test-notify-from-pf",
				Body:  getCurrentDatetime(),
				//Icon:  "https://my-server/icon.png",
			},
		},
		//APNS: &messaging.APNSConfig{},
		//FCMOptions: &messaging.FCMOptions{},
		// one of below
		Token: WebPushRegistrationToken,
		//Topic: "",
		//Condition: "",
	}
	resp, err := client.Send(context.Background(), msg)
	//batchResp, err := client.SendAllDryRun(context.Background(), []*messaging.Message{msg})
	if err != nil {
		log.Fatalf("error sending message: %v\n", err)
	}
	log.Info(resp)
}

func runClient() {

}

func main() {
	runServer()
	/*
		if len(os.Args) != 2 {
			fmt.Println("run either of: ./fcm client or ./fcm server")
		}
		mode := os.Args[1]
		switch mode {
		case "server":
			runServer()
		default:
			runClient()
		}
	*/
}
