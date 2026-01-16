package workouts_test

import (
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"phobos/internal/features/exercises"
	"phobos/internal/features/templates"
	"phobos/internal/features/workouts"
	"phobos/internal/testutil"
)

func TestHandleList_Empty(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	resp := app.Request("GET", "/workouts", "")

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body := testutil.ReadBody(t, resp)
	if !strings.Contains(body, "Active Workouts") {
		t.Error("expected page to contain 'Active Workouts'")
	}
	if !strings.Contains(body, "No active workouts") {
		t.Error("expected page to show empty state message")
	}
}

func TestHandleNew(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	resp := app.Request("GET", "/workouts/new", "")

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body := testutil.ReadBody(t, resp)
	if !strings.Contains(body, "Start New Workout") {
		t.Error("expected page to contain 'Start New Workout'")
	}
}

func TestHandleCreate(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	resp := app.Request("POST", "/workouts", "name=Monday+Push&date=2024-01-15")

	// Should redirect to the new workout
	if resp.StatusCode != http.StatusSeeOther {
		t.Errorf("expected status 303, got %d", resp.StatusCode)
	}

	// Verify workout was created
	workoutList, _ := workouts.ListInProgress(app.DB)
	if len(workoutList) != 1 {
		t.Errorf("expected 1 workout, got %d", len(workoutList))
	}
	if workoutList[0].Name != "Monday Push" {
		t.Errorf("expected name 'Monday Push', got '%s'", workoutList[0].Name)
	}
}

func TestHandleShow_InProgress(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	id, _ := workouts.Create(app.DB, "Test Workout", time.Now(), nil)

	resp := app.Request("GET", "/workouts/"+strconv.FormatInt(id, 10), "")

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body := testutil.ReadBody(t, resp)
	if !strings.Contains(body, "Test Workout") {
		t.Error("expected page to contain workout name")
	}
	if !strings.Contains(body, "Add Exercise") {
		t.Error("expected page to show add exercise form for in-progress workout")
	}
	if !strings.Contains(body, "Finish Workout") {
		t.Error("expected page to show finish button")
	}
}

func TestHandleUpdate(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	id, _ := workouts.Create(app.DB, "Old Name", time.Now(), nil)

	resp := app.HTMXRequest("PUT", "/workouts/"+strconv.FormatInt(id, 10),
		"name=New+Name&date=2024-01-20&notes=Some+notes")

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Verify update
	workout, _ := workouts.GetByID(app.DB, id)
	if workout.Name != "New Name" {
		t.Errorf("expected name 'New Name', got '%s'", workout.Name)
	}
	if workout.Notes != "Some notes" {
		t.Errorf("expected notes 'Some notes', got '%s'", workout.Notes)
	}
}

func TestHandleFinish(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	id, _ := workouts.Create(app.DB, "To Finish", time.Now(), nil)

	resp := app.HTMXRequest("POST", "/workouts/"+strconv.FormatInt(id, 10)+"/finish", "")

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Verify workout is finished
	workout, _ := workouts.GetByID(app.DB, id)
	if !workout.IsFinished() {
		t.Error("expected workout to be finished")
	}
}

func TestHandleShow_Finished(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	id, _ := workouts.Create(app.DB, "Finished Workout", time.Now(), nil)
	workouts.Finish(app.DB, id)

	resp := app.Request("GET", "/workouts/"+strconv.FormatInt(id, 10), "")

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body := testutil.ReadBody(t, resp)
	if !strings.Contains(body, "Completed") {
		t.Error("expected page to show completed status")
	}
	// Should NOT show add exercise form for finished workout
	if strings.Contains(body, "Add Exercise") {
		t.Error("expected page to NOT show add exercise form for finished workout")
	}
}

func TestHandleHistory(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	// Create and finish some workouts
	id1, _ := workouts.Create(app.DB, "Workout 1", time.Now(), nil)
	workouts.Finish(app.DB, id1)
	id2, _ := workouts.Create(app.DB, "Workout 2", time.Now(), nil)
	workouts.Finish(app.DB, id2)
	workouts.Create(app.DB, "In Progress", time.Now(), nil) // Not finished

	resp := app.Request("GET", "/workouts/history", "")

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body := testutil.ReadBody(t, resp)
	if !strings.Contains(body, "Workout 1") {
		t.Error("expected history to contain 'Workout 1'")
	}
	if !strings.Contains(body, "Workout 2") {
		t.Error("expected history to contain 'Workout 2'")
	}
	// In progress workout should not appear in history
	if strings.Contains(body, "In Progress") {
		t.Error("expected history to NOT contain in-progress workout")
	}
}

func TestHandleDelete(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	id, _ := workouts.Create(app.DB, "To Delete", time.Now(), nil)

	resp := app.HTMXRequest("DELETE", "/workouts/"+strconv.FormatInt(id, 10), "")

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Verify deletion
	workout, _ := workouts.GetByID(app.DB, id)
	if workout != nil {
		t.Error("expected workout to be deleted")
	}
}

func TestHandleAddExercise(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	workoutID, _ := workouts.Create(app.DB, "My Workout", time.Now(), nil)
	exerciseID, _ := exercises.Create(app.DB, "Squat")

	resp := app.HTMXRequest("POST", "/workouts/"+strconv.FormatInt(workoutID, 10)+"/exercises",
		"exercise_id="+strconv.FormatInt(exerciseID, 10))

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body := testutil.ReadBody(t, resp)
	if !strings.Contains(body, "Squat") {
		t.Error("expected response to contain exercise name")
	}

	// Verify it was added
	workout, _ := workouts.GetByID(app.DB, workoutID)
	if len(workout.Exercises) != 1 {
		t.Errorf("expected 1 exercise, got %d", len(workout.Exercises))
	}
}

func TestHandleAddExercise_FinishedWorkout(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	workoutID, _ := workouts.Create(app.DB, "Finished", time.Now(), nil)
	workouts.Finish(app.DB, workoutID)
	exerciseID, _ := exercises.Create(app.DB, "Squat")

	resp := app.HTMXRequest("POST", "/workouts/"+strconv.FormatInt(workoutID, 10)+"/exercises",
		"exercise_id="+strconv.FormatInt(exerciseID, 10))

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400 for finished workout, got %d", resp.StatusCode)
	}
}

func TestHandleRemoveExercise(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	workoutID, _ := workouts.Create(app.DB, "My Workout", time.Now(), nil)
	exerciseID, _ := exercises.Create(app.DB, "Squat")
	weID, _ := workouts.AddExercise(app.DB, workoutID, exerciseID)

	resp := app.HTMXRequest("DELETE", "/workouts/exercises/"+strconv.FormatInt(weID, 10), "")

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Verify removal
	workout, _ := workouts.GetByID(app.DB, workoutID)
	if len(workout.Exercises) != 0 {
		t.Errorf("expected 0 exercises, got %d", len(workout.Exercises))
	}
}

func TestHandleAddSet(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	workoutID, _ := workouts.Create(app.DB, "My Workout", time.Now(), nil)
	exerciseID, _ := exercises.Create(app.DB, "Squat")
	weID, _ := workouts.AddExercise(app.DB, workoutID, exerciseID)

	resp := app.HTMXRequest("POST", "/workouts/exercises/"+strconv.FormatInt(weID, 10)+"/sets",
		"reps=10&weight=135")

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body := testutil.ReadBody(t, resp)
	if !strings.Contains(body, "10") {
		t.Error("expected response to contain reps")
	}
	if !strings.Contains(body, "135") {
		t.Error("expected response to contain weight")
	}

	// Verify it was added
	we, _ := workouts.GetWorkoutExerciseByID(app.DB, weID)
	if len(we.Sets) != 1 {
		t.Errorf("expected 1 set, got %d", len(we.Sets))
	}
	if we.Sets[0].Reps != 10 {
		t.Errorf("expected 10 reps, got %d", we.Sets[0].Reps)
	}
	if we.Sets[0].Weight != 135 {
		t.Errorf("expected 135 lbs, got %.1f", we.Sets[0].Weight)
	}
}

func TestHandleUpdateSet(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	workoutID, _ := workouts.Create(app.DB, "My Workout", time.Now(), nil)
	exerciseID, _ := exercises.Create(app.DB, "Squat")
	weID, _ := workouts.AddExercise(app.DB, workoutID, exerciseID)
	setID, _ := workouts.AddSet(app.DB, weID, 10, 135)

	resp := app.HTMXRequest("PUT", "/workouts/sets/"+strconv.FormatInt(setID, 10),
		"reps=12&weight=145")

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Verify update
	set, _ := workouts.GetSetByID(app.DB, setID)
	if set.Reps != 12 {
		t.Errorf("expected 12 reps, got %d", set.Reps)
	}
	if set.Weight != 145 {
		t.Errorf("expected 145 lbs, got %.1f", set.Weight)
	}
}

func TestHandleDeleteSet(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	workoutID, _ := workouts.Create(app.DB, "My Workout", time.Now(), nil)
	exerciseID, _ := exercises.Create(app.DB, "Squat")
	weID, _ := workouts.AddExercise(app.DB, workoutID, exerciseID)
	setID, _ := workouts.AddSet(app.DB, weID, 10, 135)

	resp := app.HTMXRequest("DELETE", "/workouts/sets/"+strconv.FormatInt(setID, 10), "")

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Verify deletion
	set, _ := workouts.GetSetByID(app.DB, setID)
	if set != nil {
		t.Error("expected set to be deleted")
	}
}

func TestCreateFromTemplate(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	// Create template with exercises
	templateID, err := templates.Create(app.DB, "Push Day")
	if err != nil {
		t.Fatalf("failed to create template: %v", err)
	}
	exerciseID, err := exercises.Create(app.DB, "Bench Press")
	if err != nil {
		t.Fatalf("failed to create exercise: %v", err)
	}
	_, err = templates.AddExercise(app.DB, templateID, exerciseID, 3, 10)
	if err != nil {
		t.Fatalf("failed to add exercise to template: %v", err)
	}

	// Verify template was created correctly
	tmpl, err := templates.GetByID(app.DB, templateID)
	if err != nil {
		t.Fatalf("failed to get template: %v", err)
	}
	if len(tmpl.Exercises) != 1 {
		t.Fatalf("template should have 1 exercise, got %d", len(tmpl.Exercises))
	}

	// Start workout from template via redirect
	resp := app.Request("GET", "/workouts/new?template_id="+strconv.FormatInt(templateID, 10), "")

	// Should redirect to the new workout
	if resp.StatusCode != http.StatusSeeOther {
		body := testutil.ReadBody(t, resp)
		t.Fatalf("expected status 303, got %d, body: %s", resp.StatusCode, body)
	}

	// Verify workout was created with exercises from template
	workoutList, _ := workouts.ListInProgress(app.DB)
	if len(workoutList) != 1 {
		t.Fatalf("expected 1 workout, got %d", len(workoutList))
	}

	workout, _ := workouts.GetByID(app.DB, workoutList[0].ID)
	if len(workout.Exercises) != 1 {
		t.Errorf("expected 1 exercise from template, got %d", len(workout.Exercises))
	}
	if workout.Exercises[0].Exercise.Name != "Bench Press" {
		t.Errorf("expected exercise 'Bench Press', got '%s'", workout.Exercises[0].Exercise.Name)
	}
}

func TestLastWeight(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	exerciseID, _ := exercises.Create(app.DB, "Squat")

	// Create first workout, log a set, and finish it
	workout1ID, _ := workouts.Create(app.DB, "Workout 1", time.Now().AddDate(0, 0, -7), nil)
	we1ID, _ := workouts.AddExercise(app.DB, workout1ID, exerciseID)
	workouts.AddSet(app.DB, we1ID, 5, 225)
	workouts.Finish(app.DB, workout1ID)

	// Create second workout with same exercise
	workout2ID, _ := workouts.Create(app.DB, "Workout 2", time.Now(), nil)
	we2ID, _ := workouts.AddExercise(app.DB, workout2ID, exerciseID)

	// Get the workout exercise and check last weight
	we, _ := workouts.GetWorkoutExerciseByID(app.DB, we2ID)
	if we.LastWeight == nil {
		t.Fatal("expected last weight to be set")
	}
	if *we.LastWeight != 225 {
		t.Errorf("expected last weight 225, got %.1f", *we.LastWeight)
	}
}

func TestMultipleSetsWithPositions(t *testing.T) {
	t.Parallel()
	app := testutil.NewTestApp(t)
	defer app.Close()

	workoutID, _ := workouts.Create(app.DB, "My Workout", time.Now(), nil)
	exerciseID, _ := exercises.Create(app.DB, "Squat")
	weID, _ := workouts.AddExercise(app.DB, workoutID, exerciseID)

	// Add multiple sets
	workouts.AddSet(app.DB, weID, 5, 135)
	workouts.AddSet(app.DB, weID, 5, 185)
	workouts.AddSet(app.DB, weID, 5, 225)

	// Verify positions
	we, _ := workouts.GetWorkoutExerciseByID(app.DB, weID)
	if len(we.Sets) != 3 {
		t.Fatalf("expected 3 sets, got %d", len(we.Sets))
	}

	if we.Sets[0].Position != 1 {
		t.Errorf("expected position 1, got %d", we.Sets[0].Position)
	}
	if we.Sets[1].Position != 2 {
		t.Errorf("expected position 2, got %d", we.Sets[1].Position)
	}
	if we.Sets[2].Position != 3 {
		t.Errorf("expected position 3, got %d", we.Sets[2].Position)
	}
}
