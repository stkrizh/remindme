package reminder

import "errors"

type OrderBy (string)

const (
	OrderByNotSet OrderBy = ""
	OrderByIDAsc  OrderBy = "id_asc"
	OrderByIDDesc OrderBy = "id_desc"
	OrderByAtAsc  OrderBy = "at_asc"
	OrderByAtDesc OrderBy = "at_desc"
)

var ErrParseOrderBy = errors.New("invalid order")

func ParseOrderBy(value string) (OrderBy, error) {
	switch value {
	case "id_asc":
		return OrderByIDAsc, nil
	case "id_desc":
		return OrderByIDDesc, nil
	case "at_asc":
		return OrderByAtAsc, nil
	case "at_desc":
		return OrderByAtDesc, nil
	default:
		return OrderByNotSet, ErrParseOrderBy
	}
}
