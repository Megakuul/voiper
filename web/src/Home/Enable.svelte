<script>
  import Popup from "../../src/lib/component/Popup.svelte";
  import { EnableConfig, RegisterSIP } from "../../wailsjs/go/app/App";

  /** @type {{ 
   * path: string, 
   * encrypted: boolean, 
   * StateRef: boolean,
   * ExceptionRef: string
   * }} */
   let {
    path,
    encrypted,
    StateRef = $bindable(),
    ExceptionRef = $bindable(),
  } = $props()

  /** @type {string} */
  let password = $state("")

  $effect(() => {
    if (!StateRef) {
      password = ""
    }
  })

  /**
  * @param {string} path
  * @param {string} key
  */
  async function enable(path, key) {
    try {
      await EnableConfig(path, key)
      await RegisterSIP()
    } catch (err) {
      ExceptionRef = err
    }
  }
</script>

{#if StateRef}
  <Popup title={"Enable Configuration?"} onsubmit={enable(path, password)} bind:StateRef={StateRef}>
    {#if encrypted}
      <input type="password" placeholder="Password" bind:value={password} />
    {/if}
    <br>
  </Popup>
{/if}