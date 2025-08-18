package domain

import "time"

func (r Room) Valid() bool { return r.ID != 0 && r.Name != "" }

func NewBooking(roomID RoomID, createdBy UserID, tr TimeRange, note string) (Booking, error) {
	if roomID == 0 || createdBy == 0 {
		return Booking{}, ErrInvalidInputData
	}
	return Booking{
		RoomID:    roomID,
		CreatedBy: createdBy,
		Range:     tr,
		Note:      note,
	}, nil
}

// В домене ВСЕ времена — в UTC. Конвертация в локальную TZ — на краях (UI/infra).
func MustUTC(t time.Time) time.Time {
	if t.Location() != time.UTC {
		return t.UTC()
	}
	return t
}

func NewTimeRange(startUTC, endUTC time.Time) (TimeRange, error) {
	s := MustUTC(startUTC)
	e := MustUTC(endUTC)
	if !e.After(s) || e.Equal(s) || s.IsZero() || e.IsZero() {
		return TimeRange{}, ErrInvalidTimeRange
	}
	return TimeRange{Start: s, End: e}, nil
}

func (tr TimeRange) Duration() time.Duration { return tr.End.Sub(tr.Start) }

// Пересечение по полуинтервалам [a, b) && [c, d)
func (tr TimeRange) Overlaps(other TimeRange) bool {
	return tr.Start.Before(other.End) && other.Start.Before(tr.End)
}

// Попадание момента в [Start, End).
func (tr TimeRange) Contains(t time.Time) bool {
	u := MustUTC(t)
	return (u.Equal(tr.Start) || u.After(tr.Start)) && u.Before(tr.End)
}

func (tr TimeRange) IsZero() bool { return tr.Start.IsZero() || tr.End.IsZero() }
