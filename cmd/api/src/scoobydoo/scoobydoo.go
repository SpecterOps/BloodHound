package scoobydoo

import (
	"context"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/specterops/bloodhound/log"
)

type ScoobyServiceInterface interface {
	EnqueueJob(exchange, routingKey string, payload []byte) error
	//FetchJob() error
	//UpdateJob() error
	//DeleteJob() error
}

type ScoobyService struct {
	Scooby     *amqp.Connection
	ScoobyChan *amqp.Channel
	Snacks     map[string]*amqp.Queue
}

func NewScoobyService() ScoobyService {
	return ScoobyService{Snacks: map[string]*amqp.Queue{}}
}

func (s *ScoobyService) ensureScoobyChan() error {
	if s.ScoobyChan == nil || s.ScoobyChan.IsClosed() {
		var err error
		err = s.ensureScoob()
		if err != nil {
			log.Infof("ensureScoobyChan get connection %v", err)
		}
		s.ScoobyChan, err = s.Scooby.Channel()
		if err != nil {
			log.Infof("Failed to make scooby chan")
			return err
		}
	}
	return nil
}

func (s *ScoobyService) ensureQueue(queueName string) error {
	err := s.ensureScoobyChan()
	if err != nil {
		return err
	}

	if _, ok := s.Snacks[queueName]; !ok {
		q, err := s.ScoobyChan.QueueDeclare(
			queueName, // name
			true,      // durable
			false,     // delete when unused
			false,     // exclusive
			false,     // no-wait
			nil,       // arguments
		)

		if err != nil {
			log.Infof("Failed to declare a queue")
			return err
		}

		s.Snacks[queueName] = &q
	}
	return nil
}

func (s *ScoobyService) ensureScoob() error {
	if s.Scooby == nil {
		var err error
		s.Scooby, err = amqp.Dial("amqp://guest:guest@rabbitmq/")
		if err != nil {
			log.Infof("Failed to connect")
			return err
		}
	}
	return nil
}

func (s *ScoobyService) EnqueueJob(queueName string, payload []byte) error {
	err := s.ensureScoob()
	if err != nil {
		log.Warnf("Failed to connect in enqueue job %v", err)
		return err
	}

	// This calls ensure channel every time
	err = s.ensureQueue(queueName)
	if err != nil {
		log.Warnf("Failed to ensure queue for %s %v", queueName, err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = s.ScoobyChan.PublishWithContext(ctx,
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        payload,
		})

	if err != nil {
		log.Infof(" [x] Sent failed")
	}

	return nil
}
