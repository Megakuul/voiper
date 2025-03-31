<script>
  import { flip } from "svelte/animate";
  import { Palette } from "../lib/color/color.svelte";
  import { fade } from "svelte/transition";
  let {
    ExceptionRef = $bindable(),
    ...restProps
  } = $props();

  /** @type {string} */
  let search = $state("")

  /** 
   * @param {Object[]} entries 
   * @param {string} search 
   * @returns {Object[]}
   */
  function filterPhonebook(entries, search, count) {
    const filtered = [];
    for (const entry of entries) {
      if (filtered.length >= count) break
      try {
        if (JSON.stringify(entry).toLowerCase().includes(search.toLowerCase())) {
          filtered.push(entry)
        }
        continue
      } catch {
        continue
      }
    }
    return filtered
  }

  function generateDummyPhonebook(count) {
    const entries = [];
    const firstNames = ["Ada", "Bob", "Charlie", "Dana", "Eli", "Fiona", "Gus", "Hank", "Iris", "Jade"];
    const lastNames = ["Smith", "Jones", "Williams", "Brown", "Davis", "Miller", "Wilson", "Moore", "Taylor", "Anderson"];
    const domains = ["example.com", "test.net", "demo.org", "sample.co"];

    for (let i = 1; i <= count; i++) {
      // Create a somewhat unique display name
      const firstName = firstNames[Math.floor(Math.random() * firstNames.length)];
      const lastName = lastNames[Math.floor(Math.random() * lastNames.length)];
      const displayName = `${firstName} ${lastName} ${i}`; // Add index for uniqueness

      // Generate a dummy Swiss-like phone number
      // Starts with +41 7, then 8 random digits
      const randomDigits = Math.floor(10000000 + Math.random() * 90000000);
      const phone = `+417${randomDigits}`;

      // Generate a dummy email address
      const domain = domains[Math.floor(Math.random() * domains.length)];
      const user = `${firstName.toLowerCase()}.${lastName.toLowerCase()}${i}@${domain}`;

      entries.push({
        DisplayName: displayName,
        Phone: phone,
        User: user
      });
    }
    return entries;
  }

  let PhonebookEntries = generateDummyPhonebook(10000)
</script>

<div transition:fade class="flex flex-col md:flex-row gap-6 mx-6" {...restProps}>
  <div style="border-color: {Palette.fgPrimary()}; color: {Palette.fgSecondary()};" class="w-full rounded-xl border-5">
    <button title="mute">
      <svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 1024 1024"><path fill="currentColor" d="m412.16 592.128l-45.44 45.44A191.23 191.23 0 0 1 320 512V256a192 192 0 1 1 384 0v44.352l-64 64V256a128 128 0 1 0-256 0v256c0 30.336 10.56 58.24 28.16 80.128m51.968 38.592A128 128 0 0 0 640 512v-57.152l64-64V512a192 192 0 0 1-287.68 166.528zM314.88 779.968l46.144-46.08A223 223 0 0 0 480 768h64a224 224 0 0 0 224-224v-32a32 32 0 1 1 64 0v32a288 288 0 0 1-288 288v64h64a32 32 0 1 1 0 64H416a32 32 0 1 1 0-64h64v-64c-61.44 0-118.4-19.2-165.12-52.032M266.752 737.6A286.98 286.98 0 0 1 192 544v-32a32 32 0 0 1 64 0v32c0 56.832 21.184 108.8 56.064 148.288z"/><path fill="currentColor" d="M150.72 859.072a32 32 0 0 1-45.44-45.056l704-708.544a32 32 0 0 1 45.44 45.056z"/></svg>
    </button>
    <button title="hold">
      <svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 24 24"><path fill="currentColor" d="M2 6c0-1.886 0-2.828.586-3.414S4.114 2 6 2s2.828 0 3.414.586S10 4.114 10 6v12c0 1.886 0 2.828-.586 3.414S7.886 22 6 22s-2.828 0-3.414-.586S2 19.886 2 18zm12 0c0-1.886 0-2.828.586-3.414S16.114 2 18 2s2.828 0 3.414.586S22 4.114 22 6v12c0 1.886 0 2.828-.586 3.414S19.886 22 18 22s-2.828 0-3.414-.586S14 19.886 14 18z"/></svg>
    </button>
    <button title="hangup">
      <svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 24 24"><path fill="currentColor" d="M12 9c-1.6 0-3.15.25-4.6.72v3.1c0 .4-.23.74-.56.9c-.98.49-1.87 1.12-2.67 1.85c-.17.18-.42.29-.67.29c-.3 0-.55-.12-.73-.3L.29 13.08a1 1 0 0 1-.29-.7c0-.28.11-.53.29-.71C3.34 8.77 7.46 7 12 7s8.66 1.77 11.71 4.67c.18.18.29.43.29.71c0 .27-.11.52-.29.7l-2.48 2.48c-.18.18-.43.3-.73.3a.98.98 0 0 1-.68-.29c-.79-.73-1.68-1.36-2.66-1.85a1 1 0 0 1-.56-.9v-3.1C15.15 9.25 13.6 9 12 9"/></svg>
    </button>
  </div>
  <div class="w-full flex flex-col gap-4">
    <div class="relative" style="color: {Palette.fgSecondary()};">
      <input placeholder="Search" bind:value={search} style="border-color: {Palette.fgPrimary()};" 
        class="w-full font-bold py-2 px-3 rounded-lg border-2 focus:outline-0" />
      <svg class="absolute right-2 top-2.5 bottom-2.5 opacity-70" xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"><g fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2"><path fill="currentColor" fill-opacity="0" stroke-dasharray="40" stroke-dashoffset="40" d="M10.76 13.24c-2.34 -2.34 -2.34 -6.14 0 -8.49c2.34 -2.34 6.14 -2.34 8.49 0c2.34 2.34 2.34 6.14 0 8.49c-2.34 2.34 -6.14 2.34 -8.49 0Z"><animate fill="freeze" attributeName="fill-opacity" begin="0.7s" dur="0.15s" values="0;0.3"/><animate fill="freeze" attributeName="stroke-dashoffset" dur="0.5s" values="40;0"/></path><path stroke-dasharray="12" stroke-dashoffset="12" d="M10.5 13.5l-7.5 7.5"><animate fill="freeze" attributeName="stroke-dashoffset" begin="0.5s" dur="0.2s" values="12;0"/></path></g></svg>
    </div>

    <div class="w-full grid grid-cols-3 gap-3">
      {#each filterPhonebook(PhonebookEntries, search, 18) as entry (entry)}
        <div animate:flip style="border-color: {Palette.fgPrimary()}; color: {Palette.fgSecondary()};"
          class="contact cursor-pointer flex flex-col p-3 rounded-xl border-2">
          <span class="text-xl font-bold overflow-hidden text-nowrap">{entry.DisplayName}</span>
          <span title="{entry.Phone} | {entry.User}" class="relative h-8 font-bold overflow-hidden text-nowrap">
            <span class="text-sm phone absolute inset-0">{entry.Phone}</span>
            <span class="text-sm user absolute inset-0">{entry.User}</span>
          </span>
          <div class="flex flex-row justify-around gap-2">
            <button title="call" aria-label="call" class="w-full flex justify-center p-1.5 rounded-lg hover:bg-slate-400/10 transition-all duration-700">
              <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"><path fill="currentColor" fill-opacity="0" stroke="currentColor" stroke-dasharray="64" stroke-dashoffset="64" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 3c0.5 0 2.5 4.5 2.5 5c0 1 -1.5 2 -2 3c-0.5 1 0.5 2 1.5 3c0.39 0.39 2 2 3 1.5c1 -0.5 2 -2 3 -2c0.5 0 5 2 5 2.5c0 2 -1.5 3.5 -3 4c-1.5 0.5 -2.5 0.5 -4.5 0c-2 -0.5 -3.5 -1 -6 -3.5c-2.5 -2.5 -3 -4 -3.5 -6c-0.5 -2 -0.5 -3 0 -4.5c0.5 -1.5 2 -3 4 -3Z"><animate fill="freeze" attributeName="fill-opacity" begin="0.7s" dur="0.15s" values="0;0.3"/><animate fill="freeze" attributeName="stroke-dashoffset" dur="0.6s" values="64;0"/></path></svg>
            </button>
            <button title="message" aria-label="message" class="w-full flex justify-center p-1.5 rounded-lg hover:bg-slate-400/10 transition-all duration-700">
              <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"><g fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2"><path stroke-dasharray="72" stroke-dashoffset="72" d="M3 19.5v-15.5c0 -0.55 0.45 -1 1 -1h16c0.55 0 1 0.45 1 1v12c0 0.55 -0.45 1 -1 1h-14.5Z"><animate fill="freeze" attributeName="stroke-dashoffset" dur="0.6s" values="72;0"/></path><path stroke-dasharray="10" stroke-dashoffset="10" d="M8 7h8"><animate fill="freeze" attributeName="stroke-dashoffset" begin="0.7s" dur="0.2s" values="10;0"/></path><path stroke-dasharray="10" stroke-dashoffset="10" d="M8 10h8"><animate fill="freeze" attributeName="stroke-dashoffset" begin="1s" dur="0.2s" values="10;0"/></path><path stroke-dasharray="6" stroke-dashoffset="6" d="M8 13h4"><animate fill="freeze" attributeName="stroke-dashoffset" begin="1.3s" dur="0.2s" values="6;0"/></path></g></svg>
            </button>
          </div>
        </div>
      {/each}
    </div>
  </div>
</div>

<style>
  .contact .phone {
    opacity: 1;
    animation: transition-phone 8s linear infinite;
  }

  @keyframes transition-phone {
    0%, 25%, 75%, 100% {
      opacity: 1;
    }
    30%, 70% {
      opacity: 0;
    }
  }

  .contact .user {
    opacity: 0;
    animation: transition-user 8s linear infinite;
  }

  @keyframes transition-user {
    0%, 25%, 75%, 100% {
      opacity: 0;
    }
    30%, 70% {
      opacity: 1;
    }
  }
</style>