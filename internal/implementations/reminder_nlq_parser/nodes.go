package remindernlqparser

type nodeVisitor interface {
	visitAt(at at) error
	visitEvery(ever every) error
	visitOn(on on) error
	visitIn(in in) error
}

type node interface {
	accept(v nodeVisitor) error
}

type at struct {
	hour   uint
	minute uint
}

func (at at) accept(v nodeVisitor) error {
	return v.visitAt(at)
}

type period string

var (
	invalid period = period("")
	second  period = period("s")
	minute  period = period("m")
	hour    period = period("h")
	day     period = period("d")
	week    period = period("w")
	month   period = period("mo")
)

type onDay string

var (
	today     onDay = onDay("today")
	tomorrow  onDay = onDay("tomorrow")
	sunday    onDay = onDay("sunday")
	monday    onDay = onDay("monday")
	tuesday   onDay = onDay("tuesday")
	wednesday onDay = onDay("wednesday")
	thursday  onDay = onDay("thursday")
	friday    onDay = onDay("friday")
	saturday  onDay = onDay("saturday")
)

type every struct {
	p  period
	n  uint
	at *at
}

func (every every) accept(v nodeVisitor) error {
	return v.visitEvery(every)
}

type on struct {
	day onDay
	at  *at
}

func (on on) accept(v nodeVisitor) error {
	return v.visitOn(on)
}

type in struct {
	p period
	n uint
}

func (in in) accept(v nodeVisitor) error {
	return v.visitIn(in)
}
