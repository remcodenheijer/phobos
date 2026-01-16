package exercises_test

import (
	"net/http"
	"strconv"
	"strings"
	"testing"

	"phobos/internal/features/exercises"
	"phobos/internal/testutil"
)

func TestHandleList_Empty(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	resp := app.Request("GET", "/exercises", "")

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body := testutil.ReadBody(t, resp)
	if !strings.Contains(body, "Exercises") {
		t.Error("expected page to contain 'Exercises'")
	}
	if !strings.Contains(body, "No exercises yet") {
		t.Error("expected page to show empty state message")
	}
}

func TestHandleCreate(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	// Create an exercise
	resp := app.HTMXRequest("POST", "/exercises", "name=Bench+Press")

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body := testutil.ReadBody(t, resp)
	if !strings.Contains(body, "Bench Press") {
		t.Error("expected response to contain exercise name")
	}

	// Verify it was saved
	exerciseList, err := exercises.ListAll(app.DB)
	if err != nil {
		t.Fatalf("failed to list exercises: %v", err)
	}
	if len(exerciseList) != 1 {
		t.Errorf("expected 1 exercise, got %d", len(exerciseList))
	}
	if exerciseList[0].Name != "Bench Press" {
		t.Errorf("expected exercise name 'Bench Press', got '%s'", exerciseList[0].Name)
	}
}

func TestHandleCreate_EmptyName(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	resp := app.HTMXRequest("POST", "/exercises", "name=")

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestHandleList_WithExercises(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	// Create some exercises
	exercises.Create(app.DB, "Squat")
	exercises.Create(app.DB, "Deadlift")
	exercises.Create(app.DB, "Bench Press")

	resp := app.Request("GET", "/exercises", "")

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body := testutil.ReadBody(t, resp)
	if !strings.Contains(body, "Squat") {
		t.Error("expected page to contain 'Squat'")
	}
	if !strings.Contains(body, "Deadlift") {
		t.Error("expected page to contain 'Deadlift'")
	}
	if !strings.Contains(body, "Bench Press") {
		t.Error("expected page to contain 'Bench Press'")
	}
}

func TestHandleDelete(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	// Create an exercise
	id, err := exercises.Create(app.DB, "Squat")
	if err != nil {
		t.Fatalf("failed to create exercise: %v", err)
	}

	// Delete it
	resp := app.HTMXRequest("DELETE", "/exercises/"+itoa(id), "")

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Verify it was deleted
	exercise, err := exercises.GetByID(app.DB, id)
	if err != nil {
		t.Fatalf("failed to get exercise: %v", err)
	}
	if exercise != nil {
		t.Error("expected exercise to be deleted")
	}
}

func TestHandleDelete_InvalidID(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	resp := app.HTMXRequest("DELETE", "/exercises/invalid", "")

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestHandleSearch(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	// Create some exercises
	exercises.Create(app.DB, "Bench Press")
	exercises.Create(app.DB, "Incline Bench Press")
	exercises.Create(app.DB, "Squat")

	// Search for 'bench'
	resp := app.HTMXRequest("GET", "/exercises/search?q=bench", "")

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body := testutil.ReadBody(t, resp)
	if !strings.Contains(body, "Bench Press") {
		t.Error("expected results to contain 'Bench Press'")
	}
	if !strings.Contains(body, "Incline Bench Press") {
		t.Error("expected results to contain 'Incline Bench Press'")
	}
	if strings.Contains(body, "Squat") {
		t.Error("expected results to NOT contain 'Squat'")
	}
}

// Helper to convert int64 to string
func itoa(i int64) string {
	return strconv.FormatInt(i, 10)
}
