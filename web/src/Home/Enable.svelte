<script>
  import Popup from "../../src/lib/component/Popup.svelte";
  import { EnableConfig } from "../../wailsjs/go/app/App";

  /** @type {{ 
   * class: string,
   * name: string, 
   * encrypted: boolean,
   * active: boolean,
   * ExceptionRef: string
   * }} */
   let {
    class: className,
    name,
    encrypted,
    active,
    ExceptionRef = $bindable(),
  } = $props()

  /** @type {boolean} */
  let popupState = $state(false)

  /** @type {string} */
  let password = $state("")

  $effect(() => {
    if (!popupState) {
      password = ""
    }
  })

  /**
  * @param {string} name
  * @param {string} key
  */
  async function enable(name, key) {
    try {
      await EnableConfig(name, key)
    } catch (err) {
      ExceptionRef = err
    }
  }
</script>

<button title="enable" aria-label="enable" class={className} onclick={() => popupState = true}>
  {#if active}
    <svg xmlns="http://www.w3.org/2000/svg" width="40" height="40" viewBox="0 0 24 24"><g fill="none" stroke="currentColor" stroke-dasharray="32" stroke-dashoffset="32" stroke-linecap="round" stroke-linejoin="round" stroke-width="2"><path d="M7 6h2v12h-2Z"><animate fill="freeze" attributeName="stroke-dashoffset" dur="0.4s" values="32;0"/></path><path d="M15 6h2v12h-2Z"><animate fill="freeze" attributeName="stroke-dashoffset" begin="0.4s" dur="0.4s" values="32;0"/></path></g></svg>
  {:else}
    <svg xmlns="http://www.w3.org/2000/svg" width="40" height="40" viewBox="0 0 24 24"><path fill="currentColor" fill-opacity="0" stroke="currentColor" stroke-dasharray="40" stroke-dashoffset="40" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 6l10 6l-10 6Z"><animate fill="freeze" attributeName="fill-opacity" begin="0.5s" dur="0.15s" values="0;0.3"/><animate fill="freeze" attributeName="stroke-dashoffset" dur="0.5s" values="40;0"/></path></svg>
  {/if}
</button>

{#if popupState}
  <Popup title={"Enable Configuration"} onsubmit={() => enable(name, password)} bind:StateRef={popupState}>
    {#if encrypted}
      <input type="password" placeholder="Password" bind:value={password} class="w-full my-2 py-2 px-3 bg-slate-700/20 rounded-lg focus:outline-0" />
    {/if}
    <br>
  </Popup>
{/if}