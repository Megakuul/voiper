<script>
  import { tick } from "svelte";
  import * as api from "../../wailsjs/go/app/App.js";

  let {
    audio = $bindable(),
    devices = $bindable(),
    busy,
    callCount,
    run,
  } = $props();
  let testing = $state("");
  let result = $state(null);
  let resultMode = $state("");
  let heardTone = $state(null);
  let headphonesReady = $state(false);
  let saved = $state(false);
  let feedback;
  const locked = $derived(busy || Boolean(testing) || callCount > 0);
  const inputs = $derived(
    devices.filter(
      (d) =>
        (d.Kind === "capture" || d.Kind === "input") &&
        (!audio.Backend || d.Backend === audio.Backend),
    ),
  );
  const outputs = $derived(
    devices.filter(
      (d) =>
        (d.Kind === "playback" || d.Kind === "output") &&
        (!audio.Backend || d.Backend === audio.Backend),
    ),
  );
  const activeMode = $derived(testing || (result ? resultMode : ""));
  const peak = $derived(Math.round(Math.min(1, result?.InputPeak || 0) * 100));

  function changed() {
    result = null;
    heardTone = null;
    saved = false;
  }
  async function test(mode) {
    if (locked) return;
    testing = mode;
    result = null;
    heardTone = null;
    await tick();
    feedback?.scrollIntoView({ block: "nearest", behavior: "smooth" });
    await run(async () => {
      try {
        result = await api.TestAudio({ ...audio }, mode);
        resultMode = mode;
      } finally {
        testing = "";
      }
    });
  }
</script>

{#snippet testFeedback()}
  <div
    class="audio-test-feedback"
    bind:this={feedback}
    aria-live="polite"
    aria-atomic="true"
  >
    {#if testing}
      <p>
        <strong
          >{testing === "speaker"
            ? "Listen for the test tone…"
            : testing === "loopback"
              ? "Speak — you should hear yourself…"
              : "Speak into your microphone now…"}</strong
        >
      </p>
      <progress aria-label="Audio test in progress"></progress>
      <p class="muted small">The test ends automatically after five seconds.</p>
    {:else if result}
      {#if resultMode === "microphone" || resultMode === "loopback"}
        <p>
          <strong
            >{!result.InputPeak
              ? "No microphone sound detected"
              : peak >= 98
                ? "Your microphone may be too loud"
                : peak < 3
                  ? "Your microphone is very quiet"
                  : "Microphone sound detected"}</strong
          >
        </p>
        <meter
          min="0"
          max="100"
          value={peak}
          aria-label="Peak microphone level during test"
        ></meter>
        <p class="muted small">
          Peak input: {peak}%. {peak < 3
            ? "Check mute and input volume in your desktop sound settings, then test again."
            : peak >= 98
              ? "Lower input volume in your desktop sound settings and test again."
              : "This is the highest level measured during the test."}
        </p>
      {:else}
        <p><strong>Did you hear the test sound?</strong></p>
        <div class="row wrap">
          <button
            type="button"
            class:primary={heardTone === true}
            onclick={() => (heardTone = true)}>Yes, I heard it</button
          >
          <button
            type="button"
            class:primary={heardTone === false}
            onclick={() => (heardTone = false)}>No sound</button
          >
        </div>
        {#if heardTone === true}<p class="muted small">
            Your selected output is ready.
          </p>
        {:else if heardTone === false}<p class="muted small">
            Check that your output is connected and unmuted, increase desktop
            volume, or choose another device above.
          </p>{/if}
      {/if}
      {#if result.CaptureOverruns || result.PlaybackUnderruns}
        <p class="muted small">
          Audio interruptions were reported. If sound was choppy, try another
          device or close other demanding applications.
        </p>
      {/if}
    {/if}
  </div>
{/snippet}

<section class="panel narrow audio-settings">
  <header>
    <h2>Sound</h2>
    <p class="muted">
      Choose where you listen and speak. Test your selection before saving.
    </p>
  </header>
  {#if callCount > 0}
    <p class="audio-notice" role="status">
      End your calls to change or test audio devices.
    </p>
  {/if}
  <form
    onsubmit={(event) => {
      event.preventDefault();
      run(async () => {
        await api.SetAudio(audio);
        saved = true;
      });
    }}
  >
    <fieldset disabled={locked} onchange={changed}>
      <div class="audio-device">
        <div class="audio-device-heading">
          <h3>Microphone</h3>
          <button type="button" onclick={() => test("microphone")}
            >Test microphone</button
          >
        </div>
        <label for="audio-input">Input device</label>
        <select id="audio-input" bind:value={audio.InputDevice}>
          <option value="">System default microphone</option>
          {#if audio.InputDevice && !inputs.some((d) => d.ID === audio.InputDevice)}
            <option value={audio.InputDevice}
              >Saved microphone · unavailable</option
            >
          {/if}
          {#each inputs as device}<option value={device.ID}
              >{device.Name}</option
            >{/each}
        </select>
        <p class="muted small">
          Speak normally for five seconds. Your voice is measured without being
          played back or saved.
        </p>
        {#if activeMode === "microphone"}{@render testFeedback()}{/if}
      </div>

      <div class="audio-device">
        <div class="audio-device-heading">
          <h3>Speakers or headphones</h3>
          <button type="button" onclick={() => test("speaker")}
            >Play test sound</button
          >
        </div>
        <label for="audio-output">Output device</label>
        <select id="audio-output" bind:value={audio.OutputDevice}>
          <option value="">System default output</option>
          {#if audio.OutputDevice && !outputs.some((d) => d.ID === audio.OutputDevice)}
            <option value={audio.OutputDevice}
              >Saved output · unavailable</option
            >
          {/if}
          {#each outputs as device}<option value={device.ID}
              >{device.Name}</option
            >{/each}
        </select>
        <p class="muted small">
          Plays a quiet tone for five seconds. Adjust listening volume in your
          desktop sound settings.
        </p>
        {#if activeMode === "speaker"}{@render testFeedback()}{/if}
      </div>

      <div class="audio-device">
        <label for="audio-ringer">Incoming call ringtone</label>
        <select id="audio-ringer" bind:value={audio.RingerDevice}>
          <option value="">Same as speakers or headphones</option>
          {#if audio.RingerDevice && !outputs.some((d) => d.ID === audio.RingerDevice)}
            <option value={audio.RingerDevice}
              >Saved ringer · unavailable</option
            >
          {/if}
          {#each outputs as device}<option value={device.ID}
              >{device.Name}</option
            >{/each}
        </select>
        <p class="muted small">
          Use a separate speaker to hear incoming calls while your headset is
          off.
        </p>
      </div>

      <details class="audio-advanced">
        <summary>Advanced sound settings</summary>
        <div class="stack">
          <label
            >Audio service
            <select
              bind:value={audio.Backend}
              onchange={() => {
                audio.InputDevice = "";
                audio.OutputDevice = "";
                audio.RingerDevice = "";
              }}
            >
              <option value="">Automatic (recommended)</option>
              <option value="pulse">PipeWire / PulseAudio</option>
              <option value="alsa">ALSA</option>
            </select>
          </label>
          <p class="muted small">
            Automatic prefers PipeWire or PulseAudio, so calls share sound with
            other apps. ALSA devices may be exclusive.
          </p>
          <label class="toggle"
            ><input type="checkbox" bind:checked={audio.EchoCancellation} /> Cancel
            speaker echo</label
          >
          <p class="muted small">
            Useful for speakerphone calls. Leave this off if your audio service
            already cancels echo.
          </p>
          <label class="toggle"
            ><input type="checkbox" bind:checked={audio.NoiseSuppression} /> Reduce
            steady background noise</label
          >
          <label class="toggle"
            ><input
              type="checkbox"
              checked={!audio.DisableAutoRecovery}
              onchange={(event) =>
                (audio.DisableAutoRecovery = !event.currentTarget.checked)}
            /> Reconnect audio devices automatically</label
          >
          <div class="audio-loopback">
            <h3>Listen to your microphone</h3>
            <p class="muted small">
              Hear yourself for five seconds. Use headphones to avoid feedback.
            </p>
            <label class="toggle"
              ><input type="checkbox" bind:checked={headphonesReady} /> I am wearing
              headphones</label
            >
            <button
              type="button"
              disabled={!headphonesReady}
              onclick={() => test("loopback")}>Listen to microphone</button
            >
            {#if activeMode === "loopback"}{@render testFeedback()}{/if}
          </div>
        </div>
      </details>
    </fieldset>

    <footer class="row wrap">
      <button class="primary" disabled={locked}>Save sound settings</button>
      <button
        type="button"
        disabled={locked}
        onclick={() =>
          run(async () => {
            devices = await api.AudioDevices();
          })}>Refresh devices</button
      >
      {#if saved}<span class="muted small" role="status"
          >Sound settings saved</span
        >{/if}
    </footer>
  </form>
</section>

<style>
  .audio-settings header {
    margin-bottom: 1.5rem;
  }
  .audio-settings header p {
    margin-top: 0.5rem;
  }
  fieldset {
    border: 0;
    padding: 0;
    margin: 0;
    min-width: 0;
  }
  .audio-device {
    padding: 1.25rem 0;
    border-top: 1px solid var(--border);
  }
  .audio-device-heading {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
    margin-bottom: 0.9rem;
    flex-wrap: wrap;
  }
  .audio-device-heading h3 {
    margin: 0;
  }
  .audio-device select {
    width: 100%;
    margin-top: 0.4rem;
  }
  .audio-device p {
    margin: 0.65rem 0 0;
  }
  .audio-advanced {
    margin: 0.5rem 0 1.25rem;
  }
  .audio-advanced .stack {
    margin-top: 1.25rem;
  }
  .audio-loopback {
    display: grid;
    justify-items: start;
    gap: 0.8rem;
    padding-top: 1rem;
    border-top: 1px solid var(--border);
  }
  .audio-loopback h3,
  .audio-loopback p {
    margin: 0;
  }
  .audio-test-feedback:empty {
    display: none;
  }
  .audio-test-feedback {
    display: grid;
    gap: 0.8rem;
    padding: 1rem;
    margin-top: 1rem;
    border: 1px solid var(--border);
    border-radius: 0.6rem;
  }
  .audio-test-feedback p {
    margin: 0;
  }
  progress,
  meter {
    appearance: none;
    -webkit-appearance: none;
    display: block;
    width: 100%;
    height: 0.8rem;
    border: 0;
    border-radius: 999px;
    overflow: hidden;
    background: var(--border);
    accent-color: var(--accent);
  }
  meter::-webkit-meter-bar {
    height: 0.8rem;
    border: 0;
    border-radius: 999px;
    background: var(--border);
    box-shadow: none;
  }
  meter::-webkit-meter-optimum-value {
    border-radius: 999px;
    background: var(--accent);
  }
  meter::-moz-meter-bar {
    border-radius: 999px;
    background: var(--accent);
  }
  progress:indeterminate {
    background: linear-gradient(
      90deg,
      var(--border) 40%,
      var(--accent) 45%,
      var(--accent) 55%,
      var(--border) 60%
    );
    background-size: 250% 100%;
    animation: testing-audio 1.6s ease-in-out infinite alternate;
  }
  progress::-webkit-progress-bar {
    border-radius: 999px;
    background: inherit;
  }
  progress:indeterminate::-webkit-progress-value {
    background: transparent;
  }
  progress:indeterminate::-moz-progress-bar {
    background: transparent;
  }
  @keyframes testing-audio {
    from {
      background-position: 100% 0;
    }
    to {
      background-position: 0 0;
    }
  }
  @media (prefers-reduced-motion: reduce) {
    progress:indeterminate {
      animation: none;
      background-position: 50% 0;
    }
  }
  .audio-notice {
    margin-bottom: 1rem;
  }
  footer {
    padding-top: 1.25rem;
    border-top: 1px solid var(--border);
  }
</style>
