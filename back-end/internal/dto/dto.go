package dto

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type PreferencePriority struct {
	BestPrice       int `json:"best_price"`
	ClosestLocation int `json:"closest_location"`
}

type PreferenceSustainability struct {
	LocalBusiness int `json:"local_business"`
	Recyclable    int `json:"recycleable"`
	Organic       int `json:"organic"`
}

type ShoppingItem struct {
	ItemName     string  `json:"item_name"`
	ItemQuantity float64 `json:"item_quantity"`
}

type SubmitPreferenceRequest struct {
	UserLocation             Location                 `json:"user_location"`
	MaximumDistanceMiles     float64                  `json:"maximum_distance_from_user_location_miles"`
	PreferencePriority       PreferencePriority       `json:"preference_priority"`
	PreferenceSustainability PreferenceSustainability `json:"preference_sustainability"`
	BudgetGBP                float64                  `json:"budget_gbp"`
	ShoppingItemList         []ShoppingItem           `json:"shopping_item_list"`
}

type SubmitPreferenceResponse struct {
	SessionID uint64 `json:"session_id"`
}

type OptionLocation struct {
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Option struct {
	OptionID        int            `json:"option_id"`
	OptionName      string         `json:"option_name"`
	PriceGBP        float64        `json:"price_gbp"`
	Location        OptionLocation `json:"location"`
	DistanceMiles   float64        `json:"distance_miles"`
	IsRecyclable    bool           `json:"is_recycleable"`
	IsLocalBusiness bool           `json:"is_local_businesss"` // spelling kept to match the API spec
	IsOrganic       bool           `json:"is_organic"`
}

type ItemOptionResult struct {
	ResultID     int      `json:"result_id"`
	ItemName     string   `json:"item_name"`
	ItemQuantity float64  `json:"item_quantity"`
	Options      []Option `json:"options"`
}

type OptionResultResponse struct {
	OptionResult []ItemOptionResult `json:"option_result"`
}

type SubmittedOption struct {
	ResultID         int `json:"result_id"`
	SelectedOptionID int `json:"selected_option_id"`
}

type SubmitOptionRequest struct {
	SessionID       uint64            `json:"session_id"`
	SubmittedOption []SubmittedOption `json:"submitted_option"`
}

type SubmitOptionResponse struct {
	IsSuccess bool   `json:"is_success"`
	Error     string `json:"error,omitempty"`
}

type SummaryItem struct {
	ResultID       int     `json:"result_id"`
	ItemName       string  `json:"item_name"`
	ItemQuantity   float64 `json:"item_quantity"`
	SelectedOption *Option `json:"selected_option"`
}

type SummaryResponse struct {
	Summary       []SummaryItem `json:"summary"`
	TotalPriceGBP float64       `json:"total_price_gbp"`
	BudgetGBP     float64       `json:"budget_gbp"`
	WithinBudget  bool          `json:"within_budget"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
