package response

import (
	"remindme/internal/core/domain/user"
)

type Limit struct {
	Value  int64 `json:"value"`
	Actual int64 `json:"actual"`
}

func (l *Limit) FromDomain(dl user.Limit) {
	l.Value = int64(dl.Value)
	l.Actual = int64(dl.Actual)
}
