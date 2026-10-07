<script>
  import { ui } from "../ui.js";
  import Icon from "../components/Icon.svelte";
  import Dropdown from "../components/Dropdown.svelte";
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
    class="audio-test-feedback grid gap-[0.8rem] p-4 mt-4 border border-[var(--border)] rounded-[0.6rem] empty:hidden [&_p]:m-0"
    aria-live="polite"
    aria-atomic="true"
  >
    {#if testing}
      <p class="m-0 leading-[1.6]">
        <strong
          >{testing === "speaker"
            ? "Listen for the test tone…"
            : testing === "loopback"
              ? "Speak — you should hear yourself…"
              : "Speak into your microphone now…"}</strong
        >
      </p>
      <div class="test-progress h-[6px] overflow-hidden rounded-[3px] bg-[var(--border)] [contain:paint]" role="progressbar" aria-label="Audio test in progress"><span class="
        block w-[35%] h-full bg-[var(--accent)] [transform:translateX(-100%)]
        animate-[testing-audio_1.6s_ease-in-out_infinite_alternate] motion-reduce:animate-none
        motion-reduce:[transform:translateX(90%)]
      "></span></div>
      <p class="m-0 leading-[1.6] text-muted text-[0.78rem]">The test ends automatically after five seconds.</p>
    {:else if result}
      {#if resultMode === "microphone" || resultMode === "loopback"}
        <p class="m-0 leading-[1.6]">
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
          class="appearance-none block w-full h-[0.8rem] border-0 rounded-[999px] overflow-hidden bg-[var(--border)] accent-[var(--accent)]"
          min="0"
          max="100"
          value={peak}
          aria-label="Peak microphone level during test"
        ></meter>
        <p class="m-0 leading-[1.6] text-muted text-[0.78rem]">
          Peak input: {peak}%. {peak < 3
            ? "Check mute and input volume in your desktop sound settings, then test again."
            : peak >= 98
              ? "Lower input volume in your desktop sound settings and test again."
              : "This is the highest level measured during the test."}
        </p>
      {:else}
        <p class="m-0 leading-[1.6]"><strong>Did you hear the test sound?</strong></p>
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
            type="button"
            class:primary={heardTone === true}
            onclick={() => (heardTone = true)}>Yes, I heard it</button
          >
          <button class={[ui.control, `
            px-4 py-[0.65rem] rounded-[6px] bg-[#35424e] text-inherit min-h-[38px] shadow-control font-medium border
            border-[#536271] [&.primary]:bg-brand [&.primary]:text-white [&.primary]:font-semibold
            [&.primary]:border-brand-border [&.primary:hover:enabled]:bg-brand-hover
            [&:hover:not(:disabled)]:bg-[#435362] [&:hover:not(:disabled)]:border-[#727e8d]
          `]}
            type="button"
            class:primary={heardTone === false}
            onclick={() => (heardTone = false)}>No sound</button
          >
        </div>
        {#if heardTone === true}<p class="m-0 leading-[1.6] text-muted text-[0.78rem]">
            Your selected output is ready.
          </p>
        {:else if heardTone === false}<p class="m-0 leading-[1.6] text-muted text-[0.78rem]">
            Check that your output is connected and unmuted, increase desktop
            volume, or choose another device above.
          </p>{/if}
      {/if}
      {#if result.CaptureOverruns || result.PlaybackUnderruns}
        <p class="m-0 leading-[1.6] text-muted text-[0.78rem]">
          Audio interruptions were reported. If sound was choppy, try another
          device or close other demanding applications.
        </p>
      {/if}
    {/if}
  </div>
{/snippet}

<section class="
  panel narrow audio-settings p-6 bg-surface shadow-panel border-t-[#526271] border-r-border
  border-b-border border-l-border rounded-[6px] min-w-0 max-w-[700px] border max-[700px]:p-[1.15rem]
  [&_>_p]:mb-4 [&_>_button_+_button]:mt-4 [&_form_+_details_button]:mt-2 [&_form_+_details_button]:mr-2
  [&_form_+_details_button]:mb-0 [&_form_+_details_button]:ml-0 [&_>_h2:first-child]:pb-4
  [&_>_h2:first-child]:border-b [&_>_h2:first-child]:border-b-border [&_>_.row_+_.row]:mt-4
">
  <header class="mb-6">
    <h2 class={ui.h2}>Sound</h2>
    <p class="mt-2 mx-0 mb-0 leading-[1.6] text-muted">
      Choose where you listen and speak. Test your selection before saving.
    </p>
  </header>
  {#if callCount > 0}
    <p class="audio-notice mb-4 mx-0 mt-0 leading-[1.6]" role="status">
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
    <fieldset class="border-0 p-0 m-0 min-w-0" onchange={changed}>
      <div class="audio-device py-5 border-t border-[var(--border)] [&_.dropdown-trigger]:mt-[0.4rem] [&>p]:mt-[0.65rem] [&>p]:mb-0">
        <div class="audio-device-heading flex items-center justify-between gap-4 mb-[0.9rem] flex-wrap">
          <h3 class="m-0 text-base font-semibold">Microphone</h3>
          <button class={[ui.control, `
            inline-flex items-center gap-2 px-4 py-[0.65rem] rounded-[6px] bg-[#35424e] text-inherit min-h-[38px]
            shadow-control font-medium border border-[#536271] [&:hover:not(:disabled)]:bg-[#435362]
            [&:hover:not(:disabled)]:border-[#727e8d]
          `]} type="button" disabled={locked} onclick={() => test("microphone")}
            ><Icon name="microphone" size={16} /> Test microphone</button
          >
        </div>
        <label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]" for="audio-input">Input device</label>
        <Dropdown
          id="audio-input"
          label="Input device"
          bind:value={audio.InputDevice}
          disabled={locked}
          onchange={changed}
          options={[
            { value: "", label: "System default microphone" },
            ...(audio.InputDevice &&
            !inputs.some((device) => device.ID === audio.InputDevice)
              ? [
                  {
                    value: audio.InputDevice,
                    label: "Saved microphone · unavailable",
                  },
                ]
              : []),
            ...inputs.map((device) => ({
              value: device.ID,
              label: device.Name,
            })),
          ]}
        />
        <p class="m-0 leading-[1.6] text-muted text-[0.78rem]">
          Speak normally for five seconds. Your voice is measured without being
          played back or saved.
        </p>
        {#if activeMode === "microphone"}{@render testFeedback()}{/if}
      </div>

      <div class="audio-device py-5 border-t border-[var(--border)] [&_.dropdown-trigger]:mt-[0.4rem] [&>p]:mt-[0.65rem] [&>p]:mb-0">
        <div class="audio-device-heading flex items-center justify-between gap-4 mb-[0.9rem] flex-wrap">
          <h3 class="m-0 text-base font-semibold">Speakers or headphones</h3>
          <button class={[ui.control, `
            inline-flex items-center gap-2 px-4 py-[0.65rem] rounded-[6px] bg-[#35424e] text-inherit min-h-[38px]
            shadow-control font-medium border border-[#536271] [&:hover:not(:disabled)]:bg-[#435362]
            [&:hover:not(:disabled)]:border-[#727e8d]
          `]} type="button" disabled={locked} onclick={() => test("speaker")}
            ><Icon name="speaker" size={16} /> Play test sound</button
          >
        </div>
        <label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]" for="audio-output">Output device</label>
        <Dropdown
          id="audio-output"
          label="Output device"
          bind:value={audio.OutputDevice}
          disabled={locked}
          onchange={changed}
          options={[
            { value: "", label: "System default output" },
            ...(audio.OutputDevice &&
            !outputs.some((device) => device.ID === audio.OutputDevice)
              ? [
                  {
                    value: audio.OutputDevice,
                    label: "Saved output · unavailable",
                  },
                ]
              : []),
            ...outputs.map((device) => ({
              value: device.ID,
              label: device.Name,
            })),
          ]}
        />
        <p class="m-0 leading-[1.6] text-muted text-[0.78rem]">
          Plays a quiet tone for five seconds. Adjust listening volume in your
          desktop sound settings.
        </p>
        {#if activeMode === "speaker"}{@render testFeedback()}{/if}
      </div>

      <div class="audio-device py-5 border-t border-[var(--border)] [&_.dropdown-trigger]:mt-[0.4rem] [&>p]:mt-[0.65rem] [&>p]:mb-0">
        <label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]" for="audio-ringer">Incoming call ringtone</label>
        <Dropdown
          id="audio-ringer"
          label="Incoming call ringtone"
          bind:value={audio.RingerDevice}
          disabled={locked}
          onchange={changed}
          options={[
            { value: "", label: "Same as speakers or headphones" },
            ...(audio.RingerDevice &&
            !outputs.some((device) => device.ID === audio.RingerDevice)
              ? [
                  {
                    value: audio.RingerDevice,
                    label: "Saved ringer · unavailable",
                  },
                ]
              : []),
            ...outputs.map((device) => ({
              value: device.ID,
              label: device.Name,
            })),
          ]}
        />
        <p class="m-0 leading-[1.6] text-muted text-[0.78rem]">
          Use a separate speaker to hear incoming calls while your headset is
          off.
        </p>
      </div>

      <details class={[ui.details, "audio-advanced mt-2 mb-5 border-t border-t-border pt-4"]}>
        <summary class={[ui.summary, "px-0 py-1"]}>Advanced sound settings</summary>
        <div class="stack mt-5 flex flex-col gap-4">
          <label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]"
            >Audio service
            <Dropdown
              label="Audio service"
              bind:value={audio.Backend}
              disabled={locked}
              options={[
                { value: "", label: "Automatic (recommended)" },
                { value: "pulse", label: "PipeWire / PulseAudio" },
                { value: "alsa", label: "ALSA" },
              ]}
              onchange={() => {
                audio.InputDevice = "";
                audio.OutputDevice = "";
                audio.RingerDevice = "";
                changed();
              }}
            />
          </label>
          <p class="m-0 leading-[1.6] text-muted text-[0.78rem]">
            Automatic prefers PipeWire or PulseAudio, so calls share sound with
            other apps. ALSA devices may be exclusive.
          </p>
          <label class="flex flex-row gap-[0.65rem] text-[0.9rem] text-[#c8cdd5] items-center min-h-[32px] cursor-pointer"
            ><input class={ui.checkbox} disabled={locked} type="checkbox" bind:checked={audio.EchoCancellation} /> Cancel
            speaker echo</label
          >
          <p class="m-0 leading-[1.6] text-muted text-[0.78rem]">
            Useful for speakerphone calls. Leave this off if your audio service
            already cancels echo.
          </p>
          <label class="flex flex-row gap-[0.65rem] text-[0.9rem] text-[#c8cdd5] items-center min-h-[32px] cursor-pointer"
            ><input class={ui.checkbox} disabled={locked} type="checkbox" bind:checked={audio.NoiseSuppression} /> Reduce
            steady background noise</label
          >
          <label class="flex flex-row gap-[0.65rem] text-[0.9rem] text-[#c8cdd5] items-center min-h-[32px] cursor-pointer"
            ><input class={ui.checkbox}
              type="checkbox"
              disabled={locked}
              checked={!audio.DisableAutoRecovery}
              onchange={(event) =>
                (audio.DisableAutoRecovery = !event.currentTarget.checked)}
            /> Reconnect audio devices automatically</label
          >
          <div class="audio-loopback grid justify-items-start gap-[0.8rem] pt-4 border-t border-[var(--border)] [&_h3]:m-0 [&_p]:m-0">
            <h3 class="m-0 text-base font-semibold">Listen to your microphone</h3>
            <p class="m-0 leading-[1.6] text-muted text-[0.78rem]">
              Hear yourself for five seconds. Use headphones to avoid feedback.
            </p>
            <label class="flex flex-row gap-[0.65rem] text-[0.9rem] text-[#c8cdd5] items-center min-h-[32px] cursor-pointer"
              ><input class={ui.checkbox} disabled={locked} type="checkbox" bind:checked={headphonesReady} /> I am wearing
              headphones</label
            >
            <button class={ui.button}
              type="button"
              disabled={locked || !headphonesReady}
              onclick={() => test("loopback")}>Listen to microphone</button
            >
            {#if activeMode === "loopback"}{@render testFeedback()}{/if}
          </div>
        </div>
      </details>
    </fieldset>

    <footer class="
      row pt-5 border-t border-[var(--border)] flex items-center gap-[0.7rem] flex-wrap
      max-[700px]:flex-wrap [&_>_button]:shrink-0 max-[700px]:[&_>_.grow]:basis-[180px]
    ">
      <button class={[ui.primaryButton, "primary"]} disabled={locked}>Save sound settings</button>
      <button class={ui.button}
        type="button"
        disabled={locked}
        onclick={() =>
          run(async () => {
            devices = await api.AudioDevices();
          })}>Refresh devices</button
      >
      {#if saved}<span class="text-muted text-[0.78rem]" role="status"
          >Sound settings saved</span
        >{/if}
    </footer>
  </form>
</section>

<style>
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
  @keyframes -global-testing-audio {
    to { transform: translateX(285%); }
  }
</style>
