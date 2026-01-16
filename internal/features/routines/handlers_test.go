package routines_test

import (
	"net/http"
	"strconv"
	"strings"
	"testing"

	"phobos/internal/features/routines"
	"phobos/internal/features/templates"
	"phobos/internal/testutil"
)

func TestHandleList_Empty(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	resp := app.Request("GET", "/routines", "")

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body := testutil.ReadBody(t, resp)
	if !strings.Contains(body, "Routines") {
		t.Error("expected page to contain 'Routines'")
	}
	if !strings.Contains(body, "No routines yet") {
		t.Error("expected page to show empty state message")
	}
}

func TestHandleCreate(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	resp := app.HTMXRequest("POST", "/routines", "name=Push+Pull+Legs")

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body := testutil.ReadBody(t, resp)
	if !strings.Contains(body, "Push Pull Legs") {
		t.Error("expected response to contain routine name")
	}

	// Verify it was saved
	routineList, err := routines.ListAll(app.DB)
	if err != nil {
		t.Fatalf("failed to list routines: %v", err)
	}
	if len(routineList) != 1 {
		t.Errorf("expected 1 routine, got %d", len(routineList))
	}
}

func TestHandleShow(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	id, _ := routines.Create(app.DB, "My Routine")

	resp := app.Request("GET", "/routines/"+strconv.FormatInt(id, 10), "")

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body := testutil.ReadBody(t, resp)
	if !strings.Contains(body, "My Routine") {
		t.Error("expected page to contain routine name")
	}
	if !strings.Contains(body, "Add Template") {
		t.Error("expected page to show add template form")
	}
}

func TestHandleShow_NotFound(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	resp := app.Request("GET", "/routines/999", "")

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", resp.StatusCode)
	}
}

func TestHandleUpdate(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	id, _ := routines.Create(app.DB, "Old Name")

	resp := app.HTMXRequest("PUT", "/routines/"+strconv.FormatInt(id, 10), "name=New+Name")

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Verify update
	routine, _ := routines.GetByID(app.DB, id)
	if routine.Name != "New Name" {
		t.Errorf("expected name 'New Name', got '%s'", routine.Name)
	}
}

func TestHandleDelete(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	id, _ := routines.Create(app.DB, "To Delete")

	resp := app.HTMXRequest("DELETE", "/routines/"+strconv.FormatInt(id, 10), "")

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Verify deletion
	routine, _ := routines.GetByID(app.DB, id)
	if routine != nil {
		t.Error("expected routine to be deleted")
	}
}

func TestHandleAddTemplate(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	routineID, _ := routines.Create(app.DB, "PPL")
	templateID, _ := templates.Create(app.DB, "Push Day")

	resp := app.HTMXRequest("POST", "/routines/"+strconv.FormatInt(routineID, 10)+"/templates",
		"template_id="+strconv.FormatInt(templateID, 10))

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body := testutil.ReadBody(t, resp)
	if !strings.Contains(body, "Push Day") {
		t.Error("expected response to contain template name")
	}

	// Verify it was added
	routine, _ := routines.GetByID(app.DB, routineID)
	if len(routine.Templates) != 1 {
		t.Errorf("expected 1 template, got %d", len(routine.Templates))
	}
}

func TestHandleRemoveTemplate(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	routineID, _ := routines.Create(app.DB, "PPL")
	templateID, _ := templates.Create(app.DB, "Push Day")
	rtID, _ := routines.AddTemplate(app.DB, routineID, templateID)

	resp := app.HTMXRequest("DELETE", "/routines/templates/"+strconv.FormatInt(rtID, 10), "")

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Verify removal
	routine, _ := routines.GetByID(app.DB, routineID)
	if len(routine.Templates) != 0 {
		t.Errorf("expected 0 templates, got %d", len(routine.Templates))
	}
}

func TestRoutineWithMultipleTemplates(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	routineID, _ := routines.Create(app.DB, "PPL")

	pushID, _ := templates.Create(app.DB, "Push Day")
	pullID, _ := templates.Create(app.DB, "Pull Day")
	legsID, _ := templates.Create(app.DB, "Legs Day")

	routines.AddTemplate(app.DB, routineID, pushID)
	routines.AddTemplate(app.DB, routineID, pullID)
	routines.AddTemplate(app.DB, routineID, legsID)

	// Get routine and verify order
	routine, _ := routines.GetByID(app.DB, routineID)
	if len(routine.Templates) != 3 {
		t.Fatalf("expected 3 templates, got %d", len(routine.Templates))
	}

	if routine.Templates[0].Template.Name != "Push Day" {
		t.Errorf("expected first template 'Push Day', got '%s'", routine.Templates[0].Template.Name)
	}
	if routine.Templates[0].Position != 1 {
		t.Errorf("expected position 1, got %d", routine.Templates[0].Position)
	}
	if routine.Templates[1].Template.Name != "Pull Day" {
		t.Errorf("expected second template 'Pull Day', got '%s'", routine.Templates[1].Template.Name)
	}
	if routine.Templates[2].Template.Name != "Legs Day" {
		t.Errorf("expected third template 'Legs Day', got '%s'", routine.Templates[2].Template.Name)
	}
}

func TestRoutineShowIncludesStartWorkoutLinks(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	routineID, _ := routines.Create(app.DB, "PPL")
	templateID, _ := templates.Create(app.DB, "Push Day")
	routines.AddTemplate(app.DB, routineID, templateID)

	resp := app.Request("GET", "/routines/"+strconv.FormatInt(routineID, 10), "")

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body := testutil.ReadBody(t, resp)
	// Should have a link to start workout from template
	if !strings.Contains(body, "Start") {
		t.Error("expected page to contain 'Start' link for templates")
	}
	expectedLink := "/workouts/new?template_id=" + strconv.FormatInt(templateID, 10)
	if !strings.Contains(body, expectedLink) {
		t.Errorf("expected page to contain link '%s'", expectedLink)
	}
}
