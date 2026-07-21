package api

type InventoryItem struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Quantity int    `json:"quantity"`
}

type AvatarInstanceRequest struct {
	Name         string            `json:"name"`
	AvatarType   string            `json:"avatarType"`
	Attributes   map[string]string `json:"attributes,omitempty"`
	Inventory    []InventoryItem   `json:"inventory,omitempty"`
	Achievements []string          `json:"achievements,omitempty"`
}

type AvatarInstanceResponse struct {
	ID           uint              `json:"id"`
	Name         string            `json:"name"`
	AvatarType   string            `json:"avatarType"`
	Game         string            `json:"game"`
	Attributes   map[string]string `json:"attributes,omitempty"`
	Inventory    []InventoryItem   `json:"inventory,omitempty"`
	Achievements []string          `json:"achievements,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
