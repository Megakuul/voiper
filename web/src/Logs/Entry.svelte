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

<button transition:fade style="color: {color}; border-color: {color};" 
  class="w-full flex flex-col gap-1 rounded-xl p-3 my-1.5 border-2 transition-all duration-700 overflow-hidden {collapsed ? "h-12" : "h-24"}"
  onclick={() => collapsed = !collapsed}>
  <div class="flex flex-row items-center gap-2">
    <span title="{time.toString()}" class="font-bold">
      {`${time.getHours().toString().padStart(2, '0')}:${time.getMinutes().toString().padStart(2, '0')}`}
    </span>
    <span class="font-bold">{level}</span>
    {#if collapsed}
      <span transition:fade class="text-nowrap overflow-hidden">{message}</span>
    {/if}
  </div>
  <p class="text-start text-wrap text-sm text-gray-200/80">
    {message}
  </p>
  <div class="flex flex-row items-center gap-2 text-nowrap">
    <span class="text-xs text-gray-700 underline">{file}:{line}</span>
    <span class="text-xs text-gray-700 underline">{func}</span>
  </div>
</button>