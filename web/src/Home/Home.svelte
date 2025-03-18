<script>
  import Logo from '../lib/component/Logo.svelte';
  import {AddConfig, EnableConfig, ListConfigs, RegisterSIP, RemoveConfig} from '../../wailsjs/go/app/App.js'
  import Spinner from '../lib/component/Spinner.svelte';
  import { fade } from 'svelte/transition';
  import { Palette } from '../lib/color/color.svelte';
  import Popup from '../lib/component/Popup.svelte';

  let {
    ExceptionRef = $bindable(),
    ...restProps
  } = $props();

  /** @type {Object.<string, boolean>} */
  let Configs = $state({});

  /** @type {boolean} */
  let enablePopupState = $state(false);

  async function list() {
    try {
      Configs = {}
      Configs = await ListConfigs()
    } catch (err) {
      ExceptionRef = err
    }
  }

  /** 
   * @param {import('wailsjs/go/models').config.Config} config
   * @param {string} name
   * @param {string} key
   */
  async function add(config, name, key) {
    try {
      await AddConfig(config, name, key)
      await list()
    } catch (err) {
      ExceptionRef = err
    }
  }

  /** 
   * @param {string} name
   * @param {boolean} encrypted
   */
   async function remove(name, encrypted) {
    try {
      await RemoveConfig(name, encrypted)
      await list()
    } catch (err) {
      ExceptionRef = err
    }
  }

  /**
   * @param {string} name
   * @param {string} key
   */
  async function enable(name, key) {
    try {
      await EnableConfig(name, key)
      await RegisterSIP()
    } catch (err) {
      ExceptionRef = err
    }
  }

  $effect(() => {
    list()
  })
</script>

<div transition:fade class="flex flex-col items-center h-full" {...restProps}>
  <Logo class="w-2/3 max-h-1/2 mb-4"></Logo>
  
  <div class="w-1/2 min-w-[400px] flex flex-col items-center gap-4">
    {#each Object.entries(Configs) as [path, encrypted]}
      <div transition:fade style="color: {Palette.fgSecondary()}; border-color: {Palette.fgPrimary()};"
        class="w-full flex flex-row items-center gap-2 bg-slate-800/20 rounded-2xl p-6 border-2">
        <h2 class="text-2xl mr-auto" title="{path}">{path.substring(path.lastIndexOf("/")+1)}</h2>
        <button title="delete" aria-label="delete" 
          class="hover:bg-slate-600/40 p-2 rounded-xl transition-all duration-700"
          onclick={() => remove(path, encrypted)}>
          <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24"><path fill="currentColor" d="M8 9h8v10H8z" opacity="0.3"/><path fill="currentColor" d="m15.5 4l-1-1h-5l-1 1H5v2h14V4zM6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6zM8 9h8v10H8z"/></svg>
        </button>
        <button title="edit" aria-label="edit" 
          class="hover:bg-slate-600/40 p-2 rounded-xl transition-all duration-700">
          <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24"><path fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8.8 20.199A2.73 2.73 0 0 1 6.869 21H3v-3.844c0-.724.288-1.419.8-1.931m5 4.974l-5-4.974m5 4.974l9.974-9.978M3.8 15.225l9.984-9.995m0 0l1.426-1.428a2.733 2.733 0 0 1 3.867-.001l1.126 1.127a2.733 2.733 0 0 1 0 3.865l-1.428 1.428M13.783 5.23l4.991 4.991"/></svg>
        </button>
        {#if encrypted}
          <button title="decrypt" aria-label="decrypt"
            class="hover:bg-slate-600/40 p-2 rounded-xl transition-all duration-700">
            <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24"><g fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2"><path d="M2.586 17.414A2 2 0 0 0 2 18.828V21a1 1 0 0 0 1 1h3a1 1 0 0 0 1-1v-1a1 1 0 0 1 1-1h1a1 1 0 0 0 1-1v-1a1 1 0 0 1 1-1h.172a2 2 0 0 0 1.414-.586l.814-.814a6.5 6.5 0 1 0-4-4z"/><circle cx="16.5" cy="7.5" r=".5" fill="currentColor"/></g></svg>
          </button>
        {:else}
          <button title="encrypt" aria-label="encrypt" 
            class="hover:bg-slate-600/40 p-2 rounded-xl transition-all duration-700">
            <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 56 56"><path fill="currentColor" d="M27.988 51.672c.375 0 .985-.14 1.594-.469c13.313-7.476 17.906-10.64 17.906-19.195v-17.93c0-2.46-1.078-3.234-3.047-4.078c-2.765-1.148-11.718-4.36-14.484-5.32c-.633-.211-1.289-.352-1.969-.352c-.656 0-1.312.188-1.945.352c-2.766.843-11.719 4.195-14.484 5.32c-1.97.82-3.047 1.617-3.047 4.078v17.93c0 8.554 4.617 11.695 17.906 19.195c.61.328 1.195.469 1.57.469m0-4.266c-.351 0-.75-.14-1.453-.562c-10.828-6.563-14.297-8.485-14.297-15.703V14.78c0-.797.164-1.101.797-1.36c3.563-1.405 10.43-3.843 14.04-5.132q.526-.21.913-.21c.282 0 .563.07.938.21c3.61 1.29 10.406 3.89 14.039 5.133c.633.234.797.562.797 1.36V31.14c0 7.218-3.492 9.117-14.297 15.703c-.703.422-1.102.562-1.477.562m-5.836-9.093h11.672c1.64 0 2.414-.82 2.414-2.579V26.64c0-1.547-.61-2.368-1.851-2.532V21.32c0-4.312-2.578-7.218-6.399-7.218c-3.797 0-6.398 2.906-6.398 7.218v2.813c-1.219.164-1.828.984-1.828 2.508v9.093c0 1.758.773 2.578 2.39 2.578m1.875-17.25c0-2.766 1.594-4.594 3.961-4.594c2.39 0 3.961 1.828 3.961 4.593v3l-7.922.024Z"/></svg>
          </button>
        {/if}
        <button title="enable" aria-label="enable" 
          class="hover:bg-slate-600/40 p-1 rounded-xl transition-all duration-700" 
          onclick={() => enablePopupState = true}>
          <svg xmlns="http://www.w3.org/2000/svg" width="40" height="40" viewBox="0 0 24 24"><path fill="currentColor" fill-opacity="0" stroke="currentColor" stroke-dasharray="40" stroke-dashoffset="40" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 6l10 6l-10 6Z"><animate fill="freeze" attributeName="fill-opacity" begin="0.5s" dur="0.15s" values="0;0.3"/><animate fill="freeze" attributeName="stroke-dashoffset" dur="0.5s" values="40;0"/></path></svg>
        </button>
      </div>

      {#if enablePopupState}
        <Popup title={"Enable Configuration?"} width={"500px"} height={"150px"} onsubmit={async () => {
          await new Promise(resolve => setTimeout(resolve, 5000))
        }} bind:StateHook={enablePopupState}>
          <span></span>
        </Popup>
      {/if}
    {/each}

    <div class="flex flex-row gap-6 w-full">
      <button style="color: {Palette.fgSecondary()}" title="add a new config"
        class="w-full flex flex-row items-center justify-center gap-2 bg-slate-600/10 hover:bg-slate-600/20 p-2 rounded-xl transition-all duration-700"
        onclick={async () => await list()}>
        <svg xmlns="http://www.w3.org/2000/svg" width="25" height="25" viewBox="0 0 48 48"><defs><mask id="ipTAdd0"><g fill="none" stroke="#fff" stroke-linejoin="round" stroke-width="4"><rect width="36" height="36" x="6" y="6" fill="#555555" rx="3"/><path stroke-linecap="round" d="M24 16v16m-8-8h16"/></g></mask></defs><path fill="currentColor" d="M0 0h48v48H0z" mask="url(#ipTAdd0)"/></svg>
        <span class="text-lg font-bold">New Config</span>
      </button>

      <button style="color: {Palette.fgSecondary()}" title="reload configs from disk"
        class="w-full flex flex-row items-center justify-center gap-2 bg-slate-600/10 hover:bg-slate-600/20 p-2 rounded-xl transition-all duration-700"
        onclick={async () => await list()}>
        <svg xmlns="http://www.w3.org/2000/svg" width="25" height="25" viewBox="0 0 14 14"><path fill="currentColor" fill-rule="evenodd" d="M8.457 4H9.75V3a1 1 0 0 0-1-1h-3.5a1 1 0 0 0-1 1v1a1 1 0 0 1-2 0V3a3 3 0 0 1 3-3h3.5a3 3 0 0 1 3 3v1h1.293a.5.5 0 0 1 .353.854l-2.293 2.292a.5.5 0 0 1-.707 0L8.104 4.854A.5.5 0 0 1 8.457 4M2.25 10H.957a.5.5 0 0 1-.353-.854l2.292-2.292a.5.5 0 0 1 .708 0l2.292 2.292a.5.5 0 0 1-.353.854H4.25v1a1 1 0 0 0 1 1h3.5a1 1 0 0 0 1-1v-1a1 1 0 1 1 2 0v1a3 3 0 0 1-3 3h-3.5a3 3 0 0 1-3-3z" clip-rule="evenodd"/></svg>
        <span class="text-lg font-bold">Reload</span>
      </button>
    </div>
  </div>
</div>

<style>
  
</style>