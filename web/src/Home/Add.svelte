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
    <span class="flex flex-row items-center gap-2 w-full">
      <input placeholder="Name" bind:value={name} class="w-full my-2 py-2 px-3 bg-slate-700/20 rounded-lg focus:outline-0" />
      <button title="Name of the config (press to load example)" aria-label="name" class="text-slate-400" onclick={() => name = "Dagobert"}>
        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"><g fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2"><path stroke-dasharray="64" stroke-dashoffset="64" d="M12 3c4.97 0 9 4.03 9 9c0 4.97 -4.03 9 -9 9c-4.97 0 -9 -4.03 -9 -9c0 -4.97 4.03 -9 9 -9Z"><animate fill="freeze" attributeName="stroke-dashoffset" dur="0.6s" values="64;0"/></path><path stroke-dasharray="8" stroke-dashoffset="8" d="M12 7v6"><animate fill="freeze" attributeName="stroke-dashoffset" begin="0.6s" dur="0.2s" values="8;0"/></path><path stroke-dasharray="2" stroke-dashoffset="2" d="M12 17v0.01"><animate fill="freeze" attributeName="stroke-dashoffset" begin="0.8s" dur="0.2s" values="2;0"/></path></g></svg>
      </button>
    </span>
    
    <span class="flex flex-row items-center gap-2 w-full">
      <input placeholder="SIP Server" bind:value={config.Server} class="w-full my-2 py-2 px-3 bg-slate-700/20 rounded-lg focus:outline-0" />
      <button title="Location of the SIP server (press to load example)" aria-label="sip server" class="text-slate-400" onclick={() => config.Server = "10.1.1.100:5060"}>
        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"><g fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2"><path stroke-dasharray="64" stroke-dashoffset="64" d="M12 3c4.97 0 9 4.03 9 9c0 4.97 -4.03 9 -9 9c-4.97 0 -9 -4.03 -9 -9c0 -4.97 4.03 -9 9 -9Z"><animate fill="freeze" attributeName="stroke-dashoffset" dur="0.6s" values="64;0"/></path><path stroke-dasharray="8" stroke-dashoffset="8" d="M12 7v6"><animate fill="freeze" attributeName="stroke-dashoffset" begin="0.6s" dur="0.2s" values="8;0"/></path><path stroke-dasharray="2" stroke-dashoffset="2" d="M12 17v0.01"><animate fill="freeze" attributeName="stroke-dashoffset" begin="0.8s" dur="0.2s" values="2;0"/></path></g></svg>
      </button>
    </span>

    <span class="flex flex-row items-center gap-2 w-full">
      <input placeholder="SIP Display Name" bind:value={config.DisplayName} class="w-full my-2 py-2 px-3 bg-slate-700/20 rounded-lg focus:outline-0" />
      <button title="Name displayed on SIP clients (press to load example)" aria-label="sip display name" class="text-slate-400" onclick={() => config.DisplayName = "Dagobert Duck"}>
        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"><g fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2"><path stroke-dasharray="64" stroke-dashoffset="64" d="M12 3c4.97 0 9 4.03 9 9c0 4.97 -4.03 9 -9 9c-4.97 0 -9 -4.03 -9 -9c0 -4.97 4.03 -9 9 -9Z"><animate fill="freeze" attributeName="stroke-dashoffset" dur="0.6s" values="64;0"/></path><path stroke-dasharray="8" stroke-dashoffset="8" d="M12 7v6"><animate fill="freeze" attributeName="stroke-dashoffset" begin="0.6s" dur="0.2s" values="8;0"/></path><path stroke-dasharray="2" stroke-dashoffset="2" d="M12 17v0.01"><animate fill="freeze" attributeName="stroke-dashoffset" begin="0.8s" dur="0.2s" values="2;0"/></path></g></svg>
      </button>
    </span>

    <span class="flex flex-row items-center gap-2 w-full">
      <input placeholder="SIP User" bind:value={config.Username} class="w-full my-2 py-2 px-3 bg-slate-700/20 rounded-lg focus:outline-0" />
      <button title="Username for SIP authentication (press to load example)" aria-label="sip username" class="text-slate-400" onclick={() => config.Username = "dagobert.duck"}>
        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"><g fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2"><path stroke-dasharray="64" stroke-dashoffset="64" d="M12 3c4.97 0 9 4.03 9 9c0 4.97 -4.03 9 -9 9c-4.97 0 -9 -4.03 -9 -9c0 -4.97 4.03 -9 9 -9Z"><animate fill="freeze" attributeName="stroke-dashoffset" dur="0.6s" values="64;0"/></path><path stroke-dasharray="8" stroke-dashoffset="8" d="M12 7v6"><animate fill="freeze" attributeName="stroke-dashoffset" begin="0.6s" dur="0.2s" values="8;0"/></path><path stroke-dasharray="2" stroke-dashoffset="2" d="M12 17v0.01"><animate fill="freeze" attributeName="stroke-dashoffset" begin="0.8s" dur="0.2s" values="2;0"/></path></g></svg>
      </button>
    </span>

    <span class="flex flex-row items-center gap-2 w-full">
      <input placeholder="SIP Password" bind:value={config.Password} class="w-full my-2 py-2 px-3 bg-slate-700/20 rounded-lg focus:outline-0" />
      <div title="Password for SIP authentication" class="rounded-lg text-slate-400 hover:bg-slate-700/20 transition-all duration-700">
        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"><g fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2"><path stroke-dasharray="64" stroke-dashoffset="64" d="M12 3c4.97 0 9 4.03 9 9c0 4.97 -4.03 9 -9 9c-4.97 0 -9 -4.03 -9 -9c0 -4.97 4.03 -9 9 -9Z"><animate fill="freeze" attributeName="stroke-dashoffset" dur="0.6s" values="64;0"/></path><path stroke-dasharray="8" stroke-dashoffset="8" d="M12 7v6"><animate fill="freeze" attributeName="stroke-dashoffset" begin="0.6s" dur="0.2s" values="8;0"/></path><path stroke-dasharray="2" stroke-dashoffset="2" d="M12 17v0.01"><animate fill="freeze" attributeName="stroke-dashoffset" begin="0.8s" dur="0.2s" values="2;0"/></path></g></svg>
      </div>
    </span>

    <hr style="border-color: {Palette.fgPrimary()};" class="border-1 rounded-lg w-full">

    <span class="flex flex-row items-center gap-2 w-full">
      <input placeholder="Encryption Key" bind:value={password} class="w-full my-2 py-2 px-3 bg-slate-700/20 rounded-lg focus:outline-0" />
      <div title="Config encryption key (leave empty for plaintext)" class="rounded-lg text-slate-400 hover:bg-slate-700/20 transition-all duration-700">
        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"><g fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2"><path stroke-dasharray="64" stroke-dashoffset="64" d="M12 3c4.97 0 9 4.03 9 9c0 4.97 -4.03 9 -9 9c-4.97 0 -9 -4.03 -9 -9c0 -4.97 4.03 -9 9 -9Z"><animate fill="freeze" attributeName="stroke-dashoffset" dur="0.6s" values="64;0"/></path><path stroke-dasharray="8" stroke-dashoffset="8" d="M12 7v6"><animate fill="freeze" attributeName="stroke-dashoffset" begin="0.6s" dur="0.2s" values="8;0"/></path><path stroke-dasharray="2" stroke-dashoffset="2" d="M12 17v0.01"><animate fill="freeze" attributeName="stroke-dashoffset" begin="0.8s" dur="0.2s" values="2;0"/></path></g></svg>
      </div>
    </span>
  </Popup>
{/if}