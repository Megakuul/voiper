// Shared Tailwind utilities for controls used throughout the application.
const control = `
    outline-offset-3 cursor-pointer leading-[1.4] transition-colors duration-150 disabled:opacity-[0.45]
    disabled:cursor-not-allowed [&:active:not(:disabled)]:shadow-[inset_0_1px_3px_#0004]
  `;

export const ui = {
  checkbox: `
    appearance-none size-[20px] min-h-[20px] flex-[0_0_20px] m-0 p-0 border border-[#697583] rounded-[4px] bg-[#1c252e]
    bg-center bg-no-repeat bg-size-[14px_14px] shadow-none cursor-pointer transition-colors duration-150
    checked:border-accent checked:bg-brand indeterminate:border-accent indeterminate:bg-brand
    enabled:hover:border-accent-text focus:border-accent focus:outline-none focus-visible:outline-2
    focus-visible:outline-solid focus-visible:outline-accent focus-visible:outline-offset-3
    disabled:opacity-40 disabled:cursor-not-allowed
  `,
  range: `
    appearance-none w-full min-w-0 min-h-[28px] p-0 border-0 bg-transparent cursor-pointer outline-offset-3
    focus:outline-2 focus:outline-solid focus:outline-accent disabled:opacity-40
    disabled:cursor-not-allowed
  `,

  button: `
    ${control} px-4 py-[0.65rem] rounded-[6px] bg-[#35424e] text-inherit min-h-[38px] shadow-control
    font-medium border border-[#536271] [&:hover:not(:disabled)]:bg-[#435362]
    [&:hover:not(:disabled)]:border-[#727e8d]
  `,
  h2: `
    mx-0 mt-0 mb-4 text-[1.15rem] font-semibold
  `,
  control,
  dangerButton: `
    px-4 py-[0.65rem] outline-offset-3 cursor-pointer rounded-[6px] bg-[#382027] text-[#ffc2cc] min-h-[38px]
    shadow-control font-medium leading-[1.4] transition-colors duration-150 border border-[#70414c]
    disabled:opacity-[0.45] disabled:cursor-not-allowed [&:hover:not(:disabled)]:bg-[#435362]
    [&:hover:not(:disabled)]:border-[#727e8d]
    [&:active:not(:disabled)]:[box-shadow:inset_0_1px_3px_#0004]
  `,
  input: `
    outline-offset-3 min-w-0 w-full min-h-[42px] rounded-[6px] bg-[#1c252e] text-[#eef0f3] border
    border-[#4b5a68] placeholder:text-[#858d99] focus:outline-2 focus:outline-solid focus:outline-accent
    focus:border-transparent [&[type='file']]:p-2 [&::file-selector-button]:px-3
    [&::file-selector-button]:py-2 [&::file-selector-button]:rounded-[4px]
    [&::file-selector-button]:bg-[#435362] [&::file-selector-button]:text-inherit
    [&::file-selector-button]:mr-3 [&::file-selector-button]:cursor-pointer
    [&::file-selector-button]:border-0
  `,
  quietButton: `
    px-4 py-[0.65rem] outline-offset-3 cursor-pointer rounded-[6px] bg-transparent text-muted min-h-[38px]
    shadow-none font-medium leading-[1.4] transition-colors duration-150 border border-[#536271]
    disabled:opacity-[0.45] disabled:cursor-not-allowed [&:hover:not(:disabled)]:bg-[#435362]
    [&:hover:not(:disabled)]:border-[#727e8d]
    [&:active:not(:disabled)]:[box-shadow:inset_0_1px_3px_#0004]
  `,
  primaryButton: `
    px-4 py-[0.65rem] outline-offset-3 cursor-pointer rounded-[6px] bg-brand text-white min-h-[38px]
    shadow-control font-semibold leading-[1.4] transition-colors duration-150 border border-brand-border
    disabled:opacity-[0.45] disabled:cursor-not-allowed hover:enabled:bg-brand-hover
    hover:enabled:border-brand-border [&:active:not(:disabled)]:[box-shadow:inset_0_1px_3px_#0004]
  `,
  details: `
    [&_>_label]:mt-4 [&_>_form]:mt-4 [&_>_fieldset]:mt-4 [&_>_button]:mr-2 [&_>_.row_>_button]:shrink-0
    [&_>_:not(summary)_+_:not(summary)]:mt-4 [&:not([open])_>_summary]:mb-0
  `,
  summary: `
    outline-offset-3 cursor-pointer text-[#d1dce5] mb-[0.8rem] list-none
  `,
};
