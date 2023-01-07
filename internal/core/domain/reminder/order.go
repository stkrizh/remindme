package reminder

import "errors"

type OrderBy struct {
	v string
}

var (
	OrderByNotSet OrderBy = OrderBy{}
	OrderByIDAsc  OrderBy = OrderBy{v: "id_asc"}
	OrderByIDDesc OrderBy = OrderBy{v: "id_desc"}
	OrderByAtAsc  OrderBy = OrderBy{v: "at_asc"}
	OrderByAtDesc OrderBy = OrderBy{v: "at_desc"}
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
