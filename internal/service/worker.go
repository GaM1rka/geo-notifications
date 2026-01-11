package service

import (
	"bytes"
	"context"
	"encoding/json"
	"geo-notifications/internal/model"
	"geo-notifications/internal/repository"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type WebhookWorker struct {
	storage    *repository.Storage
	logger     *logrus.Logger
	webhookURL string
}

func NewWebhookWorker(storage *repository.Storage, logger *logrus.Logger, webhookURL string) *WebhookWorker {
	return &WebhookWorker{
		storage:    storage,
		logger:     logger,
		webhookURL: webhookURL,
	}
}

func (w *WebhookWorker) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			res, err := w.storage.BLPopWebhookTask(ctx, 5*time.Second, w.webhookURL)
			if err != nil {
				w.logger.WithError(err).Error("BLPop error")
				continue
			}

			var task model.WebhookPayload
			if err := json.Unmarshal([]byte(res), &task); err != nil {
				w.logger.WithError(err).Error("unmarshal webhook task error")
				continue
			}

			body, err := json.Marshal(task)
			if err != nil {
				w.logger.WithError(err).Error("marshal webhook task for http request")
				continue
			}
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.webhookURL, bytes.NewReader(body))
			if err != nil {
				w.logger.WithError(err).Error("invalid request to webhookURL")
				continue
			}
			req.Header.Set("Content-Type", "application/json")
			client := http.Client{}
			_, err = client.Do(req)
			if err != nil {
				w.logger.WithError(err).Error("problem while sending reqeust to webhookURL")
				continue
			}
		}
	}
}
