<script>
  import { Palette } from "../color/color.svelte";
  import { fade } from "svelte/transition";

  let {
    title,
    duration,
    color,
    MessageRef = $bindable(),
  } = $props();

  /** @type {number} */
  let timeout = 0;

  $effect(() => {
    if (MessageRef) {
      clearTimeout(timeout)
      timeout = setTimeout(() => {
        MessageRef = undefined;
      }, duration)
    }
  })
</script>

{#if MessageRef}
  <div transition:fade class="fixed z-40 w-full bottom-5">
    <div style="background-color: {Palette.bgPrimary()}; border-color: {color}; color: {color};"
      class="flex flex-col items-start gap-1 mx-10 p-6 rounded-2xl border-2 overflow-hidden">
        <div class="w-full flex flex-row justify-between">
          <h1 class="text-2xl font-bold">{title}</h1>
          <button title="close" aria-label="close" 
            class="p-2 rounded-xl hover:bg-slate-500/40 transition-all duration-500" 
            onclick="{() => MessageRef = undefined}">
            <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"><path fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 5L12 5L19 5M5 12H19M5 19L12 19L19 19"><animate fill="freeze" attributeName="d" dur="0.4s" values="M5 5L12 5L19 5M5 12H19M5 19L12 19L19 19;M5 5L12 12L19 5M12 12H12M5 19L12 12L19 19"/></path></svg>
          </button>
        </div>
        <p class="w-full max-h-12 overflow-x-scroll text-start">{MessageRef}</p>
    </div>
  </div>
{/if}
