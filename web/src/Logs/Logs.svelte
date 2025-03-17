<script>
  import { fade } from "svelte/transition";
  let {
    ...restProps
  } = $props();
</script>

<div transition:fade class="flex flex-col items-center h-full" {...restProps}>
  <div>
    Logs
  </div>
</div>