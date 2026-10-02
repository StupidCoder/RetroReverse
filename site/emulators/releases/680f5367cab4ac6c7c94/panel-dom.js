// Local lookup continues working after IDs are namespaced, including adapter
// content inserted after construction. Labels retain their accessible targets.
export const localElement=(root,id)=>root.querySelector('[data-local-id="'+id+'"]')||root.querySelector('#'+id)||(root.id==='render-workspace'?document.getElementById(id):null);
export function namespacePanel(root,prefix){for(const el of root.querySelectorAll('[id]')){el.dataset.localId??=el.id;el.id=prefix+'-'+el.dataset.localId;}for(const label of root.querySelectorAll('[for]')){label.dataset.localFor??=label.htmlFor;label.htmlFor=prefix+'-'+label.dataset.localFor;}}
