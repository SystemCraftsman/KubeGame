package api

type InventoryItem struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Quantity int    `json:"quantity"`
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

type ErrorResponse struct {
	Error string `json:"error"`
}
