# Season Visibility in Link Wizard

## Problem

When linking a new season of an anime in the Link Wizard (Step 3), users have no visibility into what seasons already exist for a show. The existing shows list only displays show names with mini posters, and the season number is a blind numeric input. Users must remember or look up which seasons are already linked.

## Solution

**Inline season pills + auto-fill** in the Link Wizard Step 3.

### Changes to the existing shows list

Each show row in the "Or select existing show" list gains season pills rendered below the show name. These are small badges showing `S1 · 24ep`, `S2 · 12ep`, etc.

**Data source:** `library.shows` already contains `seasons: Season[]` with `number` and `episodes` per season. Currently, `existingShows` maps shows to just their names via `.map(s => s.name)`, which feeds into a shared `existingItems` computed (union of shows/movies as `string[]`).

**Implementation approach:** Rather than changing `existingItems` (which is shared with movies and assumes `string[]`), the series template will use its own computed properties that return full `Show` objects. This avoids a type mismatch with the movies path. Specifically:
- New `filteredExistingShows` computed returns `Show[]` filtered by `showSearch`
- The series template uses `filteredExistingShows` directly instead of `filteredExistingItems`
- The movies template continues to use `filteredExistingItems` (strings) as-is

**Rendering:**
- Pills appear in a flex-wrap row beneath each show name
- Each pill shows season number and episode count
- Styled as small outline badges consistent with the app's existing Badge component
- Shows with no seasons yet display no pills (just the name, same as today)

### Changes to the season number input

When a user selects an existing show:
- The season number auto-fills to `max(existing season numbers) + 1`
- A subtle text hint "(next available)" appears beside the input
- Users can still manually override the number
- If the user manually enters an existing season number, a warning appears: "Season X already exists (Y episodes)"

If the show has no existing seasons, the default remains `1` (current behavior).

### What doesn't change

- The rest of Step 3 (name input, poster preview, suggested name logic)
- No new API calls — all season data is already fetched via `library.fetchShows()`
- No backend changes
- Steps 1, 2, 4, 5, 6 of the wizard are unaffected
- Library view and Downloads view are unaffected

## Files to modify

- `frontend/src/views/LinkWizardView.vue` — the only file that needs changes:
  - Add `filteredExistingShows` computed returning `Show[]` (for series template)
  - Update `selectExistingShow()` to accept a `Show` and auto-fill season number
  - Add `selectedShow` ref to track the currently selected show (for duplicate season warning)
  - Split the existing shows template: series uses `filteredExistingShows` with season pills, movies continues using `filteredExistingItems`
  - Add season pills to each show row
  - Add "next available" hint beside season input
  - Add duplicate season warning when user enters an existing season number
