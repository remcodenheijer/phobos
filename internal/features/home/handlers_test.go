package home_test

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"phobos/internal/features/routines"
	"phobos/internal/features/templates"
	"phobos/internal/features/workouts"
	"phobos/internal/testutil"
)

func TestHandleHome_Empty(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	resp := app.Request("GET", "/", "")

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body := testutil.ReadBody(t, resp)
	if !strings.Contains(body, "Dashboard") {
		t.Error("expected page to contain 'Dashboard'")
	}
	if !strings.Contains(body, "New Workout") {
		t.Error("expected page to contain 'New Workout' button")
	}
}

func TestHandleHome_WithActiveWorkouts(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	// Create an active workout
	workouts.Create(app.DB, "Monday Push", time.Now(), nil)

	resp := app.Request("GET", "/", "")

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body := testutil.ReadBody(t, resp)
	if !strings.Contains(body, "Active Workouts") {
		t.Error("expected page to contain 'Active Workouts' section")
	}
	if !strings.Contains(body, "Monday Push") {
		t.Error("expected page to contain active workout name")
	}
	if !strings.Contains(body, "In Progress") {
		t.Error("expected page to show 'In Progress' badge")
	}
}

func TestHandleHome_WithTemplates(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	// Create templates
	templates.Create(app.DB, "Push Day")
	templates.Create(app.DB, "Pull Day")

	resp := app.Request("GET", "/", "")

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body := testutil.ReadBody(t, resp)
	if !strings.Contains(body, "Quick Start from Template") {
		t.Error("expected page to contain 'Quick Start from Template' section")
	}
	if !strings.Contains(body, "Push Day") {
		t.Error("expected page to contain 'Push Day' template")
	}
	if !strings.Contains(body, "Pull Day") {
		t.Error("expected page to contain 'Pull Day' template")
	}
}

func TestHandleHome_WithRoutines(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	// Create routines
	routines.Create(app.DB, "PPL Program")

	resp := app.Request("GET", "/", "")

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body := testutil.ReadBody(t, resp)
	if !strings.Contains(body, "Routines") {
		t.Error("expected page to contain 'Routines' section")
	}
	if !strings.Contains(body, "PPL Program") {
		t.Error("expected page to contain routine name")
	}
}

func TestHandleHome_WithRecentWorkouts(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	// Create and finish some workouts
	id1, _ := workouts.Create(app.DB, "Workout 1", time.Now().AddDate(0, 0, -1), nil)
	workouts.Finish(app.DB, id1)
	id2, _ := workouts.Create(app.DB, "Workout 2", time.Now().AddDate(0, 0, -2), nil)
	workouts.Finish(app.DB, id2)

	resp := app.Request("GET", "/", "")

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body := testutil.ReadBody(t, resp)
	if !strings.Contains(body, "Recent Workouts") {
		t.Error("expected page to contain 'Recent Workouts' section")
	}
	if !strings.Contains(body, "Workout 1") {
		t.Error("expected page to contain 'Workout 1'")
	}
	if !strings.Contains(body, "Workout 2") {
		t.Error("expected page to contain 'Workout 2'")
	}
	if !strings.Contains(body, "View all") {
		t.Error("expected page to contain 'View all' link to history")
	}
}

func TestHandleHome_NoActiveWorkoutsSection_WhenEmpty(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	// Create only a finished workout
	id, _ := workouts.Create(app.DB, "Finished", time.Now(), nil)
	workouts.Finish(app.DB, id)

	resp := app.Request("GET", "/", "")

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body := testutil.ReadBody(t, resp)
	// Active Workouts section should not appear when there are none
	if strings.Contains(body, "Active Workouts") {
		t.Error("expected page to NOT contain 'Active Workouts' section when empty")
	}
}

func TestHandleHome_LimitsRecentWorkouts(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	// Create more than 5 finished workouts
	for i := 0; i < 10; i++ {
		id, _ := workouts.Create(app.DB, "Workout", time.Now().AddDate(0, 0, -i), nil)
		workouts.Finish(app.DB, id)
	}

	resp := app.Request("GET", "/", "")

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body := testutil.ReadBody(t, resp)
	// Should show "View all" link because there are more workouts
	if !strings.Contains(body, "View all") {
		t.Error("expected page to contain 'View all' link when there are more than 5 workouts")
	}
}
