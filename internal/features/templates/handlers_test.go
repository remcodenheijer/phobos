package templates_test

import (
	"net/http"
	"strconv"
	"strings"
	"testing"

	"phobos/internal/features/exercises"
	"phobos/internal/features/templates"
	"phobos/internal/testutil"
)

func TestHandleList_Empty(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	resp := app.Request("GET", "/templates", "")

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body := testutil.ReadBody(t, resp)
	if !strings.Contains(body, "Workout Templates") {
		t.Error("expected page to contain 'Workout Templates'")
	}
	if !strings.Contains(body, "No templates yet") {
		t.Error("expected page to show empty state message")
	}
}

func TestHandleCreate(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	resp := app.HTMXRequest("POST", "/templates", "name=Push+Day")

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body := testutil.ReadBody(t, resp)
	if !strings.Contains(body, "Push Day") {
		t.Error("expected response to contain template name")
	}

	// Verify it was saved
	templateList, err := templates.ListAll(app.DB)
	if err != nil {
		t.Fatalf("failed to list templates: %v", err)
	}
	if len(templateList) != 1 {
		t.Errorf("expected 1 template, got %d", len(templateList))
	}
}

func TestHandleShow(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	// Create a template
	id, _ := templates.Create(app.DB, "Leg Day")

	resp := app.Request("GET", "/templates/"+strconv.FormatInt(id, 10), "")

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body := testutil.ReadBody(t, resp)
	if !strings.Contains(body, "Leg Day") {
		t.Error("expected page to contain template name")
	}
	if !strings.Contains(body, "Add Exercise") {
		t.Error("expected page to show add exercise form")
	}
}

func TestHandleShow_NotFound(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	resp := app.Request("GET", "/templates/999", "")

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", resp.StatusCode)
	}
}

func TestHandleUpdate(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	id, _ := templates.Create(app.DB, "Old Name")

	resp := app.HTMXRequest("PUT", "/templates/"+strconv.FormatInt(id, 10), "name=New+Name")

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Verify update
	tmpl, _ := templates.GetByID(app.DB, id)
	if tmpl.Name != "New Name" {
		t.Errorf("expected name 'New Name', got '%s'", tmpl.Name)
	}
}

func TestHandleDelete(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	id, _ := templates.Create(app.DB, "To Delete")

	resp := app.HTMXRequest("DELETE", "/templates/"+strconv.FormatInt(id, 10), "")

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Verify deletion
	tmpl, _ := templates.GetByID(app.DB, id)
	if tmpl != nil {
		t.Error("expected template to be deleted")
	}
}

func TestHandleAddExercise(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	// Create template and exercise
	templateID, _ := templates.Create(app.DB, "Push Day")
	exerciseID, _ := exercises.Create(app.DB, "Bench Press")

	resp := app.HTMXRequest("POST", "/templates/"+strconv.FormatInt(templateID, 10)+"/exercises",
		"exercise_id="+strconv.FormatInt(exerciseID, 10)+"&target_sets=3&target_reps=10")

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body := testutil.ReadBody(t, resp)
	if !strings.Contains(body, "Bench Press") {
		t.Error("expected response to contain exercise name")
	}
	if !strings.Contains(body, "3 sets") {
		t.Error("expected response to contain target sets")
	}

	// Verify it was added
	tmpl, _ := templates.GetByID(app.DB, templateID)
	if len(tmpl.Exercises) != 1 {
		t.Errorf("expected 1 exercise, got %d", len(tmpl.Exercises))
	}
}

func TestHandleRemoveExercise(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	// Create template and exercise
	templateID, _ := templates.Create(app.DB, "Push Day")
	exerciseID, _ := exercises.Create(app.DB, "Bench Press")
	teID, _ := templates.AddExercise(app.DB, templateID, exerciseID, 3, 10)

	resp := app.HTMXRequest("DELETE", "/templates/exercises/"+strconv.FormatInt(teID, 10), "")

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Verify removal
	tmpl, _ := templates.GetByID(app.DB, templateID)
	if len(tmpl.Exercises) != 0 {
		t.Errorf("expected 0 exercises, got %d", len(tmpl.Exercises))
	}
}

func TestTemplateWithMultipleExercises(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	// Create template
	templateID, _ := templates.Create(app.DB, "Full Body")

	// Create exercises
	squatID, _ := exercises.Create(app.DB, "Squat")
	benchID, _ := exercises.Create(app.DB, "Bench Press")
	rowID, _ := exercises.Create(app.DB, "Barbell Row")

	// Add exercises to template
	templates.AddExercise(app.DB, templateID, squatID, 5, 5)
	templates.AddExercise(app.DB, templateID, benchID, 3, 8)
	templates.AddExercise(app.DB, templateID, rowID, 3, 8)

	// Get template and verify order
	tmpl, _ := templates.GetByID(app.DB, templateID)
	if len(tmpl.Exercises) != 3 {
		t.Fatalf("expected 3 exercises, got %d", len(tmpl.Exercises))
	}

	if tmpl.Exercises[0].Exercise.Name != "Squat" {
		t.Errorf("expected first exercise to be Squat, got %s", tmpl.Exercises[0].Exercise.Name)
	}
	if tmpl.Exercises[0].Position != 1 {
		t.Errorf("expected first exercise position to be 1, got %d", tmpl.Exercises[0].Position)
	}
	if tmpl.Exercises[1].Exercise.Name != "Bench Press" {
		t.Errorf("expected second exercise to be Bench Press, got %s", tmpl.Exercises[1].Exercise.Name)
	}
	if tmpl.Exercises[2].Exercise.Name != "Barbell Row" {
		t.Errorf("expected third exercise to be Barbell Row, got %s", tmpl.Exercises[2].Exercise.Name)
	}
}
