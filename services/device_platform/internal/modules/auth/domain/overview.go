package domain

// ParentOverview is the guardian-facing home dashboard.
type ParentOverview struct {
	TodayConversationCount int64   `json:"today_conversation_count"`
	TodaySpentUSD          float64 `json:"today_spent_usd"`
	RemainingBalanceUSD    float64 `json:"remaining_balance_usd"`
	BalanceUSD             float64 `json:"balance_usd"`
	DeviceCount            int64   `json:"device_count"`
	OnlineDeviceCount      int64   `json:"online_device_count"`
}
