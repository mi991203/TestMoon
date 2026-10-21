package logic

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/kmsg"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/mi991203/TestMoon/backend/testing/execution/internal/runtime"
	"github.com/mi991203/TestMoon/backend/testing/execution/model"
)

const topic = "testmoon.execution.v1"

func StartQueue(deps *runtime.Dependencies) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	producer, e := kgo.NewClient(kgo.SeedBrokers(strings.Split(deps.Config.Kafka, ",")...), kgo.RequiredAcks(kgo.AllISRAcks()))
	if e != nil {
		panic(e)
	}
	deps.Producer = producer
	ping, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	if e = producer.Ping(ping); e != nil {
		panic("Kafka unavailable")
	}
	request := kmsg.NewPtrCreateTopicsRequest()
	t := kmsg.NewCreateTopicsRequestTopic()
	t.Topic = topic
	t.NumPartitions = 3
	t.ReplicationFactor = 1
	request.Topics = append(request.Topics, t)
	response, e := producer.Request(ping, request)
	if e != nil {
		panic(e)
	}
	for _, v := range response.(*kmsg.CreateTopicsResponse).Topics {
		if v.ErrorCode != 0 && v.ErrorCode != 36 {
			panic("Kafka topic creation failed")
		}
	}
	consumer, e := kgo.NewClient(kgo.SeedBrokers(strings.Split(deps.Config.Kafka, ",")...), kgo.ConsumerGroup("testmoon-execution-v1"), kgo.ConsumeTopics(topic), kgo.DisableAutoCommit())
	if e != nil {
		panic(e)
	}
	deps.Consumer = consumer
	go func() { <-ctx.Done(); producer.Close(); consumer.Close(); stop() }()
	go publishLoop(ctx, deps)
	go consumeLoop(ctx, deps)
	go reconcileLoop(ctx, deps)
}

func publishLoop(ctx context.Context, deps *runtime.Dependencies) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		for i := 0; i < 20; i++ {
			var event model.Outbox
			e := deps.DB.Transaction(func(tx *gorm.DB) error {
				if e := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).Where("published = ? AND lease_until < ?", false, time.Now()).Order("created_at").First(&event).Error; e != nil {
					return e
				}
				return tx.Model(&event).Update("lease_until", time.Now().Add(30*time.Second)).Error
			})
			if e != nil {
				break
			}
			publish, cancel := context.WithTimeout(ctx, 10*time.Second)
			e = deps.Producer.ProduceSync(publish, &kgo.Record{Topic: topic, Key: []byte(event.RunID), Value: []byte(event.Payload)}).FirstErr()
			cancel()
			if e == nil {
				deps.DB.Model(&event).Update("published", true)
			} else {
				logrus.WithField("event_id", event.ID).Warn("Kafka publication delayed")
				break
			}
		}
	}
}

func consumeLoop(ctx context.Context, deps *runtime.Dependencies) {
	for {
		fetches := deps.Consumer.PollFetches(ctx)
		if ctx.Err() != nil {
			return
		}
		if len(fetches.Errors()) > 0 {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
			}
			continue
		}
		iter := fetches.RecordIter()
		for !iter.Done() {
			record := iter.Next()
			var event struct {
				EventID string `json:"event_id"`
				RunID   string `json:"run_id"`
				Version int    `json:"version"`
			}
			if json.Unmarshal(record.Value, &event) != nil || event.EventID == "" || event.RunID == "" || event.Version != 1 {
				logrus.Warn("invalid execution event ignored")
				_ = deps.Consumer.CommitRecords(ctx, record)
				continue
			}
			e := deps.DB.Transaction(func(tx *gorm.DB) error {
				res := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&model.ConsumedEvent{ID: event.EventID, CreatedAt: time.Now()})
				if res.Error != nil {
					return res.Error
				}
				if res.RowsAffected == 0 {
					return nil
				}
				return tx.Model(&model.Run{}).Where("id = ? AND state = ?", event.RunID, "created").Update("state", "queued").Error
			})
			if e != nil {
				logrus.Error("execution event transaction failed")
				for e != nil && ctx.Err() == nil {
					select {
					case <-ctx.Done():
						return
					case <-time.After(time.Second):
					}
					e = deps.DB.Transaction(func(tx *gorm.DB) error {
						res := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&model.ConsumedEvent{ID: event.EventID, CreatedAt: time.Now()})
						if res.Error != nil {
							return res.Error
						}
						if res.RowsAffected == 0 {
							return nil
						}
						return tx.Model(&model.Run{}).Where("id = ? AND state = ?", event.RunID, "created").Update("state", "queued").Error
					})
				}
			}
			if ctx.Err() == nil {
				if e = deps.Consumer.CommitRecords(ctx, record); e != nil {
					logrus.Warn("Kafka offset commit delayed")
				}
			}
		}
	}
}

func reconcileLoop(ctx context.Context, deps *runtime.Dependencies) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	lastCleanup := time.Time{}
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		now := time.Now()
		deps.DB.Model(&model.Run{}).Where("state = ? AND lease_until < ?", "running", now).Updates(map[string]any{"state": "error", "error": "Runner失联，未自动重试", "finished_at": now})
		deps.DB.Model(&model.Run{}).Where("state = ? AND deadline < ?", "running", now).Updates(map[string]any{"state": "timed_out", "error": "执行超时", "finished_at": now})
		deps.DB.Model(&model.Run{}).Where("state IN ? AND queue_deadline < ?", []string{"created", "queued"}, now).Updates(map[string]any{"state": "timed_out", "error": "排队超时", "finished_at": now})
		if now.Sub(lastCleanup) > time.Hour {
			cutoff := now.Add(-time.Duration(deps.Config.RetentionDays) * 24 * time.Hour)
			e := deps.DB.Transaction(func(tx *gorm.DB) error {
				var expired []model.Run
				if e := tx.Unscoped().Where("finished_at < ? AND state NOT IN ?", cutoff, []string{"created", "queued", "running"}).Limit(1000).Find(&expired).Error; e != nil {
					return e
				}
				ids := []string{}
				for _, v := range expired {
					ids = append(ids, v.ID)
				}
				if len(ids) > 0 {
					if e := tx.Unscoped().Where("run_id IN ?", ids).Delete(&model.Outbox{}).Error; e != nil {
						return e
					}
					if e := tx.Unscoped().Where("id IN ?", ids).Delete(&model.Run{}).Error; e != nil {
						return e
					}
				}
				return tx.Where("created_at < ?", cutoff).Delete(&model.ConsumedEvent{}).Error
			})
			if e == nil {
				lastCleanup = now
			} else {
				logrus.Warn("history cleanup delayed")
			}
		}
	}
}
