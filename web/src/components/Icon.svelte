<script>
  let { name, size = 20 } = $props();
  const paths = {
    overview: "M3 3h7v7H3z M14 3h7v7h-7z M3 14h7v7H3z M14 14h7v7h-7z",
    phone: "M7 3H4a1 1 0 0 0-1 1c0 9.4 7.6 17 17 17a1 1 0 0 0 1-1v-3l-5-2-2 2a15 15 0 0 1-7-7l2-2z",
    contacts: "M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2 M9 3a4 4 0 1 0 0 8 4 4 0 0 0 0-8 M18 8h4 M18 12h4 M19 16h3",
    history: "M3 11a9 9 0 1 1 2 7 M3 4v7h7 M12 7v5l3 2",
    messages: "M21 11a8 8 0 0 1-8 8H7l-4 3V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z M7 8h10 M7 12h7",
    accounts: "M4 3h16v18H4z M8 7h8 M8 11h8 M8 15h4 M16 17h.01",
    settings: "M4 6h16 M4 12h16 M4 18h16 M8 3v6 M16 9v6 M10 15v6",
    arrow: "M5 12h14 M14 7l5 5-5 5",
    incoming: "M18 6 6 18 M6 7v11h11",
    outgoing: "M6 18 18 6 M7 6h11v11",
    microphone: "M9 5a3 3 0 0 1 6 0v7a3 3 0 0 1-6 0z M5 11v1a7 7 0 0 0 14 0v-1 M12 19v3 M8 22h8",
    speaker: "M11 4 6 8H2v8h4l5 4z M15 8a6 6 0 0 1 0 8 M18 5a10 10 0 0 1 0 14",
  };
</script>

<svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true" class="shrink-0 align-middle"><path d={paths[name] || paths.phone} /></svg>
