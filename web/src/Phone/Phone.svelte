<script>
  import * as api from "../../wailsjs/go/app/App.js";
  let {
    snapshot,
    account = $bindable(""),
    target = $bindable(""),
    run,
    busy,
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
          suggestions = query ? contacts : contacts.filter((c) => c.Favorite);
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

<div class="phone-layout">
  <section class="call-list" class:has-calls={snapshot.Calls.length > 0}>
    {#if snapshot.Calls.filter((c) => c.State === "connected").length >= 2 || hasConference}<div
        class="row wrap"
      >
        <button
          disabled={busy ||
            hasConference ||
            activeCalls.length < 2 ||
            activeCalls.length > 3}
          onclick={() =>
            run(() => api.Conference(activeCalls.map((call) => call.ID)))}
          >Merge active calls</button
        >
        {#if hasConference}<button
            disabled={busy}
            onclick={() => run(() => api.LeaveConference())}
            >Separate conference</button
          >{:else}<p class="muted small">
            Merge two or three active calls. Resume held calls before merging.
          </p>{/if}
      </div>{/if}
    <div class="row spread wrap call-heading">
      <h2>{snapshot.Calls.length ? "Current calls" : "Calls"}</h2>
      {#if snapshot.Calls.length}<button
          class="quiet"
          aria-pressed={showStats}
          onclick={() => (showStats = !showStats)}
          >{showStats ? "Hide call statistics" : "Show call statistics"}</button
        >{/if}
    </div>
    {#each snapshot.Calls as call (call.ID)}
      <article class="panel call-card">
        <span class="muted small"
          >{call.Account} · {call.Direction === "incoming"
            ? "Incoming call"
            : "Outgoing call"}</span
        >
        <h2>{call.Remote}</h2>
        <div class="row">
          <span class="pill"
            >{call.Held
              ? "On hold"
              : {
                  connected: "Connected",
                  ringing: "Incoming call",
                  dialing: "Calling…",
                  connecting: "Connecting…",
                }[call.State] || call.State}</span
          >{#if call.Muted}<span class="pill">Microphone muted</span
            >{/if}{#if call.Stats?.Conference}<span class="pill"
              >Conference</span
            >{/if}
        </div>
        {#if call.TransferStatus}<p role="status">
            Transfer: {call.TransferStatus}
          </p>{/if}
        {#if call.Stats?.AudioStopped}<p class="error-text" role="alert">
            Audio device stopped. {call.Stats.AudioRecovering
              ? "Reconnecting…"
              : call.Stats.AudioRecoveryError ||
                "Automatic recovery is paused."}
            <button
              disabled={busy ||
                call.Stats.AudioRecovering ||
                call.Stats.AudioRecoveryPending}
              onclick={() => run(() => api.RetryAudio(call.ID))}
              >Retry audio</button
            >
          </p>{/if}
        {#if call.Stats?.LastError}<p class="error-text" role="alert">
            Audio: {call.Stats.LastError}
          </p>{/if}
        {#if call.Stats?.MediaWarning}<p class="error-text" role="status">
            {call.Stats.MediaWarning}
          </p>{/if}
        {#if call.State === "connected" && call.Stats?.ICECandidateType}
          <div class="row wrap">
            <button
              disabled={busy ||
                restarting[call.ID] ||
                call.Stats.ICERestartPending}
              onclick={() => restartPath(call.ID)}
              >Reconnect network audio</button
            >
            {#if restarting[call.ID] || call.Stats.ICERestartPending}<span
                class="muted small"
                role="status">Checking a new audio path…</span
              >{/if}
          </div>
          {#if restartErrors[call.ID]}<p class="error-text" role="alert">
              {restartErrors[call.ID]}
            </p>{/if}
        {/if}
        {#if call.Connected && !call.Connected.startsWith("0001-")}<p
            class="muted small"
          >
            Connected · {duration(call.Connected)}
          </p>{/if}
        {#if call.Error}<p class="error-text">{call.Error}</p>{/if}
        <div class="row wrap call-actions">
          {#if call.State === "ringing"}<button
              class="primary"
              disabled={busy}
              onclick={() => run(() => api.Answer(call.ID))}>Answer</button
            >{/if}
          {#if call.State === "connected"}<button
              class:active={call.Muted}
              disabled={busy}
              onclick={() => run(() => api.Mute(call.ID, !call.Muted))}
              aria-pressed={call.Muted}
              >{call.Muted ? "Unmute microphone" : "Mute microphone"}</button
            ><button
              class:active={call.Held}
              disabled={busy}
              onclick={() => run(() => api.Hold(call.ID, !call.Held))}
              aria-pressed={call.Held}
              >{call.Held ? "Resume call" : "Hold call"}</button
            >{/if}
          <button class="danger" onclick={() => run(() => api.Hangup(call.ID))}
            >{call.State === "ringing" ? "Decline" : "Hang up"}</button
          >
        </div>
        {#if call.State === "ringing"}<form
            class="row"
            onsubmit={(e) => {
              e.preventDefault();
              run(() => api.Redirect(call.ID, transfer[call.ID] || ""));
            }}
          >
            <input
              aria-label="Redirect destination"
              placeholder="Redirect destination"
              bind:value={transfer[call.ID]}
            /><button disabled={busy || !transfer[call.ID]}>Redirect</button>
          </form>{/if}
        {#if call.State === "connected"}
          <details class="call-details">
            <summary>Phone menu (DTMF)</summary>
            <form
              class="stack dtmf-form"
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
              <label for={`tone-${call.ID}`}>Menu digit</label>
              <div class="row">
                <input
                  id={`tone-${call.ID}`}
                  placeholder="0–9, * or #"
                  maxlength="1"
                  autocomplete="off"
                  bind:value={tones[call.ID]}
                />
                <button
                  disabled={busy ||
                    call.Held ||
                    !/^[0-9*#ABCD]$/i.test(tones[call.ID] || "")}
                  >Send digit</button
                >
              </div>
              <p class="muted small">
                Type a digit and press Enter to send it to the call’s phone
                menu.{call.Held ? " Resume the call first." : ""}
              </p>
              {#if toneNotices[call.ID]}<p class="small" role="status">
                  {toneNotices[call.ID]}
                </p>{/if}
            </form>
          </details>
          <details class="call-details">
            <summary>Transfer call</summary>
            <form
              class="row"
              onsubmit={(e) => {
                e.preventDefault();
                run(() => api.Transfer(call.ID, transfer[call.ID] || ""));
              }}
            >
              <input
                aria-label="Number or SIP address to transfer to"
                placeholder="Number or SIP address"
                bind:value={transfer[call.ID]}
              /><button disabled={busy || !transfer[call.ID]}
                >Transfer now</button
              >
            </form>
            <p class="muted small">
              To speak with the recipient first, hold this call and start a
              second call. Then choose that call below.
            </p>
            <div class="row wrap">
              <select
                aria-label="Consultation call"
                bind:value={consult[call.ID]}
                ><option value="">Choose consultation call</option
                >{#each snapshot.Calls.filter((c) => c.ID !== call.ID && c.Account === call.Account && c.State === "connected") as other}<option
                    value={other.ID}>{other.Remote}</option
                  >{/each}</select
              ><button
                disabled={busy || !consult[call.ID]}
                onclick={() =>
                  run(() => api.AttendedTransfer(call.ID, consult[call.ID]))}
                >Complete transfer</button
              >
            </div>
            <p class="muted small">
              Keep this call open until the transfer succeeds. Transfers require
              support from the phone service.
            </p>
          </details>{/if}
        {#if call.State === "connected"}<details class="call-details">
            <summary>Call volume & recording</summary><label
              >Microphone · {Math.round(
                (call.Stats.InputGain ?? 1) * 100,
              )}%<input
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
            ><label
              >Speaker · {Math.round((call.Stats.OutputGain ?? 1) * 100)}%<input
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
            ><button
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
            >{#if call.Stats.Recording}<p role="status">
                Recording · {call.Stats.RecordingPath}
              </p>{/if}
          </details>{/if}
        {#if showStats}<dl class="stats">
            {#each statistics(call.Stats || {}) as [key, value]}<div>
                <dt>{key}</dt>
                <dd>{value}</dd>
              </div>{/each}
          </dl>
          <details>
            <summary>All technical statistics</summary>
            <dl class="stats">
              {#each Object.entries(call.Stats || {}) as [key, value]}<div>
                  <dt>{key}</dt>
                  <dd>{value === null ? "Unavailable" : String(value)}</dd>
                </div>{/each}
            </dl>
          </details>{/if}
      </article>
    {:else}<div class="empty-call">
        <h3>No active calls</h3>
        <p class="muted">
          Type a number above to call. Incoming calls will appear here.
        </p>
      </div>{/each}
  </section>
  <section class="panel call-composer">
    <h2>New call</h2>
    <form
      class="stack"
      onsubmit={(event) => {
        event.preventDefault();
        if (!busy && account && target.trim())
          run(() => api.Dial(account, target.trim()));
      }}
    >
      <label for="call-destination">Name, phone number or SIP address</label>
      <div class="row dial-entry">
        <input
          id="call-destination"
          class="dial-input grow"
          placeholder="Type a number or search contacts"
          autocomplete="off"
          bind:value={target}
        />
        <button class="primary" disabled={busy || !account || !target.trim()}
          >Call</button
        >
      </div>
      <div class="row spread wrap">
        <p class="muted small">
          {account
            ? `Calling from ${account}. Press Enter to call.`
            : "Add and select an account to make a call."}
        </p>
        <button
          type="button"
          class="quiet"
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
      {#if redialNotice}<p class="muted small" role="status">
          {redialNotice}
        </p>{/if}
      {#if suggestions.length}
        <div
          class="contact-suggestions"
          aria-label={target ? "Matching contacts" : "Favorite contacts"}
        >
          <p class="muted small">
            {target ? "Matching contacts" : "Favorites"}
          </p>
          {#each suggestions as contact}
            <div class="list-item wrap">
              <button
                type="button"
                class="text-button grow"
                onclick={() => {
                  target = contact.Address;
                  if (contact.Account) account = contact.Account;
                }}
                ><strong>{contact.Name}</strong><span class="muted"
                  >{contact.Address}</span
                ></button
              >
              {#each contact.Numbers || [] as number}
                {#if target && number.Address !== contact.Address && number.Address.toLowerCase().includes(target.toLowerCase())}
                  <button
                    type="button"
                    onclick={() => {
                      target = number.Address;
                      if (contact.Account) account = contact.Account;
                    }}
                    >{number.Label || "Other number"} · {number.Address}</button
                  >
                {/if}
              {/each}
            </div>
          {/each}
        </div>
      {/if}
      {#if contactError}<p class="error-text" role="status">
          Contact search: {contactError}
        </p>{/if}
    </form>
  </section>
  <section class="phone-utilities" aria-label="Account services">
    {#if snapshot.Accounts.find((a) => a.Name === account)?.Features?.length}
      <details>
        <summary>Phone service shortcuts</summary>
        <div class="row wrap">
          {#each snapshot.Accounts.find((a) => a.Name === account).Features as feature}
            <button
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
    <details>
      <summary>Voicemail & availability</summary>
      <div class="row wrap">
        <button
          disabled={busy || !account}
          onclick={() => run(() => api.WatchVoicemail(account))}
          >Check for voicemail</button
        ><button
          disabled={busy || !account}
          onclick={() => run(() => api.DialVoicemail(account))}
          >Call voicemail</button
        >
      </div>
      {#each (snapshot.Presence || []).filter((p) => p.Account === account && p.Voicemail) as p}<p
        >
          {p.Voicemail.Waiting ? "New voicemail" : "No new voicemail"} · {p
            .Voicemail.New} new / {p.Voicemail.Old} old · {p.State}
        </p>{/each}
      <label
        >Availability message<input
          maxlength="512"
          bind:value={statusNote}
        /></label
      >
      <div class="row wrap">
        <button
          disabled={busy || !account}
          onclick={() =>
            run(() => api.PublishStatus(account, true, statusNote))}
          >Set available</button
        ><button
          disabled={busy || !account}
          onclick={() =>
            run(() => api.PublishStatus(account, false, statusNote))}
          >Set unavailable</button
        ><button
          disabled={busy || !account}
          onclick={() => run(() => api.UnpublishStatus(account))}
          >Stop publishing</button
        >
      </div>
      {#each snapshot.Accounts.filter((a) => a.Name === account && a.PresenceState) as a}
        <p class="small">
          Published status: {a.PresenceState}{a.PresenceNote
            ? ` · ${a.PresenceNote}`
            : ""}
        </p>
        {#if a.PresenceError}<p class="small error-text">
            {a.PresenceError}
          </p>{/if}
      {/each}
      <p class="muted small">
        Status refreshes while this account is enabled. Standard SIP publication
        requires server support. This does not set Swyx custom statuses.
      </p>
    </details>
    {#each snapshot.Accounts as a}<div class="account-status">
        <span class:online={a.State === "registered"} class="dot"></span><strong
          >{a.Name}</strong
        ><span class="muted">{a.State}</span>{#if a.Error}<p
            class="small error-text"
          >
            {a.Error}
          </p>{/if}
      </div>{/each}
  </section>
</div>
