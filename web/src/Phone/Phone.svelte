<script>
  import { ui } from "../ui.js";
  import Icon from "../components/Icon.svelte";
  import Dropdown from "../components/Dropdown.svelte";
  import ContactCard from "../Contacts/ContactCard.svelte";
  import { extensionOf } from "../Contacts/presence.js";
  import * as api from "../../wailsjs/go/app/App.js";
  let {
    snapshot,
    account = $bindable(""),
    target = $bindable(""),
    run,
    busy,
    workspace = false,
  } = $props();
  let suggestions = $state([]);
  let contactError = $state("");
  let redialNotice = $state("");

  $effect(() => {
    const query = target.trim();
    let cancelled = false;
    const timer = setTimeout(async () => {
      try {
        const contacts = await api.ContactsPage(query, 0, 8);
        if (!cancelled) {
          suggestions = query ? contacts : workspace ? [] : contacts.filter((c) => c.Favorite);
          contactError = "";
        }
      } catch (error) {
        if (!cancelled) contactError = String(error);
      }
    }, 150);
    return () => {
      cancelled = true;
      clearTimeout(timer);
    };
  });
  let presenceByTarget = $derived(
    new Map(
      (snapshot.Presence || [])
        .filter((p) => p.Account === account && !p.Voicemail)
        .map((p) => [p.Target, p]),
    ),
  );
  let accountRegistered = $derived(
    snapshot.Accounts.some((a) => a.Name === account && a.State === "registered"),
  );
  $effect(() => {
    const extensions = suggestions.map(extensionOf).filter(Boolean);
    if (account && accountRegistered && extensions.length)
      api.WatchVisiblePresence(account, extensions).catch((e) => (contactError = String(e)));
  });
  const statRow = "flex justify-between gap-4 py-[0.6rem] border-b border-border";
  const statValue = "m-0 min-w-0 text-right wrap-anywhere";
  let transfer = $state({});
  let tones = $state({});
  let toneNotices = $state({});
  let showStats = $state(false);
  let consult = $state({});
  let statusNote = $state("");
  let restarting = $state({});
  let restartErrors = $state({});
  let activeCalls = $derived(
    snapshot.Calls.filter((call) => call.State === "connected" && !call.Held),
  );
  let hasConference = $derived(
    snapshot.Calls.some((call) => call.Stats?.Conference),
  );

  $effect(() => {
    const active = new Set(snapshot.Calls.map((call) => call.ID));
    for (const id of Object.keys(restarting)) {
      if (!active.has(id) && !restarting[id]) {
        delete restarting[id];
        delete restartErrors[id];
      }
    }
  });

  function duration(connected) {
    const seconds = Math.max(
      0,
      Math.floor((Date.now() - new Date(connected).getTime()) / 1000),
    );
    return `${Math.floor(seconds / 60)}:${String(seconds % 60).padStart(2, "0")}`;
  }

  async function restartPath(id) {
    restarting[id] = true;
    restartErrors[id] = "";
    try {
      await api.RestartMediaPath(id);
    } catch (error) {
      restartErrors[id] = String(error);
    } finally {
      restarting[id] = false;
    }
  }

  function statistics(stats) {
    const flowLabels = {
      waiting: "Starting media",
      closed: "Closed",
      held: "On hold",
      "device-unavailable": "Audio device unavailable",
      securing: "Securing media",
      inactive: "Inactive",
      "no-media": "No RTP flow",
      "missing-outbound": "No outgoing RTP",
      "missing-inbound": "No incoming RTP",
      duplex: "Sending and receiving RTP",
      sending: "Sending RTP",
      receiving: "Receiving RTP",
    };
    const received = stats.PacketsReceived || 0;
    const lost = stats.PacketsLost || 0;
    const loss =
      received + lost
        ? `${((100 * lost) / (received + lost)).toFixed(1)}%`
        : "No packets yet";
    return [
      ["Codec", stats.Codec || "Negotiating"],
      ["Audio flow", flowLabels[stats.MediaFlow] || "Waiting for audio"],
      [
        "Encoder bitrate",
        stats.EncoderBitrate > 0
          ? `${(stats.EncoderBitrate / 1000).toFixed(1)} kbit/s`
          : "Fixed by codec",
      ],
      [
        "Audio rate",
        stats.SampleRate
          ? `${stats.SampleRate / 1000} kHz codec · ${(stats.DeviceSampleRate || stats.SampleRate) / 1000} kHz device`
          : "Waiting for audio",
      ],
      [
        "Encryption",
        stats.Codec
          ? stats.Encrypted
            ? stats.Transport?.startsWith("UDP/TLS/")
              ? "DTLS-SRTP encrypted"
              : "SDES-SRTP encrypted"
            : stats.Transport?.startsWith("UDP/TLS/")
              ? "DTLS handshake pending or closed"
              : "RTP unencrypted"
          : "Negotiating",
      ],
      ["Received packet loss", `${loss} · ${lost} lost`],
      [
        "Peer-reported packet loss",
        stats.ReceiverReportAvailable
          ? `${(stats.RemotePacketLossPercent || 0).toFixed(1)}% · report ${(stats.ReceiverReportAgeSeconds || 0).toFixed(0)} s ago`
          : "Waiting for RTCP reports",
      ],
      ["Jitter", `${(stats.JitterMilliseconds || 0).toFixed(1)} ms`],
      [
        "Round trip",
        stats.RTTMilliseconds > 0
          ? `${stats.RTTMilliseconds.toFixed(1)} ms`
          : "Waiting for RTCP reports",
      ],
      ["Playout buffer", `${stats.BufferMilliseconds || 0} ms`],
      [
        "Packets",
        `${stats.PacketsSent || 0} sent · ${received} received · ${stats.PacketsDropped || 0} dropped`,
      ],
      [
        "Network path",
        [
          stats.ICECandidateType,
          stats.ICEState,
          stats.ICERestartPending ? "reconnecting" : "",
        ]
          .filter(Boolean)
          .join(" · ") || "Direct RTP",
      ],
      [
        "Audio processing",
        `Echo cancellation ${stats.EchoCancellation ? "on" : "off"} · noise suppression ${stats.NoiseSuppression ? "on" : "off"}`,
      ],
    ];
  }
</script>

<div class={["flex flex-col", workspace ? "gap-0" : "gap-6"]}>
  <!-- Without calls, the dialer comes first. -->
  <section class={snapshot.Calls.length === 0 && "order-1"}>
    {#if snapshot.Calls.filter((c) => c.State === "connected").length >= 2 || hasConference}<div
        class={[ui.row, "flex-wrap"]}
      >
        <button class={ui.button}
          disabled={busy ||
            hasConference ||
            activeCalls.length < 2 ||
            activeCalls.length > 3}
          onclick={() =>
            run(() => api.Conference(activeCalls.map((call) => call.ID)))}
          >Merge active calls</button
        >
        {#if hasConference}<button class={ui.button}
            disabled={busy}
            onclick={() => run(() => api.LeaveConference())}
            >Separate conference</button
          >{:else}<p class={ui.hint}>
            Merge two or three active calls. Resume held calls before merging.
          </p>{/if}
      </div>{/if}
    {#if snapshot.Calls.length || !workspace}<div class="flex flex-wrap items-center justify-between gap-4 wrap-anywhere">
      <h2 class="m-0 text-[1.15rem] font-semibold">{snapshot.Calls.length ? "Current calls" : "Calls"}</h2>
      {#if snapshot.Calls.length}<button
          class={ui.quietButton}
          aria-pressed={showStats}
          onclick={() => (showStats = !showStats)}
          >{showStats ? "Hide call statistics" : "Show call statistics"}</button
        >{/if}
    </div>
    {/if}
    {#each snapshot.Calls as call (call.ID)}
      <article class={["min-w-0 rounded-[6px] border border-border [&>p]:mb-4", workspace ? "p-5 mb-6 bg-surface-inset" : "p-6 mt-4 bg-surface shadow-panel border-t-panel-edge max-[700px]:p-[1.15rem]"]}>
        <span class="text-muted text-[0.78rem]"
          >{call.Account} · {call.Direction === "incoming"
            ? "Incoming call"
            : "Outgoing call"}</span
        >
        <h2 class="mx-0 mt-4 mb-4 text-[1.15rem] font-semibold wrap-anywhere">{call.Remote}</h2>
        <div class={ui.row}>
          <span class={ui.badge}
            >{call.Held
              ? "On hold"
              : {
                  connected: "Connected",
                  ringing: "Incoming call",
                  dialing: "Calling…",
                  connecting: "Connecting…",
                }[call.State] || call.State}</span
          >{#if call.Muted}<span class={ui.badge}>Microphone muted</span
            >{/if}{#if call.Stats?.Conference}<span class={ui.badge}
              >Conference</span
            >{/if}
        </div>
        {#if call.TransferStatus}<p class={ui.text} role="status">
            Transfer: {call.TransferStatus}
          </p>{/if}
        {#if call.Stats?.AudioStopped}<p class={ui.error} role="alert">
            Audio device stopped. {call.Stats.AudioRecovering
              ? "Reconnecting…"
              : call.Stats.AudioRecoveryError ||
                "Automatic recovery is paused."}
            <button class={ui.button}
              disabled={busy ||
                call.Stats.AudioRecovering ||
                call.Stats.AudioRecoveryPending}
              onclick={() => run(() => api.RetryAudio(call.ID))}
              >Retry audio</button
            >
          </p>{/if}
        {#if call.Stats?.LastError}<p class={ui.error} role="alert">
            Audio: {call.Stats.LastError}
          </p>{/if}
        {#if call.Stats?.MediaWarning}<p class={ui.error} role="status">
            {call.Stats.MediaWarning}
          </p>{/if}
        {#if call.State === "connected" && call.Stats?.ICECandidateType}
          <div class={[ui.row, "flex-wrap"]}>
            <button class={ui.button}
              disabled={busy ||
                restarting[call.ID] ||
                call.Stats.ICERestartPending}
              onclick={() => restartPath(call.ID)}
              >Reconnect network audio</button
            >
            {#if restarting[call.ID] || call.Stats.ICERestartPending}<span
                class="text-muted text-[0.78rem]"
                role="status">Checking a new audio path…</span
              >{/if}
          </div>
          {#if restartErrors[call.ID]}<p class={ui.error} role="alert">
              {restartErrors[call.ID]}
            </p>{/if}
        {/if}
        {#if call.Connected && !call.Connected.startsWith("0001-")}<p
            class={ui.hint}
          >
            Connected · {duration(call.Connected)}
          </p>{/if}
        {#if call.Error}<p class={ui.error}>{call.Error}</p>{/if}
        <div class={[ui.row, "flex-wrap mt-6"]}>
          {#if call.State === "ringing"}<button
              class={ui.primaryButton}
              disabled={busy}
              onclick={() => run(() => api.Answer(call.ID))}>Answer</button
            >{/if}
          {#if call.State === "connected"}<button class={[ui.button, "aria-pressed:text-accent-text aria-pressed:border-accent"]}
              disabled={busy}
              onclick={() => run(() => api.Mute(call.ID, !call.Muted))}
              aria-pressed={call.Muted}
              >{call.Muted ? "Unmute microphone" : "Mute microphone"}</button
            ><button class={[ui.button, "aria-pressed:text-accent-text aria-pressed:border-accent"]}
              disabled={busy}
              onclick={() => run(() => api.Hold(call.ID, !call.Held))}
              aria-pressed={call.Held}
              >{call.Held ? "Resume call" : "Hold call"}</button
            >{/if}
          <button class={ui.dangerButton} onclick={() => run(() => api.Hangup(call.ID))}
            >{call.State === "ringing" ? "Decline" : "Hang up"}</button
          >
        </div>
        {#if call.State === "ringing"}<form
            class={ui.row}
            onsubmit={(e) => {
              e.preventDefault();
              run(() => api.Redirect(call.ID, transfer[call.ID] || ""));
            }}
          >
            <input class={ui.field}
              aria-label="Redirect destination"
              placeholder="Redirect destination"
              bind:value={transfer[call.ID]}
            /><button class={ui.button} disabled={busy || !transfer[call.ID]}>Redirect</button>
          </form>{/if}
        {#if call.State === "connected"}
          <details class={[ui.details, "mt-5 pt-4 border-t border-border"]}>
            <summary class={ui.summary}>Phone menu (DTMF)</summary>
            <form
              class="stack flex flex-col gap-3 items-stretch max-w-[440px] flex-wrap [&>label]:flex-1 [&>label]:min-w-[min(220px,_100%)]"
              onsubmit={(event) => {
                event.preventDefault();
                const digit = (tones[call.ID] || "").trim().toUpperCase();
                if (busy || call.Held || !/^[0-9*#ABCD]$/.test(digit)) return;
                run(async () => {
                  toneNotices[call.ID] = "";
                  await api.SendDTMF(call.ID, digit);
                  tones[call.ID] = "";
                  toneNotices[call.ID] = `Sent ${digit}.`;
                });
              }}
            >
              <label class={ui.label} for={`tone-${call.ID}`}>Menu digit</label>
              <div class={ui.row}>
                <input class={ui.field}
                  id={`tone-${call.ID}`}
                  placeholder="0–9, * or #"
                  maxlength="1"
                  autocomplete="off"
                  bind:value={tones[call.ID]}
                />
                <button class={ui.button}
                  disabled={busy ||
                    call.Held ||
                    !/^[0-9*#ABCD]$/i.test(tones[call.ID] || "")}
                  >Send digit</button
                >
              </div>
              <p class={ui.hint}>
                Type a digit and press Enter to send it to the call’s phone
                menu.{call.Held ? " Resume the call first." : ""}
              </p>
              {#if toneNotices[call.ID]}<p class="m-0 leading-[1.6] text-[0.78rem]" role="status">
                  {toneNotices[call.ID]}
                </p>{/if}
            </form>
          </details>
          <details class={[ui.details, "mt-5 pt-4 border-t border-border"]}>
            <summary class={ui.summary}>Transfer call</summary>
            <form
              class={ui.row}
              onsubmit={(e) => {
                e.preventDefault();
                run(() => api.Transfer(call.ID, transfer[call.ID] || ""));
              }}
            >
              <input class={ui.field}
                aria-label="Number or SIP address to transfer to"
                placeholder="Number or SIP address"
                bind:value={transfer[call.ID]}
              /><button class={ui.button} disabled={busy || !transfer[call.ID]}
                >Transfer now</button
              >
            </form>
            <p class={ui.hint}>
              To speak with the recipient first, hold this call and start a
              second call. Then choose that call below.
            </p>
            <div class={[ui.row, "flex-wrap"]}>
              <Dropdown
                label="Consultation call"
                bind:value={consult[call.ID]}
                disabled={busy}
                options={[
                  { value: "", label: "Choose consultation call" },
                  ...snapshot.Calls.filter(
                    (c) =>
                      c.ID !== call.ID &&
                      c.Account === call.Account &&
                      c.State === "connected",
                  ).map((other) => ({ value: other.ID, label: other.Remote })),
                ]}
              /><button class={ui.button}
                disabled={busy || !consult[call.ID]}
                onclick={() =>
                  run(() => api.AttendedTransfer(call.ID, consult[call.ID]))}
                >Complete transfer</button
              >
            </div>
            <p class={ui.hint}>
              Keep this call open until the transfer succeeds. Transfers require
              support from the phone service.
            </p>
          </details>{/if}
        {#if call.State === "connected"}<details class={[ui.details, "mt-5 pt-4 border-t border-border"]}>
            <summary class={ui.summary}>Call volume & recording</summary><label class={ui.label}
              >Microphone · {Math.round(
                (call.Stats.InputGain ?? 1) * 100,
              )}%<input class={ui.range}
                type="range"
                min="0"
                max="2"
                step="0.1"
                value={call.Stats.InputGain ?? 1}
                onchange={(e) =>
                  run(() =>
                    api.SetGain(
                      call.ID,
                      Number(e.currentTarget.value),
                      call.Stats.OutputGain ?? 1,
                    ),
                  )}
              /></label
            ><label class={ui.label}
              >Speaker · {Math.round((call.Stats.OutputGain ?? 1) * 100)}%<input class={ui.range}
                type="range"
                min="0"
                max="2"
                step="0.1"
                value={call.Stats.OutputGain ?? 1}
                onchange={(e) =>
                  run(() =>
                    api.SetGain(
                      call.ID,
                      call.Stats.InputGain ?? 1,
                      Number(e.currentTarget.value),
                    ),
                  )}
              /></label
            ><button class={ui.button}
              disabled={busy}
              onclick={() =>
                run(() =>
                  call.Stats.Recording
                    ? api.StopRecording(call.ID)
                    : api.StartRecording(call.ID),
                )}
              >{call.Stats.Recording
                ? "Stop recording"
                : "Record to WAV…"}</button
            >{#if call.Stats.Recording}<p class={ui.text} role="status">
                Recording · {call.Stats.RecordingPath}
              </p>{/if}
          </details>{/if}
        {#if showStats}<dl class="mt-4 text-xs tabular-nums">
            {#each statistics(call.Stats || {}) as [key, value]}<div class={statRow}>
                <dt class="text-muted">{key}</dt>
                <dd class={statValue}>{value}</dd>
              </div>{/each}
          </dl>
          <details class={[ui.details, ui.section]}>
            <summary class={ui.summary}>All technical statistics</summary>
            <dl class="mt-4 text-xs tabular-nums">
              {#each Object.entries(call.Stats || {}) as [key, value]}<div class={statRow}>
                  <dt class="text-muted">{key}</dt>
                  <dd class={statValue}>{value === null ? "Unavailable" : String(value)}</dd>
                </div>{/each}
            </dl>
          </details>{/if}
      </article>
    {:else}<div class={["mt-4 py-8 text-center rounded-[6px]", workspace ? "text-muted" : "px-6 border border-border bg-surface-secondary"]}>
        <h3 class={workspace ? "m-0 text-[0.9rem] font-normal" : ui.h3}>No active calls</h3>
        {#if !workspace}<p class="m-0 leading-[1.6] text-muted">
          Type a number above to call. Incoming calls will appear here.
        </p>{/if}
      </div>{/each}
  </section>
  <section class={["min-w-0 rounded-[6px]", workspace ? "pt-2 pb-6" : "p-6 border border-border border-t-panel-edge bg-surface shadow-panel max-[700px]:p-[1.15rem]"]}>
    <div class="flex items-center gap-4 mb-7"><span class="grid place-items-center w-[48px] h-[48px] rounded-[10px] bg-accent-surface text-accent border border-brand-border"><Icon name="phone" size={24} /></span><div><h2 class="mx-0 mt-0 mb-1 text-[1.45rem] font-semibold">{workspace ? "Make a call" : "New call"}</h2>{#if workspace}<p class={ui.hint}>Enter a number, SIP address or contact name.</p>{/if}</div></div>
    <form
      class="stack flex flex-col gap-[0.85rem]"
      onsubmit={(event) => {
        event.preventDefault();
        if (!busy && account && target.trim())
          run(() => api.Dial(account, target.trim()));
      }}
    >
      <label class={ui.label} for="call-destination">Name, phone number or SIP address</label>
      <div class="flex items-stretch gap-3 max-[700px]:flex-wrap">
        <input
          id="call-destination"
          class={[ui.input, "dial-input flex-1 p-[0.8rem] text-left text-[1.15rem] max-[700px]:basis-full"]}
          placeholder="Type a number or search contacts"
          autocomplete="off"
          bind:value={target}
        />
        <button class={[ui.primaryButton, "flex shrink-0 items-center justify-center gap-2"]} disabled={busy || !account || !target.trim()}
          ><Icon name="phone" size={18} /> Call</button
        >
      </div>
      <div class={[ui.row, "justify-between flex-wrap"]}>
        <p class={ui.hint}>
          {account
            ? `Calling from ${account}. Press Enter to call.`
            : "Add and select an account to make a call."}
        </p>
        <button
          type="button"
          class={ui.quietButton}
          disabled={busy || !account}
          onclick={() =>
            run(async () => {
              const previous = await api.LastDialed(account);
              if (previous) {
                target = previous;
                redialNotice =
                  "Last number selected. Press Enter or Call to dial.";
              } else
                redialNotice = "No outgoing calls saved for this account yet.";
            })}>Use last number</button
        >
      </div>
      {#if redialNotice}<p class={ui.hint} role="status">
          {redialNotice}
        </p>{/if}
      {#if suggestions.length}
        <div
          class="max-h-[320px] overflow-y-auto pr-1"
          aria-label={target ? "Matching contacts" : "Favorite contacts"}
        >
          <p class={ui.hint}>
            {target ? "Matching contacts" : "Favorites"}
          </p>
          {#each suggestions as contact (contact.ID)}
            <ContactCard
              {contact}
              presence={presenceByTarget.get(extensionOf(contact))}
              onDial={(address) => {
                target = address;
                if (contact.Account) account = contact.Account;
              }}
            />
          {/each}
        </div>
      {/if}
      {#if contactError}<p class={ui.error} role="status">
          Contact search: {contactError}
        </p>{/if}
    </form>
  </section>
  <section class={["order-2", workspace ? "border-t border-border" : "grid grid-cols-[repeat(auto-fit,minmax(min(280px,100%),1fr))] gap-5"]} aria-label="Account services">
    {#if snapshot.Accounts.find((a) => a.Name === account)?.Features?.length}
      <details class={[ui.details, workspace ? "mt-[1.3rem] pt-4" : ui.section]}>
        <summary class={ui.summary}>Phone service shortcuts</summary>
        <div class={[ui.row, "flex-wrap"]}>
          {#each snapshot.Accounts.find((a) => a.Name === account).Features as feature}
            <button class={ui.button}
              disabled={busy ||
                (feature.Action === "transfer" &&
                  !snapshot.Calls.some(
                    (c) =>
                      c.Account === account &&
                      c.State === "connected" &&
                      !c.Held,
                  ))}
              onclick={() =>
                run(() =>
                  api.DialFeature(
                    account,
                    snapshot.Calls.find(
                      (c) =>
                        c.Account === account &&
                        c.State === "connected" &&
                        !c.Held,
                    )?.ID || "",
                    feature.Name,
                    target,
                  ),
                )}>{feature.Name}</button
            >
          {/each}
        </div>
      </details>
    {/if}
    <details class={[ui.details, workspace ? "mt-[1.3rem] pt-4" : ui.section]}>
      <summary class={ui.summary}>Your status & voicemail</summary>
      <div class={[ui.row, "flex-wrap"]}>
        <button class={ui.button}
          disabled={busy || !account}
          onclick={() => run(() => api.WatchVoicemail(account))}
          >Check for voicemail</button
        ><button class={ui.button}
          disabled={busy || !account}
          onclick={() => run(() => api.DialVoicemail(account))}
          >Call voicemail</button
        >
      </div>
      {#each (snapshot.Presence || []).filter((p) => p.Account === account && p.Voicemail) as p}<p class={ui.text}
        >
          {p.Voicemail.Waiting ? "New voicemail" : "No new voicemail"} · {p
            .Voicemail.New} new / {p.Voicemail.Old} old · {p.State}
        </p>{/each}
      <label class={ui.label}
        >Your status message<input class={ui.field}
          maxlength="512"
          bind:value={statusNote}
        /></label
      >
      <div class={[ui.row, "flex-wrap"]}>
        <button class={ui.button}
          disabled={busy || !account}
          onclick={() =>
            run(() => api.PublishStatus(account, true, statusNote))}
          >Set available</button
        ><button class={ui.button}
          disabled={busy || !account}
          onclick={() =>
            run(() => api.PublishStatus(account, false, statusNote))}
          >Set unavailable</button
        ><button class={ui.button}
          disabled={busy || !account}
          onclick={() => run(() => api.UnpublishStatus(account))}
          >Stop publishing</button
        >
      </div>
      {#each snapshot.Accounts.filter((a) => a.Name === account && a.PresenceState) as a}
        <p class="m-0 leading-[1.6] text-[0.78rem]">
          Published status: {a.PresenceState}{a.PresenceNote
            ? ` · ${a.PresenceNote}`
            : ""}
        </p>
        {#if a.PresenceError}<p class="m-0 leading-[1.6] text-[0.78rem] text-danger">
            {a.PresenceError}
          </p>{/if}
      {/each}
      <p class={ui.hint}>
        Status refreshes while this account is enabled. Standard SIP publication
        requires server support. This does not set Swyx custom statuses.
      </p>
    </details>
  </section>
</div>
