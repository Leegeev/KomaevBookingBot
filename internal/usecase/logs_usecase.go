package usecase

import (
	"context"
	"time"

	"github.com/leegeev/KomaevBookingBot/internal/domain"
	"github.com/leegeev/KomaevBookingBot/pkg/config"
	"github.com/leegeev/KomaevBookingBot/pkg/logger"
)

type CreateLogCmd struct {
	UserID    domain.UserID
	UserName  string
	Type      string // "sogl" или "zapros"
	Date      time.Time
	Doveritel string
	Comment   string
}

type LogService struct {
	logRepo domain.LogRepository
	logger  logger.Logger
	cfg     config.Telegram
}

func NewLogService(logRepo domain.LogRepository, logger logger.Logger, cfg config.Telegram) *LogService {
	return &LogService{
		logRepo: logRepo,
		logger:  logger,
		cfg:     cfg,
	}
}

func (s *LogService) GetUser(ctx context.Context, id int64) error {
	// TODO
	return nil
}

func (s *LogService) CreateUser(ctx context.Context, id int64, FIO string) error {
	// TODO:
	return nil
}

func (s *LogService) GetSoglasheniyaByUserID(ctx context.Context, userID int64) (domain.Soglashenie, error) {
	// TODO:
	return domain.Soglashenie{}, nil
}

func (s *LogService) GetZaprosiByUserId(ctx context.Context, userID int64) (domain.Zapros, error) {
	// TODO:
	return domain.Zapros{}, nil
}

func (s *LogService) GetSoglasheniyaById(ctx context.Context, id int64) (domain.Soglashenie, error) {
	// TODO:
	return domain.Soglashenie{}, nil
}

func (s *LogService) GetZaprosById(ctx context.Context, id int64) (domain.Zapros, error) {
	// TODO:
	return domain.Zapros{}, nil
}

func (s *LogService) CreateLog(ctx context.Context, l CreateLogCmd) (int64, error) { // по type определить какую запись создать {
	// TODO:
	return 0, nil
}

func (s *LogService) CreateExcelReport(ctx context.Context) (string, error) {
	// TODO:
	return "", nil
}

/*
type LogEntry struct {
	ID        int
	UserName  string
	Action    string
	CreatedAt time.Time
}

func CreateExcelReport(logs []LogEntry) (string, error) {
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)

	// Заголовки
	headers := []string{"ID", "User", "Action", "Created At"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	// Данные
	for i, log := range logs {
		f.SetCellValue(sheet, fmt.Sprintf("A%d", i+2), log.ID)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", i+2), log.UserName)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", i+2), log.Action)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", i+2), log.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	// Сохраняем файл
	filePath := fmt.Sprintf("report_%d.xlsx", time.Now().Unix())
	if err := f.SaveAs(filePath); err != nil {
		return "", err
	}
	return filePath, nil
}

func (r *Repository) GetLogs(ctx context.Context) ([]LogEntry, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, username, action, created_at
		FROM logs
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []LogEntry
	for rows.Next() {
		var l LogEntry
		if err := rows.Scan(&l.ID, &l.UserName, &l.Action, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}

	return logs, nil
}
*/
