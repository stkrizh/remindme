package channel

import "errors"

type OrderBy struct {
	v string
}

var (
	OrderByNotSet OrderBy = OrderBy{}
	OrderByIDAsc  OrderBy = OrderBy{v: "id_asc"}
	OrderByIDDesc OrderBy = OrderBy{v: "id_desc"}
)

var ErrParseOrderBy = errors.New("invalid order")

func ParseOrderBy(value string) (OrderBy, error) {
	switch value {
	case "id_asc":
		return OrderByIDAsc, nil
	case "id_desc":
		return OrderByIDDesc, nil
	default:
		return OrderByNotSet, ErrParseOrderBy
	}
}
