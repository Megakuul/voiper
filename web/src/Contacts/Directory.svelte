<script>
  import { onMount } from "svelte";
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

<details>
  <summary>Company directory</summary>
  <p class="muted small">
    Use your administrator-provided LDAP endpoint. Connections require StartTLS
    or LDAPS with a trusted certificate. Cached contacts remain available
    offline.
  </p>
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  {#if notice}<p role="status">{notice}</p>{/if}
  <fieldset disabled={busy}>
    <details>
      <summary>Find advertised directory servers</summary>
      <label
        >Company DNS domain<input
          placeholder="example.org"
          bind:value={discoveryDomain}
        /></label
      >
      <button
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
      {#each servers as server}<button
          type="button"
          onclick={() => {
            profile.URL = server;
            password = "";
            bases = [];
            result = null;
          }}>{server}</button
        >{/each}
      <p class="muted small">
        Uses DNS service records. Selecting a result does not sign in; LDAP
        connections still require verified StartTLS. This does not assume the
        SIP server also hosts your directory.
      </p>
    </details>
    <form
      class="stack"
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
      <label
        >Directory URL<input
          required
          placeholder="ldaps://directory.example.org"
          bind:value={profile.URL}
        /></label
      >
      <button
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
      {#if bases.length}<label
          >Available bases<select
            onchange={(e) => (profile.BaseDN = e.currentTarget.value)}
            ><option value="">Select a base</option>{#each bases as base}<option
                value={base}>{base}</option
              >{/each}</select
          ></label
        >{/if}
      <label
        >Base DN<input
          required
          placeholder="dc=example,dc=org"
          bind:value={profile.BaseDN}
        /></label
      >
      <label>Bind DN (optional)<input bind:value={profile.BindDN} /></label>
      <label
        >Additional CA certificate file (optional)<input
          placeholder="System trust by default"
          bind:value={profile.CAFile}
        /></label
      >
      <label
        >Password<input
          type="password"
          autocomplete="off"
          bind:value={password}
        /></label
      >
      <label class="toggle"
        ><input type="checkbox" bind:checked={profile.UseSecretService} /> Save password
        in the desktop wallet</label
      >
      <p class="muted small">
        Leave an existing password empty to keep it. Without the wallet, enter
        the password again after restarting Voiper. LDAP credentials may differ
        from SIP credentials.
      </p>
      <label
        >Maximum entries<input
          type="number"
          min="1"
          max="5000"
          bind:value={profile.Limit}
        /></label
      >
      <label class="toggle"
        ><input type="checkbox" bind:checked={profile.Enabled} /> Keep directory contacts
        synchronized</label
      >
      <label
        >Refresh every (minutes)<input
          type="number"
          min="5"
          max="1440"
          bind:value={profile.IntervalMinutes}
        /></label
      >
      <button disabled={busy}>Save directory</button>
      <div class="row">
        <button
          type="button"
          disabled={busy}
          onclick={() =>
            run(async () => {
              result = await api.LookupDirectory(config());
            })}>Preview contacts</button
        >
        <button
          type="button"
          disabled={busy || running || !profile.Enabled}
          onclick={() =>
            run(async () => {
              await api.RefreshDirectory();
              await load();
            })}>{running ? "Refreshing…" : "Refresh saved directory"}</button
        >
        <button
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
  <p class="muted small" aria-live="polite">
    Last successful refresh: {date(status.LastSuccess)}. {status.Entries || 0} directory
    entries; {status.Preserved || 0} local edits, duplicates or hidden contacts preserved.{status.Partial
      ? " The last scan was partial; missing contacts were kept."
      : ""}
  </p>
  {#if status.Error}<p role="status">{status.Error}</p>{/if}
  <p class="muted small">
    Editing a synchronized contact keeps a local copy. Deleting it hides it from
    later refreshes. Only a complete scan can remove contacts that disappeared
    from the directory. Changing servers preserves the old phonebook as an
    imported snapshot.
  </p>
  {#if result}
    <p>
      {result.Contacts.length} entries · {result.Skipped} skipped{result.Truncated
        ? " · limit reached"
        : ""}
    </p>
    <div class="preview">
      {#each result.Contacts.slice(0, 20) as entry}<p>
          {entry.Name} · {entry.Address}
        </p>{/each}
    </div>
    <button
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
