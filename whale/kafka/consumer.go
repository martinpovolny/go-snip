package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"
	"whale/models"
	"whale/repository"

	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
)

type personMessage struct {
	ExternalID  string `json:"external_id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	DateOfBirth string `json:"date_of_birth"`
}

type Consumer struct {
	client *kgo.Client
	repo   repository.PersonRepository
}

// EnsureTopic creates the topic if it does not already exist.
func EnsureTopic(ctx context.Context, brokers []string, topic string) {
	cl, err := kgo.NewClient(kgo.SeedBrokers(brokers...))
	if err != nil {
		log.Printf("kafka: admin client: %v", err)
		return
	}
	adm := kadm.NewClient(cl)
	defer adm.Close()

	resp, err := adm.CreateTopics(ctx, 1, 1, nil, topic)
	if err != nil {
		log.Printf("kafka: create topic %s: %v", topic, err)
		return
	}
	for _, r := range resp {
		if r.Err != nil {
			log.Printf("kafka: topic %s: %v", r.Topic, r.Err)
		} else {
			log.Printf("kafka: created topic %s", r.Topic)
		}
	}
}

func NewConsumer(brokers []string, topic string, group string, repo repository.PersonRepository) (*Consumer, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ConsumerGroup(group),
		kgo.ConsumeTopics(topic),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
	)
	if err != nil {
		return nil, err
	}
	return &Consumer{client: client, repo: repo}, nil
}

func (c *Consumer) Run(ctx context.Context) {
	defer c.client.Close()
	for {
		fetches := c.client.PollFetches(ctx)
		if fetches.IsClientClosed() || ctx.Err() != nil {
			return
		}
		fetches.EachError(func(t string, p int32, err error) {
			log.Printf("kafka fetch error topic=%s partition=%d: %v", t, p, err)
		})
		fetches.EachRecord(func(r *kgo.Record) {
			log.Printf("kafka: received message topic=%s partition=%d offset=%d", r.Topic, r.Partition, r.Offset)
			var msg personMessage
			if err := json.Unmarshal(r.Value, &msg); err != nil {
				log.Printf("kafka: failed to unmarshal message: %v", err)
				return
			}
			dob, err := time.Parse(time.RFC3339, msg.DateOfBirth)
			if err != nil {
				log.Printf("kafka: invalid date_of_birth %q: %v", msg.DateOfBirth, err)
				return
			}
			person := &models.Person{
				ExternalID:  msg.ExternalID,
				Name:        msg.Name,
				Email:       msg.Email,
				DateOfBirth: dob,
			}
			if err := c.repo.Save(person); err != nil {
				log.Printf("kafka: failed to save person external_id=%s: %v", msg.ExternalID, err)
				return
			}
			log.Printf("kafka: saved person external_id=%s", msg.ExternalID)
		})
	}
}
