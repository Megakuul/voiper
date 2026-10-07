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

<div class="
  grid grid-cols-[minmax(230px,_0.8fr)_minmax(0,_2fr)] gap-4 max-[850px]:grid-cols-[1fr]
  [&_>_*]:min-w-0 [&_>_.chat]:max-w-none
">
  <section class="
    panel p-6 bg-surface shadow-panel border-t-[#526271] border-r-border border-b-border border-l-border
    rounded-[6px] min-w-0 flex flex-col gap-[0.8rem] border max-[700px]:p-[1.15rem]
    max-[850px]:max-h-[280px] max-[850px]:overflow-auto [&_>_p]:mb-4 [&_>_button_+_button]:mt-4
    [&_form_+_details_button]:mt-2 [&_form_+_details_button]:mr-2 [&_form_+_details_button]:mb-0
    [&_form_+_details_button]:ml-0 [&_>_h2:first-child]:pb-4 [&_>_h2:first-child]:border-b
    [&_>_h2:first-child]:border-b-border [&_>_.row_+_.row]:mt-4
  ">
    <h2 class={ui.h2}>Conversations</h2>
    <form
      class="row flex items-center gap-[0.7rem] max-[700px]:flex-wrap [&_>_button]:shrink-0 max-[700px]:[&_>_.grow]:basis-[180px]"
      onsubmit={(e) => {
        e.preventDefault();
        run(refresh);
      }}
    >
      <input class={[ui.input, "px-[0.8rem] py-[0.65rem]"]}
        aria-label="Search conversations"
        placeholder="Search conversations"
        bind:value={search}
      /><button class={ui.button}>Search</button>
    </form>
    <form
      class="row flex items-center gap-[0.7rem] max-[700px]:flex-wrap [&_>_button]:shrink-0 max-[700px]:[&_>_.grow]:basis-[180px]"
      onsubmit={(e) => {
        e.preventDefault();
        run(() => openConversation(target));
      }}
    >
      <input class={[ui.input, "px-[0.8rem] py-[0.65rem]"]}
        required
        aria-label="New conversation recipient"
        placeholder="Extension or SIP address"
        bind:value={target}
      /><button class={ui.button} disabled={!account}>New</button>
    </form>
    {#each conversations as c}<button
        class={[ui.control, `
          px-4 py-[0.65rem] rounded-[6px] bg-[#35424e] text-inherit min-h-[38px] shadow-control font-medium
          text-left flex flex-col gap-[0.3rem] w-full wrap-anywhere border border-[#536271]
          [&.active]:text-accent-text [&.active]:border-accent [&_strong]:flex [&_strong]:justify-between
          [&:hover:not(:disabled)]:bg-[#435362] [&:hover:not(:disabled)]:border-[#727e8d]
        `]}
        class:active={selected === c.Remote}
        onclick={() => run(() => openConversation(c.Remote))}
        ><strong
          >{c.Remote}{#if c.Unread}<span class="badge px-[0.7rem] py-[0.3rem] rounded-[4px] bg-[#435362] text-xs ml-auto">{c.Unread}</span
            >{/if}</strong
        ><span class="text-muted text-[0.78rem]">{c.Preview}</span><span class="text-muted text-[0.78rem]"
          >{new Date(c.Updated).toLocaleString()}</span
        ></button
      >{:else}<p class="px-0 py-8 m-0 leading-[1.6] text-muted text-center">No conversations yet.</p>{/each}
  </section>
  <section class="
    panel chat p-6 bg-surface shadow-panel border-t-[#526271] border-r-border border-b-border
    border-l-border rounded-[6px] min-w-0 max-w-[850px] border max-[700px]:p-[1.15rem] [&_>_p]:my-4
    [&_>_button_+_button]:mt-4 [&_form_+_details_button]:mt-2 [&_form_+_details_button]:mr-2
    [&_form_+_details_button]:mb-0 [&_form_+_details_button]:ml-0 [&_>_.quiet]:mt-4
    [&_>_h2:first-child]:pb-4 [&_>_h2:first-child]:border-b [&_>_h2:first-child]:border-b-border
    [&_>_.row_+_.row]:mt-4
  ">
    {#if selected}
      <div class="
        row flex items-center gap-[0.7rem] justify-between max-[700px]:flex-wrap [&_>_button]:shrink-0
        max-[700px]:[&_>_.grow]:basis-[180px]
      ">
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
      {#if confirmDelete}<div class="
        px-4 py-3 bg-accent-surface rounded-[6px] flex justify-between items-center mb-4 border
        border-brand-border [&_button]:px-2 [&_button]:py-[0.2rem] [&_button]:bg-transparent
        [&_button]:border-0
      ">
          <span>Delete this conversation from this device?</span><button
            class={[ui.dangerButton, "danger"]}
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
        class="row flex items-center gap-[0.7rem] max-[700px]:flex-wrap [&_>_button]:shrink-0 max-[700px]:[&_>_.grow]:basis-[180px]"
        onsubmit={(e) => {
          e.preventDefault();
          viewRevision++;
          messages = [];
          loadingOlder = false;
          run(refresh);
        }}
      >
        <input class={[ui.input, "px-[0.8rem] py-[0.65rem]"]}
          aria-label="Search messages"
          placeholder="Search this conversation"
          bind:value={messageSearch}
        /><button class={ui.button}>Search</button>
      </form>
      <p class="m-0 leading-[1.6] text-muted text-[0.78rem]">
        SIP text messaging. Server acceptance does not establish delivery or
        read status.
      </p>
      <div
        class="
          px-0 py-4 flex flex-col gap-4 min-h-[180px] max-h-[50vh] overflow-auto [&_article]:p-4
          [&_article]:self-start [&_article]:max-w-[85%] [&_article]:bg-[#303d49] [&_article]:rounded-[6px]
          [&_article_p]:whitespace-pre-wrap [&_article_p]:wrap-anywhere [&_article.outgoing]:self-end
          [&_article.outgoing]:bg-accent-surface
        "
        bind:this={thread}
        aria-label="Conversation messages"
      >
        {#if hasOlder}<button class={ui.button} disabled={busy} onclick={() => run(older)}
            >Load earlier messages</button
          >{/if}
        {#each messages as m (m.ID)}<article
            class:outgoing={m.Direction === "outgoing"}
          >
            <strong>{m.Direction === "outgoing" ? "You" : m.Remote}</strong>
            <p class="m-0 leading-[1.6]">{m.Body}</p>
            <span class="text-muted text-[0.78rem]"
              >{m.Status} · {new Date(m.Created).toLocaleString()}</span
            >{#if m.Status === "failed"}<button class={ui.button} onclick={() => (draft = m.Body)}
                >Copy to composer</button
              >{/if}
          </article>{:else}<p class="px-0 py-8 m-0 leading-[1.6] text-muted text-center">
            No messages match this view.
          </p>{/each}
      </div>
      <form
        class="row flex items-center gap-[0.7rem] max-[700px]:flex-wrap [&_>_button]:shrink-0 max-[700px]:[&_>_.grow]:basis-[180px]"
        onsubmit={(e) => {
          e.preventDefault();
          run(send);
        }}
      >
        <textarea class="
          px-[0.8rem] py-[0.65rem] outline-offset-3 min-w-0 w-full min-h-[90px] rounded-[6px] bg-[#1c252e]
          text-[#eef0f3] resize-y border border-[#4b5a68] placeholder:text-[#858d99] focus:outline-2
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
          class={[ui.primaryButton, "primary"]}
          disabled={busy || !account || !draft.trim()}>Send</button
        >
      </form>
    {:else}<p class="px-0 py-8 m-0 leading-[1.6] text-muted text-center">
        Select a conversation or enter a recipient to start one.
      </p>{/if}
  </section>
</div>
