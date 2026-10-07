<script>
  import { ui } from "../ui.js";
  import * as api from "../../wailsjs/go/app/App.js";
  let { snapshot, logs, run, busy } = $props();
</script>

<section class="
  panel live-diagnostics p-6 bg-surface shadow-panel border-t-[#526271] border-r-border border-b-border
  border-l-border rounded-[6px] min-w-0 border max-[700px]:p-[1.15rem] [&_>_p]:mb-4
  [&_>_button_+_button]:mt-4 [&_form_+_details_button]:mt-2 [&_form_+_details_button]:mr-2
  [&_form_+_details_button]:mb-0 [&_form_+_details_button]:ml-0 [&_>_h2:first-child]:pb-4
  [&_>_h2:first-child]:border-b [&_>_h2:first-child]:border-b-border [&_>_.row_+_.row]:mt-4
" aria-label="Live diagnostics">
  <div class="
    row flex items-center gap-[0.7rem] justify-between flex-wrap max-[700px]:flex-wrap
    [&_>_button]:shrink-0 max-[700px]:[&_>_.grow]:basis-[180px]
  ">
    <h2 class={ui.h2}>Live diagnostics</h2>
    <span class="text-muted text-[0.78rem]">Updates automatically</span>
  </div>
  <div class="diagnostic-accounts">
    {#each snapshot.Accounts as account (account.Name)}
      <div class="px-0 py-4 border-b border-b-border [&_p]:mt-2 [&_button]:ml-auto [&_button]:text-[0.85rem]">
        <div class="
          row flex items-center gap-[0.7rem] flex-wrap max-[700px]:flex-wrap [&_>_button]:shrink-0
          max-[700px]:[&_>_.grow]:basis-[180px]
        ">
          <strong>{account.Name}</strong>
          <span class={[
            "px-[0.7rem] py-[0.3rem] rounded-[4px] text-xs",
            account.State === "registered" ? "text-[#a7d7b9] bg-[#253b30]" :
            account.State === "failed" ? "text-[#ffc2cc] bg-[#382027]" : "bg-[#435362]",
          ]}>{account.State}</span>
          <button class={ui.button} disabled={busy || account.State === "registering"}
            onclick={() => run(() => api.RetryRegistration(account.Name))}>Retry registration</button>
        </div>
        {#if account.Error}<p class="m-0 leading-[1.6] text-[#f9a8b9]">{account.Error}</p>{/if}
        {#if account.PresenceError}<p class="m-0 leading-[1.6] text-[#f9a8b9]">Presence: {account.PresenceError}</p>{/if}
        {#if account.Capabilities?.Detail}<p class="m-0 leading-[1.6] text-muted text-[0.78rem]">{account.Capabilities.Detail}</p>{/if}
      </div>
    {:else}<p class="m-0 leading-[1.6] text-muted">No enabled accounts. Enable an account in Accounts to connect.</p>{/each}
  </div>
  {#each snapshot.Calls as call (call.ID)}
    <div class="px-0 py-4 border-b border-b-border">
      <div class="
        row flex items-center gap-[0.7rem] justify-between flex-wrap max-[700px]:flex-wrap
        [&_>_button]:shrink-0 max-[700px]:[&_>_.grow]:basis-[180px]
      "><strong>{call.Remote}</strong><span>{call.State}</span></div>
      <dl class="
        mx-0 grid grid-cols-[repeat(auto-fit,_minmax(120px,_1fr))] gap-4 mt-4 mb-0 tabular-nums
        [&_dt]:text-[0.8rem] [&_dt]:text-muted [&_dd]:mx-0 [&_dd]:mt-1 [&_dd]:mb-0
      ">
        <div><dt>Codec</dt><dd>{call.Stats?.Codec || "Negotiating"}</dd></div>
        <div><dt>Jitter</dt><dd>{(call.Stats?.JitterMilliseconds || 0).toFixed(1)} ms</dd></div>
        <div><dt>Packets lost</dt><dd>{call.Stats?.PacketsLost || 0}</dd></div>
        <div><dt>Round trip</dt><dd>{call.Stats?.RTTMilliseconds > 0 ? `${call.Stats.RTTMilliseconds.toFixed(1)} ms` : "Waiting"}</dd></div>
      </dl>
      {#if call.Error}<p class="m-0 leading-[1.6] text-[#f9a8b9]">{call.Error}</p>{/if}
    </div>
  {/each}
  <details class={[ui.details, "mt-[1.3rem] border-t border-t-border pt-4"]}>
    <summary class={[ui.summary, "px-0 py-1"]}>Recent application events ({Math.min(logs.length, 20)})</summary>
    <pre class="p-4 whitespace-pre-wrap wrap-anywhere max-h-[60vh] overflow-auto text-xs bg-[#1c252e] rounded-[6px]">{logs.slice(-20).join("\n") || "No log entries yet."}</pre>
  </details>
</section>
