import React, { useEffect, useState } from 'react';

const regions = [
  {key:'kanto', label:'カントー'}, {key:'johto', label:'ジョウト'}, {key:'hoenn', label:'ホウエン'},
  {key:'sinnoh', label:'シンオウ'}, {key:'unova', label:'イッシュ'}, {key:'kalos', label:'カロス'},
  {key:'alola', label:'アローラ'}, {key:'galar', label:'ガラル'}, {key:'paldea', label:'パルデア'}
];

type Config = { regions:string[]; allowMega:boolean; allowPrimal:boolean };
export const StartScreen: React.FC<{onStarted:(sessionId:string, config:Config)=>void; initialConfig?:Config | null}> = ({onStarted, initialConfig}) => {
  const [selected, setSelected] = useState<string[]>(initialConfig?.regions ?? regions.map(r=>r.key));
  const [allowMega, setAllowMega] = useState(initialConfig?.allowMega ?? false);
  const [allowPrimal, setAllowPrimal] = useState(initialConfig?.allowPrimal ?? false);
  useEffect(()=>{
    if(initialConfig){
      setSelected(initialConfig.regions);
      setAllowMega(initialConfig.allowMega);
      setAllowPrimal(initialConfig.allowPrimal);
    }
  },[initialConfig]);
  const toggle = (k:string)=> setSelected(s=> s.includes(k)? s.filter(x=>x!==k): [...s,k]);

  const start = async () => {
    const res = await fetch('/api/quiz/start', {method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({regions:selected, allowMega:allowMega, allowPrimal:allowPrimal})});
    const data = await res.json();
    onStarted(data.sessionId, {regions:selected, allowMega, allowPrimal});
  };

  return (
    <div style={{display:'flex', flexDirection:'column', alignItems:'center', padding:'40px 16px'}}>
      <h1 style={{fontSize:40, margin:'0 0 12px'}}>ポケモン シルエットクイズ</h1>
      <p style={{margin:'0 0 32px', fontSize:16, color:'#555'}}>地方を選んでスタート (未選択なら全て対象)</p>
      <div style={{background:'#fff', border:'1px solid #ddd', borderRadius:16, padding:32, width:'100%', maxWidth:880, boxShadow:'0 4px 14px rgba(0,0,0,0.12)'}}>
        <section style={{marginBottom:32}}>
          <h2 style={{fontSize:24, margin:'0 0 16px'}}>出題範囲 (地方)</h2>
          <div style={{display:'flex', flexWrap:'wrap', gap:12}}>
            {regions.map(r=>{
              const active = selected.includes(r.key);
              return (
                <button key={r.key} onClick={()=>toggle(r.key)}
                  style={{
                    padding:'10px 16px',
                    border:'1px solid '+(active?'#1d4ed8':'#999'),
                    background: active? '#2563eb':'#f3f4f6',
                    color: active? '#fff':'#222',
                    borderRadius:10,
                    cursor:'pointer',
                    fontSize:16,
                    boxShadow: active? '0 2px 6px rgba(0,0,0,0.25)': '0 1px 3px rgba(0,0,0,0.12)'
                  }}>{r.label}</button>
              );
            })}
          </div>
        </section>
        <section style={{marginBottom:36}}>
          <h2 style={{fontSize:24, margin:'0 0 16px'}}>特別形態</h2>
          <div style={{display:'flex', gap:32, flexWrap:'wrap'}}>
            <label style={checkLabelStyle}>
              <input type="checkbox" checked={allowMega} onChange={e=>setAllowMega(e.target.checked)} />
              <span>メガシンカ</span>
            </label>
            <label style={checkLabelStyle}>
              <input type="checkbox" checked={allowPrimal} onChange={e=>setAllowPrimal(e.target.checked)} />
              <span>ゲンシカイキ</span>
            </label>
          </div>
        </section>
        <div style={{textAlign:'center'}}>
          <button onClick={start} style={startBtnStyle}>スタート</button>
        </div>
      </div>
    </div>
  );
};

const startBtnStyle: React.CSSProperties = {
  fontSize:22,
  padding:'14px 40px',
  background:'#2563eb',
  color:'#fff',
  border:'none',
  borderRadius:12,
  cursor:'pointer',
  boxShadow:'0 3px 10px rgba(0,0,0,0.25)'
};

const checkLabelStyle: React.CSSProperties = {
  display:'flex',
  alignItems:'center',
  gap:8,
  fontSize:18
};
