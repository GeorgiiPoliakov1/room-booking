package request

type SlotListRequest struct {
	Date string `query:"date" form:"date" validate:"required,date_iso"`
}
