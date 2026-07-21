package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"gorm.io/gorm"
	"systemcraftsman.com/kubegame/internal/persistence"
)

type Handler struct {
	getDB func(game, namespace string) (*gorm.DB, error)
}

func NewHandler(getDB func(game, namespace string) (*gorm.DB, error)) *Handler {
	return &Handler{getDB: getDB}
}

func (h *Handler) CreateAvatarInstance(w http.ResponseWriter, r *http.Request) {
	game := r.PathValue("game")
	namespace := r.PathValue("namespace")

	if game == "" {
		writeError(w, http.StatusBadRequest, "missing game parameter")
		return
	}

	var req AvatarInstanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" || req.AvatarType == "" {
		writeError(w, http.StatusBadRequest, "name and avatarType are required")
		return
	}

	db, err := h.getDB(game, namespace)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to connect to game database: %v", err))
		return
	}

	if err := validateAgainstBlueprint(db, req.AvatarType, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	instance := &persistence.AvatarInstance{
		Name:       req.Name,
		AvatarName: req.AvatarType,
		Game:       game,
	}
	if result := db.Create(instance); result.Error != nil {
		writeError(w, http.StatusConflict, fmt.Sprintf("failed to create avatar instance: %v", result.Error))
		return
	}

	for name, value := range req.Attributes {
		attr := &persistence.AvatarInstanceAttribute{
			AvatarInstanceID: instance.ID,
			Name:             name,
			Value:            value,
		}
		if result := db.Create(attr); result.Error != nil {
			writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to create attribute: %v", result.Error))
			return
		}
	}

	for _, item := range req.Inventory {
		inv := &persistence.AvatarInstanceInventoryItem{
			AvatarInstanceID: instance.ID,
			Name:             item.Name,
			Type:             item.Type,
			Quantity:         item.Quantity,
		}
		if result := db.Create(inv); result.Error != nil {
			writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to create inventory item: %v", result.Error))
			return
		}
	}

	for _, ach := range req.Achievements {
		achievement := &persistence.AvatarInstanceAchievement{
			AvatarInstanceID: instance.ID,
			Name:             ach,
		}
		if result := db.Create(achievement); result.Error != nil {
			writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to create achievement: %v", result.Error))
			return
		}
	}

	for name, value := range req.Customizations {
		cust := &persistence.AvatarInstanceCustomization{
			AvatarInstanceID: instance.ID,
			Name:             name,
			Value:            value,
		}
		if result := db.Create(cust); result.Error != nil {
			writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to create customization: %v", result.Error))
			return
		}
	}

	resp := buildInstanceResponse(db, instance)
	writeJSON(w, http.StatusCreated, resp)
}

func (h *Handler) GetAvatarInstance(w http.ResponseWriter, r *http.Request) {
	game := r.PathValue("game")
	name := r.PathValue("name")
	namespace := r.PathValue("namespace")

	if game == "" || name == "" {
		writeError(w, http.StatusBadRequest, "missing game or instance name parameter")
		return
	}

	db, err := h.getDB(game, namespace)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to connect to game database: %v", err))
		return
	}

	var instance persistence.AvatarInstance
	if result := db.Where("name = ? AND game = ?", name, game).First(&instance); result.Error != nil {
		writeError(w, http.StatusNotFound, fmt.Sprintf("avatar instance %q not found", name))
		return
	}

	resp := buildInstanceResponse(db, &instance)
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) ListAvatarInstances(w http.ResponseWriter, r *http.Request) {
	game := r.PathValue("game")
	namespace := r.PathValue("namespace")

	if game == "" {
		writeError(w, http.StatusBadRequest, "missing game parameter")
		return
	}

	db, err := h.getDB(game, namespace)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to connect to game database: %v", err))
		return
	}

	var instances []persistence.AvatarInstance
	db.Where("game = ?", game).Find(&instances)

	var responses []AvatarInstanceResponse
	for i := range instances {
		responses = append(responses, *buildInstanceResponse(db, &instances[i]))
	}

	writeJSON(w, http.StatusOK, responses)
}

func (h *Handler) DeleteAvatarInstance(w http.ResponseWriter, r *http.Request) {
	game := r.PathValue("game")
	name := r.PathValue("name")
	namespace := r.PathValue("namespace")

	if game == "" || name == "" {
		writeError(w, http.StatusBadRequest, "missing game or instance name parameter")
		return
	}

	db, err := h.getDB(game, namespace)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to connect to game database: %v", err))
		return
	}

	var instance persistence.AvatarInstance
	if result := db.Where("name = ? AND game = ?", name, game).First(&instance); result.Error != nil {
		writeError(w, http.StatusNotFound, fmt.Sprintf("avatar instance %q not found", name))
		return
	}

	db.Where("avatar_instance_id = ?", instance.ID).Delete(&persistence.AvatarInstanceCustomization{})
	db.Where("avatar_instance_id = ?", instance.ID).Delete(&persistence.AvatarInstanceAttribute{})
	db.Where("avatar_instance_id = ?", instance.ID).Delete(&persistence.AvatarInstanceInventoryItem{})
	db.Where("avatar_instance_id = ?", instance.ID).Delete(&persistence.AvatarInstanceAchievement{})
	db.Delete(&instance)

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListWorlds(w http.ResponseWriter, r *http.Request) {
	game := r.PathValue("game")
	namespace := r.PathValue("namespace")

	if game == "" {
		writeError(w, http.StatusBadRequest, "missing game parameter")
		return
	}

	db, err := h.getDB(game, namespace)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to connect to game database: %v", err))
		return
	}

	var worlds []persistence.World
	db.Where("game = ?", game).Find(&worlds)

	var responses []WorldResponse
	for _, w := range worlds {
		responses = append(responses, WorldResponse{
			Name:        w.Name,
			Game:        w.Game,
			Description: w.Description,
		})
	}

	writeJSON(w, http.StatusOK, responses)
}

func (h *Handler) GetWorld(w http.ResponseWriter, r *http.Request) {
	game := r.PathValue("game")
	name := r.PathValue("name")
	namespace := r.PathValue("namespace")

	if game == "" || name == "" {
		writeError(w, http.StatusBadRequest, "missing game or world name parameter")
		return
	}

	db, err := h.getDB(game, namespace)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to connect to game database: %v", err))
		return
	}

	var world persistence.World
	if result := db.Where("name = ? AND game = ?", name, game).First(&world); result.Error != nil {
		writeError(w, http.StatusNotFound, fmt.Sprintf("world %q not found", name))
		return
	}

	writeJSON(w, http.StatusOK, WorldResponse{
		Name:        world.Name,
		Game:        world.Game,
		Description: world.Description,
	})
}

func (h *Handler) ListAreas(w http.ResponseWriter, r *http.Request) {
	game := r.PathValue("game")
	world := r.PathValue("world")
	namespace := r.PathValue("namespace")

	if game == "" || world == "" {
		writeError(w, http.StatusBadRequest, "missing game or world parameter")
		return
	}

	db, err := h.getDB(game, namespace)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to connect to game database: %v", err))
		return
	}

	var areas []persistence.Area
	db.Where("game = ? AND world = ?", game, world).Find(&areas)

	var responses []AreaResponse
	for i := range areas {
		responses = append(responses, *buildAreaResponse(db, &areas[i]))
	}

	writeJSON(w, http.StatusOK, responses)
}

func (h *Handler) GetArea(w http.ResponseWriter, r *http.Request) {
	game := r.PathValue("game")
	name := r.PathValue("name")
	namespace := r.PathValue("namespace")

	if game == "" || name == "" {
		writeError(w, http.StatusBadRequest, "missing game or area name parameter")
		return
	}

	db, err := h.getDB(game, namespace)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to connect to game database: %v", err))
		return
	}

	var area persistence.Area
	if result := db.Where("name = ? AND game = ?", name, game).First(&area); result.Error != nil {
		writeError(w, http.StatusNotFound, fmt.Sprintf("area %q not found", name))
		return
	}

	resp := buildAreaResponse(db, &area)
	writeJSON(w, http.StatusOK, resp)
}

func buildAreaResponse(db *gorm.DB, area *persistence.Area) *AreaResponse {
	resp := &AreaResponse{
		Name:        area.Name,
		Game:        area.Game,
		World:       area.World,
		Description: area.Description,
	}

	var connections []persistence.AreaConnection
	db.Where("area_name = ?", area.Name).Find(&connections)
	for _, c := range connections {
		resp.ConnectedAreas = append(resp.ConnectedAreas, c.ConnectsTo)
	}

	var props []persistence.AreaPropertyRecord
	db.Where("area_name = ?", area.Name).Find(&props)
	resp.Properties = make(map[string]string)
	for _, p := range props {
		resp.Properties[p.Name] = p.Value
	}

	return resp
}

func buildInstanceResponse(db *gorm.DB, instance *persistence.AvatarInstance) *AvatarInstanceResponse {
	resp := &AvatarInstanceResponse{
		ID:         instance.ID,
		Name:       instance.Name,
		AvatarType: instance.AvatarName,
		Game:       instance.Game,
		Attributes: make(map[string]string),
	}

	var attrs []persistence.AvatarInstanceAttribute
	db.Where("avatar_instance_id = ?", instance.ID).Find(&attrs)
	for _, a := range attrs {
		resp.Attributes[a.Name] = a.Value
	}

	var items []persistence.AvatarInstanceInventoryItem
	db.Where("avatar_instance_id = ?", instance.ID).Find(&items)
	for _, item := range items {
		resp.Inventory = append(resp.Inventory, InventoryItem{
			Name:     item.Name,
			Type:     item.Type,
			Quantity: item.Quantity,
			Equipped: item.Equipped,
		})
	}

	var achievements []persistence.AvatarInstanceAchievement
	db.Where("avatar_instance_id = ?", instance.ID).Find(&achievements)
	for _, a := range achievements {
		resp.Achievements = append(resp.Achievements, a.Name)
	}

	var customizations []persistence.AvatarInstanceCustomization
	db.Where("avatar_instance_id = ?", instance.ID).Find(&customizations)
	if len(customizations) > 0 {
		resp.Customizations = make(map[string]string)
		for _, c := range customizations {
			resp.Customizations[c.Name] = c.Value
		}
	}

	return resp
}

func (h *Handler) ListItems(w http.ResponseWriter, r *http.Request) {
	game := r.PathValue("game")
	namespace := r.PathValue("namespace")

	if game == "" {
		writeError(w, http.StatusBadRequest, "missing game parameter")
		return
	}

	db, err := h.getDB(game, namespace)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to connect to game database: %v", err))
		return
	}

	var items []persistence.ItemDefinition
	db.Where("game = ?", game).Find(&items)

	var responses []ItemResponse
	for i := range items {
		responses = append(responses, *buildItemResponse(db, &items[i]))
	}

	writeJSON(w, http.StatusOK, responses)
}

func (h *Handler) GetItem(w http.ResponseWriter, r *http.Request) {
	game := r.PathValue("game")
	name := r.PathValue("name")
	namespace := r.PathValue("namespace")

	if game == "" || name == "" {
		writeError(w, http.StatusBadRequest, "missing game or item name parameter")
		return
	}

	db, err := h.getDB(game, namespace)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to connect to game database: %v", err))
		return
	}

	var item persistence.ItemDefinition
	if result := db.Where("name = ? AND game = ?", name, game).First(&item); result.Error != nil {
		writeError(w, http.StatusNotFound, fmt.Sprintf("item %q not found", name))
		return
	}

	writeJSON(w, http.StatusOK, buildItemResponse(db, &item))
}

func (h *Handler) GrantItem(w http.ResponseWriter, r *http.Request) {
	game := r.PathValue("game")
	avatarName := r.PathValue("name")
	namespace := r.PathValue("namespace")

	if game == "" || avatarName == "" {
		writeError(w, http.StatusBadRequest, "missing game or avatar name parameter")
		return
	}

	var req GrantItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.ItemName == "" || req.Quantity <= 0 {
		writeError(w, http.StatusBadRequest, "itemName and positive quantity are required")
		return
	}

	db, err := h.getDB(game, namespace)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to connect to game database: %v", err))
		return
	}

	var instance persistence.AvatarInstance
	if result := db.Where("name = ? AND game = ?", avatarName, game).First(&instance); result.Error != nil {
		writeError(w, http.StatusNotFound, fmt.Sprintf("avatar instance %q not found", avatarName))
		return
	}

	var itemDef persistence.ItemDefinition
	if result := db.Where("name = ? AND game = ?", req.ItemName, game).First(&itemDef); result.Error != nil {
		writeError(w, http.StatusNotFound, fmt.Sprintf("item %q not found in catalog", req.ItemName))
		return
	}

	if itemDef.Stackable {
		var existing persistence.AvatarInstanceInventoryItem
		result := db.Where("avatar_instance_id = ? AND name = ?", instance.ID, req.ItemName).First(&existing)
		if result.Error == nil {
			newQty := existing.Quantity + req.Quantity
			if itemDef.MaxStack > 0 && newQty > itemDef.MaxStack {
				newQty = itemDef.MaxStack
			}
			db.Model(&existing).Update("quantity", newQty)
			resp := buildInstanceResponse(db, &instance)
			writeJSON(w, http.StatusOK, resp)
			return
		}
	}

	quantity := req.Quantity
	if itemDef.Stackable && itemDef.MaxStack > 0 && quantity > itemDef.MaxStack {
		quantity = itemDef.MaxStack
	}

	inv := &persistence.AvatarInstanceInventoryItem{
		AvatarInstanceID: instance.ID,
		Name:             req.ItemName,
		Type:             itemDef.Category,
		Quantity:         quantity,
	}
	if result := db.Create(inv); result.Error != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to grant item: %v", result.Error))
		return
	}

	resp := buildInstanceResponse(db, &instance)
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) EquipItem(w http.ResponseWriter, r *http.Request) {
	game := r.PathValue("game")
	avatarName := r.PathValue("name")
	namespace := r.PathValue("namespace")

	if game == "" || avatarName == "" {
		writeError(w, http.StatusBadRequest, "missing game or avatar name parameter")
		return
	}

	var req EquipRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.ItemName == "" {
		writeError(w, http.StatusBadRequest, "itemName is required")
		return
	}

	db, err := h.getDB(game, namespace)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to connect to game database: %v", err))
		return
	}

	var instance persistence.AvatarInstance
	if result := db.Where("name = ? AND game = ?", avatarName, game).First(&instance); result.Error != nil {
		writeError(w, http.StatusNotFound, fmt.Sprintf("avatar instance %q not found", avatarName))
		return
	}

	var invItem persistence.AvatarInstanceInventoryItem
	if result := db.Where("avatar_instance_id = ? AND name = ?", instance.ID, req.ItemName).First(&invItem); result.Error != nil {
		writeError(w, http.StatusNotFound, fmt.Sprintf("item %q not in inventory", req.ItemName))
		return
	}

	var itemDef persistence.ItemDefinition
	if result := db.Where("name = ? AND game = ?", req.ItemName, game).First(&itemDef); result.Error != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("item %q not found in catalog", req.ItemName))
		return
	}

	if itemDef.Category != "Equipment" {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("item %q is not equipment (category: %s)", req.ItemName, itemDef.Category))
		return
	}

	db.Model(&invItem).Update("equipped", true)

	resp := buildInstanceResponse(db, &instance)
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) UnequipItem(w http.ResponseWriter, r *http.Request) {
	game := r.PathValue("game")
	avatarName := r.PathValue("name")
	namespace := r.PathValue("namespace")

	if game == "" || avatarName == "" {
		writeError(w, http.StatusBadRequest, "missing game or avatar name parameter")
		return
	}

	var req EquipRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.ItemName == "" {
		writeError(w, http.StatusBadRequest, "itemName is required")
		return
	}

	db, err := h.getDB(game, namespace)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to connect to game database: %v", err))
		return
	}

	var instance persistence.AvatarInstance
	if result := db.Where("name = ? AND game = ?", avatarName, game).First(&instance); result.Error != nil {
		writeError(w, http.StatusNotFound, fmt.Sprintf("avatar instance %q not found", avatarName))
		return
	}

	var invItem persistence.AvatarInstanceInventoryItem
	if result := db.Where("avatar_instance_id = ? AND name = ?", instance.ID, req.ItemName).First(&invItem); result.Error != nil {
		writeError(w, http.StatusNotFound, fmt.Sprintf("item %q not in inventory", req.ItemName))
		return
	}

	db.Model(&invItem).Update("equipped", false)

	resp := buildInstanceResponse(db, &instance)
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) ActivatePowerup(w http.ResponseWriter, r *http.Request) {
	game := r.PathValue("game")
	avatarName := r.PathValue("name")
	namespace := r.PathValue("namespace")

	if game == "" || avatarName == "" {
		writeError(w, http.StatusBadRequest, "missing game or avatar name parameter")
		return
	}

	var req ActivatePowerupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.ItemName == "" {
		writeError(w, http.StatusBadRequest, "itemName is required")
		return
	}

	db, err := h.getDB(game, namespace)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to connect to game database: %v", err))
		return
	}

	var instance persistence.AvatarInstance
	if result := db.Where("name = ? AND game = ?", avatarName, game).First(&instance); result.Error != nil {
		writeError(w, http.StatusNotFound, fmt.Sprintf("avatar instance %q not found", avatarName))
		return
	}

	var invItem persistence.AvatarInstanceInventoryItem
	if result := db.Where("avatar_instance_id = ? AND name = ?", instance.ID, req.ItemName).First(&invItem); result.Error != nil {
		writeError(w, http.StatusNotFound, fmt.Sprintf("item %q not in inventory", req.ItemName))
		return
	}

	var itemDef persistence.ItemDefinition
	if result := db.Where("name = ? AND game = ?", req.ItemName, game).First(&itemDef); result.Error != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("item %q not found in catalog", req.ItemName))
		return
	}

	if itemDef.Category != "Powerup" {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("item %q is not a powerup (category: %s)", req.ItemName, itemDef.Category))
		return
	}

	if itemDef.Duration <= 0 {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("powerup %q has no duration defined", req.ItemName))
		return
	}

	now := time.Now().Unix()
	powerup := &persistence.ActivePowerup{
		AvatarInstanceID: instance.ID,
		ItemName:         req.ItemName,
		ActivatedAt:      now,
		ExpiresAt:        now + int64(itemDef.Duration),
	}
	if result := db.Create(powerup); result.Error != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to activate powerup: %v", result.Error))
		return
	}

	if invItem.Quantity <= 1 {
		db.Delete(&invItem)
	} else {
		db.Model(&invItem).Update("quantity", invItem.Quantity-1)
	}

	writeJSON(w, http.StatusOK, ActivePowerupResponse{
		ItemName:    powerup.ItemName,
		ActivatedAt: powerup.ActivatedAt,
		ExpiresAt:   powerup.ExpiresAt,
	})
}

func (h *Handler) ListActivePowerups(w http.ResponseWriter, r *http.Request) {
	game := r.PathValue("game")
	avatarName := r.PathValue("name")
	namespace := r.PathValue("namespace")

	if game == "" || avatarName == "" {
		writeError(w, http.StatusBadRequest, "missing game or avatar name parameter")
		return
	}

	db, err := h.getDB(game, namespace)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to connect to game database: %v", err))
		return
	}

	var instance persistence.AvatarInstance
	if result := db.Where("name = ? AND game = ?", avatarName, game).First(&instance); result.Error != nil {
		writeError(w, http.StatusNotFound, fmt.Sprintf("avatar instance %q not found", avatarName))
		return
	}

	now := time.Now().Unix()
	var powerups []persistence.ActivePowerup
	db.Where("avatar_instance_id = ? AND expires_at > ?", instance.ID, now).Find(&powerups)

	var responses []ActivePowerupResponse
	for _, p := range powerups {
		responses = append(responses, ActivePowerupResponse{
			ItemName:    p.ItemName,
			ActivatedAt: p.ActivatedAt,
			ExpiresAt:   p.ExpiresAt,
		})
	}

	writeJSON(w, http.StatusOK, responses)
}

func buildItemResponse(db *gorm.DB, item *persistence.ItemDefinition) *ItemResponse {
	resp := &ItemResponse{
		Name:      item.Name,
		Category:  item.Category,
		Rarity:    item.Rarity,
		Stackable: item.Stackable,
		MaxStack:  item.MaxStack,
		Duration:  item.Duration,
	}

	var effects []persistence.ItemEffectRecord
	db.Where("item_name = ? AND game = ?", item.Name, item.Game).Find(&effects)
	if len(effects) > 0 {
		resp.Effects = make(map[string]string)
		for _, e := range effects {
			resp.Effects[e.Attribute] = e.Modifier
		}
	}

	return resp
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{Error: message})
}
