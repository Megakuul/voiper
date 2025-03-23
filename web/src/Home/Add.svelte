<script>
  import Popup from "../lib/component/Popup.svelte";
  import { ListConfigs, SetConfig } from "../../wailsjs/go/app/App";
  import { Palette } from "../lib/color/color.svelte";

  /** @type {{ 
   * class: string
   * PostHook: function(): Promise<any>, 
   * ExceptionRef: string
   * }} */
  let {
    class: className,
    PostHook,
    ExceptionRef = $bindable(),
  } = $props()

  /** @type {boolean} */
  let popupState = $state(false)
  /** @type {string}*/
  let name = $state("")
  /** @type {string}*/
  let password = $state("")
  /** @type {import('wailsjs/go/models').config.Config} */
  let config = $state({})

  $effect(() => {
    if (!popupState) {
      name = ""
      password = ""
      config = {}
    }
  })

  /** 
   * @param {import('wailsjs/go/models').config.Config} config
   * @param {string} name
   * @param {string} key
   */
   async function add(config, name, key) {
    try {
      const configs = await ListConfigs()
      if (configs[name]) {
        throw `Config '${name}' does already exist`
      }
      await SetConfig(config, name, key)
      await PostHook()
    } catch (err) {
      ExceptionRef = err
    }
  }
</script>

<button style="color: {Palette.fgSecondary()}" title="add a new config" class={className} onclick={() => popupState = true}>
  <svg xmlns="http://www.w3.org/2000/svg" width="25" height="25" viewBox="0 0 48 48"><defs><mask id="ipTAdd0"><g fill="none" stroke="#fff" stroke-linejoin="round" stroke-width="4"><rect width="36" height="36" x="6" y="6" fill="#555555" rx="3"/><path stroke-linecap="round" d="M24 16v16m-8-8h16"/></g></mask></defs><path fill="currentColor" d="M0 0h48v48H0z" mask="url(#ipTAdd0)"/></svg>
  <span class="text-lg font-bold">New Config</span>
</button>

{#if popupState}
  <Popup title={"New Configuration"} onsubmit={() => add(config, name, password)} bind:StateRef={popupState}>
    <input placeholder="Name" bind:value={name} class="w-full my-2 py-2 px-3 bg-slate-700/20 rounded-lg focus:outline-0" />
    <input placeholder="Server" bind:value={config.Server} class="w-full my-2 py-2 px-3 bg-slate-700/20 rounded-lg focus:outline-0" />
    <input placeholder="Display Name" bind:value={config.DisplayName} class="w-full my-2 py-2 px-3 bg-slate-700/20 rounded-lg focus:outline-0" />
    <input placeholder="Username" bind:value={config.Username} class="w-full my-2 py-2 px-3 bg-slate-700/20 rounded-lg focus:outline-0" />
    <input placeholder="Password" bind:value={config.Password} class="w-full my-2 py-2 px-3 bg-slate-700/20 rounded-lg focus:outline-0" />

    <input type="password" placeholder="Password" bind:value={password} class="w-full my-2 py-2 px-3 bg-slate-700/20 rounded-lg focus:outline-0" />
  </Popup>
{/if}