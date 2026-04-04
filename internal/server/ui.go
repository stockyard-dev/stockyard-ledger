package server
import "net/http"
func(s *Server)dashboard(w http.ResponseWriter,r *http.Request){w.Header().Set("Content-Type","text/html");w.Write([]byte(dashHTML))}
const dashHTML=`<!DOCTYPE html><html><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>Ledger</title>
<style>
:root{--bg:#1a1410;--bg2:#241e18;--bg3:#2e261e;--rust:#e8753a;--leather:#a0845c;--cream:#f0e6d3;--cd:#bfb5a3;--cm:#7a7060;--gold:#d4a843;--green:#4a9e5c;--red:#c94444;--mono:'JetBrains Mono',monospace;--serif:'Libre Baskerville',serif}
*{margin:0;padding:0;box-sizing:border-box}body{background:var(--bg);color:var(--cream);font-family:var(--mono);line-height:1.5}
.hdr{padding:1rem 1.5rem;border-bottom:1px solid var(--bg3);display:flex;justify-content:space-between;align-items:center}
.hdr h1{font-size:.9rem;letter-spacing:2px}
.main{padding:1.5rem;max-width:900px;margin:0 auto}
.stats{display:grid;grid-template-columns:repeat(3,1fr);gap:.6rem;margin-bottom:1.2rem}
.st{background:var(--bg2);border:1px solid var(--bg3);padding:.8rem;text-align:center}
.st-v{font-size:1.3rem}.st-l{font-size:.55rem;color:var(--cm);text-transform:uppercase;letter-spacing:1px;margin-top:.1rem}
.filters{display:flex;gap:.3rem;margin-bottom:1rem;flex-wrap:wrap}
.fbtn{font-size:.6rem;padding:.2rem .5rem;border:1px solid var(--bg3);background:var(--bg);color:var(--cm);cursor:pointer}
.fbtn:hover{border-color:var(--leather)}.fbtn.active{border-color:var(--rust);color:var(--rust)}
.tx{display:flex;justify-content:space-between;align-items:center;padding:.6rem .8rem;border-bottom:1px solid var(--bg3);font-size:.78rem}
.tx:hover{background:var(--bg2)}
.tx-left{display:flex;gap:.8rem;align-items:center}
.tx-type{font-size:.5rem;padding:.1rem .3rem;text-transform:uppercase;letter-spacing:1px;width:42px;text-align:center}
.tx-debit{background:#c9444422;color:var(--red);border:1px solid #c9444444}
.tx-credit{background:#4a9e5c22;color:var(--green);border:1px solid #4a9e5c44}
.tx-desc{color:var(--cream)}
.tx-meta{font-size:.6rem;color:var(--cm)}
.tx-amount{font-family:var(--mono);font-size:.85rem}
.tx-amount.debit{color:var(--red)}.tx-amount.credit{color:var(--green)}
.tx-cat{font-size:.55rem;padding:.1rem .3rem;background:var(--bg3);color:var(--cm)}
.btn{font-size:.6rem;padding:.25rem .6rem;cursor:pointer;border:1px solid var(--bg3);background:var(--bg);color:var(--cd)}.btn:hover{border-color:var(--leather);color:var(--cream)}
.btn-p{background:var(--rust);border-color:var(--rust);color:var(--bg)}
.modal-bg{display:none;position:fixed;inset:0;background:rgba(0,0,0,.6);z-index:100;align-items:center;justify-content:center}.modal-bg.open{display:flex}
.modal{background:var(--bg2);border:1px solid var(--bg3);padding:1.5rem;width:400px;max-width:90vw}
.modal h2{font-size:.8rem;margin-bottom:1rem;color:var(--rust)}
.fr{margin-bottom:.5rem}.fr label{display:block;font-size:.55rem;color:var(--cm);text-transform:uppercase;letter-spacing:1px;margin-bottom:.15rem}
.fr input,.fr select,.fr textarea{width:100%;padding:.35rem .5rem;background:var(--bg);border:1px solid var(--bg3);color:var(--cream);font-family:var(--mono);font-size:.7rem}
.acts{display:flex;gap:.4rem;justify-content:flex-end;margin-top:.8rem}
.empty{text-align:center;padding:3rem;color:var(--cm);font-style:italic;font-size:.75rem}
</style></head><body>
<div class="hdr"><h1>LEDGER</h1><button class="btn btn-p" onclick="openForm()">+ Transaction</button></div>
<div class="main">
<div class="stats" id="stats"></div>
<div class="filters" id="filters"></div>
<div id="txlist"></div>
</div>
<div class="modal-bg" id="mbg" onclick="if(event.target===this)cm()"><div class="modal" id="mdl"></div></div>
<script>
const A='/api';let txs=[],filter='';
async function load(){
  const[t,s]=await Promise.all([fetch(A+'/transactions').then(r=>r.json()),fetch(A+'/stats').then(r=>r.json())]);
  txs=t.transactions||[];
  const income=txs.filter(t=>t.type==='credit').reduce((s,t)=>s+t.amount,0);
  const expense=txs.filter(t=>t.type==='debit').reduce((s,t)=>s+t.amount,0);
  document.getElementById('stats').innerHTML='<div class="st"><div class="st-v" style="color:var(--green)">$'+fmt(income/100)+'</div><div class="st-l">Income</div></div><div class="st"><div class="st-v" style="color:var(--red)">$'+fmt(expense/100)+'</div><div class="st-l">Expenses</div></div><div class="st"><div class="st-v">$'+fmt((income-expense)/100)+'</div><div class="st-l">Balance</div></div>';
  const cats=[...new Set(txs.map(t=>t.category).filter(c=>c))];
  let fh='<button class="fbtn'+(filter===''?' active':'')+'" onclick="setFilter(\'\')">All</button><button class="fbtn'+(filter==='credit'?' active':'')+'" onclick="setFilter(\'credit\')">Income</button><button class="fbtn'+(filter==='debit'?' active':'')+'" onclick="setFilter(\'debit\')">Expenses</button>';
  cats.forEach(c=>{fh+='<button class="fbtn'+(filter===c?' active':'')+'" onclick="setFilter(\''+c+'\')">'+esc(c)+'</button>';});
  document.getElementById('filters').innerHTML=fh;
  render();
}
function setFilter(f){filter=f;render();}
function render(){
  let filtered=txs;
  if(filter==='credit'||filter==='debit')filtered=txs.filter(t=>t.type===filter);
  else if(filter)filtered=txs.filter(t=>t.category===filter);
  if(!filtered.length){document.getElementById('txlist').innerHTML='<div class="empty">No transactions yet.</div>';return;}
  let h='';
  filtered.forEach(t=>{
    h+='<div class="tx"><div class="tx-left"><span class="tx-type tx-'+t.type+'">'+t.type+'</span><div><div class="tx-desc">'+esc(t.description)+'</div><div class="tx-meta">'+t.date;
    if(t.account)h+=' · '+esc(t.account);
    if(t.category)h+=' · <span class="tx-cat">'+esc(t.category)+'</span>';
    h+='</div></div></div><div style="display:flex;align-items:center;gap:.5rem"><span class="tx-amount '+t.type+'">'+(t.type==='debit'?'-':'+')+' $'+(t.amount/100).toFixed(2)+'</span><button class="btn" onclick="del(\''+t.id+'\')" style="font-size:.5rem;color:var(--cm)">✕</button></div></div>';
  });
  document.getElementById('txlist').innerHTML=h;
}
async function del(id){if(confirm('Delete?')){await fetch(A+'/transactions/'+id,{method:'DELETE'});load();}}
function openForm(){
  document.getElementById('mdl').innerHTML='<h2>New Transaction</h2><div class="fr"><label>Description</label><input id="f-d" placeholder="e.g. Monthly hosting"></div><div class="fr"><label>Amount ($)</label><input id="f-a" type="number" step="0.01" placeholder="49.99"></div><div class="fr"><label>Type</label><select id="f-t"><option value="debit">Expense (debit)</option><option value="credit">Income (credit)</option></select></div><div class="fr"><label>Category</label><input id="f-c" placeholder="e.g. hosting, salary, ads"></div><div class="fr"><label>Account</label><input id="f-ac" placeholder="e.g. checking, savings"></div><div class="fr"><label>Date</label><input id="f-dt" type="date" value="'+new Date().toISOString().split('T')[0]+'"></div><div class="fr"><label>Notes</label><textarea id="f-n" rows="2"></textarea></div><div class="acts"><button class="btn" onclick="cm()">Cancel</button><button class="btn btn-p" onclick="sub()">Add</button></div>';
  document.getElementById('mbg').classList.add('open');
}
async function sub(){const amt=Math.round(parseFloat(document.getElementById('f-a').value)*100);await fetch(A+'/transactions',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({description:document.getElementById('f-d').value,amount:amt,type:document.getElementById('f-t').value,category:document.getElementById('f-c').value,account:document.getElementById('f-ac').value,date:document.getElementById('f-dt').value,notes:document.getElementById('f-n').value})});cm();load();}
function cm(){document.getElementById('mbg').classList.remove('open');}
function fmt(n){return n<0?'-$'+Math.abs(n).toFixed(2):'$'+n.toFixed(2);}
function esc(s){if(!s)return'';const d=document.createElement('div');d.textContent=s;return d.innerHTML;}
load();
</script></body></html>`
