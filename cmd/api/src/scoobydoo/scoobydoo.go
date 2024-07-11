package scoobydoo

import (
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
	SnackMap     map[string]*amqp.Queue
}

type ScoobySnack struct {
	ContentType string
	Body []byte
}

func NewScoobyService() ScoobyService {
	return ScoobyService{SnackMap: map[string]*amqp.Queue{}}
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

func (s *ScoobyService) EnsureQueue(queueName string) error {
	err := s.ensureScoobyChan()
	if err != nil {
		return err
	}

	if _, ok := s.SnackMap[queueName]; !ok {
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

		s.SnackMap[queueName] = &q
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

func (s *ScoobyService) EnqueueJob(queueName string, snack ScoobySnack) error {
	err := s.ensureScoob()
	if err != nil {
		log.Warnf("Failed to connect in enqueue job %v", err)
		return err
	}

	// This calls ensure channel every time
	err = s.EnsureQueue(queueName)
	if err != nil {
		log.Warnf("Failed to ensure queue for %s %v", queueName, err)
		return err
	}

	err = s.ScoobyChan.Publish(
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: snack.ContentType,
			Body:        snack.Body,
		})

	if err != nil {
		log.Infof(" [x] Sent failed")
	}

	return nil
}
