package auth

type Action int

const (
	ActionCreate Action = iota
	ActionRead
	ActionUpdate
	ActionDelete
	ActionReadMine // read with ?mine query, usually filter by user_id field
)

var actionToStr = map[Action]string{
	ActionCreate:   "create",
	ActionRead:     "read",
	ActionUpdate:   "update",
	ActionDelete:   "delete",
	ActionReadMine: "read_mine",
}

func (a Action) String() string {
	return actionToStr[a]
}
