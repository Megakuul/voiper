<script>
  import { GetConfig, SetConfig } from "../../wailsjs/go/app/App";
  import Popup from "../lib/component/Popup.svelte";

  /** @type {{ 
   * class: string
   * name: string,
   * encrypted: boolean, 
   * PostHook: function(): Promise<any>,
   * ExceptionRef: string
   * }} */
  let {
    class: className,
    name,
    encrypted,
    PostHook,
    ExceptionRef = $bindable(),
  } = $props()

  /** @type {boolean} */
  let prePopupState = $state(false)
  /** @type {boolean} */
  let popupState = $state(false)
  /** @type {string}*/
  let password = $state("")
  /** @type {import('wailsjs/go/models').config.Config} */
  let config = $state(undefined)

  $effect(() => {
    if (!popupState) {
      password = ""
      config = undefined
    }
  })

  /** 
   * @param {string} name
   * @param {string} key
   */
   async function load(name, key) {
    try {
      config = await GetConfig(name, key)
      popupState = true;
    } catch (err) {
      ExceptionRef = err
    }
  }

  /** 
   * @param {import('wailsjs/go/models').config.Config} config
   * @param {string} name
   * @param {string} key
   */
   async function update(config, name, key) {
    try {
      await SetConfig(config, name, key)
      await PostHook()
    } catch (err) {
      ExceptionRef = err
    }
  }
</script>

<button title="edit" aria-label="edit" class={className} onclick={() => prePopupState = true}>
  <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24"><path fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8.8 20.199A2.73 2.73 0 0 1 6.869 21H3v-3.844c0-.724.288-1.419.8-1.931m5 4.974l-5-4.974m5 4.974l9.974-9.978M3.8 15.225l9.984-9.995m0 0l1.426-1.428a2.733 2.733 0 0 1 3.867-.001l1.126 1.127a2.733 2.733 0 0 1 0 3.865l-1.428 1.428M13.783 5.23l4.991 4.991"/></svg>
</button>

{#if prePopupState}
  <Popup title={"Unlock Configuration"} onsubmit={() => load(name, password)} bind:StateRef={prePopupState}>
    {#if encrypted}
      <input type="password" placeholder="Password" bind:value={password} class="w-full my-2 py-2 px-3 bg-slate-700/20 rounded-lg focus:outline-0" />
    {/if}
    <br>
  </Popup>
{:else if popupState}
  <Popup title={"Edit Configuration"} onsubmit={() => update(config, name, password)} bind:StateRef={popupState}>
    <input placeholder="Name" bind:value={name} class="w-full my-2 py-2 px-3 bg-slate-700/20 rounded-lg focus:outline-0" />
    <input placeholder="Server" bind:value={config.Server} class="w-full my-2 py-2 px-3 bg-slate-700/20 rounded-lg focus:outline-0" />
    <input placeholder="Display Name" bind:value={config.DisplayName} class="w-full my-2 py-2 px-3 bg-slate-700/20 rounded-lg focus:outline-0" />
    <input placeholder="Username" bind:value={config.Username} class="w-full my-2 py-2 px-3 bg-slate-700/20 rounded-lg focus:outline-0" />
    <input placeholder="Password" bind:value={config.Password} class="w-full my-2 py-2 px-3 bg-slate-700/20 rounded-lg focus:outline-0" />
  </Popup>
{/if}