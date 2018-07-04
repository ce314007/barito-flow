package flow

import (
	"strings"

	"github.com/Shopify/sarama"
)

type KafkaAdmin interface {
	RefreshTopics() error
	Topics() []string
	TopicsWithSuffix(suffix string) []string
	Exist(topic string) bool
	CreateTopic(topic string, numPartitions int32, replicationFactor int16) error
	Close()
}

type kafkaAdmin struct {
	topics []string
	client sarama.Client
	config *sarama.Config
}

func NewKafkaAdmin(brokers []string, config *sarama.Config) (KafkaAdmin, error) {

	client, err := sarama.NewClient(brokers, config)
	if err != nil {
		return nil, err
	}

	return &kafkaAdmin{
		client: client,
		config: config,
	}, nil
}

func (a *kafkaAdmin) CreateTopicIfNotExist(topic string) {
	// TODO: fetch cache topics list
	// TODO: check if topic exist
	// TODO: refresh cache if topic not exist and check again
	// TODO: create topic
}

func (a *kafkaAdmin) RefreshTopics() (err error) {
	topics, err := a.client.Topics()
	if err != nil {
		return
	}

	a.topics = topics
	return
}

func (a *kafkaAdmin) Topics() []string {
	if len(a.topics) <= 0 {
		a.RefreshTopics()
	}

	return a.topics
}

func (a *kafkaAdmin) Exist(topic string) bool {
	for _, topic0 := range a.Topics() {
		if topic0 == topic {
			return true
		}
	}

	return false
}

func (a *kafkaAdmin) TopicsWithSuffix(suffix string) (topics []string) {
	for _, topic := range a.Topics() {
		if strings.HasSuffix(topic, suffix) {
			topics = append(topics, topic)
		}
	}
	return
}

func (a *kafkaAdmin) Close() {
	a.client.Close()
}

func (a *kafkaAdmin) CreateTopic(topic string, numPartitions int32, replicationFactor int16) (err error) {

	request := a.createTopicsRequest(topic, numPartitions, replicationFactor)

	for _, broker := range a.client.Brokers() {
		err = broker.Open(a.config)
		defer broker.Close()

		if err != nil {
			return
		}

		_, err = broker.CreateTopics(request)
		if err != nil {
			return
		}
	}

	return

}

func (a *kafkaAdmin) createTopicsRequest(topic string, numPartitions int32, replicationFactor int16) *sarama.CreateTopicsRequest {

	var version int16 = 0
	if a.config.Version.IsAtLeast(sarama.V0_11_0_0) {
		version = 1
	}
	if a.config.Version.IsAtLeast(sarama.V1_0_0_0) {
		version = 2
	}

	return &sarama.CreateTopicsRequest{
		Version: version,
		TopicDetails: map[string]*sarama.TopicDetail{
			topic: &sarama.TopicDetail{
				NumPartitions:     numPartitions,
				ReplicationFactor: replicationFactor,
			},
		},
		ValidateOnly: false,
	}

}