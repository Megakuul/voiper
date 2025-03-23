<script>
  import Popup from "../../src/lib/component/Popup.svelte";
  import { RemoveConfig } from "../../wailsjs/go/app/App";

  /** @type {{
   * class: string,
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
  let popupState = $state(false)

  /** 
  * @param {string} name
  * @param {boolean} encrypted
  */
  async function remove(name, encrypted) {
    try {
      await RemoveConfig(name, encrypted)
      await PostHook()
    } catch (err) {
      ExceptionRef = err
    }
  }
</script>

<button title="delete" aria-label="delete" class={className} onclick={() => popupState = true}>
  <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24"><path fill="currentColor" d="M8 9h8v10H8z" opacity="0.3"/><path fill="currentColor" d="m15.5 4l-1-1h-5l-1 1H5v2h14V4zM6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6zM8 9h8v10H8z"/></svg>
</button>

{#if popupState}
  <Popup title={"Remove Configuration?"} onsubmit={() => remove(name, encrypted)} bind:StateRef={popupState}>
    <br>
  </Popup>
{/if}