<script>
  import { ui } from "./ui.js";
  import { Switch } from "bits-ui";
  import { onMount, tick } from "svelte";
  import * as api from "../wailsjs/go/app/App.js";
  import { EventsOn } from "../wailsjs/runtime/runtime.js";
  import Accounts from "./Home/Home.svelte";
  import Overview from "./Overview.svelte";
  import Phone from "./Phone/Phone.svelte";
  import Messages from "./Messages/Messages.svelte";
  import History from "./History/History.svelte";
  import Directory from "./Contacts/Directory.svelte";
  import CSVFields from "./Contacts/CSVFields.svelte";
  import AudioSettings from "./Audio/AudioSettings.svelte";
  import LiveDiagnostics from "./components/LiveDiagnostics.svelte";
  import Dropdown from "./components/Dropdown.svelte";
  import Sidebar from "./components/Sidebar.svelte";

  let page = $state("overview");
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
  let dndPending = $state(false);
  let dndRevision = 0;
  let account = $state("");
  let accountOptions = $state([]);
  let accountInitialized = false;
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
  let showDiagnostics = $state(false);
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
      const requestedDNDRevision = dndRevision;
      const next = await api.Snapshot();
      // Presence comes from a Go map; stable ordering avoids false UI changes.
      next.Presence?.sort((a, b) =>
        `${a.Account}:${a.Target}:${a.SubscriptionID}`.localeCompare(`${b.Account}:${b.Target}:${b.SubscriptionID}`),
      );
      for (const field of ["Accounts", "Presence", "Audio", "DND", "UnreadMessages"]) {
        // A snapshot started before a toggle must not undo its local feedback.
        if (field === "DND" && (dndPending || requestedDNDRevision !== dndRevision)) continue;
        if (JSON.stringify(next[field]) !== JSON.stringify(snapshot[field]))
          snapshot[field] = next[field];
      }
      // Registration retries must not rebuild account menus while they are open.
      const nextAccountOptions = next.Accounts.map((a) => ({ value: a.Name, label: a.Name }));
      if (JSON.stringify(nextAccountOptions) !== JSON.stringify(accountOptions))
        accountOptions = nextAccountOptions;
      // Active calls also drive elapsed time and their live media meters.
      if (next.Calls.length || snapshot.Calls.length) snapshot.Calls = next.Calls;
      if (!preferencesLoaded) {
        await loadPreferences();
        await acceptDialTarget();
      }
      if (!accountInitialized && snapshot.Accounts.length) {
        if (!account) account = snapshot.Accounts[0].Name;
        accountInitialized = true;
      }
    } catch (e) {
      error = String(e);
    } finally {
      refreshing = false;
    }
  }
  async function setDND(checked) {
    if (dndPending) return;
    const previous = snapshot.DND;
    dndPending = true;
    dndRevision++;
    snapshot.DND = checked;
    error = "";
    try {
      await api.SetDND(checked);
    } catch (e) {
      snapshot.DND = previous;
      error = String(e);
    } finally {
      dndRevision++;
      dndPending = false;
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
    if (!accountInitialized && preferences.DefaultAccount) {
      account = preferences.DefaultAccount;
      accountInitialized = true;
    }
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
    if (event.key === "Escape" && showDiagnostics) {
      showDiagnostics = false;
      document.getElementById("diagnostics-toggle")?.focus();
      return;
    }
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
<dialog class="
  p-6 m-auto max-w-[min(32rem,_90vw)] rounded-[6px] bg-surface [box-shadow:0_16px_64px_#0009] text-inherit
  border border-border [&::backdrop]:bg-[#0009] [&_>_.row]:mt-4
"
  bind:this={quitDialog}
  oncancel={() => (confirmQuit = false)}
  onclose={() => (confirmQuit = false)}
>
  <h2 class={ui.h2}>End active calls?</h2>
  <p class="m-0 leading-[1.6]">Quitting ends all calls and unregisters your accounts.</p>
  <div class="row flex items-center gap-[0.7rem] max-[700px]:flex-wrap [&_>_button]:shrink-0 max-[700px]:[&_>_.grow]:basis-[180px]">
    <button class={ui.button} onclick={() => (confirmQuit = false)}>Keep open</button><button
      class={[ui.dangerButton, "danger"]}
      onclick={() => api.Quit()}>Quit Voiper</button
    >
  </div>
</dialog>
<div class="grid grid-cols-[190px_minmax(0,_1fr)] min-h-screen w-full max-[1100px]:grid-cols-[64px_minmax(0,_1fr)]">
  <Sidebar {page} {navigate} unreadMessages={snapshot.UnreadMessages} callCount={snapshot.Calls.length} />
  <main class="
    px-[clamp(1rem,_2.5vw,_3rem)] pt-7 pb-20 min-w-0 max-[700px]:pt-5 [&_>_*]:min-w-0
    [@media(max-height:_550px)]:pt-4 [&_>_.panel_+_.panel]:mt-6
  ">
    <header class="
      flex items-center justify-between gap-6 mb-6 max-[700px]:items-start max-[700px]:flex-wrap
      max-[700px]:gap-4 [@media(max-height:_550px)]:mb-4 max-[850px]:flex-wrap [&_h1]:wrap-anywhere
    ">
      <div>
        <h1 class="m-0 text-2xl font-semibold tracking-[-0.025em]">
          {{
            overview: "Overview",
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
    <div class="ml-auto max-[850px]:ml-0">
      <div class="flex items-center gap-[0.8rem] text-[0.85rem] text-muted">
        <label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]" for="do-not-disturb">Do not disturb</label>
        <Switch.Root id="do-not-disturb" class={[ui.control, `
          p-[2px] rounded-[6px] bg-[#35424e] text-inherit min-h-[22px] [box-shadow:inset_0_1px_3px_#0004]
          font-medium relative w-[36px] h-[22px] border border-[#626d7b] data-[state=checked]:bg-brand
          data-[state=checked]:border-accent [&:hover:not(:disabled)]:bg-[#435362]
          [&:hover:not(:disabled)]:border-[#727e8d] [&[data-state='checked']:hover:not(:disabled)]:bg-brand
          [&[data-state='checked']:hover:not(:disabled)]:border-accent
        `]}
          checked={snapshot.DND}
          aria-busy={dndPending}
          onCheckedChange={setDND}>
          <Switch.Thumb class="
            block w-[16px] h-[16px] rounded-[3px] bg-[#f0f3f7] [box-shadow:0_1px_3px_#0006] relative left-0 transition-[left]
            duration-100 data-[state=checked]:left-[14px]
          " />
        </Switch.Root>
      </div>
    </div>
      {#if ["overview", "phone", "contacts", "messages"].includes(page)}
        <label class="
          flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5] w-[280px] max-w-[50%] max-[700px]:w-full
          max-[700px]:max-w-none max-[850px]:w-full max-[850px]:max-w-none
        "
          >Outgoing account<Dropdown
            label="Outgoing account"
            bind:value={account}
            options={[
              { value: "", label: "Choose account" },
              ...accountOptions,
            ]}
          /></label
        >
      {/if}
    </header>
    {#if error}<div class="
      px-4 py-3 bg-[#351e28] rounded-[6px] flex justify-between items-center mb-4 border border-[#814351]
      [&_button]:px-2 [&_button]:py-[0.2rem] [&_button]:bg-transparent [&_button]:border-0
    " role="alert">
        <span>{error}</span><button class={ui.button}
          aria-label="Dismiss error"
          onclick={() => (error = "")}>×</button
        >
      </div>{/if}
    {#if notice}<div class="
      px-4 py-3 bg-accent-surface rounded-[6px] flex justify-between items-center mb-4 border
      border-brand-border [&_button]:px-2 [&_button]:py-[0.2rem] [&_button]:bg-transparent
      [&_button]:border-0
    " role="status">
        <span>{notice}</span><button class={ui.button}
          aria-label="Dismiss notice"
          onclick={() => (notice = "")}>×</button
        >
      </div>{/if}

    {#if page === "overview"}
      <Overview {snapshot} bind:account bind:target {run} {busy} manageContacts={() => navigate("contacts")} />
    {:else if page === "phone"}
      <Phone {snapshot} bind:account bind:target {run} {busy} />
    {:else if page === "accounts"}
      <Accounts {snapshot} {run} {busy} />
    {:else if page === "contacts"}
      <div class="grid grid-cols-[minmax(270px,_1fr)_minmax(320px,_1.2fr)] gap-6 max-[1000px]:grid-cols-[1fr] [&_>_*]:min-w-0">
        <section class="
          panel p-6 bg-surface shadow-panel border-t-[#526271] border-r-border border-b-border border-l-border
          rounded-[6px] min-w-0 border max-[700px]:p-[1.15rem] [&_>_p]:mb-4 [&_>_button_+_button]:mt-4
          [&_form_+_details_button]:mt-2 [&_form_+_details_button]:mr-2 [&_form_+_details_button]:mb-0
          [&_form_+_details_button]:ml-0 [&_>_h2:first-child]:pb-4 [&_>_h2:first-child]:border-b
          [&_>_h2:first-child]:border-b-border [&_>_.row_+_.row]:mt-4
        ">
          <h2 class={ui.h2}>Phonebook</h2>
          <form
            class="row flex items-center gap-[0.7rem] max-[700px]:flex-wrap [&_>_button]:shrink-0 max-[700px]:[&_>_.grow]:basis-[180px]"
            onsubmit={(e) => {
              e.preventDefault();
              run(loadContacts);
            }}
          >
            <input class={[ui.input, "px-[0.8rem] py-[0.65rem]"]}
              aria-label="Search contacts"
              placeholder="Name or number"
              bind:value={search}
            /><button class={ui.button} type="submit">Search</button>
          </form>
          <div class="mt-2">
            {#each contacts as c}
              {@const watched = (snapshot.Presence || []).find(
                (p) =>
                  p.Account === account &&
                  !p.Voicemail &&
                  (p.Target === c.Address ||
                    p.Remote === c.Address ||
                    p.Remote.startsWith("sip:" + c.Address + "@")),
              )}
              <div class="
                px-0 py-4 gap-[0.7rem] items-center border-b border-b-border flex-wrap [&_strong]:block
                [&_strong]:wrap-anywhere [&_strong]:mb-[0.3rem] [&_span]:block [&_span]:wrap-anywhere
                [&_button]:text-[0.8rem]
              ">
                <button
                  class="
                    grow px-0 py-[0.35rem] outline-offset-3 cursor-pointer rounded-[6px] bg-transparent text-inherit
                    min-h-[38px] shadow-none font-medium leading-[1.4] transition-colors duration-150 flex-1 min-w-0
                    text-left border-0 border-[#536271] disabled:opacity-[0.45] disabled:cursor-not-allowed
                    [&:hover:not(:disabled)]:bg-[#435362] [&:hover:not(:disabled)]:border-[#727e8d]
                    [&:active:not(:disabled)]:[box-shadow:inset_0_1px_3px_#0004]
                  "
                  onclick={() => {
                    if (c.Account) account = c.Account;
                    dial(c.Address);
                  }}
                  ><strong>{c.Favorite ? "★ " : ""}{c.Name}</strong><span
                    class="text-muted">{c.Address}</span
                  >{#if c.Source === "directory-sync"}<span class="text-muted text-[0.78rem]"
                      >Company directory · updated {new Date(
                        c.Updated,
                      ).toLocaleString()}</span
                    >{/if}{#if watched}<span
                      class="text-[0.78rem]"
                      title={new Date(watched.Updated).toLocaleString()}
                      >{watched.State} · {watched.Source}{watched.Note
                        ? " · " + watched.Note
                        : ""} · updated {new Date(watched.Updated).toLocaleString()}</span
                    >{/if}</button
                ><button class={ui.button}
                  disabled={busy || !account}
                  onclick={() =>
                    run(() =>
                      watched
                        ? api.UnwatchPresence(account, c.Address)
                        : api.WatchPresence(account, c.Address),
                    )}>{watched ? "Stop watching" : "Watch status"}</button
                >{#each c.Numbers || [] as n}<button class={ui.button}
                    title={n.Address}
                    onclick={() => {
                      if (c.Account) account = c.Account;
                      dial(n.Address);
                    }}>{n.Label || n.Address}</button
                  >{/each}
                <button class={ui.button}
                  onclick={() => {
                    if (c.Account) account = c.Account;
                    target = c.Address;
                    page = "messages";
                  }}>Message</button
                >
                <button class={ui.button}
                  onclick={() =>
                    (contact = {
                      ...c,
                      Numbers: (c.Numbers || []).map((n) => ({ ...n })),
                    })}>Edit</button
                ><button
                  class={[ui.quietButton, "quiet"]}
                  onclick={() =>
                    run(async () => {
                      await api.DeleteContact(c.ID);
                      await loadContacts();
                    })}>Delete</button
                >
              </div>{:else}<p class="px-0 py-8 m-0 leading-[1.6] text-muted text-center">
                Add a contact or import your phonebook.
              </p>{/each}
          </div>
          {#if contacts.length >= 100}<button class={ui.button}
              disabled={busy}
              onclick={() => run(() => loadContacts(true))}
              >Load more contacts</button
            >{/if}
          <p class="m-0 leading-[1.6] text-muted text-[0.78rem]">
            Search supports names and all saved numbers. Imported directory
            entries remain available offline.
          </p>
        </section>
        <section class="
          panel p-6 bg-surface shadow-panel border-t-[#526271] border-r-border border-b-border border-l-border
          rounded-[6px] min-w-0 border max-[700px]:p-[1.15rem] [&_>_p]:mb-4 [&_>_button_+_button]:mt-4
          [&_form_+_details_button]:mt-2 [&_form_+_details_button]:mr-2 [&_form_+_details_button]:mb-0
          [&_form_+_details_button]:ml-0 [&_>_h2:first-child]:pb-4 [&_>_h2:first-child]:border-b
          [&_>_h2:first-child]:border-b-border [&_>_.row_+_.row]:mt-4
        ">
          <h2 class={ui.h2}>{contact.ID ? "Edit contact" : "New contact"}</h2>
          {#if contact.Source === "directory-sync"}<p class="m-0 leading-[1.6] text-muted text-[0.78rem]">
              Saving changes keeps your own copy. Directory refreshes will
              preserve it.
            </p>{/if}
          {#if contact.DirectoryID && contact.Source !== "directory-sync"}
            <p class="m-0 leading-[1.6] text-muted text-[0.78rem]">
              This is your local copy of a directory contact.
            </p>
            <button class={ui.button}
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
            class="stack flex flex-col gap-4"
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
            <label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]">Name<input class={[ui.input, "px-[0.8rem] py-[0.65rem]"]} bind:value={contact.Name} /></label><label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]"
              >Number or SIP address<input class={[ui.input, "px-[0.8rem] py-[0.65rem]"]}
                required
                bind:value={contact.Address}
              /></label
            >
            <div class="stack flex flex-col gap-4">
              <label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]"
                >Preferred account<Dropdown
                  label="Preferred account"
                  bind:value={contact.Account}
                  options={[
                    { value: "", label: "Use selected account" },
                    ...accountOptions,
                  ]}
                /></label
              >
              {#each contact.Numbers || [] as number, i}<div class="row flex items-center gap-[0.7rem] max-[700px]:flex-wrap [&_>_button]:shrink-0 max-[700px]:[&_>_.grow]:basis-[180px]">
                  <input class={[ui.input, "px-[0.8rem] py-[0.65rem]"]}
                    aria-label="Number label"
                    placeholder="Mobile / work"
                    bind:value={number.Label}
                  /><input class={[ui.input, "px-[0.8rem] py-[0.65rem]"]}
                    required
                    aria-label="Additional number"
                    placeholder="Number or SIP address"
                    bind:value={number.Address}
                  /><button class={ui.button}
                    type="button"
                    onclick={() =>
                      (contact.Numbers = contact.Numbers.filter(
                        (_, index) => index !== i,
                      ))}>Remove</button
                  >
                </div>{/each}
              <button class={ui.button}
                type="button"
                onclick={() =>
                  (contact.Numbers = [
                    ...(contact.Numbers || []),
                    { Label: "", Address: "" },
                  ])}>Add another number</button
              >
            </div>
            <label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]">Notes<textarea class="
              px-[0.8rem] py-[0.65rem] outline-offset-3 min-w-0 w-full min-h-[90px] rounded-[6px] bg-[#1c252e]
              text-[#eef0f3] resize-y border border-[#4b5a68] placeholder:text-[#858d99] focus:outline-2
              focus:outline-solid focus:outline-accent focus:border-transparent
            " bind:value={contact.Notes}></textarea></label
            ><label class="flex flex-row gap-[0.65rem] text-[0.9rem] text-[#c8cdd5] items-center min-h-[32px] cursor-pointer"
              ><input class={ui.checkbox} type="checkbox" bind:checked={contact.Favorite} /> Favorite</label
            >
            <div class="row flex items-center gap-[0.7rem] max-[700px]:flex-wrap [&_>_button]:shrink-0 max-[700px]:[&_>_.grow]:basis-[180px]">
              <button class={[ui.primaryButton, "primary"]} type="submit" disabled={busy}
                >Save contact</button
              ><button class={ui.button}
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
          <details class={[ui.details, "mt-[1.3rem] border-t border-t-border pt-4"]}>
            <summary class={[ui.summary, "px-0 py-1"]}>Import / export contacts</summary>
            <p class="m-0 leading-[1.6] text-muted text-[0.78rem]">
              CSV columns: name, address, notes; optional group and label
              preserve multiple numbers. vCard 3.0 and 4.0 are also supported.
              Contacts match by exact primary address, never by name.
            </p>
            <input class={[ui.input, "px-[0.8rem] py-[0.65rem]"]}
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
                <button class={ui.button}
                  type="button"
                  disabled={busy}
                  onclick={() => {
                    mapCSV = true;
                    clearImportPreview();
                  }}>Map CSV columns</button
                >
              {/if}
            {/if}
            <label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]">
              Existing contacts
              <Dropdown
                label="Existing contacts"
                disabled={busy}
                bind:value={mergeImportNumbers}
                options={[
                  { value: false, label: "Keep existing contacts unchanged" },
                  {
                    value: true,
                    label: "Add missing numbers to existing contacts",
                  },
                ]}
                onchange={() => {
                  clearImportPreview();
                  if (csv) run(previewContacts);
                }}
              />
            </label>
            <p class="m-0 leading-[1.6] text-muted text-[0.78rem]">
              Existing names, notes, labels, favorites and account preferences
              are always kept. Repeated primary addresses in this file use the
              first record.
            </p>
            {#if importPreview}
              <p class="m-0 leading-[1.6]" aria-live="polite">
                {importPreview.Added} new · {importPreview.Updated} updated ·
                {importPreview.Skipped} skipped · {importPreview.Invalid} invalid
              </p>
              {#if importPreview.Invalid}
                <p class="m-0 leading-[1.6] text-[#f9a8b9]">
                  Nothing will be imported until all invalid records are fixed.
                </p>
                <ul>
                  {#each importPreview.Entries.filter((entry) => entry.Action === "invalid").slice(0, 20) as entry}
                    <li>Record {entry.Record}: {entry.Detail}</li>
                  {/each}
                </ul>
                {#if importPreview.Invalid > 20}
                  <p class="m-0 leading-[1.6] text-muted text-[0.78rem]">
                    Showing the first 20 validation errors.
                  </p>
                {/if}
              {/if}
              <div class="p-4 whitespace-pre-wrap wrap-anywhere max-h-[180px] overflow-auto text-xs bg-[#1c252e] rounded-[6px]">
                <table class="w-full border-collapse text-left">
                  <thead
                    ><tr><th class="p-[0.7rem] border-b border-b-border align-top font-semibold text-muted">Record</th><th class="p-[0.7rem] border-b border-b-border align-top font-semibold text-muted">Contact</th><th class="p-[0.7rem] border-b border-b-border align-top font-semibold text-muted">Action</th></tr
                    ></thead
                  >
                  <tbody>
                    {#each importPreview.Entries.slice(0, 50) as entry}
                      <tr>
                        <td class="p-[0.7rem] border-b border-b-border align-top">{entry.Record}</td>
                        <td class="p-[0.7rem] border-b border-b-border align-top">
                          {entry.Contact.Name}<br />{entry.Contact.Address}
                          {#if entry.Contact.Numbers?.length}
                            <br /><span class="text-muted text-[0.78rem]">
                              {entry.Contact.Numbers.map((number) =>
                                number.Label
                                  ? `${number.Label}: ${number.Address}`
                                  : number.Address,
                              ).join(" · ")}
                            </span>
                          {/if}
                        </td>
                        <td class="p-[0.7rem] border-b border-b-border align-top"
                          >{entry.Action}<br /><span class="text-muted text-[0.78rem]"
                            >{entry.Detail}</span
                          ></td
                        >
                      </tr>
                    {/each}
                  </tbody>
                </table>
              </div>
              {#if importPreview.Entries.length > 50}
                <p class="m-0 leading-[1.6] text-muted text-[0.78rem]">
                  Showing the first 50 records; totals cover the entire file.
                </p>
              {/if}
              <button class={ui.button}
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
              <p class="m-0 leading-[1.6] text-muted text-[0.78rem]">
                Import checks the current phonebook again; final counts may
                change.
              </p>
            {/if}<button class={ui.button} onclick={() => run(() => api.ExportContacts("csv"))}
              >Export CSV</button
            ><button class={ui.button} onclick={() => run(() => api.ExportContacts("vcard"))}
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
      <div class="
        grid grid-cols-[minmax(0,_1.15fr)_minmax(0,_1fr)] gap-6 items-start max-[900px]:grid-cols-[1fr]
        [&_>_*]:min-w-0 [&_>_.narrow]:max-w-none max-[900px]:[&_>_.narrow]:max-w-none
      ">
        <AudioSettings
          bind:audio
          bind:devices
          {busy}
          {run}
          callCount={snapshot.Calls.length}
        />
        <section class="
          panel p-6 bg-surface shadow-panel border-t-[#526271] border-r-border border-b-border border-l-border
          rounded-[6px] min-w-0 border max-[700px]:p-[1.15rem] [&_>_p]:mb-4 [&_>_h3]:pt-5 [&_>_h3]:border-t
          [&_>_h3]:border-t-border [&_>_button_+_button]:mt-4 [&_form_>_button]:self-start
          [&_form_+_details_button]:mt-2 [&_form_+_details_button]:mr-2 [&_form_+_details_button]:mb-0
          [&_form_+_details_button]:ml-0 [&_>_h2:first-child]:pb-4 [&_>_h2:first-child]:border-b
          [&_>_h2:first-child]:border-b-border [&_>_.row_+_.row]:mt-4 [&_>_:not(:first-child)]:mt-5
        ">
          <h2 class={ui.h2}>Desktop & privacy</h2>
          <form
            class="stack flex flex-col gap-4"
            onsubmit={(e) => {
              e.preventDefault();
              run(() => api.SetPreferences(preferences));
            }}
          >
            <label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]"
              >Default outgoing account<Dropdown
                label="Default outgoing account"
                bind:value={preferences.DefaultAccount}
                options={[
                  { value: "", label: "First enabled account" },
                  ...accountOptions,
                ]}
              /></label
            >
            <label class="flex flex-row gap-[0.65rem] text-[0.9rem] text-[#c8cdd5] items-center min-h-[32px] cursor-pointer"
              ><input class={ui.checkbox}
                type="checkbox"
                bind:checked={preferences.Notifications}
              /> Desktop call and message notifications</label
            >
            <label class="flex flex-row gap-[0.65rem] text-[0.9rem] text-[#c8cdd5] items-center min-h-[32px] cursor-pointer"
              ><input class={ui.checkbox}
                type="checkbox"
                bind:checked={preferences.CloseToBackground}
                disabled={!desktopCapabilities.Tray}
              /> Closing the window keeps Voiper in the tray</label
            >
            <label class="flex flex-row gap-[0.65rem] text-[0.9rem] text-[#c8cdd5] items-center min-h-[32px] cursor-pointer"
              ><input class={ui.checkbox} type="checkbox" bind:checked={preferences.Autostart} /> Start
              Voiper when I log in</label
            >
            <label class="flex flex-col gap-2 text-[0.9rem] text-[#c8cdd5]"
              >Keep message history<Dropdown
                label="Keep message history"
                bind:value={preferences.MessageRetentionDays}
                options={[
                  { value: 0, label: "Until I delete it" },
                  { value: 30, label: "30 days" },
                  { value: 90, label: "90 days" },
                  { value: 365, label: "1 year" },
                ]}
              /></label
            >
            <p class="m-0 leading-[1.6] text-muted text-[0.78rem]">
              Saving a retention limit deletes older messages from this device.
              Autostart uses this installed executable.
            </p>
            <button class={[ui.primaryButton, "primary"]} disabled={busy}>Save preferences</button>
          </form>
          <p class="m-0 leading-[1.6] text-muted text-[0.78rem]">
            Notifications: {desktopCapabilities.Notifications
              ? "available"
              : "unavailable"} · Tray: {desktopCapabilities.Tray
              ? "available"
              : "unavailable"} · Password wallet: {desktopCapabilities.Secrets
              ? "available"
              : "unavailable"}
          </p>
          {#each desktopCapabilities.Warnings || [] as warning}<p
              class="m-0 leading-[1.6] text-muted text-[0.78rem]"
            >
              {warning}
            </p>{/each}
          <h3 class="m-0 text-base font-semibold">Global keyboard shortcuts</h3>
          <p class="m-0 leading-[1.6] text-muted text-[0.78rem]">
            Enable shortcuts for this session through your desktop's permission
            dialog. Choose keys to show the dialer or toggle mute when exactly
            one call is connected and not held. Shortcuts stop when Voiper
            quits.
          </p>
          <p class="m-0 leading-[1.6]" role="status">
            {globalShortcuts.Configuring
              ? "Waiting for desktop permission…"
              : globalShortcuts.Enabled
                ? "Enabled for this session"
                : globalShortcuts.Available
                  ? "Available · disabled"
                  : "Unavailable on this desktop"}
          </p>
          {#if globalShortcuts.Error}<p class="m-0 leading-[1.6] text-muted text-[0.78rem]">
              {globalShortcuts.Error}
            </p>{/if}
          <div class="
            row flex items-center gap-[0.7rem] flex-wrap max-[700px]:flex-wrap [&_>_button]:shrink-0
            max-[700px]:[&_>_.grow]:basis-[180px]
          ">
            <button class={ui.button}
              disabled={shortcutBusy ||
                globalShortcuts.Enabled ||
                globalShortcuts.Configuring}
              onclick={() => shortcutAction(() => api.EnableGlobalShortcuts())}
              >{globalShortcuts.Available
                ? "Enable shortcuts…"
                : "Check availability / enable…"}</button
            >
            {#if globalShortcuts.Enabled && globalShortcuts.CanConfigure}
              <button class={ui.button}
                disabled={shortcutBusy}
                onclick={() =>
                  shortcutAction(() => api.ConfigureGlobalShortcuts())}
                >Configure keys…</button
              >
            {/if}
            {#if globalShortcuts.Enabled || globalShortcuts.Configuring || shortcutBusy}
              <button class={ui.button}
                onclick={() =>
                  shortcutAction(() => api.DisableGlobalShortcuts())}
                >Disable / cancel setup</button
              >
            {/if}
          </div>
          {#if globalShortcuts.Enabled && !globalShortcuts.CanConfigure}
            <p class="m-0 leading-[1.6] text-muted text-[0.78rem]">
              To change keys on this desktop, disable and enable shortcuts
              again.
            </p>
          {/if}
          {#each globalShortcuts.Bindings || [] as binding}
            <p class="m-0 leading-[1.6]">
              {binding.Description} · {binding.Trigger ||
                "configured by desktop"}
            </p>
          {/each}
          <button class={ui.button} onclick={() => run(() => api.RequestQuit())}
            >Quit Voiper</button
          >
        </section>
      </div>
    {:else if page === "logs"}
      <LiveDiagnostics {snapshot} {logs} {run} {busy} />
      <section class="
        panel p-6 bg-surface shadow-panel border-t-[#526271] border-r-border border-b-border border-l-border
        rounded-[6px] min-w-0 border max-[700px]:p-[1.15rem] [&_>_p]:mb-4 [&_>_button_+_button]:mt-4
        [&_form_+_details_button]:mt-2 [&_form_+_details_button]:mr-2 [&_form_+_details_button]:mb-0
        [&_form_+_details_button]:ml-0 [&_>_h2:first-child]:pb-4 [&_>_h2:first-child]:border-b
        [&_>_h2:first-child]:border-b-border [&_>_.row_+_.row]:mt-4
      ">
        <div class="
          row flex items-center gap-[0.7rem] justify-between flex-wrap max-[700px]:flex-wrap
          [&_>_button]:shrink-0 max-[700px]:[&_>_.grow]:basis-[180px]
        ">
          <h2 class={ui.h2}>Recent application logs</h2>
          <div class="
            row flex items-center gap-[0.7rem] flex-wrap max-[700px]:flex-wrap [&_>_button]:shrink-0
            max-[700px]:[&_>_.grow]:basis-[180px]
          ">
            <button class={ui.button} onclick={() => run(() => api.ExportDiagnostics(false))}
              >Export diagnostics</button
            >
            <button class={[ui.quietButton, "quiet"]} onclick={() => (logs = [])}>Clear logs</button
            >
          </div>
        </div>
        <p class="m-0 leading-[1.6] text-muted text-[0.78rem]">
          Keeps the latest 500 entries. Call statistics are available in the
          active call view.
        </p>
        <pre class="p-4 whitespace-pre-wrap wrap-anywhere max-h-[60vh] overflow-auto text-xs bg-[#1c252e] rounded-[6px]">{logs.join("\n") || "No log entries."}</pre>
      </section>
    {/if}
  </main>
  {#if showDiagnostics}
    <aside id="live-diagnostics" class="
      fixed z-[105] right-4 bottom-[56px] w-[min(620px,_calc(100vw_-_2rem))] max-h-[calc(100vh_-_110px)]
      overflow-y-auto rounded-[6px] bg-surface [box-shadow:0_12px_40px_#0009]
      [animation:panel-enter_140ms_ease-out] border border-[#586575] [&_.panel]:shadow-none
      [&_.panel]:border-0
    " aria-label="Live diagnostics">
      <div class="px-4 py-2 flex items-center justify-between border-b border-b-border bg-[#303d49]">
        <span class="text-xs text-muted font-semibold">CONNECTION & CALL QUALITY</span>
        <button class={[ui.control, `
          p-[6px] rounded-[4px] bg-transparent text-muted min-h-[32px] shadow-none font-medium inline-flex
          items-center justify-center w-[32px] h-[32px] border border-transparent [&_svg]:w-[18px]
          [&_svg]:h-[18px] [&_svg]:[fill:none] [&_svg]:[stroke:currentColor] [&_svg]:[stroke-width:1.6]
          [&_svg]:[stroke-linecap:round] [&_svg]:[stroke-linejoin:round] [&:hover:not(:disabled)]:bg-[#435362]
          [&:hover:not(:disabled)]:border-[#727e8d]
        `]} aria-label="Close live diagnostics" title="Close diagnostics"
          onclick={() => { showDiagnostics = false; document.getElementById("diagnostics-toggle")?.focus(); }}>×</button>
      </div>
      <LiveDiagnostics {snapshot} {logs} {run} {busy} />
    </aside>
  {/if}
  <footer class="
    px-[clamp(1rem,_2.5vw,_3rem)] py-[4px] fixed z-[110] bottom-0 left-0 right-0 flex justify-between
    items-center gap-3 min-h-[44px] [transform:translateZ(0)] border-t border-t-[#4b5a68]
    bg-surface-secondary [box-shadow:0_-1px_4px_#0003]
  " aria-label="Connection status and diagnostics">
    <div class="
      flex items-center gap-[0.65rem] text-xs text-muted wrap-anywhere max-[450px]:gap-[0.3rem]
      max-[450px]:text-[0.7rem] [&_.dot]:shrink-0
    ">
      <span class="dot w-[7px] h-[7px] rounded-full bg-[#9c7084] [&.online]:bg-[#73d3a1]" class:online={snapshot.Accounts.some((a) => a.State === "registered")}></span>
      <span>{snapshot.Accounts.filter((a) => a.State === "registered").length}/{snapshot.Accounts.length} accounts connected</span>
      {#if snapshot.Calls.length}<span>· {snapshot.Calls.length} calls</span>{/if}
    </div>
    <div class="flex items-center gap-[0.65rem] max-[450px]:gap-[0.2rem]">
      <button id="diagnostics-toggle" class={[ui.control, `
        p-[6px] rounded-[4px] bg-transparent text-muted min-h-[32px] shadow-none font-medium inline-flex
        items-center justify-center w-[32px] h-[32px] border border-transparent [&.active]:text-accent-text
        [&.active]:bg-accent-surface [&.active]:border-brand-border [&_svg]:w-[18px] [&_svg]:h-[18px]
        [&_svg]:[fill:none] [&_svg]:[stroke:currentColor] [&_svg]:[stroke-width:1.6]
        [&_svg]:[stroke-linecap:round] [&_svg]:[stroke-linejoin:round] [&:hover:not(:disabled)]:bg-[#435362]
        [&:hover:not(:disabled)]:border-[#727e8d]
      `]} class:active={showDiagnostics}
        aria-label="Live connection diagnostics" title="Live connection diagnostics"
        aria-expanded={showDiagnostics} aria-controls="live-diagnostics"
        onclick={() => (showDiagnostics = !showDiagnostics)}>
        <svg aria-hidden="true" viewBox="0 0 24 24"><path d="M3 12h4l3-7 4 14 3-7h4" /></svg>
      </button>
      <button class={[ui.control, `
        p-[6px] rounded-[4px] bg-transparent text-muted min-h-[32px] shadow-none font-medium inline-flex
        items-center justify-center w-[32px] h-[32px] border border-transparent [&.active]:text-accent-text
        [&.active]:bg-accent-surface [&.active]:border-brand-border [&_svg]:w-[18px] [&_svg]:h-[18px]
        [&_svg]:[fill:none] [&_svg]:[stroke:currentColor] [&_svg]:[stroke-width:1.6]
        [&_svg]:[stroke-linecap:round] [&_svg]:[stroke-linejoin:round] [&:hover:not(:disabled)]:bg-[#435362]
        [&:hover:not(:disabled)]:border-[#727e8d]
      `]} class:active={page === "logs"} aria-label="Open diagnostic logs" title="Diagnostic logs"
        onclick={() => { showDiagnostics = false; navigate("logs"); }}>
        <svg aria-hidden="true" viewBox="0 0 24 24"><rect x="4" y="3" width="16" height="18" rx="2"/><path d="M8 8h8M8 12h8M8 16h5" /></svg>
      </button>
      <button class={[ui.control, `
        p-[6px] rounded-[4px] bg-transparent text-muted min-h-[32px] shadow-none font-medium inline-flex
        items-center justify-center w-[32px] h-[32px] border border-transparent [&_svg]:w-[18px]
        [&_svg]:h-[18px] [&_svg]:[fill:none] [&_svg]:[stroke:currentColor] [&_svg]:[stroke-width:1.6]
        [&_svg]:[stroke-linecap:round] [&_svg]:[stroke-linejoin:round] [&:hover:not(:disabled)]:bg-[#435362]
        [&:hover:not(:disabled)]:border-[#727e8d]
      `]} disabled={busy} aria-label="Export diagnostics" title="Export diagnostics"
        onclick={() => run(() => api.ExportDiagnostics(false))}>
        <svg aria-hidden="true" viewBox="0 0 24 24"><path d="M12 3v12m-4-4 4 4 4-4M4 16v5h16v-5" /></svg>
      </button>
    </div>
  </footer>
</div>
