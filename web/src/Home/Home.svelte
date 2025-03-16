<script>
  import Logo from '../lib/component/Logo.svelte';
  import {EnableConfig, ListConfigs, RegisterSIP} from '../../wailsjs/go/app/App.js'
  import Spinner from '../lib/component/Spinner.svelte';

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
      // Exception = err
    }
  }

  async function enable(name) {
    try {
      await EnableConfig(name, "")
    } catch (err) {
      // Exception = err
    }
  }
</script>

<div>
    <Logo width="1000"></Logo>

    <Spinner color="white" class="bg-amber-500" width="300px"></Spinner>

    <button class="btn" onclick={voip}>VOIPPPPPPPPP</button>
    <button class="btn" onclick={list}>List</button>
    <div class="flex flex-col items-center">
        {#each Object.entries(Configs) as [path, encrypted]}
        <button class="w-1/2 bg-slate-500" onclick={() => enable(path)}>{path} {encrypted}</button>
        {/each}
    </div>
</div>

<style>
  
  </style>