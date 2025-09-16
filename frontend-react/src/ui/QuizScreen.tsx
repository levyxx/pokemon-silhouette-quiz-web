import React, { useEffect, useRef, useState } from 'react';
import type { SessionState } from './App';

type Props = { session: SessionState; onSolved:(p:{pokemonId:number; answer:string})=>void; onGiveUp:(p:{pokemonId:number; answer:string})=>void; onAbort:()=>void };

interface GuessResp { correct:boolean; solved:boolean; retryAfter?:number }

export const QuizScreen: React.FC<Props> = ({session,onSolved,onGiveUp,onAbort}) => {
  const [hint, setHint] = useState<{types:string[]; region:string; first:string} | null>(null);
  const [input, setInput] = useState('');
  const [candidates, setCandidates] = useState<string[]>([]);
  const [message, setMessage] = useState('');
  const [loading, setLoading] = useState(false);
  const retryAfterRef = useRef<number>(0);

  useEffect(()=>{
    // fetch minimal hint by starting with silhouette id unknown; server currently not exposing hints endpoint.
    // Placeholder: we can derive first letter etc after first failed guess (simulate) -> simpler approach ask user to use guess flow.
  },[]);

  const submit = async () => {
    if(!input) return;
    if(retryAfterRef.current>0){ setMessage(`あと${retryAfterRef.current}秒待ってください`); return; }
    setLoading(true);
    const res = await fetch('/api/quiz/guess', {method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({sessionId: session.sessionId, answer: input})});
    const data: GuessResp = await res.json();
    setLoading(false);
    if(data.retryAfter){ retryAfterRef.current = data.retryAfter; let iv = setInterval(()=>{ retryAfterRef.current -=1; if(retryAfterRef.current<=0){clearInterval(iv); setMessage('');} else { setMessage(`あと${retryAfterRef.current}秒`);} },1000);}    
    if(data.correct){ setMessage('正解!'); onSolved({pokemonId:0, answer: input}); }
    else if(data.solved){ onSolved({pokemonId:0, answer: input}); }
    else { setMessage('はずれ'); }
  };

  const giveUp = async () => {
    const res = await fetch('/api/quiz/giveup', {method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({sessionId: session.sessionId})});
    const data = await res.json();
    onGiveUp({pokemonId:data.pokemonId, answer:data.name});
  };

  // dynamic candidate search (Japanese or English)
  useEffect(()=>{
    const controller = new AbortController();
    if(!input){ setCandidates([]); return; }
    const q = input.trim();
    if(!q){ setCandidates([]); return; }
    const t = setTimeout(()=>{
      fetch(`/api/quiz/search?prefix=${encodeURIComponent(q)}`, {signal: controller.signal})
        .then(r=> r.ok ? r.json(): [])
        .then((arr:string[])=> setCandidates(arr))
        .catch(()=>{});
    }, 200);
    return ()=>{ controller.abort(); clearTimeout(t); };
  },[input]);

  return (
    <div>
      <h2>シルエットを当てよう</h2>
      <div style={{width:300,height:300, background:'#ddd', display:'flex',alignItems:'center',justifyContent:'center'}}>
  <img src={`/api/quiz/silhouette/session/${session.sessionId}?ts=${Date.now()}`} alt="silhouette" style={{maxWidth:'100%', maxHeight:'100%'}} />
      </div>
      <div style={{marginTop:16}}>
        <input value={input} onChange={e=>setInput(e.target.value)} placeholder="ポケモン名" />
        <button onClick={submit} disabled={loading}>解答</button>
        <button onClick={giveUp}>ギブアップ</button>
        <button onClick={onAbort}>最初に戻る</button>
      </div>
      {candidates.length>0 && (
        <ul style={{border:'1px solid #ccc', maxWidth:200, padding:4, listStyle:'none'}}>
          {candidates.map(c=> <li key={c} style={{cursor:'pointer'}} onClick={()=>setInput(c)}>{c}</li>)}
        </ul>
      )}
      <div style={{marginTop:12}}>{message}</div>
    </div>
  );
};
