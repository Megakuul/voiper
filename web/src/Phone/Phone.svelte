<script>
  import { ui } from "../ui.js";
  import Icon from "../components/Icon.svelte";
  import Dropdown from "../components/Dropdown.svelte";
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
  <section class="[&_>_div_>_h2]:m-0 [&:not(.has-calls)]:order-[1]" class:has-calls={snapshot.Calls.length > 0}>
    {#if snapshot.Calls.filter((c) => c.State === "connected").length >= 2 || hasConference}<div
        class="
          row flex items-center gap-[0.7rem] flex-wrap max-[700px]:flex-wrap [&_>_button]:shrink-0
          max-[700px]:[&_>_.grow]:basis-[180px]
        "
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
          >{:else}<p class="m-0 leading-[1.6] text-muted text-[0.78rem]">
            Merge two or three active calls. Resume held calls before merging.
          </p>{/if}
      </div>{/if}
    {#if snapshot.Calls.length || !workspace}<div class="
      row flex items-center gap-4 justify-between flex-wrap wrap-anywhere max-[700px]:flex-wrap
      [&_>_button]:shrink-0 [&_h2]:m-0 max-[700px]:[&_>_.grow]:basis-[180px]
    ">
      <h2 class={ui.h2}>{snapshot.Calls.length ? "Current calls" : "Calls"}</h2>
      {#if snapshot.Calls.length}<button
          class={[ui.quietButton, "quiet"]}
          aria-pressed={showStats}
          onclick={() => (showStats = !showStats)}
          >{showStats ? "Hide call statistics" : "Show call statistics"}</button
        >{/if}
    </div>
    {/if}
    {#each snapshot.Calls as call (call.ID)}
      <article class={["min-w-0 rounded-[6px] border border-border [&>p]:mb-4 [&>h2]:mt-4 [&>h2]:wrap-anywhere", workspace ? "p-5 mb-6 bg-surface-inset" : "p-6 mt-4 bg-surface shadow-panel border-t-[#526271] max-[700px]:p-[1.15rem]"]}>
        <span class="text-muted text-[0.78rem]"
          >{call.Account} · {call.Direction === "incoming"
            ? "Incoming call"
            : "Outgoing call"}</span
        >
        <h2 class={ui.h2}>{call.Remote}</h2>
        <div class="row flex items-center gap-[0.7rem] max-[700px]:flex-wrap [&_>_button]:shrink-0 max-[700px]:[&_>_.grow]:basis-[180px]">
          <span class="px-[0.7rem] py-[0.3rem] rounded-[4px] bg-[#435362] text-xs"
            >{call.Held
              ? "On hold"
              : {
                  connected: "Connected",
                  ringing: "Incoming call",
                  dialing: "Calling…",
                  connecting: "Connecting…",
                }[call.State] || call.State}</span
          >{#if call.Muted}<span class="px-[0.7rem] py-[0.3rem] rounded-[4px] bg-[#435362] text-xs">Microphone muted</span
            >{/if}{#if call.Stats?.Conference}<span class="px-[0.7rem] py-[0.3rem] rounded-[4px] bg-[#435362] text-xs"
              >Conference</span
            >{/if}
        </div>
        {#if call.TransferStatus}<p class="m-0 leading-[1.6]" role="status">
            Transfer: {call.TransferStatus}
          </p>{/if}
        {#if call.Stats?.AudioStopped}<p class="m-0 leading-[1.6] text-[#f9a8b9]" role="alert">
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
        {#if call.Stats?.LastError}<p class="m-0 leading-[1.6] text-[#f9a8b9]" role="alert">
            Audio: {call.Stats.LastError}
          </p>{/if}
        {#if call.Stats?.MediaWarning}<p class="m-0 leading-[1.6] text-[#f9a8b9]" role="status">
            {call.Stats.MediaWarning}
          </p>{/if}
        {#if call.State === "connected" && call.Stats?.ICECandidateType}
          <div class="
            row flex items-center gap-[0.7rem] flex-wrap max-[700px]:flex-wrap [&_>_button]:shrink-0
            max-[700px]:[&_>_.grow]:basis-[180px]
          ">
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
          {#if restartErrors[call.ID]}<p class="m-0 leading-[1.6] text-[#f9a8b9]" role="alert">
              {restartErrors[call.ID]}
            </p>{/if}
        {/if}
        {#if call.Connected && !call.Connected.startsWith("0001-")}<p
            class="m-0 leading-[1.6] text-muted text-[0.78rem]"
          >
            Connected · {duration(call.Connected)}
          </p>{/if}
        {#if call.Error}<p class="m-0 leading-[1.6] text-[#f9a8b9]">{call.Error}</p>{/if}
        <div class="
          row flex items-center gap-[0.7rem] flex-wrap mt-6 max-[700px]:flex-wrap [&_>_button]:shrink-0
          max-[700px]:[&_>_.grow]:basis-[180px]
        ">
          {#if call.State === "ringing"}<button
              class={[ui.primaryButton, "primary"]}
              disabled={busy}
              onclick={() => run(() => api.Answer(call.ID))}>Answer</button
            >{/if}
          {#if call.State === "connected"}<button class={[ui.control, `
            px-4 py-[0.65rem] rounded-[6px] bg-[#35424e] text-inherit min-h-[38px] shadow-control font-medium border
            border-[#536271] [&.active]:text-accent-text [&.active]:border-accent
            [&:hover:not(:disabled)]:bg-[#435362] [&:hover:not(:disabled)]:border-[#727e8d]
          `]}
              class:active={call.Muted}
              disabled={busy}
              onclick={() => run(() => api.Mute(call.ID, !call.Muted))}
              aria-pressed={call.Muted}
              >{call.Muted ? "Unmute microphone" : "Mute microphone"}</button
            ><button class={[ui.control, `
              px-4 py-[0.65rem] rounded-[6px] bg-[#35424e] text-inherit min-h-[38px] shadow-control font-medium border
              border-[#536271] [&.active]:text-accent-text [&.active]:border-accent
              [&:hover:not(:disabled)]:bg-[#435362] [&:hover:not(:disabled)]:border-[#727e8d]
            `]}
              class:active={call.Held}
              disabled={busy}
              onclick={() => run(() => api.Hold(call.ID, !call.Held))}
              aria-pressed={call.Held}
              >{call.Held ? "Resume call" : "Hold call"}</button
            >{/if}
          <button class={[ui.dangerButton, "danger"]} onclick={() => run(() => api.Hangup(call.ID))}
            >{call.State === "ringing" ? "Decline" : "Hang up"}</button
          >
        </div>
        {#if call.State === "ringing"}<form
            class="row flex items-center gap-[0.7rem] max-[700px]:flex-wrap [&_>_button]:shrink-0 max-[700px]:[&_>_.grow]:basis-[180px]"
            onsubmit={(e) => {
              e.preventDefault();
              run(() => api.Redirect(call.ID, transfer[call.ID] || ""));
            }}
          >
            <input class={[ui.input, "px-[0.8rem] py-[0.65rem]"]}
              aria-label="Redirect destination"
              placeholder="Redirect destination"
              bind:value={transfer[call.ID]}
            /><button class={ui.button} disabled={busy || !transfer[call.ID]}>Redirect</button>
          </form>{/if}
        {#if call.State === "connected"}
          <details class={[ui.details, "mt-5 border-t border-t-border pt-4"]}>
            <summary class={[ui.summary, "px-0 py-1"]}>Phone menu (DTMF)</summary>
            <form
              class="stack flex flex-col gap-3 items-stretch max-w-[440px] flex-wrap [&_>_label]:flex-1 [&_>_label]:min-w-[min(220px,_100%)]"
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
              <label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]" for={`tone-${call.ID}`}>Menu digit</label>
              <div class="row flex items-center gap-[0.7rem] max-[700px]:flex-wrap [&_>_button]:shrink-0 max-[700px]:[&_>_.grow]:basis-[180px]">
                <input class={[ui.input, "px-[0.8rem] py-[0.65rem]"]}
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
              <p class="m-0 leading-[1.6] text-muted text-[0.78rem]">
                Type a digit and press Enter to send it to the call’s phone
                menu.{call.Held ? " Resume the call first." : ""}
              </p>
              {#if toneNotices[call.ID]}<p class="m-0 leading-[1.6] text-[0.78rem]" role="status">
                  {toneNotices[call.ID]}
                </p>{/if}
            </form>
          </details>
          <details class={[ui.details, "mt-5 border-t border-t-border pt-4"]}>
            <summary class={[ui.summary, "px-0 py-1"]}>Transfer call</summary>
            <form
              class="row flex items-center gap-[0.7rem] max-[700px]:flex-wrap [&_>_button]:shrink-0 max-[700px]:[&_>_.grow]:basis-[180px]"
              onsubmit={(e) => {
                e.preventDefault();
                run(() => api.Transfer(call.ID, transfer[call.ID] || ""));
              }}
            >
              <input class={[ui.input, "px-[0.8rem] py-[0.65rem]"]}
                aria-label="Number or SIP address to transfer to"
                placeholder="Number or SIP address"
                bind:value={transfer[call.ID]}
              /><button class={ui.button} disabled={busy || !transfer[call.ID]}
                >Transfer now</button
              >
            </form>
            <p class="m-0 leading-[1.6] text-muted text-[0.78rem]">
              To speak with the recipient first, hold this call and start a
              second call. Then choose that call below.
            </p>
            <div class="
              row flex items-center gap-[0.7rem] flex-wrap max-[700px]:flex-wrap [&_>_button]:shrink-0
              max-[700px]:[&_>_.grow]:basis-[180px]
            ">
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
            <p class="m-0 leading-[1.6] text-muted text-[0.78rem]">
              Keep this call open until the transfer succeeds. Transfers require
              support from the phone service.
            </p>
          </details>{/if}
        {#if call.State === "connected"}<details class={[ui.details, "mt-5 border-t border-t-border pt-4"]}>
            <summary class={[ui.summary, "px-0 py-1"]}>Call volume & recording</summary><label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]"
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
            ><label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]"
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
            >{#if call.Stats.Recording}<p class="m-0 leading-[1.6]" role="status">
                Recording · {call.Stats.RecordingPath}
              </p>{/if}
          </details>{/if}
        {#if showStats}<dl class="
          text-xs mt-4 tabular-nums [&_div]:px-0 [&_div]:py-[0.6rem] [&_div]:flex [&_div]:justify-between
          [&_div]:gap-4 [&_div]:border-b [&_div]:border-b-border [&_dt]:text-muted [&_dd]:m-0
          [&_dd]:wrap-anywhere [&_dd]:text-right [&_dd]:min-w-0
        ">
            {#each statistics(call.Stats || {}) as [key, value]}<div>
                <dt>{key}</dt>
                <dd>{value}</dd>
              </div>{/each}
          </dl>
          <details class={[ui.details, "mt-[1.3rem] border-t border-t-border pt-4"]}>
            <summary class={[ui.summary, "px-0 py-1"]}>All technical statistics</summary>
            <dl class="
              text-xs mt-4 tabular-nums [&_div]:px-0 [&_div]:py-[0.6rem] [&_div]:flex [&_div]:justify-between
              [&_div]:gap-4 [&_div]:border-b [&_div]:border-b-border [&_dt]:text-muted [&_dd]:m-0
              [&_dd]:wrap-anywhere [&_dd]:text-right [&_dd]:min-w-0
            ">
              {#each Object.entries(call.Stats || {}) as [key, value]}<div>
                  <dt>{key}</dt>
                  <dd>{value === null ? "Unavailable" : String(value)}</dd>
                </div>{/each}
            </dl>
          </details>{/if}
      </article>
    {:else}<div class={["mt-4 py-8 text-center rounded-[6px]", workspace ? "text-muted [&>h3]:text-[0.9rem] [&>h3]:font-normal [&>p]:hidden" : "px-6 border border-border bg-surface-secondary"]}>
        <h3 class="m-0 text-base font-semibold">No active calls</h3>
        <p class="m-0 leading-[1.6] text-muted">
          Type a number above to call. Incoming calls will appear here.
        </p>
      </div>{/each}
  </section>
  <section class={["min-w-0 rounded-[6px] [&_.stack]:gap-[0.85rem]", workspace ? "pt-2 pb-6" : "p-6 border border-border border-t-[#526271] bg-surface shadow-panel max-[700px]:p-[1.15rem]"]}>
    <div class="flex items-center gap-4 mb-7 [&_h2]:mx-0 [&_h2]:mt-0 [&_h2]:mb-1 [&_h2]:text-[1.45rem]"><span class="grid place-items-center w-[48px] h-[48px] rounded-[10px] bg-accent-surface text-accent border border-brand-border"><Icon name="phone" size={24} /></span><div><h2 class={ui.h2}>{workspace ? "Make a call" : "New call"}</h2>{#if workspace}<p class="m-0 leading-[1.6] text-muted text-[0.78rem]">Enter a number, SIP address or contact name.</p>{/if}</div></div>
    <form
      class="stack flex flex-col gap-4"
      onsubmit={(event) => {
        event.preventDefault();
        if (!busy && account && target.trim())
          run(() => api.Dial(account, target.trim()));
      }}
    >
      <label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]" for="call-destination">Name, phone number or SIP address</label>
      <div class="
        row flex items-end gap-3 max-[700px]:flex-wrap [&_>_button]:shrink-0 [&_>_button]:min-h-[52px]
        [&_>_button]:flex [&_>_button]:items-center [&_>_button]:justify-center [&_>_button]:gap-2
        [&_>_label]:flex-1 [&_>_input]:flex-1 [&_>_input]:w-0 max-[700px]:[&_>_label]:basis-full
        max-[700px]:[&_>_input]:basis-full max-[700px]:[&_>_input]:w-full
        max-[700px]:[&_>_.grow]:basis-[180px]
      ">
        <input
          id="call-destination"
          class={[ui.input, "dial-input grow p-[0.8rem] flex-1 text-left text-[1.15rem]"]}
          placeholder="Type a number or search contacts"
          autocomplete="off"
          bind:value={target}
        />
        <button class={[ui.primaryButton, "primary"]} disabled={busy || !account || !target.trim()}
          ><Icon name="phone" size={18} /> Call</button
        >
      </div>
      <div class="
        row flex items-center gap-[0.7rem] justify-between flex-wrap max-[700px]:flex-wrap
        [&_>_button]:shrink-0 max-[700px]:[&_>_.grow]:basis-[180px]
      ">
        <p class="m-0 leading-[1.6] text-muted text-[0.78rem]">
          {account
            ? `Calling from ${account}. Press Enter to call.`
            : "Add and select an account to make a call."}
        </p>
        <button
          type="button"
          class={[ui.quietButton, "quiet"]}
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
      {#if redialNotice}<p class="m-0 leading-[1.6] text-muted text-[0.78rem]" role="status">
          {redialNotice}
        </p>{/if}
      {#if suggestions.length}
        <div
          class="max-h-[220px] overflow-y-auto"
          aria-label={target ? "Matching contacts" : "Favorite contacts"}
        >
          <p class="m-0 leading-[1.6] text-muted text-[0.78rem]">
            {target ? "Matching contacts" : "Favorites"}
          </p>
          {#each suggestions as contact}
            <div class="
              px-0 py-4 flex-wrap gap-[0.7rem] items-center border-b border-b-border [&_strong]:block
              [&_strong]:wrap-anywhere [&_strong]:mb-[0.3rem] [&_span]:block [&_span]:wrap-anywhere
              [&_button]:text-[0.8rem]
            ">
              <button
                type="button"
                class="
                  grow px-0 py-[0.35rem] outline-offset-3 cursor-pointer rounded-[6px] bg-transparent text-inherit
                  min-h-[38px] shadow-none font-medium leading-[1.4] transition-colors duration-150 flex-1 min-w-0
                  text-left border-0 border-[#536271] disabled:opacity-[0.45] disabled:cursor-not-allowed
                  [&:hover:not(:disabled)]:bg-[#435362] [&:hover:not(:disabled)]:border-[#727e8d]
                  [&:active:not(:disabled)]:[box-shadow:inset_0_1px_3px_#0004]
                "
                onclick={() => {
                  target = contact.Address;
                  if (contact.Account) account = contact.Account;
                }}
                ><strong>{contact.Name}</strong><span class="text-muted"
                  >{contact.Address}</span
                ></button
              >
              {#each contact.Numbers || [] as number}
                {#if target && number.Address !== contact.Address && number.Address.toLowerCase().includes(target.toLowerCase())}
                  <button class={ui.button}
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
      {#if contactError}<p class="m-0 leading-[1.6] text-[#f9a8b9]" role="status">
          Contact search: {contactError}
        </p>{/if}
    </form>
  </section>
  <section class={["order-2", workspace ? "border-t border-border [&>details]:border-t-0" : "grid grid-cols-[repeat(auto-fit,minmax(min(280px,100%),1fr))] gap-5"]} aria-label="Account services">
    {#if snapshot.Accounts.find((a) => a.Name === account)?.Features?.length}
      <details class={[ui.details, "mt-[1.3rem] border-t border-t-border pt-4"]}>
        <summary class={[ui.summary, "px-0 py-1"]}>Phone service shortcuts</summary>
        <div class="
          row flex items-center gap-[0.7rem] flex-wrap max-[700px]:flex-wrap [&_>_button]:shrink-0
          max-[700px]:[&_>_.grow]:basis-[180px]
        ">
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
    <details class={[ui.details, "mt-[1.3rem] border-t border-t-border pt-4"]}>
      <summary class={[ui.summary, "px-0 py-1"]}>Your status & voicemail</summary>
      <div class="
        row flex items-center gap-[0.7rem] flex-wrap max-[700px]:flex-wrap [&_>_button]:shrink-0
        max-[700px]:[&_>_.grow]:basis-[180px]
      ">
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
      {#each (snapshot.Presence || []).filter((p) => p.Account === account && p.Voicemail) as p}<p class="m-0 leading-[1.6]"
        >
          {p.Voicemail.Waiting ? "New voicemail" : "No new voicemail"} · {p
            .Voicemail.New} new / {p.Voicemail.Old} old · {p.State}
        </p>{/each}
      <label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]"
        >Your status message<input class={[ui.input, "px-[0.8rem] py-[0.65rem]"]}
          maxlength="512"
          bind:value={statusNote}
        /></label
      >
      <div class="
        row flex items-center gap-[0.7rem] flex-wrap max-[700px]:flex-wrap [&_>_button]:shrink-0
        max-[700px]:[&_>_.grow]:basis-[180px]
      ">
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
        {#if a.PresenceError}<p class="m-0 leading-[1.6] text-[0.78rem] text-[#f9a8b9]">
            {a.PresenceError}
          </p>{/if}
      {/each}
      <p class="m-0 leading-[1.6] text-muted text-[0.78rem]">
        Status refreshes while this account is enabled. Standard SIP publication
        requires server support. This does not set Swyx custom statuses.
      </p>
    </details>
  </section>
</div>
