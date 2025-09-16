import React, { useState } from 'react';

const regions = [
  {key:'kanto', label:'カントー'}, {key:'johto', label:'ジョウト'}, {key:'hoenn', label:'ホウエン'},
  {key:'sinnoh', label:'シンオウ'}, {key:'unova', label:'イッシュ'}, {key:'kalos', label:'カロス'},
  {key:'alola', label:'アローラ'}, {key:'galar', label:'ガラル'}, {key:'paldea', label:'パルデア'}
];

export const StartScreen: React.FC<{onStarted:(sessionId:string)=>void}> = ({onStarted}) => {
  const [selected, setSelected] = useState<string[]>(regions.map(r=>r.key));
  const [allowMega, setAllowMega] = useState(false);
  const [allowPrimal, setAllowPrimal] = useState(false);
  const toggle = (k:string)=> setSelected(s=> s.includes(k)? s.filter(x=>x!==k): [...s,k]);

  const start = async () => {
    const res = await fetch('/api/quiz/start', {method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({regions:selected, allowMega:allowMega, allowPrimal:allowPrimal})});
    const data = await res.json();
    onStarted(data.sessionId);
  };

  return (
    <div>
      <h1>ポケモン シルエットクイズ</h1>
      <h2>出題範囲 (地方)</h2>
      <div style={{display:'flex', flexWrap:'wrap', gap:8}}>
        {regions.map(r=>{
          const active = selected.includes(r.key);
          return <button key={r.key} onClick={()=>toggle(r.key)} style={{padding:'8px 12px', border:'1px solid #444', background: active? '#1976d2':'#eee', color: active? '#fff':'#222', cursor:'pointer'}}>{r.label}</button>;
        })}
      </div>
      <h2 style={{marginTop:24}}>特別形態</h2>
      <div style={{display:'flex', gap:12}}>
        <label><input type="checkbox" checked={allowMega} onChange={e=>setAllowMega(e.target.checked)} /> メガシンカ</label>
        <label><input type="checkbox" checked={allowPrimal} onChange={e=>setAllowPrimal(e.target.checked)} /> ゲンシカイキ</label>
      </div>
      <div style={{marginTop:32}}>
        <button onClick={start} style={{fontSize:18, padding:'10px 20px'}}>スタート</button>
      </div>
    </div>
  );
};
