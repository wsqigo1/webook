package sarama

import (
	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"testing"
)

var addr = []string{"localhost:9094"}

func TestSyncProducer(t *testing.T) {
	cfg := sarama.NewConfig()
	// 同步的 Producer 一定要设置
	cfg.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(addr, cfg)
	assert.NoError(t, err)
	cfg.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	//cfg.Producer.Partitioner = sarama.NewRandomPartitioner
	//cfg.Producer.Partitioner = sarama.NewHashPartitioner
	//cfg.Producer.Partitioner = sarama.NewManualPartitioner
	//cfg.Producer.Partitioner = sarama.NewConsistentCRCHashPartitioner
	//cfg.Producer.Partitioner = sarama.NewCustomPartitioner()
	_, _, err = producer.SendMessage(&sarama.ProducerMessage{
		Topic: "test_topic",
		Value: sarama.StringEncoder("hello，这是一条消息"),
		// 会在生产者和消费者之间传递
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("key1"),
				Value: []byte("value1"),
			},
		},
		Metadata: "这是 metadata",
	})
	assert.NoError(t, err)
}

func TestAsyncProducer(t *testing.T) {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	cfg.Producer.Return.Errors = true
	producer, err := sarama.NewAsyncProducer(addr, cfg)
	assert.NoError(t, err)
	msgs := producer.Input()
	msgs <- &sarama.ProducerMessage{
		Topic: "test_topic",
		Value: sarama.StringEncoder("这是一条消息"),
		// 会在生产者和消费者之间传递的
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("key1"),
				Value: []byte("value1"),
			},
		},
		Metadata: "这是 metadata",
	}
	// 在实践中，一般是开另外一个 goroutine 来处理结果的
	select {
	case msg := <-producer.Successes():
		// 这边是成功了
		t.Log("发送成功", string(msg.Value.(sarama.StringEncoder)))
	case err := <-producer.Errors():
		// 这边是出错了
		val, _ := err.Msg.Value.Encode()
		t.Log("发送失败", err.Err, string(val))
	}
}
