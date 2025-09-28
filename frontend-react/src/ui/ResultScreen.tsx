import React, { useEffect } from 'react';
import type { SessionState } from './App';

export const ResultScreen: React.FC<{session:SessionState; onNext:()=>void; onBack:()=>void}> = ({session,onNext,onBack}) => {
  useEffect(() => {
    const handler = (keyEvent: KeyboardEvent) => {
      if (keyEvent.key === 'Enter') {
        keyEvent.preventDefault();
        onNext();
      }else if (keyEvent.key === 'Escape'){
        keyEvent.preventDefault();
        onBack();
      }
    };
    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, [onNext, onBack]);
  return (
    <div style={{display:'flex', flexDirection:'column', alignItems:'center', padding:'32px 16px'}}>
      <h2 style={{fontSize:32, marginBottom:24}}>結果</h2>
      <div style={{display:'flex', flexDirection:'column', alignItems:'center', gap:24}}>
        <div style={{width:420,height:420, background:'#fff', display:'flex',alignItems:'center',justifyContent:'center', border:'2px solid #ccc', borderRadius:16, boxShadow:'0 4px 14px rgba(0,0,0,0.15)'}}>
          <img src={`/api/quiz/artwork/${session.sessionId}?ts=${Date.now()}`} alt={session.answer} style={{maxWidth:'100%', maxHeight:'100%'}} />
        </div>
        <div style={{fontSize:28}}>答え: <strong>{session.answer}</strong></div>
      </div>
      <div style={{display:'flex', gap:20, marginTop:32}}>
        <button style={navBtn} onClick={onBack}>スタート画面 (Esc)</button>
        <button style={primaryBtn} onClick={onNext}>次の問題 (Enter)</button>
      </div>
    </div>
  );
};

const primaryBtn: React.CSSProperties = {
  fontSize:18,
  padding:'12px 22px',
  background:'#2563eb',
  color:'#fff',
  border:'none',
  borderRadius:10,
  cursor:'pointer',
  boxShadow:'0 2px 6px rgba(0,0,0,0.2)'
};
const navBtn: React.CSSProperties = {
  fontSize:16,
  padding:'10px 18px',
  background:'#f3f4f6',
  border:'1px solid #bbb',
  borderRadius:10,
  cursor:'pointer'
};
