package api

type InventoryItem struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Quantity int    `json:"quantity"`
	Equipped bool   `json:"equipped"`
}

type AvatarInstanceRequest struct {
	Name           string            `json:"name"`
	AvatarType     string            `json:"avatarType"`
	Attributes     map[string]string `json:"attributes,omitempty"`
	Inventory      []InventoryItem   `json:"inventory,omitempty"`
	Achievements   []string          `json:"achievements,omitempty"`
	Customizations map[string]string `json:"customizations,omitempty"`
}

type AvatarInstanceResponse struct {
	ID             uint              `json:"id"`
	Name           string            `json:"name"`
	AvatarType     string            `json:"avatarType"`
	Game           string            `json:"game"`
	Attributes     map[string]string `json:"attributes,omitempty"`
	Inventory      []InventoryItem   `json:"inventory,omitempty"`
	Achievements   []string          `json:"achievements,omitempty"`
	Customizations map[string]string `json:"customizations,omitempty"`
}

type WorldResponse struct {
	Name        string `json:"name"`
	Game        string `json:"game"`
	Description string `json:"description,omitempty"`
}

type AreaResponse struct {
	Name           string            `json:"name"`
	Game           string            `json:"game"`
	World          string            `json:"world"`
	Description    string            `json:"description,omitempty"`
	ConnectedAreas []string          `json:"connectedAreas,omitempty"`
	Properties     map[string]string `json:"properties,omitempty"`
}

type ItemResponse struct {
	Name      string            `json:"name"`
	Category  string            `json:"category"`
	Rarity    string            `json:"rarity,omitempty"`
	Stackable bool              `json:"stackable"`
	MaxStack  int               `json:"maxStack,omitempty"`
	Duration  int               `json:"duration,omitempty"`
	Effects   map[string]string `json:"effects,omitempty"`
}

type GrantItemRequest struct {
	ItemName string `json:"itemName"`
	Quantity int    `json:"quantity"`
}

type EquipRequest struct {
	ItemName string `json:"itemName"`
}

type ActivatePowerupRequest struct {
	ItemName string `json:"itemName"`
}

type ActivePowerupResponse struct {
	ItemName    string `json:"itemName"`
	ActivatedAt int64  `json:"activatedAt"`
	ExpiresAt   int64  `json:"expiresAt"`
}

type CurrencyResponse struct {
	Name           string `json:"name"`
	Symbol         string `json:"symbol,omitempty"`
	Tradeable      bool   `json:"tradeable"`
	MaxBalance     int64  `json:"maxBalance,omitempty"`
	InitialBalance int64  `json:"initialBalance,omitempty"`
}

type WalletBalanceResponse struct {
	Currency string `json:"currency"`
	Symbol   string `json:"symbol,omitempty"`
	Balance  int64  `json:"balance"`
}

type CreditDebitRequest struct {
	Currency string `json:"currency"`
	Amount   int64  `json:"amount"`
}

type TransferRequest struct {
	Currency string `json:"currency"`
	Amount   int64  `json:"amount"`
	ToAvatar string `json:"toAvatar"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
