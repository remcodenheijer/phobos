# CRUD Application Architecture

## Overview

A hypermedia-driven web application using Go, Templ, and TemplUI. The server owns all application state and returns HTML responses. HTMX enables partial page updates without full reloads.

## Tech Stack

| Layer | Technology |
|-------|------------|
| Language | Go |
| Router | Fiber |
| Templating | Templ |
| UI Components | TemplUI |
| Hypermedia | HTMX |
| Database | SQLite |
| Migrations | Goose or golang-migrate |

## Project Structure (Vertical Slice)

```
├── cmd/
│   └── server/
│       └── main.go           # Entry point, wiring
├── internal/
│   ├── features/
│   │   └── items/            # One folder per feature
│   │       ├── list.go       # Handler + query + list.templ
│   │       ├── create.go     # Handler + command + form.templ
│   │       ├── edit.go       # Handler + update logic
│   │       ├── delete.go     # Handler + delete logic
│   │       └── templates.templ
│   ├── shared/
│   │   ├── db/               # Database connection
│   │   └── middleware/       # Auth, logging, etc.
│   └── ui/
│       ├── layouts/          # Base page layouts
│       └── components/       # Shared TemplUI components
├── static/
│   └── vendor/               # Vendored JS (HTMX)
├── vendor/                   # Vendored Go modules
└── migrations/               # Database migrations
```

## Request Flow

```
Browser → Fiber → Handler → Database
                     ↓
                  Templ renders HTML
                     ↓
                  HTML response (full page or fragment)
```

## Route Design

| Method | Path | Returns | Purpose |
|--------|------|---------|---------|
| GET | /items | Page | List all items |
| GET | /items/new | Fragment | New item form |
| POST | /items | Fragment + HX-Trigger | Create item |
| GET | /items/{id} | Page | View single item |
| GET | /items/{id}/edit | Fragment | Edit form |
| PUT | /items/{id} | Fragment | Update item |
| DELETE | /items/{id} | Empty (200) | Delete item |

## Key Patterns

**Fragment vs Full Page**: Handlers check the `HX-Request` header via Fiber's `c.Get()`. Return a fragment for HTMX requests, a full page otherwise.

**Out-of-Band Updates**: Use `hx-swap-oob` to update multiple page regions from a single response (e.g., update a list item and a toast notification).

**Redirects**: For POST requests from non-HTMX clients, redirect with 303. For HTMX, return the updated fragment directly or use `HX-Redirect`.

**Error Handling**: Return error fragments that swap into place, or use `HX-Retarget` to display errors in a dedicated region.

## State Management

All state lives on the server. The client holds no application state—only the current HTML document. Navigation and actions are driven entirely by hypermedia controls (links and forms) present in the response.

## Design Principles

### Locality of Behaviour (LoB)

> "The behaviour of a unit of code should be as obvious as possible by looking only at that unit of code."

Behaviour should be visible where it happens. With HTMX, the button declares its own behaviour inline rather than relying on JavaScript in a separate file:

```html
<button hx-delete="/items/42" hx-target="closest tr">Delete</button>
```

LoB trades off against DRY (Don't Repeat Yourself) and SoC (Separation of Concerns). Accept some repetition when it keeps behaviour local and obvious. The closer behaviour is to the element it affects, the easier the code is to understand and maintain.

**Reference**: https://htmx.org/essays/locality-of-behaviour/

### Vendoring

Copy dependencies directly into your project rather than relying on package managers at build/deploy time.

```
static/
  vendor/
    htmx-2.0.4.min.js
```

Benefits: Your entire project is in source control. No external systems needed to build. Dependencies are visible and auditable. You can read, debug, and modify vendored code. No surprises at deployment time.

Go supports vendoring natively with `go mod vendor`. Frontend dependencies like HTMX are small enough to simply download and commit.

**Reference**: https://htmx.org/essays/vendoring/

### Mobile First

Design for the smallest screen first, then progressively enhance for larger viewports.

1. Start with base styles for mobile (no media query)
2. Add `min-width` media queries for larger screens
3. Let content dictate breakpoints, not device sizes

```css
/* Base: mobile */
.container { padding: 1rem; }

/* Larger screens */
@media (min-width: 40rem) {
  .container { padding: 2rem; max-width: 60rem; }
}
```

Benefits: Forces focus on core content and functionality. Mobile styles are simpler, so the baseline CSS is smaller. Users on slow connections get the lightest experience first.

TemplUI components are responsive by default. Use Tailwind's responsive prefixes (`sm:`, `md:`, `lg:`) to adjust layouts at different breakpoints.

### Vertical Slice Architecture

Organize code by feature, not by technical layer. Each feature contains everything it needs—handler, templates, queries—in one place.

**Traditional Layered**:
```
handlers/
  item_handler.go
services/
  item_service.go
repositories/
  item_repository.go
```

**Vertical Slice**:
```
features/
  items/
    list.go         # GET /items - handler, query, template
    create.go       # POST /items - handler, command, template
    edit.go         # GET/PUT /items/{id}/edit
    delete.go       # DELETE /items/{id}
```

Benefits: Changes are localized to one file. Adding a feature means adding code, not modifying shared layers. Each slice can make its own implementation choices.

**References**:
- https://github.com/sebajax/go-vertical-slice-architecture
- https://github.com/mehdihadeli/go-vertical-slice-template
- https://www.bensampica.com/post/verticalslice/

### Test-Driven Development (TDD)

Write tests before implementation. Red → Green → Refactor.

1. **Red**: Write a failing test that defines expected behaviour
2. **Green**: Write the minimum code to make the test pass
3. **Refactor**: Clean up while keeping tests green

Vertical slices simplify testing—each feature is self-contained with clear boundaries. Test the slice end-to-end rather than mocking layers.

**Testing strategy**:

| Type | Scope | Tools |
|------|-------|-------|
| Unit | Pure functions, domain logic | `testing`, `testify` |
| Integration | Handler + database | `testing`, SQLite in-memory |
| UI / E2E | Full browser interaction | Playwright |

Use SQLite's in-memory mode (`:memory:`) for fast, isolated integration tests. Each test gets a fresh database.

```go
db, _ := sql.Open("sqlite", ":memory:")
```

Prefer integration tests over unit tests with mocks. Testing real database queries catches more bugs than mocking repository interfaces.

**UI testing with Playwright**: Playwright automates real browsers (Chromium, Firefox, WebKit) and works well with HTMX apps. Run Playwright tests in JavaScript/TypeScript alongside your Go backend.

```
tests/
  e2e/
    items.spec.ts
    playwright.config.ts
```

UI tests verify the full loop: browser → HTMX request → Fiber handler → database → HTML response → DOM update.

---

## TemplUI Component Reference

TemplUI provides accessible, composable components for Templ. Use the CLI to add components to your project.

**Form & Input**: Button, Checkbox, Input, Select Box, Textarea, Form, Label, Radio, Switch, Slider, Date Picker, Time Picker, Tags Input, Rating, Input OTP, Calendar

**Layout & Navigation**: Accordion, Breadcrumb, Pagination, Separator, Sidebar, Tabs

**Overlays & Dialogs**: Dialog, Dropdown, Popover, Sheet, Tooltip

**Feedback & Status**: Alert, Badge, Progress, Skeleton, Toast

**Display & Media**: Aspect Ratio, Avatar, Card, Carousel, Charts, Table

**Misc**: Code, Collapsible, Copy Button, Icon

**Reference**: https://templui.io/docs/components

---

## Dependencies

- fiber (web framework)
- templ (template compiler)
- templui (component library)
- htmx (vendored in static/)
- modernc.org/sqlite or mattn/go-sqlite3

Use `go mod vendor` to vendor Go dependencies into a `/vendor` directory.
