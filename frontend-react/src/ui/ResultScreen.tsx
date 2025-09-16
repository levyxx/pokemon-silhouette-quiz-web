import React from 'react';
import type { SessionState } from './App';

export const ResultScreen: React.FC<{session:SessionState; onNext:()=>void; onBack:()=>void}> = ({session,onNext,onBack}) => {
  return (
    <div>
      <h2>結果</h2>
      <p>答え: {session.answer}</p>
      <div style={{display:'flex', gap:16}}>
        <button onClick={onBack}>スタート画面</button>
        <button onClick={onNext}>次の問題</button>
      </div>
    </div>
  );
};
