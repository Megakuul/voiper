<script>
  import Popup from "../lib/component/Popup.svelte";
  import { AddConfig } from "../../wailsjs/go/app/App";

  /** @type {{ 
   * PostHook: function(): Promise<any>, 
   * StateRef: boolean,
   * ExceptionRef: string
   * }} */
  let {
    PostHook,
    StateRef = $bindable(),
    ExceptionRef = $bindable(),
  } = $props()

  /** @type {string}*/
  let name = $state("")
  /** @type {string}*/
  let password = $state("")
  /** @type {import('wailsjs/go/models').config.Config} */
  let config = $state({})

  $effect(() => {
    if (!StateRef) {
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
      await AddConfig(config, name, key)
      await PostHook()
    } catch (err) {
      ExceptionRef = err
    }
  }
</script>

{#if StateRef}
  <Popup title={"New Configuration"} onsubmit={add(config, name, password)} bind:StateRef={StateRef}>
    <input placeholder="Name" bind:value={name} />
    <input placeholder="Server" bind:value={config.Server} />
    <input placeholder="Display Name" bind:value={config.DisplayName} />
    <input placeholder="Username" bind:value={config.Username} />
    <input placeholder="Password" bind:value={config.Password} />

    <input type="password" placeholder="Password" bind:value={password} />
  </Popup>
{/if}