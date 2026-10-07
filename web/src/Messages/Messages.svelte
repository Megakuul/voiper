<script>
  import { onMount, tick } from "svelte";
  import * as api from "../../wailsjs/go/app/App.js";
  import { EventsOn } from "../../wailsjs/runtime/runtime.js";
  let { account, target = $bindable(""), run, busy, onDial, onSave } = $props();
  let conversations = $state([]);
  let messages = $state([]);
  let draft = $state("");
  let search = $state("");
  let messageSearch = $state("");
  let selected = $state("");
  let loadingOlder = $state(false);
  let hasOlder = $state(false);
  let confirmDelete = $state(false);
  let thread = $state();
  let refreshing = false;
  let disposed = false;
  let lastAccount = "";
  let viewRevision = 0;

  async function refresh() {
    if (refreshing || disposed || !account) return;
    refreshing = true;
    const currentAccount = account;
    const currentRemote = selected;
    const revision = viewRevision;
    try {
      const nextConversations = await api.Conversations(currentAccount, search);
      if (account !== currentAccount || revision !== viewRevision || disposed)
        return;
      conversations = nextConversations;
      if (currentRemote) {
        const latest = await api.MessagesPage(
          currentAccount,
          currentRemote,
          messageSearch,
          0,
          100,
        );
        if (
          account !== currentAccount ||
          selected !== currentRemote ||
          revision !== viewRevision ||
          disposed
        )
          return;
        const nearBottom =
          !thread ||
          thread.scrollHeight - thread.scrollTop - thread.clientHeight < 80;
        const latestIDs = new Set(latest.map((m) => m.ID));
        messages = [
          ...messages.filter(
            (m) =>
              !latestIDs.has(m.ID) && (!latest.length || m.ID < latest[0].ID),
          ),
          ...latest,
        ];
        if (!loadingOlder) hasOlder = latest.length === 100;
        if (
          latest.length &&
          nearBottom &&
          !messageSearch &&
          document.hasFocus()
        )
          await api.MarkConversationRead(
            currentAccount,
            currentRemote,
            latest.at(-1).ID,
          );
        if (nearBottom) {
          await tick();
          thread?.scrollTo({ top: thread.scrollHeight });
        }
      }
    } finally {
      refreshing = false;
    }
  }
  async function openConversation(remote) {
    viewRevision++;
    selected = remote;
    target = remote;
    messages = [];
    messageSearch = "";
    loadingOlder = false;
    confirmDelete = false;
    await refresh();
  }
  async function older() {
    if (!messages.length) return;
    const currentAccount = account,
      currentRemote = selected;
    const revision = viewRevision;
    const rows = await api.MessagesPage(
      currentAccount,
      currentRemote,
      messageSearch,
      messages[0].ID,
      100,
    );
    if (
      account !== currentAccount ||
      selected !== currentRemote ||
      revision !== viewRevision ||
      disposed
    )
      return;
    messages = [...rows, ...messages];
    loadingOlder = true;
    hasOlder = rows.length === 100;
  }
  async function send() {
    const sendingAccount = account;
    const recipient = selected || target;
    const submitted = draft;
    const revision = viewRevision;
    if (!recipient || !submitted.trim()) return;
    await api.SendMessage(sendingAccount, recipient, submitted);
    if (disposed || account !== sendingAccount || revision !== viewRevision)
      return;
    if (draft === submitted) draft = "";
    selected = recipient;
    await refresh();
  }
  $effect(() => {
    if (account !== lastAccount) {
      viewRevision++;
      lastAccount = account;
      selected = "";
      messages = [];
      conversations = [];
      run(refresh);
    }
  });
  onMount(() => {
    disposed = false;
    if (target) selected = target;
    run(refresh);
    const off = EventsOn("incoming-message", () => run(refresh));
    const timer = setInterval(() => {
      refresh().catch(() => {});
    }, 2000);
    return () => {
      disposed = true;
      off();
      clearInterval(timer);
    };
  });
</script>

<div class="conversation-layout">
  <section class="panel conversation-list">
    <h2>Conversations</h2>
    <form
      class="row"
      onsubmit={(e) => {
        e.preventDefault();
        run(refresh);
      }}
    >
      <input
        aria-label="Search conversations"
        placeholder="Search conversations"
        bind:value={search}
      /><button>Search</button>
    </form>
    <form
      class="row"
      onsubmit={(e) => {
        e.preventDefault();
        run(() => openConversation(target));
      }}
    >
      <input
        required
        aria-label="New conversation recipient"
        placeholder="Extension or SIP address"
        bind:value={target}
      /><button disabled={!account}>New</button>
    </form>
    {#each conversations as c}<button
        class="conversation-choice"
        class:active={selected === c.Remote}
        onclick={() => run(() => openConversation(c.Remote))}
        ><strong
          >{c.Remote}{#if c.Unread}<span class="badge">{c.Unread}</span
            >{/if}</strong
        ><span class="muted small">{c.Preview}</span><span class="muted small"
          >{new Date(c.Updated).toLocaleString()}</span
        ></button
      >{:else}<p class="empty">No conversations yet.</p>{/each}
  </section>
  <section class="panel chat">
    {#if selected}
      <div class="row spread">
        <h2>{selected}</h2>
        <button disabled={busy} onclick={() => onDial(selected)}
          >Open dialer</button
        >
        <button disabled={busy} onclick={() => onSave(selected)}
          >Save contact</button
        >
        <button onclick={() => (confirmDelete = !confirmDelete)}
          >Delete conversation</button
        >
      </div>
      {#if confirmDelete}<div class="banner">
          <span>Delete this conversation from this device?</span><button
            class="danger"
            onclick={() =>
              run(async () => {
                const deletingAccount = account,
                  remote = selected,
                  revision = viewRevision;
                await api.DeleteConversation(deletingAccount, remote);
                if (
                  disposed ||
                  account !== deletingAccount ||
                  revision !== viewRevision
                )
                  return;
                viewRevision++;
                selected = "";
                messages = [];
                confirmDelete = false;
                await refresh();
              })}>Delete</button
          ><button onclick={() => (confirmDelete = false)}>Cancel</button>
        </div>{/if}
      <form
        class="row"
        onsubmit={(e) => {
          e.preventDefault();
          viewRevision++;
          messages = [];
          loadingOlder = false;
          run(refresh);
        }}
      >
        <input
          aria-label="Search messages"
          placeholder="Search this conversation"
          bind:value={messageSearch}
        /><button>Search</button>
      </form>
      <p class="muted small">
        SIP text messaging. Server acceptance does not establish delivery or
        read status.
      </p>
      <div
        class="messages"
        bind:this={thread}
        aria-label="Conversation messages"
      >
        {#if hasOlder}<button disabled={busy} onclick={() => run(older)}
            >Load earlier messages</button
          >{/if}
        {#each messages as m (m.ID)}<article
            class:outgoing={m.Direction === "outgoing"}
          >
            <strong>{m.Direction === "outgoing" ? "You" : m.Remote}</strong>
            <p>{m.Body}</p>
            <span class="muted small"
              >{m.Status} · {new Date(m.Created).toLocaleString()}</span
            >{#if m.Status === "failed"}<button onclick={() => (draft = m.Body)}
                >Copy to composer</button
              >{/if}
          </article>{:else}<p class="empty">
            No messages match this view.
          </p>{/each}
      </div>
      <form
        class="row"
        onsubmit={(e) => {
          e.preventDefault();
          run(send);
        }}
      >
        <textarea
          required
          aria-label="Message"
          placeholder="Write a message… (Ctrl+Enter sends)"
          maxlength="16384"
          bind:value={draft}
          onkeydown={(e) => {
            if (e.key === "Enter" && (e.ctrlKey || e.metaKey)) {
              e.preventDefault();
              if (!busy) run(send);
            }
          }}></textarea><button
          class="primary"
          disabled={busy || !account || !draft.trim()}>Send</button
        >
      </form>
    {:else}<p class="empty">
        Select a conversation or enter a recipient to start one.
      </p>{/if}
  </section>
</div>
