package api

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/gorm"

	"systemcraftsman.com/kubegame/internal/persistence"
)

var testDB *gorm.DB

func TestMain(m *testing.M) {
	if flag.Lookup("test.short") != nil && flag.Lookup("test.short").Value.String() == "true" {
		fmt.Println("skipping integration tests in short mode")
		os.Exit(0)
	}

	ctx := context.Background()

	container, err := tcpostgres.Run(ctx,
		"postgres:13",
		tcpostgres.WithDatabase("postgres"),
		tcpostgres.WithUsername("testuser"),
		tcpostgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start postgres container: %v\n", err)
		os.Exit(1)
	}
	defer container.Terminate(ctx)

	mappedPort, err := container.MappedPort(ctx, "5432")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get mapped port: %v\n", err)
		os.Exit(1)
	}

	os.Setenv("DATABASE_TYPE", "postgres")
	os.Setenv("DATABASE_NAME", "postgres")
	os.Setenv("DATABASE_PORT_DEVELOPMENT", mappedPort.Port())
	os.Setenv("APP_ENV", "development")

	testDB, err = persistence.CreateDatabaseConnection("localhost", "testuser", "testpass")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create DB connection: %v\n", err)
		os.Exit(1)
	}

	if err := persistence.RunMigrations(testDB, persistence.AllModels()...); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to run migrations: %v\n", err)
		os.Exit(1)
	}

	seedBlueprint()

	os.Exit(m.Run())
}

func seedBlueprint() {
	testDB.Create(&persistence.AvatarType{
		Name: "test-avatar", Game: "test-game", Type: "Adventurer", Description: "Test",
	})
	testDB.Create(&persistence.AttributeType{AvatarName: "test-avatar", Name: "strength", ValueType: "int"})
	testDB.Create(&persistence.AttributeType{AvatarName: "test-avatar", Name: "intelligence", ValueType: "int"})
	testDB.Create(&persistence.InventoryType{AvatarName: "test-avatar", Name: "Weapon", Category: "Equipment"})
	testDB.Create(&persistence.AchievementType{AvatarName: "test-avatar", Name: "First Blood", Description: "Win first battle"})
	testDB.Create(&persistence.CustomizationTypeRecord{AvatarName: "test-avatar", Name: "Race"})
	testDB.Create(&persistence.CustomizationTypeRecord{AvatarName: "test-avatar", Name: "Class"})
	testDB.Create(&persistence.CustomizationOption{AvatarName: "test-avatar", CustomizationName: "Race", Value: "Human"})
	testDB.Create(&persistence.CustomizationOption{AvatarName: "test-avatar", CustomizationName: "Race", Value: "Elf"})
	testDB.Create(&persistence.CustomizationOption{AvatarName: "test-avatar", CustomizationName: "Class", Value: "Warrior"})
	testDB.Create(&persistence.CustomizationOption{AvatarName: "test-avatar", CustomizationName: "Class", Value: "Mage"})
	testDB.Create(&persistence.ItemDefinition{Name: "iron-sword", Game: "test-game", Category: "Equipment", Rarity: "Common", Stackable: false})
	testDB.Create(&persistence.ItemDefinition{Name: "health-potion", Game: "test-game", Category: "Powerup", Rarity: "Common", Stackable: true, MaxStack: 10, Duration: 300})
	testDB.Create(&persistence.ItemDefinition{Name: "golden-cape", Game: "test-game", Category: "Vanity", Rarity: "Rare", Stackable: false})
	testDB.Create(&persistence.ItemEffectRecord{ItemName: "iron-sword", Game: "test-game", Attribute: "strength", Modifier: "+5"})
	testDB.Create(&persistence.CurrencyDefinition{Name: "gold", Game: "test-game", Symbol: "G", Tradeable: true, MaxBalance: 1000000, InitialBalance: 100})
	testDB.Create(&persistence.CurrencyDefinition{Name: "gems", Game: "test-game", Symbol: "D", Tradeable: false, MaxBalance: 10000, InitialBalance: 0})
}

func newTestHandler() *Handler {
	return NewHandler(func(game, namespace string) (*gorm.DB, error) {
		if game == "test-game" {
			return testDB, nil
		}
		return nil, fmt.Errorf("game %q not found", game)
	})
}

func newTestMux() *http.ServeMux {
	h := newTestHandler()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/namespaces/{namespace}/games/{game}/avatars", h.CreateAvatarInstance)
	mux.HandleFunc("GET /api/v1/namespaces/{namespace}/games/{game}/avatars", h.ListAvatarInstances)
	mux.HandleFunc("GET /api/v1/namespaces/{namespace}/games/{game}/avatars/{name}", h.GetAvatarInstance)
	mux.HandleFunc("DELETE /api/v1/namespaces/{namespace}/games/{game}/avatars/{name}", h.DeleteAvatarInstance)
	mux.HandleFunc("GET /api/v1/namespaces/{namespace}/games/{game}/items", h.ListItems)
	mux.HandleFunc("GET /api/v1/namespaces/{namespace}/games/{game}/items/{name}", h.GetItem)
	mux.HandleFunc("POST /api/v1/namespaces/{namespace}/games/{game}/avatars/{name}/inventory", h.GrantItem)
	mux.HandleFunc("POST /api/v1/namespaces/{namespace}/games/{game}/avatars/{name}/equip", h.EquipItem)
	mux.HandleFunc("POST /api/v1/namespaces/{namespace}/games/{game}/avatars/{name}/unequip", h.UnequipItem)
	mux.HandleFunc("POST /api/v1/namespaces/{namespace}/games/{game}/avatars/{name}/powerups/activate", h.ActivatePowerup)
	mux.HandleFunc("GET /api/v1/namespaces/{namespace}/games/{game}/avatars/{name}/powerups", h.ListActivePowerups)
	mux.HandleFunc("GET /api/v1/namespaces/{namespace}/games/{game}/currencies", h.ListCurrencies)
	mux.HandleFunc("GET /api/v1/namespaces/{namespace}/games/{game}/currencies/{name}", h.GetCurrency)
	mux.HandleFunc("GET /api/v1/namespaces/{namespace}/games/{game}/avatars/{name}/wallet", h.GetWallet)
	mux.HandleFunc("POST /api/v1/namespaces/{namespace}/games/{game}/avatars/{name}/wallet/credit", h.CreditWallet)
	mux.HandleFunc("POST /api/v1/namespaces/{namespace}/games/{game}/avatars/{name}/wallet/debit", h.DebitWallet)
	mux.HandleFunc("POST /api/v1/namespaces/{namespace}/games/{game}/avatars/{name}/wallet/transfer", h.TransferWallet)
	mux.HandleFunc("POST /api/v1/games/{game}/avatars", h.CreateAvatarInstance)
	mux.HandleFunc("GET /api/v1/games/{game}/avatars", h.ListAvatarInstances)
	mux.HandleFunc("GET /api/v1/games/{game}/avatars/{name}", h.GetAvatarInstance)
	mux.HandleFunc("DELETE /api/v1/games/{game}/avatars/{name}", h.DeleteAvatarInstance)
	return mux
}

func TestCreateAvatarInstance(t *testing.T) {
	mux := newTestMux()

	body := AvatarInstanceRequest{
		Name:       "parzival",
		AvatarType: "test-avatar",
		Attributes: map[string]string{
			"strength":     "50",
			"intelligence": "90",
		},
		Inventory: []InventoryItem{
			{Name: "Sword", Type: "Equipment", Quantity: 1},
		},
		Achievements: []string{"First Blood"},
	}

	b, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/v1/namespaces/default/games/test-game/avatars", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp AvatarInstanceResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Name != "parzival" {
		t.Errorf("expected name parzival, got %s", resp.Name)
	}
	if resp.AvatarType != "test-avatar" {
		t.Errorf("expected avatarType test-avatar, got %s", resp.AvatarType)
	}
	if resp.Attributes["strength"] != "50" {
		t.Errorf("expected strength 50, got %s", resp.Attributes["strength"])
	}
	if len(resp.Inventory) != 1 {
		t.Errorf("expected 1 inventory item, got %d", len(resp.Inventory))
	}
	if len(resp.Achievements) != 1 {
		t.Errorf("expected 1 achievement, got %d", len(resp.Achievements))
	}

	t.Cleanup(func() {
		testDB.Where("name = ?", "parzival").Delete(&persistence.AvatarInstance{})
	})
}

func TestGetAvatarInstance(t *testing.T) {
	mux := newTestMux()

	instance := &persistence.AvatarInstance{Name: "get-test", AvatarName: "test-avatar", Game: "test-game"}
	testDB.Create(instance)
	testDB.Create(&persistence.AvatarInstanceAttribute{AvatarInstanceID: instance.ID, Name: "strength", Value: "42"})
	t.Cleanup(func() {
		testDB.Where("avatar_instance_id = ?", instance.ID).Delete(&persistence.AvatarInstanceAttribute{})
		testDB.Delete(instance)
	})

	req := httptest.NewRequest("GET", "/api/v1/namespaces/default/games/test-game/avatars/get-test", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp AvatarInstanceResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Name != "get-test" {
		t.Errorf("expected name get-test, got %s", resp.Name)
	}
	if resp.Attributes["strength"] != "42" {
		t.Errorf("expected strength 42, got %s", resp.Attributes["strength"])
	}
}

func TestGetAvatarInstanceNotFound(t *testing.T) {
	mux := newTestMux()

	req := httptest.NewRequest("GET", "/api/v1/namespaces/default/games/test-game/avatars/nonexistent", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestListAvatarInstances(t *testing.T) {
	mux := newTestMux()

	i1 := &persistence.AvatarInstance{Name: "list-test-1", AvatarName: "test-avatar", Game: "test-game"}
	i2 := &persistence.AvatarInstance{Name: "list-test-2", AvatarName: "test-avatar", Game: "test-game"}
	testDB.Create(i1)
	testDB.Create(i2)
	t.Cleanup(func() {
		testDB.Delete(i1)
		testDB.Delete(i2)
	})

	req := httptest.NewRequest("GET", "/api/v1/namespaces/default/games/test-game/avatars", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp []AvatarInstanceResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if len(resp) < 2 {
		t.Errorf("expected at least 2 instances, got %d", len(resp))
	}
}

func TestDeleteAvatarInstance(t *testing.T) {
	mux := newTestMux()

	instance := &persistence.AvatarInstance{Name: "delete-test", AvatarName: "test-avatar", Game: "test-game"}
	testDB.Create(instance)

	req := httptest.NewRequest("DELETE", "/api/v1/namespaces/default/games/test-game/avatars/delete-test", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", w.Code, w.Body.String())
	}

	var check persistence.AvatarInstance
	result := testDB.Where("name = ?", "delete-test").First(&check)
	if result.Error == nil {
		t.Error("expected instance to be deleted, but it still exists")
	}
}

func TestCreateAvatarInstanceInvalidAttribute(t *testing.T) {
	mux := newTestMux()

	body := AvatarInstanceRequest{
		Name:       "invalid-attr",
		AvatarType: "test-avatar",
		Attributes: map[string]string{
			"nonexistent_attr": "100",
		},
	}

	b, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/v1/namespaces/default/games/test-game/avatars", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreateAvatarInstanceInvalidInventory(t *testing.T) {
	mux := newTestMux()

	body := AvatarInstanceRequest{
		Name:       "invalid-inv",
		AvatarType: "test-avatar",
		Inventory: []InventoryItem{
			{Name: "Magic Wand", Type: "NonexistentCategory", Quantity: 1},
		},
	}

	b, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/v1/namespaces/default/games/test-game/avatars", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreateAvatarInstanceInvalidAchievement(t *testing.T) {
	mux := newTestMux()

	body := AvatarInstanceRequest{
		Name:         "invalid-ach",
		AvatarType:   "test-avatar",
		Achievements: []string{"Nonexistent Achievement"},
	}

	b, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/v1/namespaces/default/games/test-game/avatars", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestShorthandPathDefaultsToDefaultNamespace(t *testing.T) {
	mux := newTestMux()

	instance := &persistence.AvatarInstance{Name: "shorthand-test", AvatarName: "test-avatar", Game: "test-game"}
	testDB.Create(instance)
	t.Cleanup(func() {
		testDB.Delete(instance)
	})

	req := httptest.NewRequest("GET", "/api/v1/games/test-game/avatars/shorthand-test", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreateAvatarInstanceWithCustomizations(t *testing.T) {
	mux := newTestMux()

	body := AvatarInstanceRequest{
		Name:       "custom-avatar",
		AvatarType: "test-avatar",
		Customizations: map[string]string{
			"Race":  "Elf",
			"Class": "Mage",
		},
	}

	b, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/v1/namespaces/default/games/test-game/avatars", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp AvatarInstanceResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Customizations["Race"] != "Elf" {
		t.Errorf("expected Race=Elf, got %s", resp.Customizations["Race"])
	}
	if resp.Customizations["Class"] != "Mage" {
		t.Errorf("expected Class=Mage, got %s", resp.Customizations["Class"])
	}

	t.Cleanup(func() {
		var instance persistence.AvatarInstance
		testDB.Where("name = ?", "custom-avatar").First(&instance)
		testDB.Where("avatar_instance_id = ?", instance.ID).Delete(&persistence.AvatarInstanceCustomization{})
		testDB.Delete(&instance)
	})
}

func TestCreateAvatarInstanceInvalidCustomizationType(t *testing.T) {
	mux := newTestMux()

	body := AvatarInstanceRequest{
		Name:       "invalid-cust-type",
		AvatarType: "test-avatar",
		Customizations: map[string]string{
			"Alignment": "Chaotic Good",
		},
	}

	b, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/v1/namespaces/default/games/test-game/avatars", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreateAvatarInstanceInvalidCustomizationOption(t *testing.T) {
	mux := newTestMux()

	body := AvatarInstanceRequest{
		Name:       "invalid-cust-opt",
		AvatarType: "test-avatar",
		Customizations: map[string]string{
			"Race": "Dragon",
		},
	}

	b, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/v1/namespaces/default/games/test-game/avatars", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestListItems(t *testing.T) {
	mux := newTestMux()

	req := httptest.NewRequest("GET", "/api/v1/namespaces/default/games/test-game/items", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp []ItemResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if len(resp) < 3 {
		t.Errorf("expected at least 3 items, got %d", len(resp))
	}
}

func TestGetItem(t *testing.T) {
	mux := newTestMux()

	req := httptest.NewRequest("GET", "/api/v1/namespaces/default/games/test-game/items/iron-sword", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp ItemResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Name != "iron-sword" {
		t.Errorf("expected iron-sword, got %s", resp.Name)
	}
	if resp.Effects["strength"] != "+5" {
		t.Errorf("expected strength +5, got %s", resp.Effects["strength"])
	}
}

func TestGrantItem(t *testing.T) {
	mux := newTestMux()

	instance := &persistence.AvatarInstance{Name: "grant-test", AvatarName: "test-avatar", Game: "test-game"}
	testDB.Create(instance)
	t.Cleanup(func() {
		testDB.Where("avatar_instance_id = ?", instance.ID).Delete(&persistence.AvatarInstanceInventoryItem{})
		testDB.Delete(instance)
	})

	body := GrantItemRequest{ItemName: "iron-sword", Quantity: 1}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/v1/namespaces/default/games/test-game/avatars/grant-test/inventory", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp AvatarInstanceResponse
	json.NewDecoder(w.Body).Decode(&resp)

	found := false
	for _, item := range resp.Inventory {
		if item.Name == "iron-sword" {
			found = true
		}
	}
	if !found {
		t.Error("expected iron-sword in inventory")
	}
}

func TestGrantItemNotInCatalog(t *testing.T) {
	mux := newTestMux()

	instance := &persistence.AvatarInstance{Name: "grant-invalid-test", AvatarName: "test-avatar", Game: "test-game"}
	testDB.Create(instance)
	t.Cleanup(func() {
		testDB.Delete(instance)
	})

	body := GrantItemRequest{ItemName: "Nonexistent Item", Quantity: 1}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/v1/namespaces/default/games/test-game/avatars/grant-invalid-test/inventory", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestEquipAndUnequipItem(t *testing.T) {
	mux := newTestMux()

	instance := &persistence.AvatarInstance{Name: "equip-test", AvatarName: "test-avatar", Game: "test-game"}
	testDB.Create(instance)
	inv := &persistence.AvatarInstanceInventoryItem{AvatarInstanceID: instance.ID, Name: "iron-sword", Type: "Equipment", Quantity: 1}
	testDB.Create(inv)
	t.Cleanup(func() {
		testDB.Where("avatar_instance_id = ?", instance.ID).Delete(&persistence.AvatarInstanceInventoryItem{})
		testDB.Delete(instance)
	})

	equipBody := EquipRequest{ItemName: "iron-sword"}
	b, _ := json.Marshal(equipBody)
	req := httptest.NewRequest("POST", "/api/v1/namespaces/default/games/test-game/avatars/equip-test/equip", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("equip: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp AvatarInstanceResponse
	json.NewDecoder(w.Body).Decode(&resp)

	for _, item := range resp.Inventory {
		if item.Name == "iron-sword" && !item.Equipped {
			t.Error("expected iron-sword to be equipped")
		}
	}

	unequipBody := EquipRequest{ItemName: "iron-sword"}
	b, _ = json.Marshal(unequipBody)
	req = httptest.NewRequest("POST", "/api/v1/namespaces/default/games/test-game/avatars/equip-test/unequip", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("unequip: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	json.NewDecoder(w.Body).Decode(&resp)
	for _, item := range resp.Inventory {
		if item.Name == "iron-sword" && item.Equipped {
			t.Error("expected iron-sword to be unequipped")
		}
	}
}

func TestEquipNonEquipment(t *testing.T) {
	mux := newTestMux()

	instance := &persistence.AvatarInstance{Name: "equip-vanity-test", AvatarName: "test-avatar", Game: "test-game"}
	testDB.Create(instance)
	inv := &persistence.AvatarInstanceInventoryItem{AvatarInstanceID: instance.ID, Name: "golden-cape", Type: "Vanity", Quantity: 1}
	testDB.Create(inv)
	t.Cleanup(func() {
		testDB.Where("avatar_instance_id = ?", instance.ID).Delete(&persistence.AvatarInstanceInventoryItem{})
		testDB.Delete(instance)
	})

	equipBody := EquipRequest{ItemName: "golden-cape"}
	b, _ := json.Marshal(equipBody)
	req := httptest.NewRequest("POST", "/api/v1/namespaces/default/games/test-game/avatars/equip-vanity-test/equip", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestActivatePowerup(t *testing.T) {
	mux := newTestMux()

	instance := &persistence.AvatarInstance{Name: "powerup-test", AvatarName: "test-avatar", Game: "test-game"}
	testDB.Create(instance)
	inv := &persistence.AvatarInstanceInventoryItem{AvatarInstanceID: instance.ID, Name: "health-potion", Type: "Powerup", Quantity: 3}
	testDB.Create(inv)
	t.Cleanup(func() {
		testDB.Where("avatar_instance_id = ?", instance.ID).Delete(&persistence.ActivePowerup{})
		testDB.Where("avatar_instance_id = ?", instance.ID).Delete(&persistence.AvatarInstanceInventoryItem{})
		testDB.Delete(instance)
	})

	body := ActivatePowerupRequest{ItemName: "health-potion"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/v1/namespaces/default/games/test-game/avatars/powerup-test/powerups/activate", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp ActivePowerupResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.ItemName != "health-potion" {
		t.Errorf("expected Health Potion, got %s", resp.ItemName)
	}
	if resp.ExpiresAt <= resp.ActivatedAt {
		t.Error("expiresAt should be after activatedAt")
	}

	var remaining persistence.AvatarInstanceInventoryItem
	testDB.Where("avatar_instance_id = ? AND name = ?", instance.ID, "health-potion").First(&remaining)
	if remaining.Quantity != 2 {
		t.Errorf("expected quantity 2 after activation, got %d", remaining.Quantity)
	}
}

func TestGameNotFound(t *testing.T) {
	mux := newTestMux()

	req := httptest.NewRequest("GET", "/api/v1/namespaces/default/games/nonexistent-game/avatars", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestListCurrencies(t *testing.T) {
	mux := newTestMux()

	req := httptest.NewRequest("GET", "/api/v1/namespaces/default/games/test-game/currencies", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp []CurrencyResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if len(resp) < 2 {
		t.Fatalf("expected at least 2 currencies, got %d", len(resp))
	}
}

func TestGetCurrency(t *testing.T) {
	mux := newTestMux()

	req := httptest.NewRequest("GET", "/api/v1/namespaces/default/games/test-game/currencies/gold", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp CurrencyResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Name != "gold" {
		t.Fatalf("expected gold, got %s", resp.Name)
	}
	if resp.Symbol != "G" {
		t.Fatalf("expected symbol G, got %s", resp.Symbol)
	}
	if !resp.Tradeable {
		t.Fatal("expected gold to be tradeable")
	}
}

func TestCreditAndDebitWallet(t *testing.T) {
	mux := newTestMux()

	instance := &persistence.AvatarInstance{Name: "wallet-test", AvatarName: "test-avatar", Game: "test-game"}
	testDB.Create(instance)
	defer func() {
		testDB.Where("avatar_instance_id = ?", instance.ID).Delete(&persistence.AvatarCurrencyBalance{})
		testDB.Delete(instance)
	}()

	creditReq := CreditDebitRequest{Currency: "gold", Amount: 500}
	body, _ := json.Marshal(creditReq)
	req := httptest.NewRequest("POST", "/api/v1/namespaces/default/games/test-game/avatars/wallet-test/wallet/credit", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("credit: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var creditResp WalletBalanceResponse
	json.NewDecoder(w.Body).Decode(&creditResp)
	if creditResp.Balance != 500 {
		t.Fatalf("expected balance 500, got %d", creditResp.Balance)
	}

	debitReq := CreditDebitRequest{Currency: "gold", Amount: 200}
	body, _ = json.Marshal(debitReq)
	req = httptest.NewRequest("POST", "/api/v1/namespaces/default/games/test-game/avatars/wallet-test/wallet/debit", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("debit: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var debitResp WalletBalanceResponse
	json.NewDecoder(w.Body).Decode(&debitResp)
	if debitResp.Balance != 300 {
		t.Fatalf("expected balance 300, got %d", debitResp.Balance)
	}
}

func TestDebitInsufficientBalance(t *testing.T) {
	mux := newTestMux()

	instance := &persistence.AvatarInstance{Name: "debit-fail-test", AvatarName: "test-avatar", Game: "test-game"}
	testDB.Create(instance)
	defer func() {
		testDB.Where("avatar_instance_id = ?", instance.ID).Delete(&persistence.AvatarCurrencyBalance{})
		testDB.Delete(instance)
	}()

	testDB.Create(&persistence.AvatarCurrencyBalance{AvatarInstanceID: instance.ID, CurrencyName: "gold", Balance: 50})

	debitReq := CreditDebitRequest{Currency: "gold", Amount: 200}
	body, _ := json.Marshal(debitReq)
	req := httptest.NewRequest("POST", "/api/v1/namespaces/default/games/test-game/avatars/debit-fail-test/wallet/debit", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestTransferWallet(t *testing.T) {
	mux := newTestMux()

	sender := &persistence.AvatarInstance{Name: "sender-test", AvatarName: "test-avatar", Game: "test-game"}
	receiver := &persistence.AvatarInstance{Name: "receiver-test", AvatarName: "test-avatar", Game: "test-game"}
	testDB.Create(sender)
	testDB.Create(receiver)
	defer func() {
		testDB.Where("avatar_instance_id = ?", sender.ID).Delete(&persistence.AvatarCurrencyBalance{})
		testDB.Where("avatar_instance_id = ?", receiver.ID).Delete(&persistence.AvatarCurrencyBalance{})
		testDB.Delete(sender)
		testDB.Delete(receiver)
	}()

	testDB.Create(&persistence.AvatarCurrencyBalance{AvatarInstanceID: sender.ID, CurrencyName: "gold", Balance: 1000})

	transferReq := TransferRequest{Currency: "gold", Amount: 300, ToAvatar: "receiver-test"}
	body, _ := json.Marshal(transferReq)
	req := httptest.NewRequest("POST", "/api/v1/namespaces/default/games/test-game/avatars/sender-test/wallet/transfer", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp WalletBalanceResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Balance != 700 {
		t.Fatalf("expected sender balance 700, got %d", resp.Balance)
	}

	var receiverBalance persistence.AvatarCurrencyBalance
	testDB.Where("avatar_instance_id = ? AND currency_name = ?", receiver.ID, "gold").First(&receiverBalance)
	if receiverBalance.Balance != 300 {
		t.Fatalf("expected receiver balance 300, got %d", receiverBalance.Balance)
	}
}

func TestTransferNonTradeableCurrency(t *testing.T) {
	mux := newTestMux()

	sender := &persistence.AvatarInstance{Name: "transfer-nontrade-test", AvatarName: "test-avatar", Game: "test-game"}
	testDB.Create(sender)
	defer func() {
		testDB.Where("avatar_instance_id = ?", sender.ID).Delete(&persistence.AvatarCurrencyBalance{})
		testDB.Delete(sender)
	}()

	testDB.Create(&persistence.AvatarCurrencyBalance{AvatarInstanceID: sender.ID, CurrencyName: "gems", Balance: 1000})

	transferReq := TransferRequest{Currency: "gems", Amount: 100, ToAvatar: "someone"}
	body, _ := json.Marshal(transferReq)
	req := httptest.NewRequest("POST", "/api/v1/namespaces/default/games/test-game/avatars/transfer-nontrade-test/wallet/transfer", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for non-tradeable, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetWallet(t *testing.T) {
	mux := newTestMux()

	instance := &persistence.AvatarInstance{Name: "get-wallet-test", AvatarName: "test-avatar", Game: "test-game"}
	testDB.Create(instance)
	defer func() {
		testDB.Where("avatar_instance_id = ?", instance.ID).Delete(&persistence.AvatarCurrencyBalance{})
		testDB.Delete(instance)
	}()

	testDB.Create(&persistence.AvatarCurrencyBalance{AvatarInstanceID: instance.ID, CurrencyName: "gold", Balance: 750})
	testDB.Create(&persistence.AvatarCurrencyBalance{AvatarInstanceID: instance.ID, CurrencyName: "gems", Balance: 25})

	req := httptest.NewRequest("GET", "/api/v1/namespaces/default/games/test-game/avatars/get-wallet-test/wallet", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp []WalletBalanceResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if len(resp) != 2 {
		t.Fatalf("expected 2 wallet entries, got %d", len(resp))
	}
}

func TestCreditExceedMaxBalance(t *testing.T) {
	mux := newTestMux()

	instance := &persistence.AvatarInstance{Name: "max-balance-test", AvatarName: "test-avatar", Game: "test-game"}
	testDB.Create(instance)
	defer func() {
		testDB.Where("avatar_instance_id = ?", instance.ID).Delete(&persistence.AvatarCurrencyBalance{})
		testDB.Delete(instance)
	}()

	testDB.Create(&persistence.AvatarCurrencyBalance{AvatarInstanceID: instance.ID, CurrencyName: "gems", Balance: 9900})

	creditReq := CreditDebitRequest{Currency: "gems", Amount: 200}
	body, _ := json.Marshal(creditReq)
	req := httptest.NewRequest("POST", "/api/v1/namespaces/default/games/test-game/avatars/max-balance-test/wallet/credit", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for max balance exceeded, got %d: %s", w.Code, w.Body.String())
	}
}
