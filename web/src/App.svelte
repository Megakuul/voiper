<script>
  import Logs from './Logs/Logs.svelte';
  import Home from './Home/Home.svelte';
  import Phone from './Phone/Phone.svelte';
  import { Palette } from './lib/color/color.svelte';
  import Message from './lib/component/Message.svelte';

  let Page = $state("home");

  /** @type {string} */
  let Exception = $state(undefined);
</script>

<Message title={"Error"} color={Palette.fgError()} duration={5000} bind:MessageRef={Exception}></Message>

<center style="background-color: {Palette.bgPrimary()};" class="h-full box-border">
  <div class="flex flex-row justify-around gap-4 m-6 max-w-[1200px]">
    <button id="home-nav"
      style="color: {Palette.fgSecondary()}; border-color: {Palette.fgPrimary()};" 
      class="flex flex-row items-center gap-2 hover:bg-slate-600/10 py-3 px-20 rounded-2xl border-2 transition-all duration-700"
      onclick={() => Page = "home"}>
      <svg xmlns="http://www.w3.org/2000/svg" width="35" height="35" viewBox="0 0 24 24"><path fill="currentColor" fill-opacity="0" d="M5 8.5l7 -5.5l7 5.5v12.5h-4v-8l-1 -1h-4l-1 1v8h-4v-12.5Z"><animate fill="freeze" attributeName="fill-opacity" begin="1.1s" dur="0.15s" values="0;0.3"/></path><g fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2"><path stroke-dasharray="16" stroke-dashoffset="16" d="M4.5 21.5h15"><animate fill="freeze" attributeName="stroke-dashoffset" dur="0.2s" values="16;0"/></path><path stroke-dasharray="16" stroke-dashoffset="16" d="M4.5 21.5v-13.5M19.5 21.5v-13.5"><animate fill="freeze" attributeName="stroke-dashoffset" begin="0.2s" dur="0.2s" values="16;0"/></path><path stroke-dasharray="28" stroke-dashoffset="28" d="M2 10l10 -8l10 8"><animate fill="freeze" attributeName="stroke-dashoffset" begin="0.4s" dur="0.4s" values="28;0"/></path><path stroke-dasharray="24" stroke-dashoffset="24" d="M9.5 21.5v-9h5v9"><animate fill="freeze" attributeName="stroke-dashoffset" begin="0.7s" dur="0.4s" values="24;0"/></path></g></svg>
      <span class="font-bold">Home</span>
    </button>
    <button id="phone-nav"
      style="color: {Palette.fgSecondary()}; border-color: {Palette.fgPrimary()};"
      class="flex flex-row items-center gap-2 hover:bg-slate-600/10 py-3 px-20 rounded-2xl border-2 transition-all duration-700"
      onclick={() => Page = "phone"}>
      <svg xmlns="http://www.w3.org/2000/svg" width="35" height="35" viewBox="0 0 24 24"><path fill="currentColor" fill-opacity="0" stroke="currentColor" stroke-dasharray="64" stroke-dashoffset="64" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 3c0.5 0 2.5 4.5 2.5 5c0 1 -1.5 2 -2 3c-0.5 1 0.5 2 1.5 3c0.39 0.39 2 2 3 1.5c1 -0.5 2 -2 3 -2c0.5 0 5 2 5 2.5c0 2 -1.5 3.5 -3 4c-1.5 0.5 -2.5 0.5 -4.5 0c-2 -0.5 -3.5 -1 -6 -3.5c-2.5 -2.5 -3 -4 -3.5 -6c-0.5 -2 -0.5 -3 0 -4.5c0.5 -1.5 2 -3 4 -3Z"><animate fill="freeze" attributeName="fill-opacity" begin="0.7s" dur="0.15s" values="0;0.3"/><animate fill="freeze" attributeName="stroke-dashoffset" dur="0.6s" values="64;0"/></path></svg>
      <span class="font-bold">Phone</span>
    </button>
    <button id="logs-nav"
      style="color: {Palette.fgSecondary()}; border-color: {Palette.fgPrimary()};" 
      class="flex flex-row items-center gap-2 hover:bg-slate-600/10 py-3 px-20 rounded-2xl border-2 transition-all duration-700" 
      onclick={() => Page = "logs"}>
      <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"><g fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2"><path fill="currentColor" fill-opacity="0" stroke-dasharray="64" stroke-dashoffset="64" d="M22 3v14c0 0.55 -0.45 1 -1 1h-14c-0.55 0 -1 -0.45 -1 -1v-14c0 -0.55 0.45 -1 1 -1h14c0.55 0 1 0.45 1 1Z"><animate fill="freeze" attributeName="fill-opacity" begin="1.9s" dur="0.15s" values="0;0.3"/><animate fill="freeze" attributeName="stroke-dashoffset" dur="0.6s" values="64;0"/></path><path stroke-dasharray="10" stroke-dashoffset="10" d="M10 6h8"><animate fill="freeze" attributeName="stroke-dashoffset" begin="0.7s" dur="0.2s" values="10;0"/></path><path stroke-dasharray="10" stroke-dashoffset="10" d="M10 10h8"><animate fill="freeze" attributeName="stroke-dashoffset" begin="0.9s" dur="0.2s" values="10;0"/></path><path stroke-dasharray="6" stroke-dashoffset="6" d="M10 14h5"><animate fill="freeze" attributeName="stroke-dashoffset" begin="1.1s" dur="0.2s" values="6;0"/></path><path stroke-dasharray="36" stroke-dashoffset="36" d="M2 6v15c0 0.55 0.45 1 1 1h15"><animate fill="freeze" attributeName="stroke-dashoffset" begin="1.4s" dur="0.5s" values="36;0"/></path></g></svg>
      <span class="font-bold">Logs</span>
    </button>
  </div>
  
  {#if Page=="home"}
    <Home bind:ExceptionRef={Exception}></Home>
  {:else if Page=="phone"}
    <Phone bind:ExceptionRef={Exception}></Phone>
  {:else if Page=="logs"}
    <Logs bind:ExceptionRef={Exception}></Logs>
  {/if}
</center>

<style>
  #home-nav:hover svg {
    animation: jump-house 3s ease infinite;
  }

  @keyframes jump-house {
    0%, 25%, 35%, 50%, 60%, 80%, 90% {
      transform: translateY(0);
    }
    30% {
      transform: translateY(-2px) translateX(1px) rotateZ(2deg);
    }
    55% {
      transform: translateY(-3px) translateX(-1px) rotateZ(-2deg);
    }
    85% {
      transform: translateY(-1.5px) translateX(1px) rotateZ(2deg);
    }
  }

  #phone-nav:hover svg {
    animation: shake-phone 1s ease infinite;
  }

  @keyframes shake-phone {
    0%, 24%, 42%, 67%, 83% {
      transform: rotateZ(3deg);
    }
    10%, 35%, 50%, 74%, 92%, 100% {
      transform: rotateZ(-2deg);
    }
  }

  #logs-nav:hover svg {
    animation: rotate-logs 2s ease infinite;
  }

  @keyframes rotate-logs {
    0% {
      opacity: 1;
      transform: translateX(0) rotateZ(0deg);
    }
    50% {
      opacity: 0;
      transform: translateX(-10px) rotateZ(1deg);
    }
    51% {
      opacity: 0;
      transform: translateX(0) translateY(-10px) rotateZ(1deg);
    }
    100% {
      transform: translateX(0) translateY(0) rotateZ(0deg);
    }
  }
</style>

