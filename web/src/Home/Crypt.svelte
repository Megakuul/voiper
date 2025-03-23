<script>
  import Popup from "../../src/lib/component/Popup.svelte";
  import { GetConfig, RemoveConfig, SetConfig } from "../../wailsjs/go/app/App";

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
  async function encrypt(name, key) {
    try {
      const config = await GetConfig(name, "")
      await SetConfig(config, name, key)
      await RemoveConfig(name, false)
      PostHook()
    } catch (err) {
      ExceptionRef = err
    }
  }

  /**
  * @param {string} name
  * @param {string} key
  */
  async function decrypt(name, key) {
    try {
      const config = await GetConfig(name, key)
      await SetConfig(config, name, "")
      await RemoveConfig(name, true)
      PostHook()
    } catch (err) {
      ExceptionRef = err
    }
  }
</script>

{#if encrypted}
  <button title="decrypt" aria-label="decrypt" class={className} onclick={() => popupState = true}>
    <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 56 56"><path fill="currentColor" d="M27.988 51.672c.375 0 .985-.14 1.594-.469c13.313-7.476 17.906-10.64 17.906-19.195v-17.93c0-2.46-1.078-3.234-3.047-4.078c-2.765-1.148-11.718-4.36-14.484-5.32c-.633-.211-1.289-.352-1.969-.352c-.656 0-1.312.188-1.945.352c-2.766.843-11.719 4.195-14.484 5.32c-1.97.82-3.047 1.617-3.047 4.078v17.93c0 8.554 4.617 11.695 17.906 19.195c.61.328 1.195.469 1.57.469m0-4.266c-.351 0-.75-.14-1.453-.562c-10.828-6.563-14.297-8.485-14.297-15.703V14.78c0-.797.164-1.101.797-1.36c3.563-1.405 10.43-3.843 14.04-5.132q.526-.21.913-.21c.282 0 .563.07.938.21c3.61 1.29 10.406 3.89 14.039 5.133c.633.234.797.562.797 1.36V31.14c0 7.218-3.492 9.117-14.297 15.703c-.703.422-1.102.562-1.477.562m-5.836-9.093h11.672c1.64 0 2.414-.82 2.414-2.579V26.64c0-1.547-.61-2.368-1.851-2.532V21.32c0-4.312-2.578-7.218-6.399-7.218c-3.797 0-6.398 2.906-6.398 7.218v2.813c-1.219.164-1.828.984-1.828 2.508v9.093c0 1.758.773 2.578 2.39 2.578m1.875-17.25c0-2.766 1.594-4.594 3.961-4.594c2.39 0 3.961 1.828 3.961 4.593v3l-7.922.024Z"/></svg>
  </button>
  {#if popupState}
    <Popup title={"Decrypt Configuration"} onsubmit={() => decrypt(name, password)} bind:StateRef={popupState}>
      <input type="password" placeholder="Password" bind:value={password} class="w-full my-2 py-2 px-3 bg-slate-700/20 rounded-lg focus:outline-0"/>
    </Popup>
  {/if}
{:else}
  <button title="encrypt" aria-label="encrypt" class={className} onclick={() => popupState = true}>
    <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24"><g fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2"><path d="M2.586 17.414A2 2 0 0 0 2 18.828V21a1 1 0 0 0 1 1h3a1 1 0 0 0 1-1v-1a1 1 0 0 1 1-1h1a1 1 0 0 0 1-1v-1a1 1 0 0 1 1-1h.172a2 2 0 0 0 1.414-.586l.814-.814a6.5 6.5 0 1 0-4-4z"/><circle cx="16.5" cy="7.5" r=".5" fill="currentColor"/></g></svg>
  </button>
  {#if popupState}
    <Popup title={"Encrypt Configuration"} onsubmit={() => encrypt(name, password)} bind:StateRef={popupState}>
      <input type="password" placeholder="Password" bind:value={password} class="w-full my-2 py-2 px-3 bg-slate-700/20 rounded-lg focus:outline-0" />
    </Popup>
  {/if}
{/if}

