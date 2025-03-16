/** 
 * @typedef {Object} Palette
 * @property {() => string} bgPrimary
 * @property {() => string} bgSecondary
 * @property {() => string} fgPrimary
 * @property {() => string} fgSecondary
 */

/** @type {Palette} */
export let Palette = $state(NewStarshipPalette());

export function NewStarshipPalette() {
  return {
    bgPrimary: () => { return "#000000" },
    bgSecondary: () => { return "#000000" },
    fgPrimary: () => { return "#1B1A55" },
    fgSecondary: () => { return "#535C91" },
  }
}