# Tasks: Context Window Protections

1. **Task 1: Define Summary Data Structure**
   - **File:** `internal/output/output.go` (or `internal/output/summary.go`)
   - **Description:** Create the `SummaryOutput` struct containing `TotalItems`, `ReturnedItems`, and `HasMore`.

2. **Task 2: Add Global Flags**
   - **File:** `internal/cmd/root.go`
   - **Description:** Add `--fields` (string slice) and `--summary` (boolean) to the root Cobra command so they are available globally. Bind them to Viper.

3. **Task 3: Implement Client-Side Field Filtering**
   - **File:** `internal/output/output.go`
   - **Description:** Update the JSON printing logic. If `--fields` is provided, marshal the original object to JSON, unmarshal it into a `map[string]interface{}` (or a slice of maps), remove all keys not present in the `--fields` slice, and re-marshal it before printing.

4. **Task 4: Implement Summary Output Logic**
   - **File:** `internal/output/output.go`
   - **Description:** Update the JSON printing logic. If `--summary` is true, intercept the data (which should be a slice), count its length, and output the `SummaryOutput` struct instead of the raw data array.

5. **Task 5: Enforce Safe Default Pagination**
   - **File:** `internal/cmd/root.go` (or wherever standard list flags are added)
   - **Description:** When setting up standard pagination flags, check if the output format is JSON. If it is JSON and the user did not explicitly provide a `--limit` flag, set the default limit to `50`.

6. **Task 6: Write Tests**
   - **File:** `internal/output/output_test.go`
   - **Description:** Write unit tests for the field filtering logic and the summary logic to ensure they correctly manipulate the output JSON without altering the underlying Go structs.
