<script>
  import { fade } from "svelte/transition";

  /** @type {{
   * color: string
   * time: Date
   * level: string
   * message: string
   * func: string
   * file: string
   * line: number
  }}*/
  let {
    color,
    time,
    level,
    message,
    func,
    file,
    line,
  } = $props()

  /** @type {boolean} */
  let collapsed = $state(true)
</script>

<div transition:fade style="color: {color}; border-color: {color};" 
  class="w-full flex flex-col gap-1 rounded-xl p-3 my-1.5 border-2 transition-all duration-700 overflow-hidden {collapsed ? "h-14" : "h-30"}">
  <div class="flex flex-row items-center gap-2">
    <span title="{time.toString()}" class="font-bold">
      {`${time.getHours().toString().padStart(2, '0')}:${time.getMinutes().toString().padStart(2, '0')}`}
    </span>
    <span class="font-bold">{level}</span>
    {#if collapsed}
      <span transition:fade class="text-nowrap overflow-hidden">{message}</span>
    {/if}
    <button title="open" aria-label="open"
      class="ml-auto p-1 rounded-md hover:bg-slate-700/20 transition-all duration-700" 
      onclick={() => collapsed = !collapsed} >
      <svg style="rotate: {collapsed ? "180deg" : "0deg"};" class="transition-all duration-700" xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"><g fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2"><path stroke-dasharray="16" stroke-dashoffset="16" d="M12 19l0 -13.5"><animate fill="freeze" attributeName="stroke-dashoffset" dur="0.2s" values="16;0"/></path><path stroke-dasharray="10" stroke-dashoffset="10" d="M12 5l5 5M12 5l-5 5"><animate fill="freeze" attributeName="stroke-dashoffset" begin="0.2s" dur="0.2s" values="10;0"/></path></g></svg>
    </button>
  </div>
  <p class="text-start text-wrap text-sm text-gray-200/80">
    {message}
  </p>
  <div class="flex flex-row items-center gap-2 text-nowrap">
    <span class="text-xs text-gray-700 underline">{file}:{line}</span>
    <span class="text-xs text-gray-700 underline">{func}</span>
  </div>
</div>