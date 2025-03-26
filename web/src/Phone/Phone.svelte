<script>
  import { fade } from "svelte/transition";
  let {
    ExceptionRef = $bindable(),
    ...restProps
  } = $props();


  let PhonebookEntries = [
    {
      DisplayName: "Dagobert Duck",
      Address: "187",
    }
  ]

</script>

<div transition:fade class="flex flex-col md:flex-row items-center h-full" {...restProps}>
  <div class="w-full h-full">
    Phone
  </div>
  <div>
    {#each PhonebookEntries as entry}
      
    {/each}
  </div>
</div>