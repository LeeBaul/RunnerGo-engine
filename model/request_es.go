package model

import (
	"context"
	"github.com/olivere/elastic/v7"
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

//
//func QueryReport(client *elastic.Client, index string, reportId string) (result *EsSceneTestResultDataMsg) {
//	if client == nil {
//		return
//	}
//	ctx := context.Background()
//	queryEs := elastic.NewBoolQuery()
//	queryEs = queryEs.Must(elastic.NewMatchQuery("report_id", reportId))
//	res, err := client.Search(index).Query(queryEs).Sort("time_stamp", true).Pretty(true).Do(ctx)
//	if err != nil {
//		log2.Logger.Error("获取报告: "+reportId+" ,数据失败：", err)
//		return
//	}
//	if res == nil || len(res.Hits.Hits) < 1 {
//		log2.Logger.Error("获取报告: " + reportId + " ,数据为空")
//		return
//	}
//	err = json.Unmarshal(res.Hits.Hits[0].Source, &result)
//	if err != nil {
//		log2.Logger.Error("获取报告: "+reportId+" ,数据转换失败", err)
//		return
//	}
//	return
//}
