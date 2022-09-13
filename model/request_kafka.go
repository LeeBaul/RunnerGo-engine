package model

import (
	"encoding/json"
	"github.com/Shopify/sarama"
	"kp-runner/log"
	"strconv"
)

/*
 将需要的测试数据写入到kafka中
*/

// SendKafkaMsg 发送消息到kafka
func SendKafkaMsg(kafkaProducer sarama.SyncProducer, resultDataMsgCh chan *ResultDataMsg, topic string) {
	defer kafkaProducer.Close()
	for {
		if resultDataMsg, ok := <-resultDataMsgCh; ok {
			msg, err := json.Marshal(resultDataMsg)
			if err != nil {
				log.Logger.Error("json转换失败", err)
				break
			}
			DataMsg := &sarama.ProducerMessage{}
			DataMsg.Topic = topic
			DataMsg.Key = sarama.StringEncoder(strconv.FormatInt(resultDataMsg.PlanId, 10) + "-" + resultDataMsg.SceneId)
			DataMsg.Value = sarama.StringEncoder(msg)
			_, _, err = kafkaProducer.SendMessage(DataMsg)
			if err != nil {
				log.Logger.Error("向kafka发送消息失败", err)
				break
			}
		} else {
			break
		}

	}
}

// NewKafkaProducer 构建生产者
func NewKafkaProducer(addrs []string) (kafkaProducer sarama.SyncProducer, err error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll           // 发送完数据需要leader和follow都确认
	config.Producer.Partitioner = sarama.NewHashPartitioner    // 设置选择分区的策略为Hash,当设置key时，所有的key的消息都在一个分区Partitioner里
	config.Producer.Return.Successes = true                    // 成功交付的消息将在success channel返回
	kafkaProducer, err = sarama.NewSyncProducer(addrs, config) // 生产者客户端
	if err != nil {
		return
	}
	return
}
