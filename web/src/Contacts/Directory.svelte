<script>
  import { ui } from "../ui.js";
  import { onMount } from "svelte";
  import Dropdown from "../components/Dropdown.svelte";
  import * as api from "../../wailsjs/go/app/App.js";
  import { EventsOn } from "../../wailsjs/runtime/runtime.js";

  let { onChanged } = $props();
  let profile = $state({
    URL: "",
    BaseDN: "",
    BindDN: "",
    CAFile: "",
    Limit: 1000,
    IntervalMinutes: 60,
    Enabled: false,
    UseSecretService: false,
  });
  let password = $state("");
  let status = $state({});
  let running = $state(false);
  let busy = $state(false);
  let error = $state("");
  let notice = $state("");
  let result = $state(null);
  let bases = $state([]);
  let discoveryDomain = $state("");
  let servers = $state([]);
  let disposed = false;

  async function load(initial = false) {
    const state = await api.DirectoryState();
    if (disposed) return;
    if (initial) profile = state.Profile;
    status = state.Status;
    running = state.Running;
  }
  async function run(action) {
    busy = true;
    error = "";
    notice = "";
    try {
      await action();
    } catch (e) {
      error = String(e);
    } finally {
      busy = false;
    }
  }
  function config() {
    return { ...profile, Password: password };
  }
  function date(value) {
    return value && !value.startsWith("0001-")
      ? new Date(value).toLocaleString()
      : "Never";
  }
  onMount(() => {
    run(() => load(true));
    const stop = EventsOn("directory-changed", () => {
      load()
        .then(() => onChanged())
        .catch((e) => {
          error = String(e);
        });
    });
    return () => {
      disposed = true;
      password = "";
      stop();
    };
  });
</script>

<details class={[ui.details, "mt-[1.3rem] border-t border-t-border pt-4"]}>
  <summary class={[ui.summary, "px-0 py-1"]}>Company directory</summary>
  <p class="m-0 leading-[1.6] text-muted text-[0.78rem]">
    Use your administrator-provided LDAP endpoint. Connections require StartTLS
    or LDAPS with a trusted certificate. Cached contacts remain available
    offline.
  </p>
  {#if error}<p class="m-0 leading-[1.6]" role="alert">{error}</p>{/if}
  {#if notice}<p class="m-0 leading-[1.6]" role="status">{notice}</p>{/if}
  <fieldset class="p-0 m-0 min-w-0 border-0" disabled={busy}>
    <details class={[ui.details, "mt-[1.3rem] border-t border-t-border pt-4"]}>
      <summary class={[ui.summary, "px-0 py-1"]}>Find advertised directory servers</summary>
      <label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]"
        >Company DNS domain<input class={[ui.input, "px-[0.8rem] py-[0.65rem]"]}
          placeholder="example.org"
          bind:value={discoveryDomain}
        /></label
      >
      <button class={ui.button}
        type="button"
        disabled={busy || !discoveryDomain}
        onclick={() =>
          run(async () => {
            servers = await api.DiscoverDirectoryServers(discoveryDomain);
            if (!servers.length)
              notice =
                "No LDAP service is advertised for this domain. Enter the administrator-provided endpoint below.";
          })}>Find servers</button
      >
      {#each servers as server}<button class={ui.button}
          type="button"
          onclick={() => {
            profile.URL = server;
            password = "";
            bases = [];
            result = null;
          }}>{server}</button
        >{/each}
      <p class="m-0 leading-[1.6] text-muted text-[0.78rem]">
        Uses DNS service records. Selecting a result does not sign in; LDAP
        connections still require verified StartTLS. This does not assume the
        SIP server also hosts your directory.
      </p>
    </details>
    <form
      class="stack flex flex-col gap-4"
      oninput={() => {
        result = null;
      }}
      onsubmit={(e) => {
        e.preventDefault();
        run(async () => {
          await api.SaveDirectoryProfile(profile, password);
          password = "";
          notice = profile.Enabled
            ? "Directory saved. Refresh runs in the background."
            : "Directory saved. Automatic refresh is off.";
          await load();
        });
      }}
    >
      <label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]"
        >Directory URL<input class={[ui.input, "px-[0.8rem] py-[0.65rem]"]}
          required
          placeholder="ldaps://directory.example.org"
          bind:value={profile.URL}
        /></label
      >
      <button class={ui.button}
        type="button"
        disabled={busy || !profile.URL}
        onclick={() =>
          run(async () => {
            bases = await api.DiscoverDirectory(config());
            if (bases.length === 1) profile.BaseDN = bases[0];
            if (!bases.length)
              notice =
                "The server did not advertise directory bases. Ask your administrator for the base DN.";
          })}>Find directory bases</button
      >
      {#if bases.length}<label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]"
          >Available bases<Dropdown
            label="Available bases"
            value={bases.includes(profile.BaseDN) ? profile.BaseDN : ""}
            disabled={busy}
            options={[
              { value: "", label: "Select a base" },
              ...bases.map((base) => ({ value: base, label: base })),
            ]}
            onchange={(value) => {
              profile.BaseDN = value;
              result = null;
            }}
          /></label
        >{/if}
      <label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]"
        >Base DN<input class={[ui.input, "px-[0.8rem] py-[0.65rem]"]}
          required
          placeholder="dc=example,dc=org"
          bind:value={profile.BaseDN}
        /></label
      >
      <label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]">Bind DN (optional)<input class={[ui.input, "px-[0.8rem] py-[0.65rem]"]} bind:value={profile.BindDN} /></label>
      <label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]"
        >Additional CA certificate file (optional)<input class={[ui.input, "px-[0.8rem] py-[0.65rem]"]}
          placeholder="System trust by default"
          bind:value={profile.CAFile}
        /></label
      >
      <label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]"
        >Password<input class={[ui.input, "px-[0.8rem] py-[0.65rem]"]}
          type="password"
          autocomplete="off"
          bind:value={password}
        /></label
      >
      <label class="flex flex-row gap-[0.65rem] text-[0.9rem] text-[#c8cdd5] items-center min-h-[32px] cursor-pointer"
        ><input class={ui.checkbox} type="checkbox" bind:checked={profile.UseSecretService} /> Save password
        in the desktop wallet</label
      >
      <p class="m-0 leading-[1.6] text-muted text-[0.78rem]">
        Leave an existing password empty to keep it. Without the wallet, enter
        the password again after restarting Voiper. LDAP credentials may differ
        from SIP credentials.
      </p>
      <label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]"
        >Maximum entries<input class={[ui.input, "px-[0.8rem] py-[0.65rem]"]}
          type="number"
          min="1"
          max="5000"
          bind:value={profile.Limit}
        /></label
      >
      <label class="flex flex-row gap-[0.65rem] text-[0.9rem] text-[#c8cdd5] items-center min-h-[32px] cursor-pointer"
        ><input class={ui.checkbox} type="checkbox" bind:checked={profile.Enabled} /> Keep directory contacts
        synchronized</label
      >
      <label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]"
        >Refresh every (minutes)<input class={[ui.input, "px-[0.8rem] py-[0.65rem]"]}
          type="number"
          min="5"
          max="1440"
          bind:value={profile.IntervalMinutes}
        /></label
      >
      <button class={ui.button} disabled={busy}>Save directory</button>
      <div class="row flex items-center gap-[0.7rem] max-[700px]:flex-wrap [&_>_button]:shrink-0 max-[700px]:[&_>_.grow]:basis-[180px]">
        <button class={ui.button}
          type="button"
          disabled={busy}
          onclick={() =>
            run(async () => {
              result = await api.LookupDirectory(config());
            })}>Preview contacts</button
        >
        <button class={ui.button}
          type="button"
          disabled={busy || running || !profile.Enabled}
          onclick={() =>
            run(async () => {
              await api.RefreshDirectory();
              await load();
            })}>{running ? "Refreshing…" : "Refresh saved directory"}</button
        >
        <button class={ui.button}
          type="button"
          disabled={busy}
          onclick={() =>
            run(async () => {
              await api.DeleteDirectoryPassword();
              notice =
                "Saved directory password removed from the desktop wallet.";
            })}>Forget wallet password</button
        >
      </div>
    </form>
  </fieldset>
  <p class="m-0 leading-[1.6] text-muted text-[0.78rem]" aria-live="polite">
    Last successful refresh: {date(status.LastSuccess)}. {status.Entries || 0} directory
    entries; {status.Preserved || 0} local edits, duplicates or hidden contacts preserved.{status.Partial
      ? " The last scan was partial; missing contacts were kept."
      : ""}
  </p>
  {#if status.Error}<p class="m-0 leading-[1.6]" role="status">{status.Error}</p>{/if}
  <p class="m-0 leading-[1.6] text-muted text-[0.78rem]">
    Editing a synchronized contact keeps a local copy. Deleting it hides it from
    later refreshes. Only a complete scan can remove contacts that disappeared
    from the directory. Changing servers preserves the old phonebook as an
    imported snapshot.
  </p>
  {#if result}
    <p class="m-0 leading-[1.6]">
      {result.Contacts.length} entries · {result.Skipped} skipped{result.Truncated
        ? " · limit reached"
        : ""}
    </p>
    <div class="p-4 whitespace-pre-wrap wrap-anywhere max-h-[180px] overflow-auto text-xs bg-[#1c252e] rounded-[6px]">
      {#each result.Contacts.slice(0, 20) as entry}<p class="m-0 leading-[1.6]">
          {entry.Name} · {entry.Address}
        </p>{/each}
    </div>
    <button class={ui.button}
      disabled={busy}
      onclick={() =>
        run(async () => {
          const count = await api.ImportDirectory(result.Contacts);
          notice = `Imported ${count} contacts as a one-time snapshot`;
          result = null;
          await onChanged();
        })}>Import preview once</button
    >
  {/if}
</details>
