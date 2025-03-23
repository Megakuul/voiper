<script>
  import { fade } from "svelte/transition";

  /** @type {{
   * LogEntriesRef: Object[]
   * ExceptionRef: string
   * }} */
  let {
    LogEntriesRef = $bindable(),
    ExceptionRef = $bindable(),
    ...restProps
  } = $props();
</script>

<div transition:fade class="flex flex-col items-center h-full" {...restProps}>
  <div></div>
  <div class="bg-slate-700/20 rounded-xl m-8">
    Logs
    <div class="flex flex-col">
      {#each LogEntriesRef.slice(-50).reverse() as log}
        <span>{log}</span>
      {/each}
    </div>
  </div>

</div>