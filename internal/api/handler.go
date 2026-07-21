package api

import (
	"encoding/json"
	"fmt"
	"net/http"

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
		})
	}

	var achievements []persistence.AvatarInstanceAchievement
	db.Where("avatar_instance_id = ?", instance.ID).Find(&achievements)
	for _, a := range achievements {
		resp.Achievements = append(resp.Achievements, a.Name)
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
