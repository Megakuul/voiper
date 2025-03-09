<script>
  import logo from './assets/images/logo-universal.png'
  import {ListConfigs, RegisterSIP} from '../wailsjs/go/app/App.js'

  /** @type {string} */
  let Exception = $state(undefined);

  /** @type {Object.<string, boolean>} */
  let Configs = $state({});

  async function voip() {
    try {
      await RegisterSIP()
      return "WORKED"
    } catch (err) {
      return err
    }
  }

  async function list() {
    try {
      Configs = {}
      Configs = await ListConfigs()
    } catch (err) {
      Exception = err
    }
  }
</script>

<main>
  <img alt="Wails logo" id="logo" src="{logo}">
    <button class="btn" onclick={voip}>VOIPPPPPPPPP</button>
    <button class="btn" onclick={list}>List</button>
  <div class="flex flex-col items-center">
    {#each Object.entries(Configs) as [path, encrypted]}
      <div class="w-1/2 bg-slate-500">{path} {encrypted}</div>
    {/each}
  </div>

  {#if Exception}
    <div>ALARM {Exception}</div>
  {/if}
</main>

<style>

  #logo {
    display: block;
    width: 50%;
    height: 50%;
    margin: auto;
    padding: 10% 0 0;
    background-position: center;
    background-repeat: no-repeat;
    background-size: 100% 100%;
    background-origin: content-box;
  }

</style>
