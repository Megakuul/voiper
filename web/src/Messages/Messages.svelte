<script>
  import { ui } from "../ui.js";
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
      if (JSON.stringify(nextConversations) !== JSON.stringify(conversations))
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
        const nextMessages = [
          ...messages.filter(
            (m) =>
              !latestIDs.has(m.ID) && (!latest.length || m.ID < latest[0].ID),
          ),
          ...latest,
        ];
        const messagesChanged = JSON.stringify(nextMessages) !== JSON.stringify(messages);
        if (messagesChanged) messages = nextMessages;
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
        if (nearBottom && messagesChanged) {
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

<div class="grid grid-cols-[minmax(230px,_0.8fr)_minmax(0,_2fr)] gap-4 max-[850px]:grid-cols-[1fr] *:min-w-0">
  <section class={[ui.panel, "flex flex-col gap-[0.8rem] max-[850px]:max-h-[280px] max-[850px]:overflow-auto"]}>
    <h2 class={ui.h2}>Conversations</h2>
    <form
      class={ui.row}
      onsubmit={(e) => {
        e.preventDefault();
        run(refresh);
      }}
    >
      <input class={ui.field}
        aria-label="Search conversations"
        placeholder="Search conversations"
        bind:value={search}
      /><button class={ui.button}>Search</button>
    </form>
    <form
      class={ui.row}
      onsubmit={(e) => {
        e.preventDefault();
        run(() => openConversation(target));
      }}
    >
      <input class={ui.field}
        required
        aria-label="New conversation recipient"
        placeholder="Extension or SIP address"
        bind:value={target}
      /><button class={ui.button} disabled={!account}>New</button>
    </form>
    {#each conversations as c}<button
        class={[ui.button, "flex w-full flex-col gap-[0.3rem] text-left wrap-anywhere", selected === c.Remote && "text-accent-text border-accent"]}
        onclick={() => run(() => openConversation(c.Remote))}
        ><strong class="flex justify-between"
          >{c.Remote}{#if c.Unread}<span class={[ui.badge, "ml-auto"]}>{c.Unread}</span
            >{/if}</strong
        ><span class="text-muted text-[0.78rem]">{c.Preview}</span><span class="text-muted text-[0.78rem]"
          >{new Date(c.Updated).toLocaleString()}</span
        ></button
      >{:else}<p class={ui.empty}>No conversations yet.</p>{/each}
  </section>
  <section class={[ui.card, "[&>p]:my-4 [&>.quiet]:mt-4"]}>
    {#if selected}
      <div class={[ui.row, "justify-between"]}>
        <h2 class={ui.h2}>{selected}</h2>
        <button class={ui.button} disabled={busy} onclick={() => onDial(selected)}
          >Open dialer</button
        >
        <button class={ui.button} disabled={busy} onclick={() => onSave(selected)}
          >Save contact</button
        >
        <button class={ui.button} onclick={() => (confirmDelete = !confirmDelete)}
          >Delete conversation</button
        >
      </div>
      {#if confirmDelete}<div class="flex flex-wrap items-center gap-2 mt-4 mb-4 px-4 py-3 rounded-[6px] border border-brand-border bg-accent-surface">
          <span class="mr-auto">Delete this conversation from this device?</span><button
            class={ui.dangerButton}
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
          ><button class={ui.button} onclick={() => (confirmDelete = false)}>Cancel</button>
        </div>{/if}
      <form
        class={ui.row}
        onsubmit={(e) => {
          e.preventDefault();
          viewRevision++;
          messages = [];
          loadingOlder = false;
          run(refresh);
        }}
      >
        <input class={ui.field}
          aria-label="Search messages"
          placeholder="Search this conversation"
          bind:value={messageSearch}
        /><button class={ui.button}>Search</button>
      </form>
      <p class={ui.hint}>
        SIP text messaging. Server acceptance does not establish delivery or
        read status.
      </p>
      <div
        class="flex flex-col gap-4 min-h-[180px] max-h-[50vh] overflow-auto py-4"
        bind:this={thread}
        aria-label="Conversation messages"
      >
        {#if hasOlder}<button class={ui.button} disabled={busy} onclick={() => run(older)}
            >Load earlier messages</button
          >{/if}
        {#each messages as m (m.ID)}<article
            class={["max-w-[85%] p-4 rounded-[6px]", m.Direction === "outgoing" ? "self-end bg-accent-surface" : "self-start bg-[#303d49]"]}
          >
            <strong>{m.Direction === "outgoing" ? "You" : m.Remote}</strong>
            <p class={[ui.text, "whitespace-pre-wrap wrap-anywhere"]}>{m.Body}</p>
            <span class="text-muted text-[0.78rem]"
              >{m.Status} · {new Date(m.Created).toLocaleString()}</span
            >{#if m.Status === "failed"}<button class={ui.button} onclick={() => (draft = m.Body)}
                >Copy to composer</button
              >{/if}
          </article>{:else}<p class={ui.empty}>
            No messages match this view.
          </p>{/each}
      </div>
      <form
        class={ui.row}
        onsubmit={(e) => {
          e.preventDefault();
          run(send);
        }}
      >
        <textarea class="
          px-[0.8rem] py-[0.65rem] outline-offset-3 min-w-0 w-full min-h-[90px] rounded-[6px] bg-field
          text-foreground resize-y border border-field-border placeholder:text-[#858d99] focus:outline-2
          focus:outline-solid focus:outline-accent focus:border-transparent
        "
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
          class={ui.primaryButton}
          disabled={busy || !account || !draft.trim()}>Send</button
        >
      </form>
    {:else}<p class={ui.empty}>
        Select a conversation or enter a recipient to start one.
      </p>{/if}
  </section>
</div>
