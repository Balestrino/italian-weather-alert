'use strict';
(() => {
 const navigation = document.querySelector('.navigation');
 const menu = navigation?.querySelector('summary');
 const narrow = matchMedia('(max-width: 700px)');
 if (navigation && narrow.matches) navigation.open = false;
 narrow.addEventListener('change', event => {if (navigation) {if(event.matches && navigation.contains(document.activeElement)) menu.focus();navigation.open = !event.matches;}});
 const input = document.querySelector('#page-search');
 const search = document.querySelector('[data-page-search]');
 const rows = [...document.querySelectorAll('[data-record]')];
 const output = document.querySelector('#search-status');
 const shortcuts = document.querySelector('#shortcuts-enabled');
 const help = document.querySelector('[data-shortcut-help]');
 if (search && rows.length) search.hidden = false;
 if (help) help.hidden = false;
 if (shortcuts) {
  try { shortcuts.checked = localStorage.getItem('iwa-shortcuts') !== 'off'; } catch (_) { /* Storage is optional. */ }
  shortcuts.addEventListener('change', () => {try {localStorage.setItem('iwa-shortcuts', shortcuts.checked ? 'on' : 'off');} catch (_) {}});
 }
 function filter() {
  const needle = input.value.trim().toLocaleLowerCase('it');
  let visible = 0;
  for (const row of rows) {row.hidden = !row.textContent.toLocaleLowerCase('it').includes(needle);if (!row.hidden) visible++;}
  if (output) output.textContent = `${visible} di ${rows.length} record in questa pagina`;
 }
 input?.addEventListener('input', filter);
 if (input && rows.length) filter();
 document.addEventListener('keydown', event => {
  if (event.altKey || event.ctrlKey || event.metaKey || event.shiftKey || event.isComposing) return;
  const editing = event.target instanceof Element && !!event.target.closest('input, textarea, select, [contenteditable]:not([contenteditable="false"])');
  if (event.key === 'Escape') {
   if (event.target === input && input.value) {input.value = '';filter();event.preventDefault();return;}
   if (!editing && navigation?.open && matchMedia('(max-width: 700px)').matches) {navigation.open = false;menu.focus();event.preventDefault();}
   return;
  }
  if (!editing && event.key === '/' && shortcuts?.checked && search && !search.hidden) {event.preventDefault();input.focus();}
 });
})();
