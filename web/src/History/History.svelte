<script>
  import { onMount } from "svelte";
  import * as api from "../../wailsjs/go/app/App.js";
  let { account, run, busy, onDial, onSave } = $props();
  let history = $state([]),
    search = $state(""),
    allAccounts = $state(true),
    hasMore = $state(false),
    confirmClear = $state(false);
  async function load(more = false) {
    const before = more && history.length ? history.at(-1).ID : 0;
    const rows = await api.HistoryPage(
      allAccounts ? "" : account,
      search,
      before,
      100,
    );
    history = more ? [...history, ...rows] : rows;
    hasMore = rows.length === 100;
  }
  onMount(() => {
    run(() => load());
  });
</script>

<section class="panel">
  <div class="row spread">
    <h2>Recent calls</h2>
    <button onclick={() => (confirmClear = !confirmClear)}>Clear history</button
    >
  </div>
  {#if confirmClear}<div class="banner">
      <span>Remove all call history on this device?</span><button
        class="danger"
        onclick={() =>
          run(async () => {
            await api.ClearHistory();
            confirmClear = false;
            await load();
          })}>Clear all</button
      ><button onclick={() => (confirmClear = false)}>Cancel</button>
    </div>{/if}
  <form
    class="row wrap"
    onsubmit={(e) => {
      e.preventDefault();
      run(() => load());
    }}
  >
    <input
      aria-label="Search history"
      placeholder="Number or SIP address"
      bind:value={search}
    /><label class="toggle"
      ><input type="checkbox" bind:checked={allAccounts} /> All accounts</label
    ><button>Search</button>
  </form>
  <div class="list">
    {#each history as h (h.ID)}<div class="list-item">
        <div class="grow">
          <strong>{h.Remote}</strong><span class="muted"
            >{h.Direction} · {h.Status} · {h.Account}</span
          ><span class="muted small"
            >{new Date(h.Started).toLocaleString()} · {Math.max(
              0,
              Math.round((new Date(h.Ended) - new Date(h.Started)) / 1000),
            )} seconds</span
          >
        </div>
        <button onclick={() => onDial(h)}>Call again</button><button
          onclick={() => onSave(h)}>Save contact</button
        ><button
          class="quiet"
          disabled={busy}
          onclick={() =>
            run(async () => {
              await api.DeleteHistory(h.ID);
              history = history.filter((c) => c.ID !== h.ID);
            })}>Delete</button
        >
      </div>{:else}<p class="empty">No calls match this view.</p>{/each}
  </div>
  {#if hasMore}<button disabled={busy} onclick={() => run(() => load(true))}
      >Load earlier calls</button
    >{/if}
</section>
