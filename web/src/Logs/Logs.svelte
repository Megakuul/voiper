<script>
  import { Palette } from "../lib/color/color.svelte";
  import { fade } from "svelte/transition";
    import Entry from "./Entry.svelte";
    import { flip } from "svelte/animate";

  /** @type {{
   * LogEntriesRef: Object[]
   * ExceptionRef: string
   * }} */
  let {
    LogEntriesRef = $bindable(),
    ExceptionRef = $bindable(),
    ...restProps
  } = $props();

  /** @type {boolean} */
  let reversed = $state(false)

  /** @type {number} */
  let limit = $state(50)

  /** @type {boolean} */
  let debug = $state(true)
  /** @type {boolean} */
  let info = $state(true)
  /** @type {boolean} */
  let warn = $state(true)
  /** @type {boolean} */
  let error = $state(true)

  let exceptionBuffer = "";
  $effect(() => {
    ExceptionRef = exceptionBuffer
  })

  /** @param {string} entry @returns {any} */
  function parseEntry(entry) {
    try {
      return JSON.parse(entry)
    } catch (err) {
      exceptionBuffer = err
    }
    return undefined
  }
</script>

<div transition:fade class="flex flex-col items-center gap-6 h-full px-8 pt-10" {...restProps}>
  <div style="color: {Palette.fgSecondary()}; border-color: {Palette.fgPrimary()};"
    class="w-full flex flex-row gap-3 bg-slate-800/20 rounded-xl p-4 border-2">
    <button title="reverse" class="relative h-11 w-11 hover:bg-slate-900 rounded-xl transition-all duration-700" onclick={() => reversed = !reversed}>
      {#if reversed}
        <svg class="absolute top-1/2 left-1/2 translate-[-50%]" transition:fade xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"><g fill="none" stroke="currentColor" stroke-dasharray="12" stroke-dashoffset="12" stroke-linecap="round" stroke-linejoin="round" stroke-width="2"><path d="M12 2l-7 7M12 2l7 7"><animate fill="freeze" attributeName="stroke-dashoffset" dur="0.3s" values="12;0"/></path><path d="M12 8l-7 7M12 8l7 7"><animate fill="freeze" attributeName="stroke-dashoffset" begin="0.3s" dur="0.3s" values="12;0"/></path><path d="M12 14l-7 7M12 14l7 7"><animate fill="freeze" attributeName="stroke-dashoffset" begin="0.6s" dur="0.3s" values="12;0"/></path></g></svg>
      {:else}
        <svg class="absolute top-1/2 left-1/2 translate-[-50%]" transition:fade xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"><g fill="none" stroke="currentColor" stroke-dasharray="12" stroke-dashoffset="12" stroke-linecap="round" stroke-linejoin="round" stroke-width="2"><path d="M12 22l-7 -7M12 22l7 -7"><animate fill="freeze" attributeName="stroke-dashoffset" dur="0.3s" values="12;0"/></path><path d="M12 16l-7 -7M12 16l7 -7"><animate fill="freeze" attributeName="stroke-dashoffset" begin="0.3s" dur="0.3s" values="12;0"/></path><path d="M12 10l-7 -7M12 10l7 -7"><animate fill="freeze" attributeName="stroke-dashoffset" begin="0.6s" dur="0.3s" values="12;0"/></path></g></svg>
      {/if}
    </button>
    <button title="show 50" class:bg-slate-900={limit===50}
      class="h-11 w-13 hover:bg-slate-900 rounded-xl transition-all duration-700" 
      onclick={() => limit = 50}>
      <span class="font-extrabold">50</span>
    </button>
    <button title="show 200" class:bg-slate-900={limit===200}
      class="h-11 w-13 hover:bg-slate-900 rounded-xl transition-all duration-700" 
      onclick={() => limit = 200}>
      <span class="font-extrabold">200</span>
    </button>
    <button title="show all" class:bg-slate-900={limit===undefined}
      class="h-11 w-13 hover:bg-slate-900 rounded-xl transition-all duration-700" 
      onclick={() => limit = undefined}>
      <span class="font-extrabold">All</span>
    </button>

    <button title="show debug" class:bg-slate-900={debug}
      class="ml-auto h-11 w-20 hover:bg-slate-900 rounded-xl transition-all duration-700" 
      onclick={() => debug = !debug}>
      <span class="font-extrabold text-[#4a5565]">Debug</span>
    </button>
    <button title="show informational" class:bg-slate-900={info}
      class="h-11 w-20 hover:bg-slate-900 rounded-xl transition-all duration-700" 
      onclick={() => info = !info}>
      <span class="font-extrabold text-[#432dd7]">Info</span>
    </button>
    <button title="show warnings" class:bg-slate-900={warn}
      class="h-11 w-20 hover:bg-slate-900 rounded-xl transition-all duration-700" 
      onclick={() => warn = !warn}>
      <span class="font-extrabold text-[#bb4d00]">Warn</span>
    </button>
    <button title="show errors" class:bg-slate-900={error}
      class="h-11 w-20 hover:bg-slate-900 rounded-xl transition-all duration-700" 
      onclick={() => error = !error}>
      <span class="font-extrabold text-[#82181a]">Error</span>
    </button>
  </div>
  <div style="color: {Palette.fgSecondary()}; border-color: {Palette.fgPrimary()};"
    class="w-full h-9/12 bg-slate-700/20 rounded-xl p-6 border-2 overflow-y-scroll overflow-x-hidden">
    <div class="flex flex-col">
      {#each reversed ? LogEntriesRef.slice(0, limit).reverse() : LogEntriesRef.slice(0, limit) as log (log)}
        {@const entry = parseEntry(log)}
        <span animate:flip={{ duration: 600 }}>
          {#if debug && entry?.level === "DEBUG"}
            <Entry color="#4a5565" time={new Date(entry?.time)} level={entry?.level} message={entry?.msg}
              func={entry?.source?.function} file={entry?.source?.file} line={entry?.source?.line}>
            </Entry>
          {:else if info && entry?.level === "INFO"}
            <Entry color="#432dd7" time={new Date(entry?.time)} level={entry?.level} message={entry?.msg}
              func={entry?.source?.function} file={entry?.source?.file} line={entry?.source?.line}>
            </Entry>
          {:else if warn && entry?.level === "WARN"}
            <Entry color="#bb4d00" time={new Date(entry?.time)} level={entry?.level} message={entry?.msg}
              func={entry?.source?.function} file={entry?.source?.file} line={entry?.source?.line}>
            </Entry>
          {:else if error && entry?.level === "ERROR"}
            <Entry color="#82181a" time={new Date(entry?.time)} level={entry?.level} message={entry?.msg}
              func={entry?.source?.function} file={entry?.source?.file} line={entry?.source?.line}>
            </Entry>
          {/if}
        </span>
      {/each}
    </div>
  </div>
  <button style="color: {Palette.fgSecondary()}" title="clear collected logs"
    class="w-full flex flex-row items-center justify-center gap-2 bg-slate-600/10 hover:bg-slate-600/20 p-2 rounded-xl transition-all duration-700"
    onclick={() => LogEntriesRef = []}>
    <svg xmlns="http://www.w3.org/2000/svg" width="28" height="28" viewBox="0 0 48 48"><defs><mask id="ipTClearFormat0"><g fill="none" stroke="#fff"><path fill="#555555" stroke-linejoin="round" stroke-width="4.302" d="M44.782 24.17L31.918 7.1L14.135 20.5L27.5 37l3.356-2.336z"/><path stroke-linejoin="round" stroke-width="4.302" d="m27.5 37l-3.839 3.075l-10.563-.001l-2.6-3.45l-6.433-8.536L14.5 20.225"/><path stroke-linecap="round" stroke-width="4.5" d="M13.206 40.072h31.36"/></g></mask></defs><path fill="currentColor" d="M0 0h48v48H0z" mask="url(#ipTClearFormat0)"/></svg>
    <span class="text-lg font-bold">Clear</span>
  </button>
</div>