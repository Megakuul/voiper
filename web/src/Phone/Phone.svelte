<script>
  import { fade } from "svelte/transition";
  let {
    ExceptionRef = $bindable(),
    ...restProps
  } = $props();
</script>

<div transition:fade class="flex flex-col items-center h-full" {...restProps}>
  <div>
    Phone
  </div>
</div>