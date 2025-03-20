<script>
  import Popup from "../../src/lib/component/Popup.svelte";
  import { RemoveConfig } from "../../wailsjs/go/app/App";

  let {
    /** @type {string} */ path,
    /** @type {boolean} */ encrypted,
    /** @type {function(): Promise<any>} */ PostHook,
    /** @type {string} */ StateRef = $bindable(),
    /** @type {string} */ ExceptionRef = $bindable(),
  } = $props()

  /** 
  * @param {string} path
  * @param {boolean} encrypted
  */
  async function remove(path, encrypted) {
    try {
      await RemoveConfig(path, encrypted)
      await PostHook()
    } catch (err) {
      ExceptionRef = err
    }
  }
</script>

{#if StateRef}
  <Popup title={"Remove Configuration?"} onsubmit={remove(path, encrypted)} bind:StateRef={StateRef}>
    <br>
  </Popup>
{/if}