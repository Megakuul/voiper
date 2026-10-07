<script>
  import { ui } from "./ui.js";
  import Phone from "./Phone/Phone.svelte";
  import Icon from "./components/Icon.svelte";
  import * as api from "../wailsjs/go/app/App.js";

  let { snapshot, account = $bindable(""), target = $bindable(""), run, busy, manageContacts } = $props();
  let recent = $state([]);
  let loading = $state(false);
  let error = $state("");
  const callIDs = $derived(snapshot.Calls.map((call) => call.ID).join(","));

  $effect(() => {
    // Reload when a call starts/ends, not for every live media statistics update.
    callIDs;
    let cancelled = false;
    loading = true;
    api.HistoryPage("", "", 0, 50).then((rows) => {
      if (cancelled) return;
      const seen = new Set();
      recent = rows.filter((row) => {
        const key = `${row.Account}:${row.Remote}`;
        if (seen.has(key)) return false;
        seen.add(key);
        return true;
      }).slice(0, 8);
      error = "";
    }).catch((e) => { if (!cancelled) error = String(e); })
      .finally(() => { if (!cancelled) loading = false; });
    return () => { cancelled = true; };
  });
</script>

<div class="
  grid grid-cols-[minmax(0,_1fr)_290px] rounded-[9px] bg-surface [box-shadow:0_4px_12px_#0003] border
  border-[#4c5b69] [&_>_*]:min-w-0 max-[1100px]:grid-cols-[minmax(0,_1fr)_245px]
  max-[850px]:grid-cols-[minmax(0,_1fr)]
">
  <div class="min-w-0 p-8 max-[1100px]:p-6 max-[850px]:p-5">
    <Phone {snapshot} bind:account bind:target {run} {busy} workspace />
  </div>
  <aside class="
    px-[1.1rem] py-7 min-w-0 border-l border-l-border bg-surface-secondary rounded-[0_9px_9px_0]
    max-[850px]:border-l-0 max-[850px]:border-t max-[850px]:border-t-border
    max-[850px]:rounded-[0_0_9px_9px]
  " aria-label="Recently called people">
    <div class="flex items-center gap-[0.6rem] mb-[0.4rem] [&_h2]:m-0 [&_h2]:text-base"><Icon name="history" /><h2 class={ui.h2}>Recent people</h2></div>
    <p class="m-0 leading-[1.6] text-muted text-[0.78rem]">Pick up where you left off.</p>
    {#if error}<p class="m-0 leading-[1.6] text-[#f9a8b9]" role="alert">{error}</p>{/if}
    <div class="mx-0 my-5" aria-busy={loading}>
      {#each recent as person (person.ID)}
        <button class={[ui.control, `
          px-[0.3rem] py-[0.85rem] border-t-0 border-t-[#536271] border-r-0 border-r-[#536271] border-b
          border-b-border border-l-0 border-l-[#536271] rounded-[6px] bg-transparent text-inherit min-h-[38px]
          shadow-none font-medium flex items-center gap-[0.65rem] w-full text-left
          [&:hover:not(:disabled)]:bg-[#435362] [&:hover:not(:disabled)]:border-[#727e8d]
        `]} title={`Select ${person.Remote} to call`}
          onclick={() => { target = person.Remote; account = person.Account; }}>
          <span class="grid place-items-center w-[32px] h-[32px] shrink-0 rounded-[6px] bg-accent-surface text-accent"><Icon name={person.Direction === "incoming" ? "incoming" : "outgoing"} size={18} /></span>
          <span class="
            flex-1 min-w-0 [&_strong]:block [&_strong]:text-[0.85rem] [&_strong]:wrap-anywhere [&_>_span]:block
            [&_>_span]:text-muted [&_>_span]:text-[0.7rem] [&_>_span]:mt-[0.2rem]
          "><strong>{person.Remote}</strong><span>{person.Account} · {new Date(person.Started).toLocaleDateString()}</span></span>
          <Icon name="phone" size={16} />
        </button>
      {:else}<div class="px-2 py-8 grid justify-items-center gap-4 text-muted text-center text-[0.85rem]"><Icon name="history" size={28} /><p class="m-0 leading-[1.6]">{loading ? "Loading recent calls…" : "Your recent calls will appear here."}</p></div>{/each}
    </div>
    <button class={[ui.control, `
      px-4 py-[0.65rem] rounded-[6px] bg-transparent text-inherit min-h-[38px] shadow-none font-medium flex
      w-full items-center justify-center gap-[0.6rem] border border-[#536271]
      [&:hover:not(:disabled)]:bg-[#435362] [&:hover:not(:disabled)]:border-[#727e8d]
    `]} onclick={manageContacts}><Icon name="contacts" size={18} /> All contacts <Icon name="arrow" size={16} /></button>
  </aside>
</div>
