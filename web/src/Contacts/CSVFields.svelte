<script>
  import { untrack } from "svelte";
  import * as api from "../../wailsjs/go/app/App.js";

  let { text, onMapped, disabled = false } = $props();
  let columns = $state([]);
  let mapping = $state({
    Name: "",
    AdditionalName: "",
    Address: "",
    Notes: "",
    Numbers: [],
  });
  let busy = $state(false);
  let error = $state("");
  let notice = $state("");
  let revision = 0;

  $effect(() => {
    const source = text;
    const current = ++revision;
    columns = [];
    error = "";
    notice = "";
    busy = false;
    untrack(() => onMapped(""));
    if (source) {
      busy = true;
      api.CSVColumns(source).then(
        (headers) => {
          if (current !== revision) return;
          columns = headers;
          const canonical = (name) =>
            headers.find((header) => header.toLowerCase() === name) || "";
          mapping = {
            Name: canonical("name"),
            AdditionalName: "",
            Address: canonical("address"),
            Notes: canonical("notes"),
            Numbers: [],
          };
          busy = false;
        },
        (failure) => {
          if (current !== revision) return;
          error = String(failure);
          busy = false;
        },
      );
    }
    return () => {
      revision++;
    };
  });

  function invalidate() {
    revision++;
    notice = "";
    error = "";
    onMapped("");
  }

  function addNumber() {
    mapping.Numbers = [...mapping.Numbers, { Column: "", Label: "" }];
    invalidate();
  }

  function removeNumber(index) {
    mapping.Numbers = mapping.Numbers.filter(
      (_, position) => position !== index,
    );
    invalidate();
  }

  async function apply() {
    const current = revision;
    busy = true;
    error = "";
    notice = "";
    try {
      const canonical = await api.MapContactCSV(text, $state.snapshot(mapping));
      if (current !== revision) return;
      await onMapped(canonical);
      if (current === revision)
        notice = "Columns mapped. Review the import preview before saving.";
    } catch (failure) {
      if (current === revision) error = String(failure);
    } finally {
      if (current === revision) busy = false;
    }
  }
</script>

<fieldset disabled={busy || disabled}>
  <legend>CSV column mapping</legend>
  <p class="muted small">
    Use a comma-separated CSV with a header row. Choose which columns contain
    names, numbers and notes. Mapping keeps one contact per source row.
  </p>
  {#if busy}<p role="status">Processing CSV…</p>{/if}
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  {#if notice}<p role="status">{notice}</p>{/if}
  {#if columns.length}
    <div class="fields">
      <label>
        Name or first name
        <select bind:value={mapping.Name} onchange={invalidate}>
          <option value="">No name column</option>
          {#each columns as column}<option value={column}>{column}</option
            >{/each}
        </select>
      </label>
      <label>
        Last name to append
        <select bind:value={mapping.AdditionalName} onchange={invalidate}>
          <option value="">None</option>
          {#each columns as column}<option value={column}>{column}</option
            >{/each}
        </select>
      </label>
      <label>
        Preferred primary number
        <select bind:value={mapping.Address} onchange={invalidate}>
          <option value="">Use the first additional number</option>
          {#each columns as column}<option value={column}>{column}</option
            >{/each}
        </select>
      </label>
      <label>
        Notes
        <select bind:value={mapping.Notes} onchange={invalidate}>
          <option value="">None</option>
          {#each columns as column}<option value={column}>{column}</option
            >{/each}
        </select>
      </label>
    </div>
    {#each mapping.Numbers as number, index}
      <div class="number">
        <label>
          Additional number {index + 1}
          <select bind:value={number.Column} onchange={invalidate}>
            <option value="">Choose a column</option>
            {#each columns as column}<option value={column}>{column}</option
              >{/each}
          </select>
        </label>
        <label>
          Label
          <input
            bind:value={number.Label}
            oninput={invalidate}
            maxlength="80"
            placeholder="Work, mobile, home…"
          />
        </label>
        <button type="button" onclick={() => removeNumber(index)}>
          Remove number {index + 1}
        </button>
      </div>
    {/each}
    <p class="muted small">
      The first nonempty number becomes primary. Empty and repeated numbers are
      skipped; rows with no number appear as errors in the preview. Labels apply
      to additional numbers. Missing names use the primary number.
    </p>
    <div class="actions">
      <button
        type="button"
        disabled={mapping.Numbers.length >= 20}
        onclick={addNumber}>Add number column</button
      >
      <button
        type="button"
        disabled={!mapping.Address &&
          !mapping.Numbers.some((number) => number.Column)}
        onclick={apply}>Apply mapping</button
      >
    </div>
  {/if}
</fieldset>

<style>
  fieldset {
    min-width: 0;
    border: 1px solid var(--border, #64748b);
    border-radius: 0.5rem;
    padding: 1rem;
  }
  .fields,
  .number {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(min(12rem, 100%), 1fr));
    gap: 0.75rem;
    margin-bottom: 0.75rem;
    align-items: end;
  }
  label {
    display: flex;
    flex-direction: column;
    gap: 0.3rem;
    min-width: 0;
  }
  select,
  input {
    width: 100%;
    min-width: 0;
  }
  .actions {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
  }
</style>
