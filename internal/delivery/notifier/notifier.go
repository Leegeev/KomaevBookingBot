package notifier

import (
	"context"
	"fmt"

	"github.com/leegeev/KomaevBookingBot/pkg/logger"
	"github.com/robfig/cron/v3"
)

/*

есть сущность, которая каждый день в 10.00 присылает сообщение в беседу
сохраняет айди этого сообщения.

внутри него у него есть handler

*/
// wake() - будет вызываться из Telegram pkg
// Внутри wake() будет запрашивать у Telegram расписание и редактировать ежедневное сообщение
// Это никак не связано с таймером, он вызывается по команде из Telegram

// Start будет запускать планировщик в отдельной горутине
// Которая будет отправлять сообщение с той же строкой, что и Wake
// и хранить айди этого сообщение в пямяти

type Notifier struct {
	log  logger.Logger
	cron *cron.Cron
}

func New(log logger.Logger) *Notifier {
	c := cron.New(cron.WithSeconds()) // с поддержкой секунд (по желанию)
	return &Notifier{
		log:  log,
		cron: c,
	}
}

// AddJob добавляет задачу по расписанию (cron-синтаксис, например "0 0 10 * * *" — каждый день в 10:00:00).
func (n *Notifier) AddJob(ctx context.Context, spec string, postSchedule func()) error {
	_, err := n.cron.AddFunc(spec, postSchedule)
	if err != nil {
		n.log.Error("failed to add cron job", "err", err)
		return fmt.Errorf("failed to add cron job: %w", err)
	}
	return nil
}

// Start запускает планировщик в отдельной горутине
func (n *Notifier) Start(ctx context.Context) {
	n.cron.Start()
	go func() {
		<-ctx.Done()
		n.log.Info("stopping notifier...")
		n.cron.Stop()
	}()
}
