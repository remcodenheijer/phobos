# Fitness Workout Tracker — Specification

## Overview

A personal application for logging strength training workouts. Users define their own exercises, create reusable workout templates, group templates into routines, and record workouts in real time or after completion.

---

## Core Entities

### Exercise

A named movement that can be included in workouts and templates.

| Field | Description |
|-------|-------------|
| Name | The exercise name (e.g., "Bench Press," "Deadlift") |

### Workout

A single training session consisting of one or more exercises performed with sets.

| Field | Description |
|-------|-------------|
| Name | A label for the workout (e.g., "Monday Push Day") |
| Date | When the workout occurred |
| Notes | Optional free-text field |
| Status | Either **In Progress** or **Finished** |
| Exercises | Ordered list of exercises performed, each with its logged sets |

### Logged Set

An individual set performed for an exercise within a workout.

| Field | Description |
|-------|-------------|
| Reps | Number of repetitions |
| Weight | Weight used |

### Workout Template

A reusable blueprint for creating new workouts.

| Field | Description |
|-------|-------------|
| Name | A label for the template (e.g., "Leg Day") |
| Exercises | Ordered list of exercises, each with target sets and reps |

### Routine

A collection of workout templates that form a training program.

| Field | Description |
|-------|-------------|
| Name | A label for the routine (e.g., "Push/Pull/Legs Program") |
| Templates | Ordered list of workout templates |

---

## Functional Requirements

### Exercise Management

- Create a new exercise by name
- View list of all exercises
- Delete an exercise

### Workout Logging

- Start a new workout (status: In Progress), either blank or from a template
- Add exercises from the exercise list to the current workout
- For each exercise, display the most recent weight used for that exercise (from workout history)
- Log individual sets (reps and weight) for each exercise
- Add or edit the workout name, date, and notes while in progress
- Mark the workout as Finished
- Once finished, the workout becomes read-only

### Workout History

- View a chronological list of all finished workouts (newest first)
- View details of any past workout
- Delete a past workout

### Template Management

- Create a template by defining exercises with target sets and reps
- View list of all templates
- Edit a template
- Delete a template
- Start a new workout from a template (pre-populates exercises and targets)

### Routine Management

- Create a routine by selecting and ordering workout templates
- View list of all routines
- Edit a routine (add, remove, or reorder templates)
- Delete a routine
- Browse templates within a routine and start a workout from any of them

---

## Out of Scope

- Multiple users or accounts
- Exercise categorization
- Progress tracking, analytics, or goals
- Workout editing after completion
