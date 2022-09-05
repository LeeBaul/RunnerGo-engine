package model

import (
	"fmt"
	"sync"
	"testing"
)

func TestSendKafkaMsg(t *testing.T) {
	address := "172.17.101.188:9092"
	kafkaProducer, err := NewKafkaProducer([]string{address})
	if err != nil {
		fmt.Println("kafka连接失败", err)
		return
	}

	resultDataMsgCh := make(chan *ResultDataMsg, 100)

	resultDataMsg := new(ResultDataMsg)

	for i := 0; i < 100; i++ {
		resultDataMsg.TargetId = int64(i)
		resultDataMsgCh <- resultDataMsg
		fmt.Println(resultDataMsg)
	}
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go SendKafkaMsg(kafkaProducer, resultDataMsgCh, "StressTestData", wg)
	wg.Wait()

}
