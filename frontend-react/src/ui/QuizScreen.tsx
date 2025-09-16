import React, { useEffect, useRef, useState } from 'react';
import type { SessionState } from './App';

type Props = { session: SessionState; onSolved:(p:{pokemonId:number; answer:string})=>void; onGiveUp:(p:{pokemonId:number; answer:string})=>void; onAbort:()=>void };

interface GuessResp { correct:boolean; solved:boolean; retryAfter?:number }

export const QuizScreen: React.FC<Props> = ({session,onSolved,onGiveUp,onAbort}) => {
  const [hint, setHint] = useState<{types:string[]; region:string; firstLetter:string} | null>(null);
  const [showType, setShowType] = useState(false);
  const [showRegion, setShowRegion] = useState(false);
  const [showFirst, setShowFirst] = useState(false);
  const [input, setInput] = useState('');
  const inputRef = useRef<HTMLInputElement | null>(null);
  const [candidates, setCandidates] = useState<string[]>([]);
  const [message, setMessage] = useState('');
  const [loading, setLoading] = useState(false);
  const retryAfterRef = useRef<number>(0);

  const ensureHint = async () => {
    if (hint) return;
    const res = await fetch(`/api/quiz/hint/${session.sessionId}`);
    if(res.ok){ setHint(await res.json()); }
  };
  const revealType = async () => { await ensureHint(); setShowType(true); };
  const revealRegion = async () => { await ensureHint(); setShowRegion(true); };
  const revealFirst = async () => { await ensureHint(); setShowFirst(true); };

  useEffect(()=>{
    // 自動フォーカス（遷移直後）
    inputRef.current?.focus();
  },[]);

  // Enter キーで解答
  useEffect(()=>{
    const handler = (e:KeyboardEvent) => {
      if(e.key === 'Enter') {
        e.preventDefault();
        submit();
      }
    };
    window.addEventListener('keydown', handler);
    return ()=> window.removeEventListener('keydown', handler);
  },[input, session]);

  const submit = async () => {
    if(!input) return;
    if(retryAfterRef.current>0){
      setMessage('回答は5秒空けてください');
      return;
    }
    setLoading(true);
    const res = await fetch('/api/quiz/guess', {method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({sessionId: session.sessionId, answer: input})});
    const data: GuessResp = await res.json();
    setLoading(false);
    if(data.retryAfter){ retryAfterRef.current = data.retryAfter; setMessage('回答は5秒空けてください'); }
    else if(data.correct){ setMessage('正解!'); onSolved({pokemonId:0, answer: input}); }
    else if(data.solved){ onSolved({pokemonId:0, answer: input}); }
    else { setMessage('はずれ'); }
    // 5秒後にクールダウン解除 (バックエンドと同じ間隔) -> 自動でメッセージクリアはしない
    if(data.retryAfter){ setTimeout(()=>{ retryAfterRef.current = 0; }, data.retryAfter * 1000); }
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
    <div style={{display:'flex', flexDirection:'column', alignItems:'center', padding:'24px 16px'}}>
      <h2 style={{fontSize:32, marginBottom:24}}>シルエットを当てよう</h2>
      <div style={{display:'flex', gap:32, alignItems:'flex-start', justifyContent:'center', width:'100%', maxWidth:1000}}>
        <div style={{flex:'0 0 auto', width:400, height:400, background:'#eee', border:'2px solid #ccc', borderRadius:12, display:'flex',alignItems:'center',justifyContent:'center', boxShadow:'0 4px 12px rgba(0,0,0,0.15)'}}>
          <img src={`/api/quiz/silhouette/session/${session.sessionId}?ts=${Date.now()}`} alt="silhouette" style={{maxWidth:'100%', maxHeight:'100%'}} />
        </div>
        <div style={{flex:'1 1 auto', minWidth:260}}>
          <div style={{display:'flex', gap:12, marginBottom:16}}>
            <input ref={inputRef} value={input} onChange={e=>setInput(e.target.value)} placeholder="ポケモン名" style={{flex:1, fontSize:20, padding:'10px 14px', border:'2px solid #888', borderRadius:8}} />
            <button onClick={submit} disabled={loading} style={btnStyle}>解答</button>
          </div>
          <div style={{display:'flex', gap:12, marginBottom:8}}>
            <button onClick={giveUp} style={secBtnStyle}>ギブアップ</button>
            <button onClick={onAbort} style={secBtnStyle}>最初に戻る</button>
          </div>
          <div style={{display:'flex', gap:12, marginBottom:12}}>
            <button onClick={revealType} disabled={showType} style={secBtnStyle}>タイプ</button>
            <button onClick={revealRegion} disabled={showRegion} style={secBtnStyle}>地方</button>
            <button onClick={revealFirst} disabled={showFirst} style={secBtnStyle}>最初の文字</button>
          </div>
          {candidates.length>0 && (
            <ul style={{border:'1px solid #ccc', maxWidth:300, padding:8, listStyle:'none', margin:0, background:'#fff', borderRadius:8, boxShadow:'0 2px 6px rgba(0,0,0,0.15)'}}>
              {candidates.map(c=> <li key={c} style={{cursor:'pointer', padding:'4px 6px', borderRadius:4}} onClick={()=>setInput(c)}>{c}</li>)}
            </ul>
          )}
          <div style={{marginTop:16, fontSize:16, color: message==='回答は5秒空けてください' ? 'red':'#222'}}>{message}</div>
        </div>
        {(showType || showRegion || showFirst) && (
          <div style={{flex:'0 0 220px', background:'#fafafa', border:'1px solid #ddd', borderRadius:12, padding:16, boxShadow:'0 2px 8px rgba(0,0,0,0.1)'}}>
            <h3 style={{marginTop:0, fontSize:20}}>ヒント</h3>
            {!hint ? <p>取得中...</p> : (
              <ul style={{paddingLeft:18, margin:0, fontSize:16}}>
                {showType && <li>タイプ: {hint.types.join(', ')}</li>}
                {showRegion && <li>地方: {hint.region}</li>}
                {showFirst && <li>最初の文字: {hint.firstLetter}</li>}
              </ul>
            )}
          </div>
        )}
      </div>
    </div>
  );
};

const btnStyle: React.CSSProperties = {
  fontSize:18,
  padding:'10px 18px',
  background:'#2563eb',
  color:'#fff',
  border:'none',
  borderRadius:8,
  cursor:'pointer',
  boxShadow:'0 2px 6px rgba(0,0,0,0.2)'
};
const secBtnStyle: React.CSSProperties = {
  fontSize:16,
  padding:'8px 14px',
  background:'#f3f4f6',
  border:'1px solid #bbb',
  borderRadius:8,
  cursor:'pointer'
};
