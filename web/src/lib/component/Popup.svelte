<script>
  import { on } from "svelte/events";
  import { Palette } from "../color/color.svelte";
  import { fade } from "svelte/transition";

  /** @type {{ 
  * title: string
  * onsubmit: function(): Promise<void>
  * StateRef: boolean,
  * children: any,
  * }} */
  let {
    title,
    onsubmit,
    StateRef = $bindable(),
    children,
  } = $props();

  let loadingState = $state(false)

  async function submit() {
    loadingState = true;
    await onsubmit();
    loadingState = false;
    StateRef = false;
  }

  $effect(() => {
    if (!document) return

    const cleanupKeydown = on(document, "keydown", (e) => {
      if (e.key === "Escape") {
        setTimeout(() => StateRef = false, 0) // use timeout to ensure svelte reacts
      } else if (e.key === "Enter") {
        setTimeout(() => submit(), 0) // use timeout to ensure svelte reacts
      }
    })

    const cleanupWheel = on(document, "wheel", (e) => {
      e.preventDefault()
      return
    }, {passive: false})

    const cleanupTouch = on(document, "touchmove", (e) => {
      e.preventDefault()
      return
    }, {passive: false})

    return () => {
      cleanupKeydown()
      cleanupWheel()
      cleanupTouch()
    }
  })
</script>

<div transition:fade class="fixed z-50 top-0 left-0 h-full w-screen flex flex-col items-center justify-center bg-slate-950/80">
  <div style="background-color: {Palette.bgPrimary()}; border-color: {Palette.fgPrimary()};" 
    class="flex flex-col items-center gap-2 min-w-md p-6 rounded-2xl border-2 overflow-hidden">
    <h1 class="mt-1 text-2xl font-bold">{title}</h1>
    {@render children()}
    <div class="mt-auto w-full flex flex-row justify-around gap-6">
      <button class="w-1/2 py-2 bg-slate-700/30 rounded-lg hover:bg-slate-700/20 transition-all duration-500" 
        onclick={() => StateRef = false}>
        Cancel
      </button>
      <button class="w-1/2 flex flex-row items-center justify-center gap-2 py-2 bg-slate-700/50 rounded-lg hover:bg-slate-700/40 transition-all duration-500" 
        onclick={() => submit()}>
        Confirm
        {#if loadingState}
          <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24"><g stroke="currentColor" stroke-width="1"><circle cx="12" cy="12" r="9.5" fill="none" stroke-linecap="round" stroke-width="3"><animate attributeName="stroke-dasharray" calcMode="spline" dur="1.5s" keySplines="0.42,0,0.58,1;0.42,0,0.58,1;0.42,0,0.58,1" keyTimes="0;0.475;0.95;1" repeatCount="indefinite" values="0 150;42 150;42 150;42 150"/><animate attributeName="stroke-dashoffset" calcMode="spline" dur="1.5s" keySplines="0.42,0,0.58,1;0.42,0,0.58,1;0.42,0,0.58,1" keyTimes="0;0.475;0.95;1" repeatCount="indefinite" values="0;-16;-59;-59"/></circle><animateTransform attributeName="transform" dur="2s" repeatCount="indefinite" type="rotate" values="0 12 12;360 12 12"/></g></svg>
        {/if}
      </button>
    </div>
  </div>
</div>