<script>
  import { onMount, tick } from "svelte";
  import * as api from "../wailsjs/go/app/App.js";
  import { EventsOn } from "../wailsjs/runtime/runtime.js";
  import Accounts from "./Home/Home.svelte";
  import Phone from "./Phone/Phone.svelte";
  import Messages from "./Messages/Messages.svelte";
  import History from "./History/History.svelte";
  import Directory from "./Contacts/Directory.svelte";
  import CSVFields from "./Contacts/CSVFields.svelte";
  import AudioSettings from "./Audio/AudioSettings.svelte";
  import logo from "./assets/images/icon.svg";

  let page = $state("phone");
  $effect(() => {
    const selectedPage = page;
    tick().then(() => {
      if (page === selectedPage) window.scrollTo({ top: 0, left: 0 });
    });
  });
  let confirmQuit = $state(false);
  let quitDialog;
  $effect(() => {
    if (quitDialog) {
      if (confirmQuit && !quitDialog.open) quitDialog.showModal();
      else if (!confirmQuit && quitDialog.open) quitDialog.close();
    }
  });
  let snapshot = $state({ Accounts: [], Calls: [], DND: false, Audio: {} });
  let account = $state("");
  let target = $state("");
  let error = $state("");
  let notice = $state("");
  let busy = $state(false);
  let contacts = $state([]);

  let search = $state("");
  let contact = $state({
    ID: 0,
    Name: "",
    Address: "",
    Notes: "",
    Favorite: false,
  });

  let devices = $state([]);
  let audio = $state({
    Backend: "",
    InputDevice: "",
    OutputDevice: "",
    RingerDevice: "",
  });
  let preferences = $state({
    DefaultAccount: "",
    Notifications: true,
    CloseToBackground: false,
    Autostart: false,
    MessageRetentionDays: 0,
  });
  let desktopCapabilities = $state({});
  let globalShortcuts = $state({
    Available: false,
    Enabled: false,
    Configuring: false,
    Bindings: [],
  });
  let shortcutBusy = $state(false);
  let preferencesLoaded = false;
  let csv = $state("");
  let rawCSV = $state("");
  let mapCSV = $state(false);
  let importFormat = $state("csv");
  let importPreview = $state(null);
  let importSource = $state(null);
  let importRevision = 0;
  let mergeImportNumbers = $state(false);
  let logs = $state([]);
  let refreshing = false;
  let disposed = false;

  function clearImportPreview() {
    importRevision++;
    importPreview = null;
    importSource = null;
  }
  async function previewContacts() {
    clearImportPreview();
    if (!csv) return;
    const revision = importRevision;
    const source = {
      format: importFormat,
      text: csv,
      merge: mergeImportNumbers,
    };
    const preview = await api.PreviewContactImport(
      source.format,
      source.text,
      source.merge,
    );
    if (revision !== importRevision) return;
    importSource = source;
    importPreview = preview;
  }

  async function refresh() {
    if (refreshing || disposed) return;
    refreshing = true;
    try {
      snapshot = await api.Snapshot();
      if (!preferencesLoaded) {
        await loadPreferences();
        await acceptDialTarget();
      }
      if (!account && snapshot.Accounts.length)
        account = snapshot.Accounts[0].Name;
    } catch (e) {
      error = String(e);
    } finally {
      refreshing = false;
    }
  }
  async function run(action) {
    busy = true;
    error = "";
    notice = "";
    try {
      return await action();
    } catch (e) {
      error = String(e);
    } finally {
      busy = false;
      await refresh();
    }
  }
  async function loadContacts(more = false) {
    const rows = await api.ContactsPage(
      search,
      more ? contacts.length : 0,
      100,
    );
    contacts = more ? [...contacts, ...rows] : rows;
  }
  async function navigate(next) {
    page = next;
    await run(async () => {
      if (next === "contacts") await loadContacts();
      if (next === "settings") {
        await loadPreferences();
        await loadShortcuts();
        audio = { ...snapshot.Audio };
        devices = await api.AudioDevices();
      }
    });
  }
  function dial(address) {
    target = address;
    page = "phone";
  }
  function saveCorrespondent(accountName, remote) {
    contact = {
      ID: 0,
      Name: "",
      Address: remote,
      Notes: "",
      Favorite: false,
      Numbers: [],
      Account: accountName,
    };
    navigate("contacts");
  }
  async function loadPreferences() {
    const settings = await api.DesktopSettings();
    preferences = settings.Preferences;
    preferencesLoaded = true;
    desktopCapabilities = settings.Capabilities;
    if (!account && preferences.DefaultAccount)
      account = preferences.DefaultAccount;
  }
  async function acceptDialTarget() {
    const address = await api.TakeDialTarget();
    if (address) dial(address);
  }
  async function loadShortcuts() {
    globalShortcuts = await api.GetGlobalShortcuts();
  }
  async function shortcutAction(action) {
    shortcutBusy = true;
    error = "";
    try {
      await action();
    } catch (e) {
      error = String(e);
    } finally {
      try {
        await loadShortcuts();
      } catch (e) {
        error = String(e);
      }
      shortcutBusy = false;
    }
  }
  async function keyboard(event) {
    if (!(event.ctrlKey || event.metaKey) || event.altKey || event.shiftKey)
      return;
    const pages = [
      "phone",
      "contacts",
      "history",
      "messages",
      "accounts",
      "settings",
    ];
    const index = Number(event.key) - 1;
    if (index >= 0 && index < pages.length) {
      event.preventDefault();
      await navigate(pages[index]);
    }
    if (event.key.toLowerCase() === "k") {
      event.preventDefault();
      page = "phone";
      await tick();
      document.querySelector(".dial-input")?.focus();
    }
  }
  onMount(() => {
    disposed = false;
    const off = [
      EventsOn("phone-changed", refresh),
      EventsOn("confirm-quit", () => (confirmQuit = true)),
      EventsOn("desktop-ready", () =>
        run(async () => {
          await loadPreferences();
          await acceptDialTarget();
          await loadShortcuts();
        }),
      ),
      EventsOn("desktop-warning", (text) => (notice = String(text))),
      EventsOn("shortcuts-changed", () =>
        loadShortcuts().catch((e) => (error = String(e))),
      ),
      EventsOn("show-dialer", async () => {
        page = "phone";
        await tick();
        document.querySelector(".dial-input")?.focus();
      }),
      EventsOn("dial-target", () => run(acceptDialTarget)),
      EventsOn("phone-error", (e) => (error = String(e))),
      EventsOn("incoming-call", (remote) => {
        notice = `Incoming call from ${remote.Remote || remote}`;
        page = "phone";
      }),
      EventsOn(
        "incoming-message",
        (remote) => (notice = `Message from ${remote.Remote || remote}`),
      ),
      EventsOn("log", (line) => {
        logs = [...logs.slice(-499), String(line)];
      }),
    ];
    run(async () => {
      await loadPreferences();
      await acceptDialTarget();
    });
    const interval = setInterval(refresh, 1000);
    return () => {
      disposed = true;
      clearInterval(interval);
      off.forEach((stop) => stop());
    };
  });
</script>

<svelte:window onkeydown={keyboard} />
<svelte:head><title>Voiper</title></svelte:head>
<dialog
  bind:this={quitDialog}
  oncancel={() => (confirmQuit = false)}
  onclose={() => (confirmQuit = false)}
>
  <h2>End active calls?</h2>
  <p>Quitting ends all calls and unregisters your accounts.</p>
  <div class="row">
    <button onclick={() => (confirmQuit = false)}>Keep open</button><button
      class="danger"
      onclick={() => api.Quit()}>Quit Voiper</button
    >
  </div>
</dialog>
<div class="shell">
  <aside class="sidebar">
    <a class="brand" href="#phone" onclick={() => navigate("phone")}
      ><img src={logo} alt="Voiper" class="brand-logo" /></a
    >
    <nav aria-label="Main navigation">
      {#each [["phone", "Phone"], ["contacts", "Contacts"], ["history", "Recent calls"], ["messages", "Messages"], ["accounts", "Accounts"], ["settings", "Audio & settings"], ["logs", "Diagnostics"]] as [key, label]}
        <button
          class:active={page === key}
          aria-current={page === key ? "page" : undefined}
          onclick={() => navigate(key)}
          >{label}{#if key === "messages" && snapshot.UnreadMessages}<span
              class="badge">{snapshot.UnreadMessages}</span
            >{/if}{#if key === "phone" && snapshot.Calls.length}<span
              class="badge">{snapshot.Calls.length}</span
            >{/if}</button
        >
      {/each}
    </nav>
    <div class="sidebar-bottom">
      <label class="toggle"
        ><input
          type="checkbox"
          checked={snapshot.DND}
          onchange={(e) => run(() => api.SetDND(e.currentTarget.checked))}
        /> Do not disturb</label
      >
    </div>
  </aside>
  <main>
    <header class="topbar">
      <div>
        <h1>
          {{
            phone: "Phone",
            contacts: "Your contacts",
            history: "Recent calls",
            messages: "Messages",
            accounts: "SIP accounts",
            settings: "Audio & settings",
            logs: "Diagnostics",
          }[page]}
        </h1>
      </div>
      {#if ["phone", "contacts", "messages"].includes(page)}
        <label class="account-select"
          >Outgoing account<select bind:value={account}
            ><option value="">Choose account</option
            >{#each snapshot.Accounts as a}<option value={a.Name}
                >{a.Name} · {a.State}</option
              >{/each}</select
          ></label
        >
      {/if}
    </header>
    {#if error}<div class="banner error" role="alert">
        <span>{error}</span><button
          aria-label="Dismiss error"
          onclick={() => (error = "")}>×</button
        >
      </div>{/if}
    {#if notice}<div class="banner" role="status">
        <span>{notice}</span><button
          aria-label="Dismiss notice"
          onclick={() => (notice = "")}>×</button
        >
      </div>{/if}

    {#if page === "phone"}
      <Phone {snapshot} bind:account bind:target {run} {busy} />
    {:else if page === "accounts"}
      <Accounts {snapshot} {run} />
    {:else if page === "contacts"}
      <div class="two-column">
        <section class="panel">
          <h2>Phonebook</h2>
          <form
            class="row"
            onsubmit={(e) => {
              e.preventDefault();
              run(loadContacts);
            }}
          >
            <input
              aria-label="Search contacts"
              placeholder="Name or number"
              bind:value={search}
            /><button type="submit">Search</button>
          </form>
          <div class="list">
            {#each contacts as c}
              {@const watched = (snapshot.Presence || []).find(
                (p) =>
                  p.Account === account &&
                  !p.Voicemail &&
                  (p.Target === c.Address ||
                    p.Remote === c.Address ||
                    p.Remote.startsWith("sip:" + c.Address + "@")),
              )}
              <div class="list-item">
                <button
                  class="text-button grow"
                  onclick={() => {
                    if (c.Account) account = c.Account;
                    dial(c.Address);
                  }}
                  ><strong>{c.Favorite ? "★ " : ""}{c.Name}</strong><span
                    class="muted">{c.Address}</span
                  >{#if c.Source === "directory-sync"}<span class="small muted"
                      >Company directory · updated {new Date(
                        c.Updated,
                      ).toLocaleString()}</span
                    >{/if}{#if watched}<span
                      class="small"
                      title={new Date(watched.Updated).toLocaleString()}
                      >{watched.State} · {watched.Source}{watched.Note
                        ? " · " + watched.Note
                        : ""} · {Math.max(
                        0,
                        Math.floor(
                          (Date.now() - new Date(watched.Updated).getTime()) /
                            1000,
                        ),
                      )}s ago</span
                    >{/if}</button
                ><button
                  disabled={busy || !account}
                  onclick={() =>
                    run(() =>
                      watched
                        ? api.UnwatchPresence(account, c.Address)
                        : api.WatchPresence(account, c.Address),
                    )}>{watched ? "Stop watching" : "Watch status"}</button
                >{#each c.Numbers || [] as n}<button
                    title={n.Address}
                    onclick={() => {
                      if (c.Account) account = c.Account;
                      dial(n.Address);
                    }}>{n.Label || n.Address}</button
                  >{/each}
                <button
                  onclick={() => {
                    if (c.Account) account = c.Account;
                    target = c.Address;
                    page = "messages";
                  }}>Message</button
                >
                <button
                  onclick={() =>
                    (contact = {
                      ...c,
                      Numbers: (c.Numbers || []).map((n) => ({ ...n })),
                    })}>Edit</button
                ><button
                  class="quiet"
                  onclick={() =>
                    run(async () => {
                      await api.DeleteContact(c.ID);
                      await loadContacts();
                    })}>Delete</button
                >
              </div>{:else}<p class="empty">
                Add a contact or import your phonebook.
              </p>{/each}
          </div>
          {#if contacts.length >= 100}<button
              disabled={busy}
              onclick={() => run(() => loadContacts(true))}
              >Load more contacts</button
            >{/if}
          <p class="muted small">
            Search supports names and all saved numbers. Imported directory
            entries remain available offline.
          </p>
        </section>
        <section class="panel">
          <h2>{contact.ID ? "Edit contact" : "New contact"}</h2>
          {#if contact.Source === "directory-sync"}<p class="muted small">
              Saving changes keeps your own copy. Directory refreshes will
              preserve it.
            </p>{/if}
          {#if contact.DirectoryID && contact.Source !== "directory-sync"}
            <p class="muted small">
              This is your local copy of a directory contact.
            </p>
            <button
              disabled={busy}
              onclick={() =>
                run(async () => {
                  await api.RestoreDirectoryContact(contact.ID);
                  contact = {
                    ID: 0,
                    Name: "",
                    Address: "",
                    Notes: "",
                    Favorite: false,
                  };
                  notice =
                    "Directory name and numbers restored. Personal notes and favorites kept.";
                  await loadContacts();
                })}>Restore directory name and numbers</button
            >
          {/if}
          <form
            class="stack"
            onsubmit={(e) => {
              e.preventDefault();
              run(async () => {
                await api.SaveContact(contact);
                contact = {
                  ID: 0,
                  Name: "",
                  Address: "",
                  Notes: "",
                  Favorite: false,
                };
                await loadContacts();
              });
            }}
          >
            <label>Name<input bind:value={contact.Name} /></label><label
              >Number or SIP address<input
                required
                bind:value={contact.Address}
              /></label
            >
            <div class="stack">
              <label
                >Preferred account<select bind:value={contact.Account}
                  ><option value="">Use selected account</option
                  >{#each snapshot.Accounts as a}<option value={a.Name}
                      >{a.Name}</option
                    >{/each}</select
                ></label
              >
              {#each contact.Numbers || [] as number, i}<div class="row">
                  <input
                    aria-label="Number label"
                    placeholder="Mobile / work"
                    bind:value={number.Label}
                  /><input
                    required
                    aria-label="Additional number"
                    placeholder="Number or SIP address"
                    bind:value={number.Address}
                  /><button
                    type="button"
                    onclick={() =>
                      (contact.Numbers = contact.Numbers.filter(
                        (_, index) => index !== i,
                      ))}>Remove</button
                  >
                </div>{/each}
              <button
                type="button"
                onclick={() =>
                  (contact.Numbers = [
                    ...(contact.Numbers || []),
                    { Label: "", Address: "" },
                  ])}>Add another number</button
              >
            </div>
            <label>Notes<textarea bind:value={contact.Notes}></textarea></label
            ><label class="toggle"
              ><input type="checkbox" bind:checked={contact.Favorite} /> Favorite</label
            >
            <div class="row">
              <button class="primary" type="submit" disabled={busy}
                >Save contact</button
              ><button
                type="button"
                onclick={() =>
                  (contact = {
                    ID: 0,
                    Name: "",
                    Address: "",
                    Notes: "",
                    Favorite: false,
                  })}>Clear</button
              >
            </div>
          </form>
          <details>
            <summary>Import / export contacts</summary>
            <p class="muted small">
              CSV columns: name, address, notes; optional group and label
              preserve multiple numbers. vCard 3.0 and 4.0 are also supported.
              Contacts match by exact primary address, never by name.
            </p>
            <input
              aria-label="CSV or vCard file"
              type="file"
              disabled={busy}
              accept=".csv,.vcf,.vcard,text/csv,text/vcard"
              onchange={(e) =>
                run(async () => {
                  const f = e.currentTarget.files?.[0];
                  csv = "";
                  rawCSV = "";
                  mapCSV = false;
                  clearImportPreview();
                  if (f) {
                    if (f.size > 4194304) throw Error("File exceeds 4 MiB");
                    importFormat = /\.(vcf|vcard)$/i.test(f.name)
                      ? "vcard"
                      : "csv";
                    csv = await f.text();
                    if (importFormat === "csv") {
                      rawCSV = csv;
                      const columns = await api.CSVColumns(rawCSV);
                      mapCSV = !columns.some(
                        (column) => column.toLowerCase() === "address",
                      );
                      if (mapCSV) {
                        csv = "";
                        notice =
                          "Choose the name and number columns before previewing this CSV.";
                        return;
                      }
                    }
                    await previewContacts();
                  }
                })}
            />
            {#if importFormat === "csv" && rawCSV}
              {#if mapCSV}
                <CSVFields
                  disabled={busy}
                  text={rawCSV}
                  onMapped={async (mapped) => {
                    csv = mapped;
                    clearImportPreview();
                    if (mapped) await previewContacts();
                  }}
                />
              {:else}
                <button
                  type="button"
                  disabled={busy}
                  onclick={() => {
                    mapCSV = true;
                    clearImportPreview();
                  }}>Map CSV columns</button
                >
              {/if}
            {/if}
            <label>
              Existing contacts
              <select
                disabled={busy}
                bind:value={mergeImportNumbers}
                onchange={(e) => {
                  mergeImportNumbers = e.currentTarget.value === "true";
                  clearImportPreview();
                  if (csv)
                    run(async () => {
                      await previewContacts();
                    });
                }}
              >
                <option value={false}>Keep existing contacts unchanged</option>
                <option value={true}
                  >Add missing numbers to existing contacts</option
                >
              </select>
            </label>
            <p class="muted small">
              Existing names, notes, labels, favorites and account preferences
              are always kept. Repeated primary addresses in this file use the
              first record.
            </p>
            {#if importPreview}
              <p aria-live="polite">
                {importPreview.Added} new · {importPreview.Updated} updated ·
                {importPreview.Skipped} skipped · {importPreview.Invalid} invalid
              </p>
              {#if importPreview.Invalid}
                <p class="error-text">
                  Nothing will be imported until all invalid records are fixed.
                </p>
                <ul>
                  {#each importPreview.Entries.filter((entry) => entry.Action === "invalid").slice(0, 20) as entry}
                    <li>Record {entry.Record}: {entry.Detail}</li>
                  {/each}
                </ul>
                {#if importPreview.Invalid > 20}
                  <p class="muted small">
                    Showing the first 20 validation errors.
                  </p>
                {/if}
              {/if}
              <div class="preview">
                <table>
                  <thead
                    ><tr><th>Record</th><th>Contact</th><th>Action</th></tr
                    ></thead
                  >
                  <tbody>
                    {#each importPreview.Entries.slice(0, 50) as entry}
                      <tr>
                        <td>{entry.Record}</td>
                        <td>
                          {entry.Contact.Name}<br />{entry.Contact.Address}
                          {#if entry.Contact.Numbers?.length}
                            <br /><span class="muted small">
                              {entry.Contact.Numbers.map((number) =>
                                number.Label
                                  ? `${number.Label}: ${number.Address}`
                                  : number.Address,
                              ).join(" · ")}
                            </span>
                          {/if}
                        </td>
                        <td
                          >{entry.Action}<br /><span class="muted small"
                            >{entry.Detail}</span
                          ></td
                        >
                      </tr>
                    {/each}
                  </tbody>
                </table>
              </div>
              {#if importPreview.Entries.length > 50}
                <p class="muted small">
                  Showing the first 50 records; totals cover the entire file.
                </p>
              {/if}
              <button
                disabled={busy ||
                  importPreview.Invalid > 0 ||
                  importPreview.Added + importPreview.Updated === 0}
                onclick={() =>
                  run(async () => {
                    const source = importSource;
                    if (!source) return;
                    const result = await api.ImportContactFile(
                      source.format,
                      source.text,
                      source.merge,
                    );
                    importPreview = result;
                    if (!result.Committed) return;
                    csv = "";
                    rawCSV = "";
                    mapCSV = false;
                    clearImportPreview();
                    notice = `Imported ${result.Added} contacts, added numbers to ${result.Updated}, skipped ${result.Skipped}`;
                    await loadContacts();
                  })}>Confirm import</button
              >
              <p class="muted small">
                Import checks the current phonebook again; final counts may
                change.
              </p>
            {/if}<button onclick={() => run(() => api.ExportContacts("csv"))}
              >Export CSV</button
            ><button onclick={() => run(() => api.ExportContacts("vcard"))}
              >Export vCard</button
            >
          </details>
          <Directory onChanged={() => loadContacts()} />
        </section>
      </div>
    {:else if page === "history"}
      <History
        {account}
        {run}
        {busy}
        onDial={(h) => {
          account = h.Account;
          dial(h.Remote);
        }}
        onSave={(h) => saveCorrespondent(h.Account, h.Remote)}
      />
    {:else if page === "messages"}
      <Messages
        {account}
        bind:target
        {run}
        {busy}
        onDial={dial}
        onSave={(remote) => saveCorrespondent(account, remote)}
      />
    {:else if page === "settings"}
      <div class="settings-layout">
        <AudioSettings
          bind:audio
          bind:devices
          {busy}
          {run}
          callCount={snapshot.Calls.length}
        />
        <section class="panel desktop-settings">
          <h2>Desktop & privacy</h2>
          <form
            class="stack"
            onsubmit={(e) => {
              e.preventDefault();
              run(() => api.SetPreferences(preferences));
            }}
          >
            <label
              >Default outgoing account<select
                bind:value={preferences.DefaultAccount}
                ><option value="">First enabled account</option
                >{#each snapshot.Accounts as a}<option value={a.Name}
                    >{a.Name}</option
                  >{/each}</select
              ></label
            >
            <label class="toggle"
              ><input
                type="checkbox"
                bind:checked={preferences.Notifications}
              /> Desktop call and message notifications</label
            >
            <label class="toggle"
              ><input
                type="checkbox"
                bind:checked={preferences.CloseToBackground}
                disabled={!desktopCapabilities.Tray}
              /> Closing the window keeps Voiper in the tray</label
            >
            <label class="toggle"
              ><input type="checkbox" bind:checked={preferences.Autostart} /> Start
              Voiper when I log in</label
            >
            <label
              >Keep message history<select
                bind:value={preferences.MessageRetentionDays}
                ><option value={0}>Until I delete it</option><option value={30}
                  >30 days</option
                ><option value={90}>90 days</option><option value={365}
                  >1 year</option
                ></select
              ></label
            >
            <p class="muted small">
              Saving a retention limit deletes older messages from this device.
              Autostart uses this installed executable.
            </p>
            <button class="primary" disabled={busy}>Save preferences</button>
          </form>
          <p class="muted small">
            Notifications: {desktopCapabilities.Notifications
              ? "available"
              : "unavailable"} · Tray: {desktopCapabilities.Tray
              ? "available"
              : "unavailable"} · Password wallet: {desktopCapabilities.Secrets
              ? "available"
              : "unavailable"}
          </p>
          {#each desktopCapabilities.Warnings || [] as warning}<p
              class="muted small"
            >
              {warning}
            </p>{/each}
          <h3>Global keyboard shortcuts</h3>
          <p class="muted small">
            Enable shortcuts for this session through your desktop's permission
            dialog. Choose keys to show the dialer or toggle mute when exactly
            one call is connected and not held. Shortcuts stop when Voiper
            quits.
          </p>
          <p role="status">
            {globalShortcuts.Configuring
              ? "Waiting for desktop permission…"
              : globalShortcuts.Enabled
                ? "Enabled for this session"
                : globalShortcuts.Available
                  ? "Available · disabled"
                  : "Unavailable on this desktop"}
          </p>
          {#if globalShortcuts.Error}<p class="muted small">
              {globalShortcuts.Error}
            </p>{/if}
          <div class="row wrap">
            <button
              disabled={shortcutBusy ||
                globalShortcuts.Enabled ||
                globalShortcuts.Configuring}
              onclick={() => shortcutAction(() => api.EnableGlobalShortcuts())}
              >{globalShortcuts.Available
                ? "Enable shortcuts…"
                : "Check availability / enable…"}</button
            >
            {#if globalShortcuts.Enabled && globalShortcuts.CanConfigure}
              <button
                disabled={shortcutBusy}
                onclick={() =>
                  shortcutAction(() => api.ConfigureGlobalShortcuts())}
                >Configure keys…</button
              >
            {/if}
            {#if globalShortcuts.Enabled || globalShortcuts.Configuring || shortcutBusy}
              <button
                onclick={() =>
                  shortcutAction(() => api.DisableGlobalShortcuts())}
                >Disable / cancel setup</button
              >
            {/if}
          </div>
          {#if globalShortcuts.Enabled && !globalShortcuts.CanConfigure}
            <p class="muted small">
              To change keys on this desktop, disable and enable shortcuts
              again.
            </p>
          {/if}
          {#each globalShortcuts.Bindings || [] as binding}
            <p>
              {binding.Description} · {binding.Trigger ||
                "configured by desktop"}
            </p>
          {/each}
          <button onclick={() => run(() => api.RequestQuit())}
            >Quit Voiper</button
          >
        </section>
      </div>
    {:else if page === "logs"}
      <section class="panel">
        <div class="row spread wrap">
          <h2>Recent application logs</h2>
          <div class="row wrap">
            <button onclick={() => run(() => api.ExportDiagnostics(false))}
              >Export diagnostics</button
            >
            <button class="quiet" onclick={() => (logs = [])}>Clear logs</button
            >
          </div>
        </div>
        <p class="muted small">
          Keeps the latest 500 entries. Call statistics are available in the
          active call view.
        </p>
        <pre class="log">{logs.join("\n") || "No log entries."}</pre>
      </section>
    {/if}
  </main>
</div>
