/**
 * @fileoverview Shared keyboard event target guards for keep-alive pages.
 *
 * Pages mounted inside PageDeck remain alive in the DOM while hidden, so global
 * window keyboard listeners must ignore events while inactive or when focus is
 * inside editable form fields.
 */

/**
 * Checks whether a keyboard event originated from an editable text control.
 *
 * @param target - Event target from a window or document keyboard listener.
 * @returns True if typing should pass through untouched.
 */
export function isEditableEventTarget(target: EventTarget | null): boolean {
  if (!(target instanceof HTMLElement)) return false
  const tag = target.tagName
  if (tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT') return true
  return target.isContentEditable
}
