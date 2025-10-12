package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/leegeev/KomaevBookingBot/internal/domain"
	"github.com/leegeev/KomaevBookingBot/pkg/config"
	"github.com/leegeev/KomaevBookingBot/pkg/logger"
	excelize "github.com/xuri/excelize/v2"
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

// ─────────────────────────────────────────────────────────────
//                 Основные методы Usecase
// ─────────────────────────────────────────────────────────────

func (s *LogService) GetUser(ctx context.Context, id int64) (domain.User, error) {
	s.logger.Info("Getting user", "userID", id)
	if id <= 0 {
		s.logger.Error("Invalid user ID", "userID", id)
		return domain.User{}, domain.ErrInvalidInputData
	}

	user, err := s.logRepo.GetUser(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get user", "err", err)
		return domain.User{}, err
	}

	s.logger.Info("User found", "userID", id, "fio", user.FIO)
	return user, nil
}

func (s *LogService) CreateUser(ctx context.Context, id int64, FIO string) error {
	s.logger.Info("Creating user", "userID", id, "fio", FIO)
	if id <= 0 || FIO == "" {
		s.logger.Error("Invalid input data", "userID", id, "fio", FIO)
		return domain.ErrInvalidInputData
	}

	if err := s.logRepo.CreateUser(ctx, id, FIO); err != nil {
		s.logger.Error("Failed to create user", "err", err)
		return err
	}

	s.logger.Info("User created or updated successfully", "userID", id)
	return nil
}

// Получить соглашения пользователя
func (s *LogService) GetSoglasheniyaByUserID(ctx context.Context, userID int64) ([]domain.Soglashenie, error) {
	s.logger.Info("Getting soglasheniya by userID", "userID", userID)
	list, err := s.logRepo.GetSoglasheniyaByUserID(ctx, domain.UserID(userID))
	if err != nil {
		s.logger.Error("Failed to get soglasheniya", "err", err)
		return nil, err
	}
	s.logger.Info("Found soglasheniya", "count", len(list))
	return list, nil
}

// Получить запросы пользователя
func (s *LogService) GetZaprosiByUserID(ctx context.Context, userID int64) ([]domain.Zapros, error) {
	s.logger.Info("Getting zaprosy by userID", "userID", userID)
	if userID <= 0 {
		return nil, domain.ErrInvalidInputData
	}

	list, err := s.logRepo.GetZaprosiByUserID(ctx, domain.UserID(userID))
	if err != nil {
		s.logger.Error("Failed to get zaprosy", "err", err)
		return nil, err
	}
	s.logger.Info("Found zaprosy", "count", len(list))
	return list, nil
}

// Получить одно соглашение по ID
func (s *LogService) GetSoglasheniyaById(ctx context.Context, id int64) (domain.Soglashenie, error) {
	s.logger.Info("Getting soglashenie by ID", "id", id)
	if id <= 0 {
		return domain.Soglashenie{}, domain.ErrInvalidInputData
	}

	record, err := s.logRepo.GetSoglashenieByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get soglashenie", "err", err)
		return domain.Soglashenie{}, err
	}
	return record, nil
}

// Получить один запрос по ID
func (s *LogService) GetZaprosById(ctx context.Context, id int64) (domain.Zapros, error) {
	s.logger.Info("Getting zapros by ID", "id", id)
	if id <= 0 {
		return domain.Zapros{}, domain.ErrInvalidInputData
	}

	record, err := s.logRepo.GetZaprosByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get zapros", "err", err)
		return domain.Zapros{}, err
	}
	return record, nil
}

// Создание записи (соглашения или запроса)
func (s *LogService) CreateLog(ctx context.Context, cmd CreateLogCmd) (int64, error) {
	s.logger.Info("Creating log entry", "user", cmd.UserName, "type", cmd.Type)

	switch cmd.Type {
	case "sogl":
		sogl := domain.Soglashenie{
			UserID:    cmd.UserID,
			UserName:  cmd.UserName,
			Date:      cmd.Date,
			Doveritel: cmd.Doveritel,
			Comment:   cmd.Comment,
			CreatedAt: time.Now(),
		}

		id, err := s.logRepo.CreateSoglashenie(ctx, sogl)
		if err != nil {
			s.logger.Error("Failed to create soglashenie", "err", err)
			return 0, err
		}
		s.logger.Info("Soglashenie created successfully", "id", id)
		return id, nil

	case "zapros":
		z := domain.Zapros{
			UserID:    cmd.UserID,
			UserName:  cmd.UserName,
			Date:      cmd.Date,
			Doveritel: cmd.Doveritel,
			Comment:   cmd.Comment,
			CreatedAt: time.Now(),
		}

		id, err := s.logRepo.CreateZapros(ctx, z)
		if err != nil {
			s.logger.Error("Failed to create zapros", "err", err)
			return 0, err
		}
		s.logger.Info("Zapros created successfully", "id", id)
		return id, nil

	default:
		s.logger.Warn("Unknown log type", "type", cmd.Type)
		return 0, domain.ErrInvalidInputData
	}
}

// ─────────────────────────────────────────────────────────────
//            Дополнительно (опциональные методы)
// ─────────────────────────────────────────────────────────────

// Пример Excel-отчёта (заглушка, для реализации позже)
// func (s *LogService) CreateExcelReport(ctx context.Context) (string, error) {
// 	s.logger.Info("Generating Excel report for logs")
// 	// TODO: реализовать выгрузку Excel отчёта из логов
// 	return "", errors.New("not implemented")
// }

type LogEntry struct {
	ID        int
	UserName  string
	Action    string
	CreatedAt time.Time
}

// CreateExcelReport формирует два Excel-файла (запросы и соглашения) за последний год.
func (s *LogService) CreateExcelReport(ctx context.Context) (string, string, error) {
	s.logger.Info("Generating Excel reports for last year")

	// 1️⃣ Берём период за последний год
	oneYearAgo := time.Now().AddDate(-1, 0, 0)

	// 2️⃣ Получаем данные из репозитория
	zaprosy, err := s.logRepo.GetZaprosiAfterDate(ctx, oneYearAgo)
	if err != nil {
		s.logger.Error("Failed to get zaprosy", "err", err)
		return "", "", err
	}

	soglasheniya, err := s.logRepo.GetSoglasheniyaAfterDate(ctx, oneYearAgo)
	if err != nil {
		s.logger.Error("Failed to get soglasheniya", "err", err)
		return "", "", err
	}

	// 3️⃣ Формируем отчёт для запросов
	zaprosFilePath, err := s.createZaprosExcel(zaprosy)
	if err != nil {
		s.logger.Error("Failed to create zapros Excel", "err", err)
		return "", "", err
	}

	// 4️⃣ Формируем отчёт для соглашений
	soglFilePath, err := s.createSoglasheniyaExcel(soglasheniya)
	if err != nil {
		s.logger.Error("Failed to create soglasheniya Excel", "err", err)
		return "", "", err
	}

	s.logger.Info("Excel reports generated successfully",
		"zaprosFile", zaprosFilePath,
		"soglFile", soglFilePath)

	return zaprosFilePath, soglFilePath, nil
}

func (s *LogService) createZaprosExcel(zaprosy []domain.Zapros) (string, error) {
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)

	headers := []string{"ID", "User ID", "User Name", "Date", "Doveritel", "Comment", "Created At"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	for i, z := range zaprosy {
		row := i + 2
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), z.ID)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), z.UserID)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), z.UserName)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), z.Date.Format("2006-01-02"))
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), z.Doveritel)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), z.Comment)
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), z.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	filePath := fmt.Sprintf("zaprosy_report_%d.xlsx", time.Now().Unix())
	if err := f.SaveAs(filePath); err != nil {
		return "", err
	}
	return filePath, nil
}

func (s *LogService) createSoglasheniyaExcel(soglasheniya []domain.Soglashenie) (string, error) {
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)

	headers := []string{"ID", "User ID", "User Name", "Date", "Doveritel", "Comment", "Created At"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	for i, sgl := range soglasheniya {
		row := i + 2
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), sgl.ID)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), sgl.UserID)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), sgl.UserName)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), sgl.Date.Format("2006-01-02"))
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), sgl.Doveritel)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), sgl.Comment)
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), sgl.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	filePath := fmt.Sprintf("soglasheniya_report_%d.xlsx", time.Now().Unix())
	if err := f.SaveAs(filePath); err != nil {
		return "", err
	}
	return filePath, nil
}
