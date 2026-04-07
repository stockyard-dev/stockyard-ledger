package server

import "net/http"

func (s *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(dashHTML))
}

const dashHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>Ledger</title>
<link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;700&display=swap" rel="stylesheet">
<style>
:root{--bg:#1a1410;--bg2:#241e18;--bg3:#2e261e;--rust:#e8753a;--leather:#a0845c;--cream:#f0e6d3;--cd:#bfb5a3;--cm:#7a7060;--gold:#d4a843;--green:#4a9e5c;--red:#c94444;--orange:#d4843a;--blue:#5b8dd9;--mono:'JetBrains Mono',monospace}
*{margin:0;padding:0;box-sizing:border-box}
body{background:var(--bg);color:var(--cream);font-family:var(--mono);line-height:1.5;font-size:13px}
.hdr{padding:.8rem 1.5rem;border-bottom:1px solid var(--bg3);display:flex;justify-content:space-between;align-items:center;gap:1rem;flex-wrap:wrap}
.hdr h1{font-size:.9rem;letter-spacing:2px}
.hdr h1 span{color:var(--rust)}
.main{padding:1.2rem 1.5rem;max-width:1000px;margin:0 auto}
.stats{display:grid;grid-template-columns:repeat(4,1fr);gap:.5rem;margin-bottom:1rem}
.st{background:var(--bg2);border:1px solid var(--bg3);padding:.7rem;text-align:center}
.st-v{font-size:1.2rem;font-weight:700;color:var(--gold)}
.st-v.green{color:var(--green)}
.st-v.red{color:var(--red)}
.st-l{font-size:.5rem;color:var(--cm);text-transform:uppercase;letter-spacing:1px;margin-top:.2rem}
.toolbar{display:flex;gap:.5rem;margin-bottom:1rem;flex-wrap:wrap;align-items:center}
.search{flex:1;min-width:180px;padding:.4rem .6rem;background:var(--bg2);border:1px solid var(--bg3);color:var(--cream);font-family:var(--mono);font-size:.7rem}
.search:focus{outline:none;border-color:var(--leather)}
.filter-sel{padding:.4rem .5rem;background:var(--bg2);border:1px solid var(--bg3);color:var(--cream);font-family:var(--mono);font-size:.65rem}
.tx{display:flex;justify-content:space-between;align-items:center;padding:.6rem .9rem;background:var(--bg2);border:1px solid var(--bg3);border-bottom:none;font-size:.72rem;gap:.6rem;cursor:pointer;transition:border-color .15s}
.tx:last-child{border-bottom:1px solid var(--bg3)}
.tx:hover{border-left-color:var(--leather)}
.tx.void{opacity:.5;text-decoration:line-through}
.tx-left{display:flex;gap:.7rem;align-items:center;flex:1;min-width:0}
.tx-type{font-size:.5rem;padding:.15rem .35rem;text-transform:uppercase;letter-spacing:1px;width:46px;text-align:center;font-weight:700;flex-shrink:0}
.tx-debit{background:#c9444422;color:var(--red);border:1px solid #c9444466}
.tx-credit{background:#4a9e5c22;color:var(--green);border:1px solid #4a9e5c66}
.tx-body{min-width:0;flex:1}
.tx-desc{color:var(--cream);font-weight:500;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.tx-meta{font-size:.55rem;color:var(--cm);display:flex;gap:.5rem;flex-wrap:wrap;margin-top:.15rem;align-items:center}
.tx-cat{font-size:.5rem;padding:.05rem .3rem;background:var(--bg3);color:var(--cd);font-weight:700}
.tx-status{font-size:.5rem;padding:.05rem .3rem;border:1px solid var(--bg3);color:var(--cm)}
.tx-status.pending{border-color:var(--orange);color:var(--orange)}
.tx-status.reconciled{border-color:var(--green);color:var(--green)}
.tx-status.void{border-color:var(--cm);color:var(--cm)}
.tx-extra{font-size:.55rem;color:var(--cd);margin-top:.3rem;display:flex;gap:.5rem;flex-wrap:wrap}
.tx-extra-pair{display:flex;gap:.2rem}
.tx-extra-label{color:var(--cm);text-transform:uppercase;letter-spacing:.5px}
.tx-extra-val{color:var(--cream)}
.tx-amount{font-family:var(--mono);font-size:.85rem;font-weight:700;flex-shrink:0;text-align:right}
.tx-amount.debit{color:var(--red)}
.tx-amount.credit{color:var(--green)}
.btn{font-family:var(--mono);font-size:.6rem;padding:.25rem .5rem;cursor:pointer;border:1px solid var(--bg3);background:var(--bg);color:var(--cd);transition:.15s}
.btn:hover{border-color:var(--leather);color:var(--cream)}
.btn-p{background:var(--rust);border-color:var(--rust);color:#fff}
.btn-p:hover{opacity:.85;color:#fff}
.btn-sm{font-size:.55rem;padding:.2rem .4rem}
.modal-bg{display:none;position:fixed;inset:0;background:rgba(0,0,0,.65);z-index:100;align-items:center;justify-content:center}
.modal-bg.open{display:flex}
.modal{background:var(--bg2);border:1px solid var(--bg3);padding:1.5rem;width:480px;max-width:92vw;max-height:90vh;overflow-y:auto}
.modal h2{font-size:.8rem;margin-bottom:1rem;color:var(--rust);letter-spacing:1px}
.fr{margin-bottom:.6rem}
.fr label{display:block;font-size:.55rem;color:var(--cm);text-transform:uppercase;letter-spacing:1px;margin-bottom:.2rem}
.fr input,.fr select,.fr textarea{width:100%;padding:.4rem .5rem;background:var(--bg);border:1px solid var(--bg3);color:var(--cream);font-family:var(--mono);font-size:.7rem}
.fr input:focus,.fr select:focus,.fr textarea:focus{outline:none;border-color:var(--leather)}
.row2{display:grid;grid-template-columns:1fr 1fr;gap:.5rem}
.fr-section{margin-top:1rem;padding-top:.8rem;border-top:1px solid var(--bg3)}
.fr-section-label{font-size:.55rem;color:var(--rust);text-transform:uppercase;letter-spacing:1px;margin-bottom:.5rem}
.acts{display:flex;gap:.4rem;justify-content:flex-end;margin-top:1rem}
.acts .btn-del{margin-right:auto;color:var(--red);border-color:#3a1a1a}
.acts .btn-del:hover{border-color:var(--red);color:var(--red)}
.empty{text-align:center;padding:3rem;color:var(--cm);font-style:italic;font-size:.85rem}
.count-label{font-size:.6rem;color:var(--cm);margin-bottom:.5rem}
@media(max-width:600px){.stats{grid-template-columns:repeat(2,1fr)}}
</style>
</head>
<body>

<div class="hdr">
<h1 id="dash-title"><span>&#9670;</span> LEDGER</h1>
<button class="btn btn-p" onclick="openNew()">+ Transaction</button>
</div>

<div class="main">
<div class="stats" id="stats"></div>
<div class="toolbar">
<input class="search" id="search" placeholder="Search description, notes, account..." oninput="debouncedRender()">
<select class="filter-sel" id="type-filter" onchange="render()">
<option value="">All Types</option>
<option value="credit">Credits (Income)</option>
<option value="debit">Debits (Expenses)</option>
</select>
<select class="filter-sel" id="status-filter" onchange="render()">
<option value="">All Statuses</option>
<option value="posted">Posted</option>
<option value="pending">Pending</option>
<option value="reconciled">Reconciled</option>
<option value="void">Void</option>
</select>
<select class="filter-sel" id="category-filter" onchange="render()">
<option value="">All Categories</option>
</select>
<select class="filter-sel" id="account-filter" onchange="render()">
<option value="">All Accounts</option>
</select>
</div>
<div class="count-label" id="count"></div>
<div id="list"></div>
</div>

<div class="modal-bg" id="mbg" onclick="if(event.target===this)closeModal()">
<div class="modal" id="mdl"></div>
</div>

<script>
var A='/api';
var RESOURCE='transactions';

var fields=[
{name:'description',label:'Description',type:'text',required:true},
{name:'amount',label:'Amount ($)',type:'money',required:true},
{name:'type',label:'Type',type:'select',options:['debit','credit'],required:true},
{name:'category',label:'Category',type:'select_or_text',options:[]},
{name:'account',label:'Account',type:'select_or_text',options:[]},
{name:'date',label:'Date',type:'date'},
{name:'status',label:'Status',type:'select',options:['posted','pending','reconciled','void']},
{name:'notes',label:'Notes',type:'textarea'}
];

var transactions=[],txExtras={},editId=null,searchTimer=null;

// ─── Money helpers ────────────────────────────────────────────────

function fmtMoney(cents){
if(!cents&&cents!==0)return'$0';
var dollars=cents/100;
var negative=dollars<0;
var abs=Math.abs(dollars);
var str;
if(abs>=1000000)str='$'+(abs/1000000).toFixed(1)+'M';
else if(abs>=1000)str='$'+(abs/1000).toFixed(1)+'k';
else str='$'+abs.toFixed(0);
return negative?'-'+str:str;
}

function fmtMoneyFull(cents){
if(!cents&&cents!==0)return'$0.00';
var negative=cents<0;
var abs=Math.abs(cents)/100;
var str='$'+abs.toLocaleString('en-US',{minimumFractionDigits:2,maximumFractionDigits:2});
return negative?'-'+str:str;
}

function parseMoney(str){
if(!str)return 0;
var n=parseFloat(String(str).replace(/[^\d.-]/g,''));
if(isNaN(n))return 0;
return Math.round(n*100);
}

function fmtDate(s){
if(!s)return'';
try{
var d=new Date(s+'T12:00:00');
if(isNaN(d.getTime()))return s;
return d.toLocaleDateString('en-US',{month:'short',day:'numeric',year:'2-digit'});
}catch(e){return s}
}

function fieldByName(n){
for(var i=0;i<fields.length;i++)if(fields[i].name===n)return fields[i];
return null;
}

function debouncedRender(){
clearTimeout(searchTimer);
searchTimer=setTimeout(render,200);
}

// ─── Loading ──────────────────────────────────────────────────────

async function load(){
try{
var resps=await Promise.all([
fetch(A+'/transactions').then(function(r){return r.json()}),
fetch(A+'/stats').then(function(r){return r.json()})
]);
transactions=resps[0].transactions||[];
renderStats(resps[1]||{});

try{
var ex=await fetch(A+'/extras/'+RESOURCE).then(function(r){return r.json()});
txExtras=ex||{};
transactions.forEach(function(t){
var x=txExtras[t.id];
if(!x)return;
Object.keys(x).forEach(function(k){if(t[k]===undefined)t[k]=x[k]});
});
}catch(e){txExtras={}}

populateFilters();
}catch(e){
console.error('load failed',e);
transactions=[];
}
render();
}

function populateFilters(){
// Categories
var catSel=document.getElementById('category-filter');
if(catSel){
var current=catSel.value;
var seen={};
var cats=[];
transactions.forEach(function(t){
if(t.category&&!seen[t.category]){seen[t.category]=true;cats.push(t.category)}
});
cats.sort();
catSel.innerHTML='<option value="">All Categories</option>'+cats.map(function(c){return'<option value="'+esc(c)+'"'+(c===current?' selected':'')+'>'+esc(c)+'</option>'}).join('');
}
// Accounts
var acctSel=document.getElementById('account-filter');
if(acctSel){
var currentA=acctSel.value;
var seenA={};
var accts=[];
transactions.forEach(function(t){
if(t.account&&!seenA[t.account]){seenA[t.account]=true;accts.push(t.account)}
});
accts.sort();
acctSel.innerHTML='<option value="">All Accounts</option>'+accts.map(function(a){return'<option value="'+esc(a)+'"'+(a===currentA?' selected':'')+'>'+esc(a)+'</option>'}).join('');
}
}

function renderStats(s){
var total=s.total||0;
var credits=s.total_credits||0;
var debits=s.total_debits||0;
var balance=s.balance||0;
var balClass=balance>=0?'green':'red';
document.getElementById('stats').innerHTML=
'<div class="st"><div class="st-v green">'+fmtMoney(credits)+'</div><div class="st-l">Income</div></div>'+
'<div class="st"><div class="st-v red">'+fmtMoney(debits)+'</div><div class="st-l">Expenses</div></div>'+
'<div class="st"><div class="st-v '+balClass+'">'+fmtMoney(balance)+'</div><div class="st-l">Balance</div></div>'+
'<div class="st"><div class="st-v">'+total+'</div><div class="st-l">Transactions</div></div>';
}

function render(){
var q=(document.getElementById('search').value||'').toLowerCase();
var tf=document.getElementById('type-filter').value;
var sf=document.getElementById('status-filter').value;
var cf=document.getElementById('category-filter').value;
var af=document.getElementById('account-filter').value;

var f=transactions;
if(q)f=f.filter(function(t){
return(t.description||'').toLowerCase().includes(q)||
(t.notes||'').toLowerCase().includes(q)||
(t.account||'').toLowerCase().includes(q);
});
if(tf)f=f.filter(function(t){return t.type===tf});
if(sf)f=f.filter(function(t){return t.status===sf});
if(cf)f=f.filter(function(t){return t.category===cf});
if(af)f=f.filter(function(t){return t.account===af});

document.getElementById('count').textContent=f.length+' transaction'+(f.length!==1?'s':'')+(transactions.length!==f.length?' (of '+transactions.length+')':'');

if(!f.length){
var msg=window._emptyMsg||'No transactions yet. Add your first one.';
document.getElementById('list').innerHTML='<div class="empty">'+esc(msg)+'</div>';
return;
}

var h='';
f.forEach(function(t){h+=txHTML(t)});
document.getElementById('list').innerHTML=h;
}

function txHTML(t){
var sign=t.type==='debit'?'-':'+';
var cls='tx '+t.type;
if(t.status==='void')cls+=' void';

var h='<div class="'+cls+'" onclick="openEdit(\''+esc(t.id)+'\')">';
h+='<div class="tx-left">';
h+='<span class="tx-type tx-'+esc(t.type)+'">'+esc(t.type)+'</span>';
h+='<div class="tx-body">';
h+='<div class="tx-desc">'+esc(t.description)+'</div>';
h+='<div class="tx-meta">';
if(t.date)h+='<span>'+esc(fmtDate(t.date))+'</span>';
if(t.account)h+='<span>'+esc(t.account)+'</span>';
if(t.category)h+='<span class="tx-cat">'+esc(t.category)+'</span>';
if(t.status&&t.status!=='posted')h+='<span class="tx-status '+esc(t.status)+'">'+esc(t.status)+'</span>';
h+='</div>';

// Custom field display
var customRows='';
fields.forEach(function(f){
if(!f.isCustom)return;
var v=t[f.name];
if(v===undefined||v===null||v==='')return;
customRows+='<span class="tx-extra-pair"><span class="tx-extra-label">'+esc(f.label)+':</span> <span class="tx-extra-val">'+esc(String(v))+'</span></span>';
});
if(customRows)h+='<div class="tx-extra">'+customRows+'</div>';

h+='</div></div>';
h+='<span class="tx-amount '+esc(t.type)+'">'+sign+' '+fmtMoneyFull(t.amount)+'</span>';
h+='</div>';
return h;
}

// ─── Modal ────────────────────────────────────────────────────────

function fieldHTML(f,value){
var v=value;
if(v===undefined||v===null)v='';
var req=f.required?' *':'';
var ph=f.placeholder?(' placeholder="'+esc(f.placeholder)+'"'):'';
var h='<div class="fr"><label>'+esc(f.label)+req+'</label>';

if(f.type==='select'){
h+='<select id="f-'+f.name+'">';
if(!f.required)h+='<option value="">Select...</option>';
(f.options||[]).forEach(function(o){
var sel=(String(v)===String(o))?' selected':'';
var disp=String(o).charAt(0).toUpperCase()+String(o).slice(1).replace(/_/g,' ');
h+='<option value="'+esc(String(o))+'"'+sel+'>'+esc(disp)+'</option>';
});
h+='</select>';
}else if(f.type==='select_or_text'){
h+='<input list="dl-'+f.name+'" type="text" id="f-'+f.name+'" value="'+esc(String(v))+'"'+ph+'>';
h+='<datalist id="dl-'+f.name+'">';
var opts=(f.options||[]).slice();
// Pull existing values from current data
transactions.forEach(function(td){
var existing=td[f.name];
if(existing&&opts.indexOf(existing)===-1)opts.push(existing);
});
opts.forEach(function(o){h+='<option value="'+esc(String(o))+'">'});
h+='</datalist>';
}else if(f.type==='textarea'){
h+='<textarea id="f-'+f.name+'" rows="2"'+ph+'>'+esc(String(v))+'</textarea>';
}else if(f.type==='money'){
var dollars=v?(v/100).toFixed(2):'';
h+='<input type="text" id="f-'+f.name+'" value="'+esc(dollars)+'" placeholder="0.00">';
}else if(f.type==='number'||f.type==='integer'){
h+='<input type="number" id="f-'+f.name+'" value="'+esc(String(v))+'"'+ph+'>';
}else{
var inputType=f.type||'text';
h+='<input type="'+esc(inputType)+'" id="f-'+f.name+'" value="'+esc(String(v))+'"'+ph+'>';
}
h+='</div>';
return h;
}

function formHTML(tx){
var t=tx||{};
var isEdit=!!tx;
var h='<h2>'+(isEdit?'EDIT TRANSACTION':'NEW TRANSACTION')+'</h2>';

h+=fieldHTML(fieldByName('description'),t.description);
h+='<div class="row2">'+fieldHTML(fieldByName('amount'),t.amount)+fieldHTML(fieldByName('type'),t.type||'debit')+'</div>';
h+='<div class="row2">'+fieldHTML(fieldByName('category'),t.category)+fieldHTML(fieldByName('account'),t.account)+'</div>';
h+='<div class="row2">'+fieldHTML(fieldByName('date'),t.date||new Date().toISOString().split('T')[0])+fieldHTML(fieldByName('status'),t.status||'posted')+'</div>';
h+=fieldHTML(fieldByName('notes'),t.notes);

var customFields=fields.filter(function(f){return f.isCustom});
if(customFields.length){
var label=window._customSectionLabel||'Additional Details';
h+='<div class="fr-section"><div class="fr-section-label">'+esc(label)+'</div>';
customFields.forEach(function(f){h+=fieldHTML(f,t[f.name])});
h+='</div>';
}

h+='<div class="acts">';
if(isEdit){
h+='<button class="btn btn-del" onclick="delTx()">Delete</button>';
}
h+='<button class="btn" onclick="closeModal()">Cancel</button>';
h+='<button class="btn btn-p" onclick="submit()">'+(isEdit?'Save':'Add')+'</button>';
h+='</div>';
return h;
}

function openNew(){
editId=null;
document.getElementById('mdl').innerHTML=formHTML();
document.getElementById('mbg').classList.add('open');
var d=document.getElementById('f-description');
if(d)d.focus();
}

function openEdit(id){
var t=null;
for(var i=0;i<transactions.length;i++){if(transactions[i].id===id){t=transactions[i];break}}
if(!t)return;
editId=id;
document.getElementById('mdl').innerHTML=formHTML(t);
document.getElementById('mbg').classList.add('open');
}

function closeModal(){
document.getElementById('mbg').classList.remove('open');
editId=null;
}

async function submit(){
var descEl=document.getElementById('f-description');
if(!descEl||!descEl.value.trim()){alert('Description is required');return}
var amtEl=document.getElementById('f-amount');
if(!amtEl||!amtEl.value.trim()){alert('Amount is required');return}

var body={};
var extras={};
fields.forEach(function(f){
var el=document.getElementById('f-'+f.name);
if(!el)return;
var val;
if(f.type==='money')val=parseMoney(el.value);
else if(f.type==='number'||f.type==='integer')val=parseFloat(el.value)||0;
else val=el.value.trim();
if(f.isCustom)extras[f.name]=val;
else body[f.name]=val;
});

var savedId=editId;
try{
if(editId){
var r1=await fetch(A+'/transactions/'+editId,{method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)});
if(!r1.ok){var e1=await r1.json().catch(function(){return{}});alert(e1.error||'Save failed');return}
}else{
var r2=await fetch(A+'/transactions',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)});
if(!r2.ok){var e2=await r2.json().catch(function(){return{}});alert(e2.error||'Add failed');return}
var created=await r2.json();
savedId=created.id;
}
if(savedId&&Object.keys(extras).length){
await fetch(A+'/extras/'+RESOURCE+'/'+savedId,{method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify(extras)}).catch(function(){});
}
}catch(e){
alert('Network error: '+e.message);
return;
}
closeModal();
load();
}

async function delTx(){
if(!editId)return;
if(!confirm('Delete this transaction?'))return;
await fetch(A+'/transactions/'+editId,{method:'DELETE'});
closeModal();
load();
}

function esc(s){
if(s===undefined||s===null)return'';
var d=document.createElement('div');
d.textContent=String(s);
return d.innerHTML;
}

document.addEventListener('keydown',function(e){if(e.key==='Escape')closeModal()});

// ─── Personalization ──────────────────────────────────────────────

(function loadPersonalization(){
fetch('/api/config').then(function(r){return r.json()}).then(function(cfg){
if(!cfg||typeof cfg!=='object')return;

if(cfg.dashboard_title){
var h1=document.getElementById('dash-title');
if(h1)h1.innerHTML='<span>&#9670;</span> '+esc(cfg.dashboard_title);
document.title=cfg.dashboard_title;
}

if(cfg.empty_state_message)window._emptyMsg=cfg.empty_state_message;
if(cfg.primary_label)window._customSectionLabel=cfg.primary_label+' Details';

if(Array.isArray(cfg.categories)){
var catField=fieldByName('category');
if(catField)catField.options=cfg.categories;
}
if(Array.isArray(cfg.accounts)){
var acctField=fieldByName('account');
if(acctField)acctField.options=cfg.accounts;
}

if(Array.isArray(cfg.custom_fields)){
cfg.custom_fields.forEach(function(cf){
if(!cf||!cf.name||!cf.label)return;
if(fieldByName(cf.name))return;
fields.push({
name:cf.name,
label:cf.label,
type:cf.type||'text',
options:cf.options||[],
isCustom:true
});
});
}
}).catch(function(){
}).finally(function(){
load();
});
})();
</script>
</body>
</html>`
