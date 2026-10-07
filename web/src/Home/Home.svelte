<script>
  import { ui } from "../ui.js";
  import { onMount } from "svelte";
  import Dropdown from "../components/Dropdown.svelte";
  import * as api from "../../wailsjs/go/app/App.js";
  let { snapshot, run, busy } = $props();
  let configs = $state({});
  let name = $state("");
  let key = $state("");
  let keys = $state({});
  let editing = $state(false);
  const empty = () => ({
    Server: "",
    ForwardAlways: "",
    ForwardBusy: "",
    ForwardNoAnswer: "",
    NoAnswerSeconds: 20,
    Features: [],
    Codecs: [],
    TLSCAFile: "",
    ICEPolicy: "disabled",
    DelayedOffer: false,
    ICEServers: [],
    MaxRedirects: 0,
    Port: 5060,
    Username: "",
    Password: "",
    UseSecretService: false,
    AutoEnable: false,
    DisplayName: "",
    Domain: "",
    AuthUsername: "",
    Transport: "udp",
    MediaSecurity: "disabled",
    SymmetricRTP: false,
    LocalAddress: "",
    MediaAddress: "",
    OutboundProxy: "",
    PresenceMode: "auto",
    MessagingMode: "auto",
    DTMFMode: "auto",
    Voicemail: "",
  });
  let cfg = $state(empty());
  async function load() {
    configs = await api.ListConfigs();
  }
  onMount(() => {
    run(load);
  });
  async function edit(n) {
    cfg = { ...empty(), ...(await api.GetConfig(n, keys[n] || "")) };
    name = n;
    key = keys[n] || "";
    editing = true;
  }
</script>

<div class="
  accounts-layout grid grid-cols-[minmax(230px,0.75fr)_minmax(0,1.5fr)] items-start gap-[24px]
  max-[850px]:grid-cols-[minmax(0,1fr)]
">
  <section class="
    panel account-list p-6 bg-surface shadow-panel border-t-[#526271] border-r-border border-b-border
    border-l-border rounded-[6px] min-w-0 border max-[700px]:p-[1.15rem]
    [&_>_button_+_button]:mt-4 [&_form_+_details_button]:mt-2 [&_form_+_details_button]:mr-2
    [&_form_+_details_button]:mb-0 [&_form_+_details_button]:ml-0 [&_>_h2:first-child]:pb-4
    [&_>_h2:first-child]:border-b [&_>_h2:first-child]:border-b-border [&_>_.row_+_.row]:mt-4
  ">
    <h2 class={ui.h2}>Your accounts</h2>
    <p class="mt-[8px] mb-[24px] mx-0 leading-[1.6] text-muted text-[0.78rem]">
      Enable an account to make and receive calls. You can use more than one at
      a time.
    </p>
    {#each Object.entries(configs) as [n, encrypted]}
      {@const active = snapshot.Accounts.find((a) => a.Name === n)}
      <div class="
        account-item grid gap-[16px] py-[20px] border-b-0 [&+.account-item]:border-t
        [&+.account-item]:border-[var(--border)] [&>p]:m-0 [&>p]:[overflow-wrap:anywhere] px-0
        [&_>_label]:mb-[0.9rem] [&_details]:p-0 [&_details]:m-0 [&_details]:border-0 [&_>_.row]:mb-[0.9rem]
      ">
        <div class="
          row flex items-center gap-[0.7rem] justify-between max-[700px]:flex-wrap [&_>_button]:shrink-0
          max-[700px]:[&_>_.grow]:basis-[180px]
        ">
          <strong class="[overflow-wrap:anywhere]">{n}</strong><span class="px-[0.7rem] py-[0.3rem] rounded-[4px] bg-[#435362] text-xs"
            >{active?.State || "disabled"}</span
          >
        </div>
        {#if encrypted}<label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]"
            >Unlock key<input class={[ui.input, "px-[0.8rem] py-[0.65rem]"]}
              type="password"
              autocomplete="off"
              bind:value={keys[n]}
            /></label
          >{/if}
        <div class="
          row flex items-center gap-[0.7rem] flex-wrap max-[700px]:flex-wrap [&_>_button]:shrink-0
          max-[700px]:[&_>_.grow]:basis-[180px]
        ">
          <button class={[ui.control, `
            px-4 py-[0.65rem] rounded-[6px] bg-[#35424e] text-inherit min-h-[38px] shadow-control font-medium border
            border-[#536271] [&.primary]:bg-brand [&.primary]:text-white [&.primary]:font-semibold
            [&.primary]:border-brand-border [&.primary:hover:enabled]:bg-brand-hover
            [&:hover:not(:disabled)]:bg-[#435362] [&:hover:not(:disabled)]:border-[#727e8d]
          `]}
            class:primary={!active}
            onclick={() =>
              run(async () => {
                if (active) await api.DisableConfig(n);
                else await api.EnableConfig(n, keys[n] || "");
              })}>{active ? "Disable" : "Enable"}</button
          ><button class={ui.button} onclick={() => run(() => edit(n))}>Edit</button>
          {#if active}<button class={ui.button}
              disabled={busy || active.State === "registering"}
              onclick={() => run(() => api.RetryRegistration(n))}
              >Retry registration</button
            >{/if}
          <details class={[ui.details, "account-removal w-full [&_button]:w-full mt-[1.3rem] border-t border-t-border pt-4"]}>
            <summary class={[ui.summary, "px-0 py-1"]}>Remove account…</summary>
            <div class="stack pt-[12px] flex flex-col gap-4">
              <button class={ui.button} onclick={() => run(() => api.DeleteSavedPassword(n))}
                >Delete saved wallet password</button
              ><button
                class={[ui.dangerButton, "danger"]}
                onclick={() =>
                  run(async () => {
                    await api.RemoveConfig(n, encrypted);
                    await load();
                  })}>Remove this account</button
              >
            </div>
          </details>
        </div>
        {#if active?.Error}<p class="m-0 leading-[1.6] text-[#f9a8b9]">{active.Error}</p>{/if}
        {#if active?.Capabilities}<p class="m-0 leading-[1.6] text-muted text-[0.78rem]">
            {active.Capabilities.Detail}
          </p>{/if}
      </div>
    {:else}<p class="px-0 py-8 m-0 leading-[1.6] text-muted text-center">Add your SIP account to get started.</p>{/each}
  </section>
  <section class="
    panel account-editor p-6 bg-surface shadow-panel border-t-[#526271] border-r-border border-b-border
    border-l-border rounded-[6px] min-w-0 border max-[700px]:p-[1.15rem]
    [&_>_button_+_button]:mt-4 [&_form_+_details_button]:mt-2 [&_form_+_details_button]:mr-2
    [&_form_+_details_button]:mb-0 [&_form_+_details_button]:ml-0 [&_>_h2:first-child]:pb-4
    [&_>_h2:first-child]:border-b [&_>_h2:first-child]:border-b-border [&_>_.row_+_.row]:mt-4
  ">
    <div class="
      row flex items-center gap-[0.7rem] justify-between flex-wrap max-[700px]:flex-wrap
      [&_>_button]:shrink-0 max-[700px]:[&_>_.grow]:basis-[180px]
    ">
      <h2 class={ui.h2}>{editing ? "Edit account" : "Add an account"}</h2>
      {#if editing}
        <button class={ui.button}
          type="button"
          onclick={() => {
            cfg = empty();
            name = "";
            key = "";
            editing = false;
          }}>Add another account</button
        >
      {/if}
    </div>
    <p class="editor-intro mt-[8px] mb-[24px] mx-0 leading-[1.6] text-muted">
      Your provider supplies the server, extension and password.
    </p>
    <form
      class="stack account-form gap-[20px] flex flex-col"
      onsubmit={(e) => {
        e.preventDefault();
        run(async () => {
          await api.SetConfig(cfg, name, key);
          await load();
        });
      }}
    >
      <div class="form-section grid gap-[20px]">
        <label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]"
          >Account name<input class={[ui.input, "px-[0.8rem] py-[0.65rem]"]}
            required
            bind:value={name}
            readonly={editing}
          /></label
        >
        <div class="row flex items-center gap-[0.7rem] max-[700px]:flex-wrap [&_>_button]:shrink-0 max-[700px]:[&_>_.grow]:basis-[180px]">
          <label class="grow flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5] flex-1 min-w-0"
            >SIP server<input class={[ui.input, "px-[0.8rem] py-[0.65rem]"]}
              required
              placeholder="pbx.example.com"
              bind:value={cfg.Server}
            /></label
          ><label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5] max-w-[100px]"
            >Port<input class={[ui.input, "px-[0.8rem] py-[0.65rem]"]}
              type="number"
              min="0"
              max="65535"
              bind:value={cfg.Port}
            /></label
          >
        </div>
        <label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]"
          >Username / extension<input class={[ui.input, "px-[0.8rem] py-[0.65rem]"]}
            required
            bind:value={cfg.Username}
          /></label
        ><label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]"
          >SIP password<input class={[ui.input, "px-[0.8rem] py-[0.65rem]"]}
            type="password"
            autocomplete="new-password"
            bind:value={cfg.Password}
          /></label
        ><label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]">Display name<input class={[ui.input, "px-[0.8rem] py-[0.65rem]"]} bind:value={cfg.DisplayName} /></label>
      </div>
      <div class="form-section account-preferences grid gap-[12px] py-[20px] border-y border-[var(--border)] [&_p]:m-0 [&_p]:pl-[28px]">
        <label class="flex flex-row gap-[0.65rem] text-[0.9rem] text-[#c8cdd5] items-center min-h-[32px] cursor-pointer"
          ><input class={ui.checkbox} type="checkbox" bind:checked={cfg.AutoEnable} /> Register automatically
          on startup</label
        >
        <p class="m-0 leading-[1.6] text-muted text-[0.78rem]">
          Encrypted account files still require manual unlocking. Desktop wallet
          accounts may ask to unlock the wallet.
        </p>
        <label class="flex flex-row gap-[0.65rem] text-[0.9rem] text-[#c8cdd5] items-center min-h-[32px] cursor-pointer"
          ><input class={ui.checkbox} type="checkbox" bind:checked={cfg.UseSecretService} /> Store SIP
          password in the desktop wallet</label
        >
        {#if cfg.UseSecretService}<p class="m-0 leading-[1.6] text-muted text-[0.78rem]">
            The account file contains no SIP password. Leave the password blank
            to keep the saved wallet entry; enabling the account may ask to
            unlock your wallet.
          </p>{/if}
      </div>
      <label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]"
        >Transport<Dropdown
          label="Transport"
          bind:value={cfg.Transport}
          options={[
            { value: "udp", label: "UDP" },
            { value: "tcp", label: "TCP" },
            { value: "tls", label: "TLS · certificate verified" },
          ]}
        /></label
      >
      <div class="form-actions grid justify-items-start gap-[12px] [&_p]:m-0">
        <button class={[ui.primaryButton, "primary"]} type="submit">Save account</button>
        <p class="m-0 leading-[1.6] text-muted text-[0.78rem]">
          Re-enable this account after saving to apply changes.
        </p>
      </div>
      <h3 class="advanced-heading mt-[12px] mb-0 text-base mx-0 font-semibold">Optional settings</h3>
      <details class={[ui.details, "m-0 border border-[var(--border)] rounded-[6px] p-0 [&_p]:m-0"]}>
        <summary class={[ui.summary, "py-[16px] px-[18px] font-medium max-[600px]:p-[14px]"]}>SIP connection</summary>
        <div class="stack pt-[4px] px-[18px] pb-[20px] gap-[18px] max-[600px]:pt-0 max-[600px]:px-[14px] max-[600px]:pb-[16px] flex flex-col">
          <label class="flex flex-row gap-[0.65rem] text-[0.9rem] text-[#c8cdd5] items-center min-h-[32px] cursor-pointer"
            ><input class={ui.checkbox} type="checkbox" bind:checked={cfg.DelayedOffer} /> Let the server
            offer audio (delayed offer)</label
          >
          <p class="m-0 leading-[1.6] text-muted text-[0.78rem]">
            Enable if your provider supplies the audio offer. Reliable
            provisional offers are answered separately for each endpoint;
            microphone capture starts after the final answer.
          </p>
          <label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]"
            >SIP domain<input class={[ui.input, "px-[0.8rem] py-[0.65rem]"]}
              placeholder="Defaults to server"
              bind:value={cfg.Domain}
            /></label
          ><label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]"
            >Authentication username<input class={[ui.input, "px-[0.8rem] py-[0.65rem]"]}
              placeholder="Defaults to username"
              bind:value={cfg.AuthUsername}
            /></label
          ><label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]">Outbound proxy<input class={[ui.input, "px-[0.8rem] py-[0.65rem]"]} bind:value={cfg.OutboundProxy} /></label
          ><label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]"
            >Local signaling address<input class={[ui.input, "px-[0.8rem] py-[0.65rem]"]}
              placeholder="Automatic"
              bind:value={cfg.LocalAddress}
            /></label
          ><label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]"
            >Advertised media IP<input class={[ui.input, "px-[0.8rem] py-[0.65rem]"]}
              placeholder="Determined from route to server"
              bind:value={cfg.MediaAddress}
            /></label
          >
        </div>
      </details>
      <details class={[ui.details, "m-0 border border-[var(--border)] rounded-[6px] p-0 [&_p]:m-0"]}>
        <summary class={[ui.summary, "py-[16px] px-[18px] font-medium max-[600px]:p-[14px]"]}>Call encryption</summary>
        <div class="stack pt-[4px] px-[18px] pb-[20px] gap-[18px] max-[600px]:pt-0 max-[600px]:px-[14px] max-[600px]:pb-[16px] flex flex-col">
          <label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]"
            >Media encryption<Dropdown
              label="Media encryption"
              bind:value={cfg.MediaSecurity}
              options={[
                { value: "disabled", label: "Plain RTP · compatibility" },
                { value: "optional", label: "Prefer SDES-SRTP · requires TLS" },
                {
                  value: "required",
                  label: "Require SDES-SRTP · requires TLS",
                },
                {
                  value: "dtls",
                  label: "Require DTLS-SRTP · requires TLS and RTCP mux",
                },
              ]}
              onchange={(nextValue) => {
                if (nextValue !== "required") cfg.SymmetricRTP = false;
              }}
            /></label
          >
          <label class="flex flex-row gap-[0.65rem] text-[0.9rem] text-[#c8cdd5] items-center min-h-[32px] cursor-pointer">
            <input class={ui.checkbox}
              type="checkbox"
              bind:checked={cfg.SymmetricRTP}
              disabled={cfg.MediaSecurity !== "required"}
            />
            Allow authenticated SRTP port changes
          </label>
          <p class="m-0 leading-[1.6] text-muted text-[0.78rem]">
            Requires SDES-SRTP over TLS. Accepts a new port only after packet
            authentication from the same IP address. ICE takes precedence; plain
            RTP and DTLS paths remain fixed.
          </p>
        </div>
      </details>
      <details class={[ui.details, "m-0 border border-[var(--border)] rounded-[6px] p-0 [&_p]:m-0"]}>
        <summary class={[ui.summary, "py-[16px] px-[18px] font-medium max-[600px]:p-[14px]"]}>Provider compatibility and voicemail</summary>
        <div class="stack pt-[4px] px-[18px] pb-[20px] gap-[18px] max-[600px]:pt-0 max-[600px]:px-[14px] max-[600px]:pb-[16px] flex flex-col">
          <label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]"
            >DTMF<Dropdown
              label="DTMF"
              bind:value={cfg.DTMFMode}
              options={[
                { value: "auto", label: "Automatic" },
                { value: "rtp", label: "RTP telephone events" },
                { value: "info", label: "SIP INFO · Swyx compatibility" },
              ]}
            /></label
          >
          <label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]"
            >Presence<Dropdown
              label="Presence"
              bind:value={cfg.PresenceMode}
              options={[
                { value: "auto", label: "Automatic" },
                { value: "standard", label: "Standard SIP" },
                { value: "dialog", label: "Dialog / busy lamp" },
                { value: "swyx", label: "Swyx" },
                { value: "disabled", label: "Disabled" },
              ]}
            /></label
          >
          <label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]"
            >Messaging<Dropdown
              label="Messaging"
              bind:value={cfg.MessagingMode}
              options={[
                { value: "auto", label: "Automatic" },
                { value: "standard", label: "Standard SIP MESSAGE" },
                { value: "swyx", label: "Swyx" },
                { value: "disabled", label: "Disabled" },
              ]}
            /></label
          >
          <label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]">Voicemail number<input class={[ui.input, "px-[0.8rem] py-[0.65rem]"]} bind:value={cfg.Voicemail} /></label>
        </div>
      </details>
      <details class={[ui.details, "m-0 border border-[var(--border)] rounded-[6px] p-0 [&_p]:m-0"]}>
        <summary class={[ui.summary, "py-[16px] px-[18px] font-medium max-[600px]:p-[14px]"]}>Call forwarding</summary>
        <div class="stack pt-[4px] px-[18px] pb-[20px] gap-[18px] max-[600px]:pt-0 max-[600px]:px-[14px] max-[600px]:pb-[16px] flex flex-col">
          <label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]"
            >Always forward to<input class={[ui.input, "px-[0.8rem] py-[0.65rem]"]}
              bind:value={cfg.ForwardAlways}
              placeholder="Disabled when empty"
            /></label
          >
          <label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]"
            >Forward when busy or DND<input class={[ui.input, "px-[0.8rem] py-[0.65rem]"]}
              bind:value={cfg.ForwardBusy}
            /></label
          >
          <label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]"
            >Forward when unanswered<input class={[ui.input, "px-[0.8rem] py-[0.65rem]"]}
              bind:value={cfg.ForwardNoAnswer}
            /></label
          >
          <label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]"
            >Ring for seconds<input class={[ui.input, "px-[0.8rem] py-[0.65rem]"]}
              type="number"
              min="5"
              max="300"
              bind:value={cfg.NoAnswerSeconds}
            /></label
          >
          <p class="m-0 leading-[1.6] text-muted text-[0.78rem]">
            Forwarding sends a SIP redirect; your PBX must support it. Settings
            apply after re-enabling this account.
          </p>
          <label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]"
            >Follow remote redirects<Dropdown
              label="Follow remote redirects"
              bind:value={cfg.MaxRedirects}
              options={[
                { value: 0, label: "Do not follow automatically (default)" },
                { value: 3, label: "Follow up to 3 SIP destinations" },
                { value: 5, label: "Follow up to 5 SIP destinations" },
              ]}
            /></label
          >
        </div>
      </details>
      <details class={[ui.details, "m-0 border border-[var(--border)] rounded-[6px] p-0 [&_p]:m-0"]}>
        <summary class={[ui.summary, "py-[16px] px-[18px] font-medium max-[600px]:p-[14px]"]}>Audio codec preference</summary>
        <div class="stack pt-[4px] px-[18px] pb-[20px] gap-[18px] max-[600px]:pt-0 max-[600px]:px-[14px] max-[600px]:pb-[16px] flex flex-col">
          <label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]"
            >Codec preference<input class={[ui.input, "px-[0.8rem] py-[0.65rem]"]}
              value={(cfg.Codecs || []).join(", ")}
              onchange={(e) =>
                (cfg.Codecs = e.target.value
                  .split(",")
                  .map((v) => v.trim().toUpperCase())
                  .filter(Boolean))}
              placeholder="OPUS, G722, PCMA, PCMU"
            /></label
          >
          <p class="m-0 leading-[1.6] text-muted text-[0.78rem]">
            Leave empty for the default quality preference. Both peers negotiate
            a common codec.
          </p>
        </div>
      </details>
      <details class={[ui.details, "m-0 border border-[var(--border)] rounded-[6px] p-0 [&_p]:m-0"]}>
        <summary class={[ui.summary, "py-[16px] px-[18px] font-medium max-[600px]:p-[14px]"]}>Network traversal and certificates</summary>
        <div class="stack pt-[4px] px-[18px] pb-[20px] gap-[18px] max-[600px]:pt-0 max-[600px]:px-[14px] max-[600px]:pb-[16px] flex flex-col">
          <label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]"
            >Additional TLS CA file<input class={[ui.input, "px-[0.8rem] py-[0.65rem]"]}
              bind:value={cfg.TLSCAFile}
              placeholder="System trust by default"
            /></label
          >
          <label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]"
            >ICE traversal<Dropdown
              label="ICE traversal"
              bind:value={cfg.ICEPolicy}
              options={[
                { value: "disabled", label: "Disabled · PBX compatibility" },
                {
                  value: "auto",
                  label: "Offer ICE; allow ordinary RTP when peer omits it",
                },
                {
                  value: "required",
                  label: "Require ICE and RTCP multiplexing",
                },
              ]}
            /></label
          >
          {#each cfg.ICEServers || [] as server, i}
            <div class="stack repeated-fields p-[18px] border border-[var(--border)] rounded-[6px] [&_button]:self-start flex flex-col gap-4">
              <label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]"
                >STUN / TURN URLs<input class={[ui.input, "px-[0.8rem] py-[0.65rem]"]}
                  value={(server.URLs || []).join(", ")}
                  onchange={(e) =>
                    (server.URLs = e.target.value
                      .split(",")
                      .map((v) => v.trim())
                      .filter(Boolean))}
                  placeholder="stun:host:3478, turn:host:3478"
                /></label
              >
              <label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]">TURN username<input class={[ui.input, "px-[0.8rem] py-[0.65rem]"]} bind:value={server.Username} /></label>
              <label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]"
                >TURN credential<input class={[ui.input, "px-[0.8rem] py-[0.65rem]"]}
                  type="password"
                  autocomplete="new-password"
                  bind:value={server.Credential}
                /></label
              >
              <button class={ui.button}
                type="button"
                onclick={() =>
                  (cfg.ICEServers = cfg.ICEServers.filter((_, n) => n !== i))}
                >Remove server</button
              >
            </div>
          {/each}
          <button class={ui.button}
            type="button"
            disabled={(cfg.ICEServers || []).length >= 8}
            onclick={() =>
              (cfg.ICEServers = [
                ...(cfg.ICEServers || []),
                { URLs: [], Username: "", Credential: "" },
              ])}>Add STUN / TURN server</button
          >
          <p class="m-0 leading-[1.6] text-muted text-[0.78rem]">
            TURN credentials are saved in the account file; use an encryption
            key to encrypt the whole profile.
          </p>
        </div>
      </details>
      <details class={[ui.details, "m-0 border border-[var(--border)] rounded-[6px] p-0 [&_p]:m-0"]}>
        <summary class={[ui.summary, "py-[16px] px-[18px] font-medium max-[600px]:p-[14px]"]}>PBX feature codes</summary>
        <div class="stack pt-[4px] px-[18px] pb-[20px] gap-[18px] max-[600px]:pt-0 max-[600px]:px-[14px] max-[600px]:pb-[16px] flex flex-col">
          <p class="m-0 leading-[1.6] text-muted text-[0.78rem]">
            Configure your PBX's pickup, parking, unpark or conference codes.
            Use {"{number}"} in a target to insert the dial field.
          </p>
          {#each cfg.Features || [] as feature, i}
            <div class="stack repeated-fields p-[18px] border border-[var(--border)] rounded-[6px] [&_button]:self-start flex flex-col gap-4">
              <label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]"
                >Feature name<input class={[ui.input, "px-[0.8rem] py-[0.65rem]"]}
                  bind:value={feature.Name}
                  placeholder="Call pickup"
                /></label
              >
              <label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]"
                >Target<input class={[ui.input, "px-[0.8rem] py-[0.65rem]"]}
                  bind:value={feature.Target}
                  placeholder="*8"
                /></label
              >
              <label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]"
                >Action<Dropdown
                  label="Feature action"
                  bind:value={feature.Action}
                  options={[
                    { value: "dial", label: "Dial" },
                    { value: "transfer", label: "Transfer an active call" },
                  ]}
                /></label
              >
              <button class={ui.button}
                type="button"
                onclick={() =>
                  (cfg.Features = cfg.Features.filter((_, n) => n !== i))}
                >Remove feature</button
              >
            </div>
          {/each}
          <button class={ui.button}
            type="button"
            disabled={(cfg.Features || []).length >= 20}
            onclick={() =>
              (cfg.Features = [
                ...(cfg.Features || []),
                { Name: "", Target: "", Action: "dial" },
              ])}>Add feature code</button
          >
        </div>
      </details>
      <details class={[ui.details, "m-0 border border-[var(--border)] rounded-[6px] p-0 [&_p]:m-0"]}>
        <summary class={[ui.summary, "py-[16px] px-[18px] font-medium max-[600px]:p-[14px]"]}>Account file encryption</summary>
        <div class="stack pt-[4px] px-[18px] pb-[20px] gap-[18px] max-[600px]:pt-0 max-[600px]:px-[14px] max-[600px]:pb-[16px] flex flex-col">
          <label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]"
            >Configuration encryption key<input class={[ui.input, "px-[0.8rem] py-[0.65rem]"]}
              type="password"
              autocomplete="new-password"
              placeholder="Empty stores the account without encryption"
              bind:value={key}
            /></label
          >
          <p class="m-0 leading-[1.6] text-muted text-[0.78rem]">
            Keep the same encryption mode when editing an existing account.
          </p>
        </div>
      </details>
      <div class="form-actions grid justify-items-start gap-[12px] [&_p]:m-0">
        <button class={[ui.primaryButton, "primary"]} type="submit">Save account</button>
      </div>
    </form>
  </section>
</div>
