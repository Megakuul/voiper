<script>
  import { onMount } from "svelte";
  import * as api from "../../wailsjs/go/app/App.js";
  let { snapshot, run } = $props();
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

<div class="accounts-layout">
  <section class="panel account-list">
    <h2>Your accounts</h2>
    <p class="muted small">
      Enable an account to make and receive calls. You can use more than one at
      a time.
    </p>
    {#each Object.entries(configs) as [n, encrypted]}
      {@const active = snapshot.Accounts.find((a) => a.Name === n)}
      <div class="account-item">
        <div class="row spread">
          <strong>{n}</strong><span class="pill"
            >{active?.State || "disabled"}</span
          >
        </div>
        {#if encrypted}<label
            >Unlock key<input
              type="password"
              autocomplete="off"
              bind:value={keys[n]}
            /></label
          >{/if}
        <div class="row wrap">
          <button
            class:primary={!active}
            onclick={() =>
              run(async () => {
                if (active) await api.DisableConfig(n);
                else await api.EnableConfig(n, keys[n] || "");
              })}>{active ? "Disable" : "Enable"}</button
          ><button onclick={() => run(() => edit(n))}>Edit</button>
          {#if active?.State === "failed"}<button
              onclick={() => run(() => api.EnableConfig(n, keys[n] || ""))}
              >Retry registration</button
            >{/if}
          <details class="account-removal">
            <summary>Remove account…</summary>
            <div class="stack">
              <button onclick={() => run(() => api.DeleteSavedPassword(n))}
                >Delete saved wallet password</button
              ><button
                class="danger"
                onclick={() =>
                  run(async () => {
                    await api.RemoveConfig(n, encrypted);
                    await load();
                  })}>Remove this account</button
              >
            </div>
          </details>
        </div>
        {#if active?.Error}<p class="error-text">{active.Error}</p>{/if}
        {#if active?.Capabilities}<p class="muted small">
            {active.Capabilities.Detail}
          </p>{/if}
      </div>
    {:else}<p class="empty">Add your SIP account to get started.</p>{/each}
  </section>
  <section class="panel account-editor">
    <div class="row spread wrap">
      <h2>{editing ? "Edit account" : "Add an account"}</h2>
      {#if editing}
        <button
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
    <p class="muted editor-intro">
      Your provider supplies the server, extension and password.
    </p>
    <form
      class="stack account-form"
      onsubmit={(e) => {
        e.preventDefault();
        run(async () => {
          await api.SetConfig(cfg, name, key);
          await load();
        });
      }}
    >
      <div class="form-section">
        <label
          >Account name<input
            required
            bind:value={name}
            readonly={editing}
          /></label
        >
        <div class="row">
          <label class="grow"
            >SIP server<input
              required
              placeholder="pbx.example.com"
              bind:value={cfg.Server}
            /></label
          ><label class="port"
            >Port<input
              type="number"
              min="0"
              max="65535"
              bind:value={cfg.Port}
            /></label
          >
        </div>
        <label
          >Username / extension<input
            required
            bind:value={cfg.Username}
          /></label
        ><label
          >SIP password<input
            type="password"
            autocomplete="new-password"
            bind:value={cfg.Password}
          /></label
        ><label>Display name<input bind:value={cfg.DisplayName} /></label>
      </div>
      <div class="form-section account-preferences">
        <label class="toggle"
          ><input type="checkbox" bind:checked={cfg.AutoEnable} /> Register automatically
          on startup</label
        >
        <p class="muted small">
          Encrypted account files still require manual unlocking. Desktop wallet
          accounts may ask to unlock the wallet.
        </p>
        <label class="toggle"
          ><input type="checkbox" bind:checked={cfg.UseSecretService} /> Store SIP
          password in the desktop wallet</label
        >
        {#if cfg.UseSecretService}<p class="muted small">
            The account file contains no SIP password. Leave the password blank
            to keep the saved wallet entry; enabling the account may ask to
            unlock your wallet.
          </p>{/if}
      </div>
      <label
        >Transport<select bind:value={cfg.Transport}
          ><option value="udp">UDP</option><option value="tcp">TCP</option
          ><option value="tls">TLS · certificate verified</option></select
        ></label
      >
      <div class="form-actions">
        <button class="primary" type="submit">Save account</button>
        <p class="muted small">
          Re-enable this account after saving to apply changes.
        </p>
      </div>
      <h3 class="advanced-heading">Optional settings</h3>
      <details>
        <summary>SIP connection</summary>
        <div class="stack">
          <label class="toggle"
            ><input type="checkbox" bind:checked={cfg.DelayedOffer} /> Let the server
            offer audio (delayed offer)</label
          >
          <p class="muted small">
            Enable if your provider supplies the audio offer. Reliable
            provisional offers are answered separately for each endpoint;
            microphone capture starts after the final answer.
          </p>
          <label
            >SIP domain<input
              placeholder="Defaults to server"
              bind:value={cfg.Domain}
            /></label
          ><label
            >Authentication username<input
              placeholder="Defaults to username"
              bind:value={cfg.AuthUsername}
            /></label
          ><label>Outbound proxy<input bind:value={cfg.OutboundProxy} /></label
          ><label
            >Local signaling address<input
              placeholder="Automatic"
              bind:value={cfg.LocalAddress}
            /></label
          ><label
            >Advertised media IP<input
              placeholder="Determined from route to server"
              bind:value={cfg.MediaAddress}
            /></label
          >
        </div>
      </details>
      <details>
        <summary>Call encryption</summary>
        <div class="stack">
          <label
            >Media encryption<select
              bind:value={cfg.MediaSecurity}
              onchange={(event) => {
                if (event.currentTarget.value !== "required")
                  cfg.SymmetricRTP = false;
              }}
              ><option value="disabled">Plain RTP · compatibility</option
              ><option value="optional">Prefer SDES-SRTP · requires TLS</option
              ><option value="required">Require SDES-SRTP · requires TLS</option
              ><option value="dtls"
                >Require DTLS-SRTP · requires TLS and RTCP mux</option
              ></select
            ></label
          >
          <label class="toggle">
            <input
              type="checkbox"
              bind:checked={cfg.SymmetricRTP}
              disabled={cfg.MediaSecurity !== "required"}
            />
            Allow authenticated SRTP port changes
          </label>
          <p class="muted small">
            Requires SDES-SRTP over TLS. Accepts a new port only after packet
            authentication from the same IP address. ICE takes precedence; plain
            RTP and DTLS paths remain fixed.
          </p>
        </div>
      </details>
      <details>
        <summary>Provider compatibility and voicemail</summary>
        <div class="stack">
          <label
            >DTMF<select bind:value={cfg.DTMFMode}
              ><option value="auto">Automatic</option><option value="rtp"
                >RTP telephone events</option
              ><option value="info">SIP INFO · Swyx compatibility</option
              ></select
            ></label
          >
          <label
            >Presence<select bind:value={cfg.PresenceMode}
              ><option value="auto">Automatic</option><option value="standard"
                >Standard SIP</option
              ><option value="dialog">Dialog / busy lamp</option><option
                value="swyx">Swyx</option
              ><option value="disabled">Disabled</option></select
            ></label
          >
          <label
            >Messaging<select bind:value={cfg.MessagingMode}
              ><option value="auto">Automatic</option><option value="standard"
                >Standard SIP MESSAGE</option
              ><option value="swyx">Swyx</option><option value="disabled"
                >Disabled</option
              ></select
            ></label
          >
          <label>Voicemail number<input bind:value={cfg.Voicemail} /></label>
        </div>
      </details>
      <details>
        <summary>Call forwarding</summary>
        <div class="stack">
          <label
            >Always forward to<input
              bind:value={cfg.ForwardAlways}
              placeholder="Disabled when empty"
            /></label
          >
          <label
            >Forward when busy or DND<input
              bind:value={cfg.ForwardBusy}
            /></label
          >
          <label
            >Forward when unanswered<input
              bind:value={cfg.ForwardNoAnswer}
            /></label
          >
          <label
            >Ring for seconds<input
              type="number"
              min="5"
              max="300"
              bind:value={cfg.NoAnswerSeconds}
            /></label
          >
          <p class="muted small">
            Forwarding sends a SIP redirect; your PBX must support it. Settings
            apply after re-enabling this account.
          </p>
          <label
            >Follow remote redirects<select bind:value={cfg.MaxRedirects}
              ><option value={0}>Do not follow automatically (default)</option
              ><option value={3}>Follow up to 3 SIP destinations</option><option
                value={5}>Follow up to 5 SIP destinations</option
              ></select
            ></label
          >
        </div>
      </details>
      <details>
        <summary>Audio codec preference</summary>
        <div class="stack">
          <label
            >Codec preference<input
              value={(cfg.Codecs || []).join(", ")}
              onchange={(e) =>
                (cfg.Codecs = e.target.value
                  .split(",")
                  .map((v) => v.trim().toUpperCase())
                  .filter(Boolean))}
              placeholder="OPUS, G722, PCMA, PCMU"
            /></label
          >
          <p class="muted small">
            Leave empty for the default quality preference. Both peers negotiate
            a common codec.
          </p>
        </div>
      </details>
      <details>
        <summary>Network traversal and certificates</summary>
        <div class="stack">
          <label
            >Additional TLS CA file<input
              bind:value={cfg.TLSCAFile}
              placeholder="System trust by default"
            /></label
          >
          <label
            >ICE traversal<select bind:value={cfg.ICEPolicy}
              ><option value="disabled">Disabled · PBX compatibility</option
              ><option value="auto"
                >Offer ICE; allow ordinary RTP when peer omits it</option
              ><option value="required"
                >Require ICE and RTCP multiplexing</option
              ></select
            ></label
          >
          {#each cfg.ICEServers || [] as server, i}
            <div class="stack repeated-fields">
              <label
                >STUN / TURN URLs<input
                  value={(server.URLs || []).join(", ")}
                  onchange={(e) =>
                    (server.URLs = e.target.value
                      .split(",")
                      .map((v) => v.trim())
                      .filter(Boolean))}
                  placeholder="stun:host:3478, turn:host:3478"
                /></label
              >
              <label>TURN username<input bind:value={server.Username} /></label>
              <label
                >TURN credential<input
                  type="password"
                  autocomplete="new-password"
                  bind:value={server.Credential}
                /></label
              >
              <button
                type="button"
                onclick={() =>
                  (cfg.ICEServers = cfg.ICEServers.filter((_, n) => n !== i))}
                >Remove server</button
              >
            </div>
          {/each}
          <button
            type="button"
            disabled={(cfg.ICEServers || []).length >= 8}
            onclick={() =>
              (cfg.ICEServers = [
                ...(cfg.ICEServers || []),
                { URLs: [], Username: "", Credential: "" },
              ])}>Add STUN / TURN server</button
          >
          <p class="muted small">
            TURN credentials are saved in the account file; use an encryption
            key to encrypt the whole profile.
          </p>
        </div>
      </details>
      <details>
        <summary>PBX feature codes</summary>
        <div class="stack">
          <p class="muted small">
            Configure your PBX's pickup, parking, unpark or conference codes.
            Use {"{number}"} in a target to insert the dial field.
          </p>
          {#each cfg.Features || [] as feature, i}
            <div class="stack repeated-fields">
              <label
                >Feature name<input
                  bind:value={feature.Name}
                  placeholder="Call pickup"
                /></label
              >
              <label
                >Target<input
                  bind:value={feature.Target}
                  placeholder="*8"
                /></label
              >
              <label
                >Action<select bind:value={feature.Action}
                  ><option value="dial">Dial</option><option value="transfer"
                    >Transfer an active call</option
                  ></select
                ></label
              >
              <button
                type="button"
                onclick={() =>
                  (cfg.Features = cfg.Features.filter((_, n) => n !== i))}
                >Remove feature</button
              >
            </div>
          {/each}
          <button
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
      <details>
        <summary>Account file encryption</summary>
        <div class="stack">
          <label
            >Configuration encryption key<input
              type="password"
              autocomplete="new-password"
              placeholder="Empty stores the account without encryption"
              bind:value={key}
            /></label
          >
          <p class="muted small">
            Keep the same encryption mode when editing an existing account.
          </p>
        </div>
      </details>
      <div class="form-actions">
        <button class="primary" type="submit">Save account</button>
      </div>
    </form>
  </section>
</div>

<style>
  .accounts-layout {
    display: grid;
    grid-template-columns: minmax(230px, 0.75fr) minmax(0, 1.5fr);
    align-items: start;
    gap: 24px;
  }
  .account-list > .small {
    margin: 8px 0 24px;
  }
  .account-item {
    display: grid;
    gap: 16px;
    padding: 20px 0;
    border-bottom: 0;
  }
  .account-item + .account-item {
    border-top: 1px solid var(--border, #343434);
  }
  .account-item > p {
    margin: 0;
    overflow-wrap: anywhere;
  }
  .account-item strong {
    overflow-wrap: anywhere;
  }
  .account-removal {
    width: 100%;
  }
  .account-removal > .stack {
    padding-top: 12px;
  }
  .account-removal button {
    width: 100%;
  }
  .editor-intro {
    margin: 8px 0 24px;
  }
  .account-form {
    gap: 20px;
  }
  .form-section {
    display: grid;
    gap: 20px;
  }
  .account-preferences {
    gap: 12px;
    padding: 20px 0;
    border-top: 1px solid var(--border, #343434);
    border-bottom: 1px solid var(--border, #343434);
  }
  .account-preferences p {
    margin: 0;
    padding-left: 28px;
  }
  .account-form > details {
    margin: 0;
    border: 1px solid var(--border, #343434);
    border-radius: 8px;
    padding: 0;
  }
  .account-form > details > summary {
    padding: 16px 18px;
    font-weight: 500;
  }
  .account-form > details > .stack {
    padding: 4px 18px 20px;
    gap: 18px;
  }
  .account-form > details p {
    margin: 0;
  }
  .repeated-fields {
    padding: 18px;
    border: 1px solid var(--border, #343434);
    border-radius: 6px;
  }
  .repeated-fields button {
    align-self: start;
  }
  .form-actions {
    display: grid;
    justify-items: start;
    gap: 12px;
  }
  .form-actions p {
    margin: 0;
  }
  .advanced-heading {
    margin: 12px 0 0;
    font-size: 1rem;
  }
  @media (max-width: 850px) {
    .accounts-layout {
      grid-template-columns: minmax(0, 1fr);
    }
  }
  @media (max-width: 600px) {
    .account-form > details > summary {
      padding: 14px;
    }
    .account-form > details > .stack {
      padding: 0 14px 16px;
    }
  }
</style>
