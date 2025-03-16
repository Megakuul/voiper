<script>
  import { fade } from 'svelte/transition';
  import Config from './Config/Config.svelte';
  import Home from './Home/Home.svelte';
  import Phone from './Phone/Phone.svelte';
    import { Palette } from './lib/color/color.svelte';

  /** @type {string} */
  let Exception = $state(undefined);

  let Page = $state("home");
</script>


<div style="background-color: {Palette.bgPrimary()};">
  <div class="flex flex-row justify-center gap-5">
    <button onclick={() => Page = "home"}>Home</button>
    <button onclick={() => Page = "phone"}>Phone</button>
    <button onclick={() => Page = "config"}>Config</button>
  </div>
  
  {#if Page=="home"}
    <main transition:fade>
      <Home></Home>
    </main>
  {:else if Page=="phone"}
    <main transition:fade>
      <Phone></Phone>
    </main>
  {:else if Page=="config"}
    <main transition:fade>
      <Config></Config>
    </main>
  {/if}
  
  {#if Exception}
    <div>ALARM {Exception}</div>
  {/if}
</div>

