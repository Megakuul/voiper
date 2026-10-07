<script>
  import { ui } from "../ui.js";
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

<section class="
  panel p-6 bg-surface shadow-panel border-t-[#526271] border-r-border border-b-border border-l-border
  rounded-[6px] min-w-0 border max-[700px]:p-[1.15rem] [&_>_p]:mb-4 [&_>_button_+_button]:mt-4
  [&_form_+_details_button]:mt-2 [&_form_+_details_button]:mr-2 [&_form_+_details_button]:mb-0
  [&_form_+_details_button]:ml-0 [&_>_h2:first-child]:pb-4 [&_>_h2:first-child]:border-b
  [&_>_h2:first-child]:border-b-border [&_>_.row_+_.row]:mt-4
">
  <div class="
    row flex items-center gap-[0.7rem] justify-between max-[700px]:flex-wrap [&_>_button]:shrink-0
    max-[700px]:[&_>_.grow]:basis-[180px]
  ">
    <h2 class={ui.h2}>Recent calls</h2>
    <button class={ui.button} onclick={() => (confirmClear = !confirmClear)}>Clear history</button
    >
  </div>
  {#if confirmClear}<div class="
    px-4 py-3 bg-accent-surface rounded-[6px] flex justify-between items-center mb-4 border
    border-brand-border [&_button]:px-2 [&_button]:py-[0.2rem] [&_button]:bg-transparent
    [&_button]:border-0
  ">
      <span>Remove all call history on this device?</span><button
        class={[ui.dangerButton, "danger"]}
        onclick={() =>
          run(async () => {
            await api.ClearHistory();
            confirmClear = false;
            await load();
          })}>Clear all</button
      ><button class={ui.button} onclick={() => (confirmClear = false)}>Cancel</button>
    </div>{/if}
  <form
    class="
      row flex items-center gap-[0.7rem] flex-wrap max-[700px]:flex-wrap [&_>_button]:shrink-0
      max-[700px]:[&_>_.grow]:basis-[180px]
    "
    onsubmit={(e) => {
      e.preventDefault();
      run(() => load());
    }}
  >
    <input class={[ui.input, "px-[0.8rem] py-[0.65rem]"]}
      aria-label="Search history"
      placeholder="Number or SIP address"
      bind:value={search}
    /><label class="flex flex-row gap-[0.65rem] text-[0.9rem] text-[#c8cdd5] items-center min-h-[32px] cursor-pointer"
      ><input class={ui.checkbox} type="checkbox" bind:checked={allAccounts} /> All accounts</label
    ><button class={ui.button}>Search</button>
  </form>
  <div class="mt-2">
    {#each history as h (h.ID)}<div class="
      px-0 py-4 gap-[0.7rem] items-center border-b border-b-border flex-wrap [&_strong]:block
      [&_strong]:wrap-anywhere [&_strong]:mb-[0.3rem] [&_span]:block [&_span]:wrap-anywhere
      [&_button]:text-[0.8rem]
    ">
        <div class="grow flex-1 min-w-0">
          <strong>{h.Remote}</strong><span class="text-muted"
            >{h.Direction} · {h.Status} · {h.Account}</span
          ><span class="text-muted text-[0.78rem]"
            >{new Date(h.Started).toLocaleString()} · {Math.max(
              0,
              Math.round((new Date(h.Ended) - new Date(h.Started)) / 1000),
            )} seconds</span
          >
        </div>
        <button class={ui.button} onclick={() => onDial(h)}>Call again</button><button class={ui.button}
          onclick={() => onSave(h)}>Save contact</button
        ><button
          class={[ui.quietButton, "quiet"]}
          disabled={busy}
          onclick={() =>
            run(async () => {
              await api.DeleteHistory(h.ID);
              history = history.filter((c) => c.ID !== h.ID);
            })}>Delete</button
        >
      </div>{:else}<p class="px-0 py-8 m-0 leading-[1.6] text-muted text-center">No calls match this view.</p>{/each}
  </div>
  {#if hasMore}<button class={ui.button} disabled={busy} onclick={() => run(() => load(true))}
      >Load earlier calls</button
    >{/if}
</section>
