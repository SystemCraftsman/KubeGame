package api

import (
	"bytes"
	"context"
	"encoding/json"
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

func TestGameNotFound(t *testing.T) {
	mux := newTestMux()

	req := httptest.NewRequest("GET", "/api/v1/namespaces/default/games/nonexistent-game/avatars", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}
