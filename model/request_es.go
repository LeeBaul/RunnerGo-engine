package model

import (
	"context"
	"github.com/olivere/elastic"
	log2 "kp-runner/log"
	"log"
	"os"
	"time"
)

func NewEsClient(host, user, password string) (client *elastic.Client) {
	client, _ = elastic.NewClient(
		elastic.SetURL(host),
		elastic.SetSniff(false),
		elastic.SetBasicAuth(user, password),
		elastic.SetErrorLog(log.New(os.Stdout, "APP", log.Lshortfile)),
		elastic.SetHealthcheckInterval(30*time.Second),
	)
	_, _, err := client.Ping(host).Do(context.Background())
	if err != nil {
		log2.Logger.Error("es连接失败", err)
	}
	return
}
