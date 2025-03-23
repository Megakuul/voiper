<script>
  import Logo from '../lib/component/Logo.svelte';
  import { ListConfigs } from '../../wailsjs/go/app/App.js'
  import { fade } from 'svelte/transition';
  import { Palette } from '../lib/color/color.svelte';
  import Enable from './Enable.svelte';
  import Add from './Add.svelte';
  import Remove from './Remove.svelte';
  import Edit from './Edit.svelte';
  import Crypt from './Crypt.svelte';
    import { flip } from 'svelte/animate';

  let {
    ExceptionRef = $bindable(),
    ...restProps
  } = $props();

  /** @type {Object.<string, boolean>} */
  let Configs = $state({});

  async function list() {
    try {
      Configs = {}
      Configs = await ListConfigs()
    } catch (err) {
      ExceptionRef = err
    }
  }

  $effect(() => {
    list()
  })
</script>

<div transition:fade class="flex flex-col items-center min-h-full" {...restProps}>
  <Logo class="w-7/12 max-h-1/3 mb-4"></Logo>
  
  <div class="w-1/2 min-w-[400px] flex flex-col items-center gap-4 mb-5">
    {#each Object.entries(Configs) as [name, encrypted] (name)}
      <div transition:fade animate:flip style="color: {Palette.fgSecondary()}; border-color: {Palette.fgPrimary()};"
        class="w-full flex flex-row items-center gap-2 bg-slate-800/20 rounded-2xl p-6 border-2">
        <h2 class="text-2xl mr-auto">{name}</h2>
        <Remove name={name} encrypted={encrypted} PostHook={list} bind:ExceptionRef={ExceptionRef}
          class="hover:bg-slate-600/40 p-2 rounded-xl transition-all duration-700">
        </Remove>
        <Edit name={name} encrypted={encrypted} PostHook={list} bind:ExceptionRef={ExceptionRef}
          class="hover:bg-slate-600/40 p-2 rounded-xl transition-all duration-700">
        </Edit>
        <Crypt name={name} encrypted={encrypted} PostHook={list} bind:ExceptionRef={ExceptionRef}
          class="hover:bg-slate-600/40 p-2 rounded-xl transition-all duration-700">
        </Crypt>
        <Enable name={name} encrypted={encrypted} bind:ExceptionRef={ExceptionRef} 
          class="hover:bg-slate-600/40 p-1 rounded-xl transition-all duration-700">
        </Enable>
      </div>
    {/each}
  </div>
  <div class="w-1/2 min-w-[400px] flex flex-row gap-6 mb-5">
    <Add PostHook={list} bind:ExceptionRef={ExceptionRef} 
      class="w-full flex flex-row items-center justify-center gap-2 bg-slate-600/10 hover:bg-slate-600/20 p-2 rounded-xl transition-all duration-700">
    </Add>
    <button style="color: {Palette.fgSecondary()}" title="reload configs from disk"
      class="w-full flex flex-row items-center justify-center gap-2 bg-slate-600/10 hover:bg-slate-600/20 p-2 rounded-xl transition-all duration-700"
      onclick={async () => await list()}>
      <svg xmlns="http://www.w3.org/2000/svg" width="25" height="25" viewBox="0 0 14 14"><path fill="currentColor" fill-rule="evenodd" d="M8.457 4H9.75V3a1 1 0 0 0-1-1h-3.5a1 1 0 0 0-1 1v1a1 1 0 0 1-2 0V3a3 3 0 0 1 3-3h3.5a3 3 0 0 1 3 3v1h1.293a.5.5 0 0 1 .353.854l-2.293 2.292a.5.5 0 0 1-.707 0L8.104 4.854A.5.5 0 0 1 8.457 4M2.25 10H.957a.5.5 0 0 1-.353-.854l2.292-2.292a.5.5 0 0 1 .708 0l2.292 2.292a.5.5 0 0 1-.353.854H4.25v1a1 1 0 0 0 1 1h3.5a1 1 0 0 0 1-1v-1a1 1 0 1 1 2 0v1a3 3 0 0 1-3 3h-3.5a3 3 0 0 1-3-3z" clip-rule="evenodd"/></svg>
      <span class="text-lg font-bold">Reload</span>
    </button>
  </div>
</div>