package model

import (
	"encoding/json"
	"github.com/Shopify/sarama"
	"kp-runner/log"
	"time"
)

/*
 将需要的测试数据写入到kafka中
*/

// SendKafkaMsg 发送消息到kafka
func SendKafkaMsg(kafkaProducer sarama.SyncProducer, resultDataMsgCh chan *ResultDataMsg, topic string, partition int32, reportId string) {
	defer kafkaProducer.Close()
	num := int64(0)
	for {
		if resultDataMsg, ok := <-resultDataMsgCh; ok {
			msg, err := json.Marshal(resultDataMsg)
			if err != nil {
				log.Logger.Error("json转换失败", err)
				break
			}
			if num == 0 {
				num = resultDataMsg.MachineNum
			}

			DataMsg := &sarama.ProducerMessage{}
			DataMsg.Topic = topic
			DataMsg.Partition = partition
			DataMsg.Value = sarama.StringEncoder(msg)
			_, _, err = kafkaProducer.SendMessage(DataMsg)
			if err != nil {
				log.Logger.Error("向kafka发送消息失败", err)
				break
			}
		} else {
			// 发送结束消息
			result := new(ResultDataMsg)
			result.ReportId = reportId
			result.End = true
			result.MachineNum = num
			result.Timestamp = time.Now().UnixMilli()
			msg, err := json.Marshal(result)
			if err != nil {
				log.Logger.Error("json转换失败", err)
				break
			}
			DataMsg := &sarama.ProducerMessage{}
			DataMsg.Topic = topic
			DataMsg.Partition = partition
			DataMsg.Value = sarama.StringEncoder(msg)
			_, _, err = kafkaProducer.SendMessage(DataMsg)
			if err != nil {
				log.Logger.Error("向kafka发送消息失败", err)
				break
			}

			log.Logger.Info(result.ReportId, " 报告消息发送结束")

			return

		}

	}
}

// NewKafkaProducer 构建生产者
func NewKafkaProducer(addrs []string) (kafkaProducer sarama.SyncProducer, err error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll           // 发送完数据需要leader和follow都确认
	config.Producer.Partitioner = sarama.NewManualPartitioner  // 设置选择分区的策略为Hash,当设置key时，所有的key的消息都在一个分区Partitioner里
	config.Producer.Return.Successes = true                    // 成功交付的消息将在success channel返回
	kafkaProducer, err = sarama.NewSyncProducer(addrs, config) // 生产者客户端
	if err != nil {
		return
	}
	return
}

// SendStopMsg 发送结束任务消息
//func SendStopMsg(host, reportId string) {
//	ctx := context.TODO()
//
//	systemRoots, err := x509.SystemCertPool()
//	if err != nil {
//		panic(errors.Wrap(err, "cannot load root CA certs"))
//	}
//	creds := credentials.NewTLS(&tls.Config{
//		RootCAs: systemRoots,
//	})
//
//	conn, err := grpc.Dial(host, grpc.WithTransportCredentials(creds))
//	defer func() {
//		grpcErr := conn.Close()
//		if grpcErr != nil {
//			log.Logger.Error("关闭grpc连接失败:", grpcErr)
//		}
//	}()
//	grpcClient := services.NewKpControllerClient(conn)
//	req := new(services.NotifyStopStressReq)
//	req.ReportID, _ = strconv.ParseInt(reportId, 10, 64)
//
//	_, err = grpcClient.NotifyStopStress(ctx, req)
//	if err != nil {
//		log2.Logger.Error("发送停止任务失败", err)
//		return
//	}
//	log2.Logger.Info(reportId, "任务结束， 消息已发送")
//}
