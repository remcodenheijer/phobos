package workouts

import (
	"strconv"
	"time"

	"phobos/internal/features/exercises"
	"phobos/internal/shared/htmx"
	"phobos/internal/shared/middleware"

	"github.com/gofiber/fiber/v2"
)

// HandleList displays active (in-progress) workouts
func HandleList(c *fiber.Ctx) error {
	db := middleware.GetDB(c)

	workouts, err := ListInProgress(db)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to load workouts")
	}

	return htmx.Render(c, WorkoutsPage(workouts))
}

// HandleHistory displays finished workouts
func HandleHistory(c *fiber.Ctx) error {
	db := middleware.GetDB(c)

	workouts, err := ListFinished(db)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to load history")
	}

	return htmx.Render(c, HistoryPage(workouts))
}

// HandleNew displays the new workout form
func HandleNew(c *fiber.Ctx) error {
	db := middleware.GetDB(c)

	// Check if starting from a template
	templateIDStr := c.Query("template_id")
	if templateIDStr != "" {
		templateID, err := strconv.ParseInt(templateIDStr, 10, 64)
		if err == nil {
			// Create workout from template
			id, err := CreateFromTemplate(db, "Workout from Template", time.Now(), templateID)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).SendString("Failed to create workout: " + err.Error())
			}
			return c.Redirect("/workouts/"+strconv.FormatInt(id, 10), fiber.StatusSeeOther)
		}
	}

	allExercises, err := exercises.ListAll(db)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to load exercises")
	}

	return htmx.Render(c, NewWorkoutPage(allExercises))
}

// HandleCreate creates a new workout
func HandleCreate(c *fiber.Ctx) error {
	db := middleware.GetDB(c)

	name := c.FormValue("name")
	if name == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Name is required")
	}

	dateStr := c.FormValue("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid date")
	}

	id, err := Create(db, name, date, nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to create workout")
	}

	return c.Redirect("/workouts/"+strconv.FormatInt(id, 10), fiber.StatusSeeOther)
}

// HandleShow displays a single workout
func HandleShow(c *fiber.Ctx) error {
	db := middleware.GetDB(c)

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
	}

	workout, err := GetByID(db, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to load workout")
	}
	if workout == nil {
		return c.Status(fiber.StatusNotFound).SendString("Workout not found")
	}

	// Load template targets if applicable
	if workout.TemplateID != nil {
		for i := range workout.Exercises {
			targetSets, targetReps, _ := GetTemplateTargets(db, *workout.TemplateID, workout.Exercises[i].ExerciseID)
			workout.Exercises[i].TargetSets = targetSets
			workout.Exercises[i].TargetReps = targetReps
		}
	}

	allExercises, err := exercises.ListAll(db)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to load exercises")
	}

	return htmx.Render(c, WorkoutDetailPage(workout, allExercises))
}

// HandleUpdate modifies a workout
func HandleUpdate(c *fiber.Ctx) error {
	db := middleware.GetDB(c)

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
	}

	name := c.FormValue("name")
	if name == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Name is required")
	}

	dateStr := c.FormValue("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid date")
	}

	notes := c.FormValue("notes")

	if err := Update(db, id, name, date, notes); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to update workout")
	}

	htmx.Trigger(c, "workoutUpdated")
	return c.SendString("")
}

// HandleDelete removes a workout
func HandleDelete(c *fiber.Ctx) error {
	db := middleware.GetDB(c)

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
	}

	if err := Delete(db, id); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to delete workout")
	}

	return c.SendString("")
}

// HandleFinish marks a workout as complete
func HandleFinish(c *fiber.Ctx) error {
	db := middleware.GetDB(c)

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
	}

	if err := Finish(db, id); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to finish workout")
	}

	return htmx.Redirect(c, "/workouts/history")
}

// HandleAddExercise adds an exercise to a workout
func HandleAddExercise(c *fiber.Ctx) error {
	db := middleware.GetDB(c)

	workoutID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid workout ID")
	}

	// Check workout status
	workout, err := GetByID(db, workoutID)
	if err != nil || workout == nil {
		return c.Status(fiber.StatusNotFound).SendString("Workout not found")
	}
	if workout.IsFinished() {
		return c.Status(fiber.StatusBadRequest).SendString("Cannot modify finished workout")
	}

	exerciseID, err := strconv.ParseInt(c.FormValue("exercise_id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid exercise ID")
	}

	id, err := AddExercise(db, workoutID, exerciseID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to add exercise")
	}

	we, err := GetWorkoutExerciseByID(db, id)
	if err != nil || we == nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to load exercise")
	}

	return htmx.Render(c, WorkoutExerciseCard(*we, false, workout.TemplateID))
}

// HandleRemoveExercise removes an exercise from a workout
func HandleRemoveExercise(c *fiber.Ctx) error {
	db := middleware.GetDB(c)

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
	}

	// Check workout status
	we, err := GetWorkoutExerciseByID(db, id)
	if err != nil || we == nil {
		return c.Status(fiber.StatusNotFound).SendString("Exercise not found")
	}

	workout, err := GetByID(db, we.WorkoutID)
	if err != nil || workout == nil || workout.IsFinished() {
		return c.Status(fiber.StatusBadRequest).SendString("Cannot modify finished workout")
	}

	if err := RemoveExercise(db, id); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to remove exercise")
	}

	return c.SendString("")
}

// HandleAddSet adds a set to a workout exercise
func HandleAddSet(c *fiber.Ctx) error {
	db := middleware.GetDB(c)

	workoutExerciseID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid exercise ID")
	}

	// Check workout status
	we, err := GetWorkoutExerciseByID(db, workoutExerciseID)
	if err != nil || we == nil {
		return c.Status(fiber.StatusNotFound).SendString("Exercise not found")
	}

	workout, err := GetByID(db, we.WorkoutID)
	if err != nil || workout == nil || workout.IsFinished() {
		return c.Status(fiber.StatusBadRequest).SendString("Cannot modify finished workout")
	}

	reps, err := strconv.Atoi(c.FormValue("reps"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid reps")
	}

	weight, err := strconv.ParseFloat(c.FormValue("weight"), 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid weight")
	}

	id, err := AddSet(db, workoutExerciseID, reps, weight)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to add set")
	}

	set, err := GetSetByID(db, id)
	if err != nil || set == nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to load set")
	}

	return htmx.Render(c, SetRow(*set, false))
}

// HandleUpdateSet modifies an existing set
func HandleUpdateSet(c *fiber.Ctx) error {
	db := middleware.GetDB(c)

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
	}

	reps, err := strconv.Atoi(c.FormValue("reps"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid reps")
	}

	weight, err := strconv.ParseFloat(c.FormValue("weight"), 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid weight")
	}

	if err := UpdateSet(db, id, reps, weight); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to update set")
	}

	return c.SendString("")
}

// HandleDeleteSet removes a set
func HandleDeleteSet(c *fiber.Ctx) error {
	db := middleware.GetDB(c)

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
	}

	if err := DeleteSet(db, id); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to delete set")
	}

	return c.SendString("")
}
